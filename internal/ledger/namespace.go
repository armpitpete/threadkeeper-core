package ledger

import (
	"context"
	"fmt"
	pathpkg "path"
	"strings"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
	"github.com/armpitpete/threadkeeper-core/internal/genesis"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/policy"
)

const schemaPrefix = "config/schemas"

func validateAuthoritativeLedgerNamespace(ctx context.Context, r *gitledger.Reader, history []gitledger.Commit) error {
	for _, commit := range history {
		changes, err := r.CommitTreeChanges(ctx, commit.ID)
		if err != nil {
			return err
		}
		for _, change := range changes {
			if err := validateAuthoritativeLedgerPath(change.Path); err != nil {
				return fmt.Errorf("%w at commit %s status %s", err, commit.ID, change.Status)
			}
		}
	}
	return nil
}

func validateAuthoritativeLedgerPath(path string) error {
	switch path {
	case genesis.LedgerPath, actorauth.LedgerPolicyPath:
		return nil
	}
	if strings.HasPrefix(path, schemaPrefix+"/") {
		if err := validateVersionedJSONLedgerPath(path, schemaPrefix); err != nil {
			return err
		}
		return nil
	}
	if strings.HasPrefix(path, policy.ReducerBindingPrefix+"/") {
		if err := validateVersionedJSONLedgerPath(path, policy.ReducerBindingPrefix); err != nil {
			return err
		}
		return nil
	}
	if strings.HasPrefix(path, "events/") {
		if err := gitledger.ValidateEventPath(path); err != nil {
			return fmt.Errorf("INTEGRITY_FAILURE: invalid Core v1 event path %q: %w", path, err)
		}
		return nil
	}
	return fmt.Errorf("INTEGRITY_FAILURE: unrecognized Core v1 ledger path %q", path)
}

func validateVersionedJSONLedgerPath(path, prefix string) error {
	if !strings.HasPrefix(path, prefix+"/") ||
		!strings.HasSuffix(path, ".json") ||
		path != pathpkg.Clean(path) ||
		strings.Contains(path, "\\") ||
		strings.ContainsAny(path, "\x00\r\n") {
		return fmt.Errorf("INTEGRITY_FAILURE: invalid Core v1 ledger path %q", path)
	}
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("INTEGRITY_FAILURE: invalid Core v1 ledger path %q", path)
		}
	}
	return nil
}
