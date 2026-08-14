package gitledger

import (
	"context"
	"fmt"
	"strings"

	"github.com/armpitpete/threadkeeper-core/internal/genesis"
)

// ReadRegularFile reads one exact tree path only after requiring the accepted
// durable representation to be an ordinary 100644 blob. Semantic authority
// records such as Genesis must use this instead of treating a symlink or
// executable-mode blob as equivalent content.
func (r *Reader) ReadRegularFile(ctx context.Context, commit, path string) ([]byte, error) {
	if err := r.requireRegularBlobAt(ctx, commit, path); err != nil {
		return nil, err
	}
	if path == genesis.LedgerPath {
		out, err := r.run(ctx, "rev-list", "--parents", "-n", "1", commit)
		if err != nil {
			return nil, err
		}
		fields := strings.Fields(strings.TrimSpace(string(out)))
		if len(fields) == 0 || len(fields) > 2 || !strings.EqualFold(fields[0], commit) {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: malformed Genesis commit identity")
		}
		if len(fields) == 1 {
			additions, err := r.EventAdditions(ctx, commit)
			if err != nil {
				return nil, err
			}
			if len(additions) != 0 {
				return nil, fmt.Errorf("GENESIS_ROOT_INVALID: root commit must not contain durable events")
			}
		}
	}
	return r.ReadFile(ctx, commit, path)
}
