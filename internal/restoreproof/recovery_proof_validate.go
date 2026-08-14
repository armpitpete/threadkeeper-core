package restoreproof

import (
	"fmt"

	"github.com/armpitpete/threadkeeper-core/internal/ledger"
)

func ValidateRecoveryProof(proof ledger.RecoveryProof) error {
	if proof.LedgerCommit == "" || proof.AuthoritativeRef == "" || proof.GitObjectFormat == "" || proof.GenesisCommit == "" || proof.ProjectID == "" || proof.LedgerID == "" || proof.GenesisContentSHA256 == "" || proof.ActorPolicyVersion == "" || proof.ActorPolicyRootContentSHA256 == "" || proof.GovernedRecordsSHA256 == "" || proof.ReplaySHA256 == "" {
		return fmt.Errorf("RECOVERY_PROOF_INVALID: proof lacks required authority identity")
	}
	if proof.HistoryCommitCount <= 0 || proof.EventCount < 0 || proof.ReducerBindingCount < 0 || proof.GovernedRecordCount < 0 {
		return fmt.Errorf("RECOVERY_PROOF_INVALID: proof contains invalid negative/zero counts")
	}
	return nil
}
