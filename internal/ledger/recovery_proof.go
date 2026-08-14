package ledger

import (
	"context"
	"fmt"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

// RecoveryProof is a compact, machine-readable fingerprint of one fully
// validated authoritative ledger replay. It contains only values derivable
// from the ledger; it does not create a second authority source.
type RecoveryProof struct {
	LedgerCommit          string `json:"ledger_commit"`
	AuthoritativeRef      string `json:"authoritative_ref"`
	GitObjectFormat       string `json:"git_object_format"`
	GenesisCommit         string `json:"genesis_commit"`
	ProjectID             string `json:"project_id"`
	LedgerID              string `json:"ledger_id"`
	GenesisContentSHA256  string `json:"genesis_content_sha256"`
	HistoryCommitCount    int    `json:"history_commit_count"`
	EventCount            int    `json:"event_count"`
	ReducerBindingCount   int    `json:"reducer_binding_count"`
	GovernedRecordCount   int    `json:"governed_record_count"`
	GovernedRecordsSHA256 string `json:"governed_records_sha256"`
	ReplaySHA256          string `json:"replay_sha256"`
}

// ProveRecovery runs the complete read-only integrity/replay path and emits a
// stable proof suitable for comparing an original authority store with a
// restored copy.
func ProveRecovery(ctx context.Context, r *gitledger.Reader) (*RecoveryProof, error) {
	manifest, err := Replay(ctx, r)
	if err != nil {
		return nil, err
	}
	return &RecoveryProof{
		LedgerCommit:          manifest.LedgerCommit,
		AuthoritativeRef:      manifest.AuthoritativeRef,
		GitObjectFormat:       manifest.GitObjectFormat,
		GenesisCommit:         manifest.GenesisCommit,
		ProjectID:             manifest.GenesisRoot.ProjectID,
		LedgerID:              manifest.GenesisRoot.LedgerID,
		GenesisContentSHA256:  manifest.GenesisRoot.ContentSHA256,
		HistoryCommitCount:    manifest.HistoryCommitCount,
		EventCount:            manifest.EventCount,
		ReducerBindingCount:   manifest.ReducerBindingCount,
		GovernedRecordCount:   manifest.GovernedRecordCount,
		GovernedRecordsSHA256: manifest.GovernedRecordsSHA256,
		ReplaySHA256:          manifest.ReplaySHA256,
	}, nil
}

// CompareRecoveryProofs requires semantic and historical equivalence. A
// restored repository with a different Genesis identity, ref name, head, event
// set or projection does not silently qualify as the same recovered authority.
func CompareRecoveryProofs(original, restored RecoveryProof) error {
	if original != restored {
		return fmt.Errorf("RECOVERY_PROOF_MISMATCH: original=%+v restored=%+v", original, restored)
	}
	return nil
}
