package restoreproof

import (
	"github.com/armpitpete/threadkeeper-core/internal/ledger"
	"github.com/armpitpete/threadkeeper-core/internal/recoveryproof"
)

func ValidateRecoveryProof(proof ledger.RecoveryProof) error {
	return recoveryproof.Validate(proof)
}
