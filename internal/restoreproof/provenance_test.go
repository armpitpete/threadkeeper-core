package restoreproof

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/digest"
)

func TestDecodeProvenanceAcceptsCanonicalBoundDocument(t *testing.T) {
	raw := provenanceFixture(t, nil)
	p, err := DecodeProvenance(raw)
	if err != nil {
		t.Fatal(err)
	}
	if p.SchemaVersion != ProvenanceSchemaV1 || p.ContentSHA256 == "" {
		t.Fatalf("incomplete provenance: %#v", p)
	}
}

func TestDecodeProvenanceRejectsSameAuthorityDomain(t *testing.T) {
	raw := provenanceFixture(t, map[string]any{"secondary_authority_domain_id": "authority:primary"})
	_, err := DecodeProvenance(raw)
	if err == nil || !strings.Contains(err.Error(), "must differ") {
		t.Fatalf("same authority domain accepted: %v", err)
	}
}

func TestDecodeProvenanceRejectsUnknownSelfCertificationField(t *testing.T) {
	raw := provenanceFixture(t, map[string]any{"operational_independence_verified": true})
	_, err := DecodeProvenance(raw)
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("self-certification field accepted: %v", err)
	}
}

func TestDecodeProvenanceRejectsDuplicateTopLevelMember(t *testing.T) {
	raw := []byte(`{"schema_version":"threadkeeper.secondary-restore-provenance.v1","schema_version":"threadkeeper.secondary-restore-provenance.v1"}`)
	raw = []byte(strings.ReplaceAll(string(raw), `\"`, `"`))
	_, err := DecodeProvenance(raw)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		t.Fatalf("duplicate top-level provenance member accepted: %v", err)
	}
}

func TestDecodeProvenanceRejectsMissingNullDuplicateEvidenceAndTimeReversal(t *testing.T) {
	base := provenanceMap()
	delete(base, "secondary_operator_id")
	if _, err := DecodeProvenance(completeProvenance(t, base)); err == nil || !strings.Contains(err.Error(), "required field") {
		t.Fatalf("missing field accepted: %v", err)
	}

	base = provenanceMap()
	base["secondary_operator_id"] = nil
	if _, err := DecodeProvenance(completeProvenance(t, base)); err == nil || !strings.Contains(err.Error(), "must not be null") {
		t.Fatalf("null field accepted: %v", err)
	}

	base = provenanceMap()
	base["external_evidence_refs"] = []string{"evidence:a", "evidence:a"}
	if _, err := DecodeProvenance(completeProvenance(t, base)); err == nil || !strings.Contains(err.Error(), "duplicate external evidence") {
		t.Fatalf("duplicate evidence ref accepted: %v", err)
	}

	base = provenanceMap()
	base["captured_at"] = "2026-08-14T15:00:00Z"
	base["restored_at"] = "2026-08-14T14:59:59Z"
	if _, err := DecodeProvenance(completeProvenance(t, base)); err == nil || !strings.Contains(err.Error(), "precedes") {
		t.Fatalf("restore-before-capture accepted: %v", err)
	}
}

func TestDecodeProvenanceRejectsUnsortedEvidenceReferences(t *testing.T) {
	raw := provenanceFixture(t, map[string]any{"external_evidence_refs": []string{"evidence:z", "evidence:a"}})
	_, err := DecodeProvenance(raw)
	if err == nil || !strings.Contains(err.Error(), "must be sorted") {
		t.Fatalf("unsorted evidence refs accepted: %v", err)
	}
}

func provenanceFixture(t *testing.T, overrides map[string]any) []byte {
	t.Helper()
	value := provenanceMap()
	for key, item := range overrides {
		value[key] = item
	}
	return completeProvenance(t, value)
}

func provenanceMap() map[string]any {
	return map[string]any{
		"schema_version":                 ProvenanceSchemaV1,
		"primary_authority_domain_id":    "authority:primary",
		"secondary_authority_domain_id":  "authority:secondary",
		"secondary_location_id":          "location:secondary-a",
		"secondary_operator_id":          "operator:secondary-a",
		"backup_set_id":                  "backup:set-001",
		"backup_artifact_id":             "artifact:ledger-001",
		"backup_artifact_sha256":         strings.Repeat("a", 64),
		"original_recovery_proof_sha256": strings.Repeat("b", 64),
		"captured_at":                    "2026-08-14T14:00:00Z",
		"restored_at":                    "2026-08-14T14:30:00Z",
		"external_evidence_refs":         []string{"evidence:provider-receipt", "evidence:restore-log"},
	}
}

func completeProvenance(t *testing.T, value map[string]any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	completed, _, err := digest.Complete(raw)
	if err != nil {
		t.Fatal(err)
	}
	return completed
}
