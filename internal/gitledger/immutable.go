package gitledger

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
)

// ImmutableJSONAdditions validates that every change beneath prefix is an
// addition of a JSON file. It returns the added paths in lexical order.
// Versioned schemas and reducer bindings use this rule so an accepted identity
// cannot be silently redefined later in history.
func (r *Reader) ImmutableJSONAdditions(ctx context.Context, commit, prefix string) ([]string, error) {
	if prefix == "" || strings.ContainsAny(prefix, "\x00\r\n") {
		return nil, fmt.Errorf("invalid immutable JSON prefix")
	}
	out, err := r.run(ctx, "diff-tree", "--root", "--no-commit-id", "--name-status", "-r", "-z", commit, "--", prefix)
	if err != nil {
		return nil, err
	}
	tokens := bytes.Split(out, []byte{0})
	paths := []string{}
	wantPrefix := strings.TrimSuffix(prefix, "/") + "/"
	for i := 0; i < len(tokens); {
		if len(tokens[i]) == 0 {
			i++
			continue
		}
		status := string(tokens[i])
		i++
		if i >= len(tokens) || len(tokens[i]) == 0 {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: malformed Git name-status output for %s", prefix)
		}
		path := string(tokens[i])
		i++
		if status != "A" {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: versioned ledger file %q changed with status %q; %s is append-only", path, status, prefix)
		}
		if !strings.HasPrefix(path, wantPrefix) {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: unexpected immutable tree path %q for %s", path, prefix)
		}
		if !strings.HasSuffix(path, ".json") {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: versioned ledger file %q is not JSON", path)
		}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}
