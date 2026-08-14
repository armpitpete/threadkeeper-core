package ledger

import (
	"context"
	"fmt"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/loadproof"
)

type ReadLoadEvidence struct {
	Recovery  RecoveryProof      `json:"recovery_proof"`
	Resources loadproof.Evidence `json:"resources"`
}

// ProveReadLoad measures repeated full authoritative recovery/replay under the
// declared resource envelope. Every concurrent operation must reproduce the
// exact baseline recovery proof. This is read-only and makes no production
// capacity claim beyond the envelope actually supplied and measured.
func ProveReadLoad(ctx context.Context, r *gitledger.Reader, envelope loadproof.Envelope) (*ReadLoadEvidence, error) {
	baseline, err := ProveRecovery(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("LOAD_REPLAY_BASELINE_FAILED: %w", err)
	}

	resources, err := loadproof.Run(ctx, envelope, func(workCtx context.Context, _, _ int) error {
		proof, err := ProveRecovery(workCtx, r)
		if err != nil {
			return err
		}
		if err := CompareRecoveryProofs(*baseline, *proof); err != nil {
			return fmt.Errorf("LOAD_REPLAY_DIVERGED: %w", err)
		}
		return nil
	})
	if err != nil {
		return &ReadLoadEvidence{Recovery: *baseline, Resources: resources}, err
	}
	return &ReadLoadEvidence{Recovery: *baseline, Resources: resources}, nil
}
