package restoreproof

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/ledger"
	"github.com/armpitpete/threadkeeper-core/internal/recoveryproof"
)

const OperationalIndependenceRequiresExternalReview = "requires_external_review"

type Report struct {
	CoreEquivalencePassed         bool                 `json:"core_equivalence_passed"`
	OperationalIndependenceStatus string               `json:"operational_independence_status"`
	RestoredStoragePath           string               `json:"restored_storage_path"`
	RestoredAuthoritativeRef      string               `json:"restored_authoritative_ref"`
	ProvenanceContentSHA256       string               `json:"provenance_content_sha256"`
	PrimaryAuthorityDomainID      string               `json:"primary_authority_domain_id"`
	SecondaryAuthorityDomainID    string               `json:"secondary_authority_domain_id"`
	SecondaryLocationID           string               `json:"secondary_location_id"`
	SecondaryOperatorID           string               `json:"secondary_operator_id"`
	BackupSetID                   string               `json:"backup_set_id"`
	BackupArtifactID              string               `json:"backup_artifact_id"`
	BackupArtifactSHA256          string               `json:"backup_artifact_sha256"`
	ExternalEvidenceRefs          []string             `json:"external_evidence_refs"`
	OriginalRecoveryProofSHA256   string               `json:"original_recovery_proof_sha256"`
	RestoredRecoveryProofSHA256   string               `json:"restored_recovery_proof_sha256"`
	OriginalRecoveryProof         ledger.RecoveryProof `json:"original_recovery_proof"`
	RestoredRecoveryProof         ledger.RecoveryProof `json:"restored_recovery_proof"`
}

func DecodeRecoveryProof(raw []byte) (ledger.RecoveryProof, error) {
	return recoveryproof.Decode(raw)
}

func RecoveryProofSHA256(proof ledger.RecoveryProof) (string, error) {
	raw, err := json.Marshal(proof)
	if err != nil {
		return "", fmt.Errorf("marshal recovery proof: %w", err)
	}
	canonical, err := canonicaljson.Canonicalize(raw)
	if err != nil {
		return "", fmt.Errorf("canonicalize recovery proof: %w", err)
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}

// Verify recomputes the restored ledger's RecoveryProof through the hardened
// Reader and compares it exactly with the pre-restore proof. Core verifies only
// authority-state equivalence. Distinct provenance identifiers and evidence
// references are declarations to be externally reviewed; they never cause Core
// to report operational independence as verified.
func Verify(ctx context.Context, r *gitledger.Reader, original ledger.RecoveryProof, provenance Provenance) (*Report, error) {
	if r == nil {
		return nil, fmt.Errorf("RESTORE_VERIFY_INVALID: restored ledger Reader is required")
	}
	if err := ValidateRecoveryProof(original); err != nil {
		return nil, err
	}
	if err := provenance.Validate(); err != nil {
		return nil, err
	}
	originalSHA, err := RecoveryProofSHA256(original)
	if err != nil {
		return nil, err
	}
	if provenance.OriginalRecoveryProofSHA256 != originalSHA {
		return nil, fmt.Errorf("RESTORE_PROVENANCE_MISMATCH: provenance original_recovery_proof_sha256=%s proof=%s", provenance.OriginalRecoveryProofSHA256, originalSHA)
	}

	restored, err := ledger.ProveRecovery(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("RESTORE_RECOVERY_PROOF_FAILED: %w", err)
	}
	restoredSHA, err := RecoveryProofSHA256(*restored)
	if err != nil {
		return nil, err
	}
	report := &Report{
		OperationalIndependenceStatus: OperationalIndependenceRequiresExternalReview,
		RestoredStoragePath:           r.GitDir(),
		RestoredAuthoritativeRef:      r.Ref(),
		ProvenanceContentSHA256:       provenance.ContentSHA256,
		PrimaryAuthorityDomainID:      provenance.PrimaryAuthorityDomainID,
		SecondaryAuthorityDomainID:    provenance.SecondaryAuthorityDomainID,
		SecondaryLocationID:           provenance.SecondaryLocationID,
		SecondaryOperatorID:           provenance.SecondaryOperatorID,
		BackupSetID:                   provenance.BackupSetID,
		BackupArtifactID:              provenance.BackupArtifactID,
		BackupArtifactSHA256:          provenance.BackupArtifactSHA256,
		ExternalEvidenceRefs:          append([]string(nil), provenance.ExternalEvidenceRefs...),
		OriginalRecoveryProofSHA256:   originalSHA,
		RestoredRecoveryProofSHA256:   restoredSHA,
		OriginalRecoveryProof:         original,
		RestoredRecoveryProof:         *restored,
	}
	if err := ledger.CompareRecoveryProofs(original, *restored); err != nil {
		return report, fmt.Errorf("RESTORE_AUTHORITY_MISMATCH: %w", err)
	}
	report.CoreEquivalencePassed = true
	return report, nil
}