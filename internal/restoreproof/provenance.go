package restoreproof

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

const ProvenanceSchemaV1 = "threadkeeper.secondary-restore-provenance.v1"

var safeIdentifier = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/@+\-]{0,255}$`)

var requiredProvenanceFields = []string{
	"schema_version",
	"primary_authority_domain_id",
	"secondary_authority_domain_id",
	"secondary_location_id",
	"secondary_operator_id",
	"backup_set_id",
	"backup_artifact_id",
	"backup_artifact_sha256",
	"original_recovery_proof_sha256",
	"captured_at",
	"restored_at",
	"external_evidence_refs",
	"content_sha256",
}

type Provenance struct {
	SchemaVersion               string   `json:"schema_version"`
	PrimaryAuthorityDomainID    string   `json:"primary_authority_domain_id"`
	SecondaryAuthorityDomainID  string   `json:"secondary_authority_domain_id"`
	SecondaryLocationID         string   `json:"secondary_location_id"`
	SecondaryOperatorID         string   `json:"secondary_operator_id"`
	BackupSetID                 string   `json:"backup_set_id"`
	BackupArtifactID            string   `json:"backup_artifact_id"`
	BackupArtifactSHA256        string   `json:"backup_artifact_sha256"`
	OriginalRecoveryProofSHA256 string   `json:"original_recovery_proof_sha256"`
	CapturedAt                  string   `json:"captured_at"`
	RestoredAt                  string   `json:"restored_at"`
	ExternalEvidenceRefs        []string `json:"external_evidence_refs"`
	ContentSHA256               string   `json:"content_sha256"`
}

func DecodeProvenance(raw []byte) (Provenance, error) {
	if err := strictjson.Validate(raw); err != nil {
		return Provenance{}, fmt.Errorf("RESTORE_PROVENANCE_INVALID: %w", err)
	}
	canonical, err := canonicaljson.Canonicalize(raw)
	if err != nil {
		return Provenance{}, fmt.Errorf("RESTORE_PROVENANCE_INVALID: canonicalize: %w", err)
	}
	if !bytes.Equal(raw, canonical) {
		return Provenance{}, fmt.Errorf("RESTORE_PROVENANCE_INVALID: document must be RFC 8785 canonical JSON")
	}
	if err := digest.Verify(raw); err != nil {
		return Provenance{}, fmt.Errorf("RESTORE_PROVENANCE_INVALID: %w", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return Provenance{}, fmt.Errorf("RESTORE_PROVENANCE_INVALID: decode fields: %w", err)
	}
	for _, name := range requiredProvenanceFields {
		value, ok := fields[name]
		if !ok {
			return Provenance{}, fmt.Errorf("RESTORE_PROVENANCE_INVALID: required field %q is missing", name)
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return Provenance{}, fmt.Errorf("RESTORE_PROVENANCE_INVALID: required field %q must not be null", name)
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var p Provenance
	if err := decoder.Decode(&p); err != nil {
		return Provenance{}, fmt.Errorf("RESTORE_PROVENANCE_INVALID: decode: %w", err)
	}
	if err := p.Validate(); err != nil {
		return Provenance{}, err
	}
	return p, nil
}

// Validate rechecks semantic and digest integrity for callers that construct a
// Provenance value directly rather than through DecodeProvenance. This prevents
// the typed API from becoming a bypass around the strict durable-document path.
func (p Provenance) Validate() error {
	raw, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("RESTORE_PROVENANCE_INVALID: marshal: %w", err)
	}
	if err := digest.Verify(raw); err != nil {
		return fmt.Errorf("RESTORE_PROVENANCE_INVALID: %w", err)
	}
	if p.SchemaVersion != ProvenanceSchemaV1 {
		return fmt.Errorf("RESTORE_PROVENANCE_INVALID: schema_version %q is unsupported", p.SchemaVersion)
	}
	for name, value := range map[string]string{
		"primary_authority_domain_id":   p.PrimaryAuthorityDomainID,
		"secondary_authority_domain_id": p.SecondaryAuthorityDomainID,
		"secondary_location_id":         p.SecondaryLocationID,
		"secondary_operator_id":         p.SecondaryOperatorID,
		"backup_set_id":                 p.BackupSetID,
		"backup_artifact_id":            p.BackupArtifactID,
	} {
		if !safeIdentifier.MatchString(value) {
			return fmt.Errorf("RESTORE_PROVENANCE_INVALID: %s is unsafe or empty", name)
		}
	}
	if p.PrimaryAuthorityDomainID == p.SecondaryAuthorityDomainID {
		return fmt.Errorf("RESTORE_PROVENANCE_INVALID: primary and secondary authority-domain IDs must differ")
	}
	if err := requireSHA256("backup_artifact_sha256", p.BackupArtifactSHA256); err != nil {
		return err
	}
	if err := requireSHA256("original_recovery_proof_sha256", p.OriginalRecoveryProofSHA256); err != nil {
		return err
	}
	captured, err := time.Parse(time.RFC3339Nano, p.CapturedAt)
	if err != nil {
		return fmt.Errorf("RESTORE_PROVENANCE_INVALID: captured_at must be RFC3339: %w", err)
	}
	restored, err := time.Parse(time.RFC3339Nano, p.RestoredAt)
	if err != nil {
		return fmt.Errorf("RESTORE_PROVENANCE_INVALID: restored_at must be RFC3339: %w", err)
	}
	if restored.Before(captured) {
		return fmt.Errorf("RESTORE_PROVENANCE_INVALID: restored_at precedes captured_at")
	}
	if len(p.ExternalEvidenceRefs) == 0 {
		return fmt.Errorf("RESTORE_PROVENANCE_INVALID: at least one external_evidence_ref is required")
	}
	if !sort.StringsAreSorted(p.ExternalEvidenceRefs) {
		return fmt.Errorf("RESTORE_PROVENANCE_INVALID: external_evidence_refs must be sorted")
	}
	seen := map[string]struct{}{}
	for _, ref := range p.ExternalEvidenceRefs {
		if ref == "" || len(ref) > 2048 || strings.ContainsAny(ref, "\x00\r\n") {
			return fmt.Errorf("RESTORE_PROVENANCE_INVALID: external evidence reference is unsafe or empty")
		}
		if _, exists := seen[ref]; exists {
			return fmt.Errorf("RESTORE_PROVENANCE_INVALID: duplicate external evidence reference %q", ref)
		}
		seen[ref] = struct{}{}
	}
	return nil
}

func requireSHA256(name, value string) error {
	if len(value) != 64 {
		return fmt.Errorf("RESTORE_PROVENANCE_INVALID: %s must be 64 lowercase hex characters", name)
	}
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 || strings.ToLower(value) != value {
		return fmt.Errorf("RESTORE_PROVENANCE_INVALID: %s must be 64 lowercase hex characters", name)
	}
	return nil
}
