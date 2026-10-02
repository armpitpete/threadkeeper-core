package gitledger

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
)

type TreeChange struct {
	Status string
	Path   string
}

// CommitTreeChanges returns every committed tree-path change for one commit.
// Rename detection is disabled so each changed path is represented directly as
// an add/delete pair rather than a two-path rename record.
func (r *Reader) CommitTreeChanges(ctx context.Context, commit string) ([]TreeChange, error) {
	out, err := r.run(ctx,
		"diff-tree",
		"--root",
		"--no-commit-id",
		"--name-status",
		"-r",
		"--no-renames",
		"-z",
		commit,
	)
	if err != nil {
		return nil, err
	}
	tokens := bytes.Split(out, []byte{0})
	changes := []TreeChange{}
	for i := 0; i < len(tokens); {
		if len(tokens[i]) == 0 {
			i++
			continue
		}
		status := string(tokens[i])
		i++
		if i >= len(tokens) || len(tokens[i]) == 0 {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: malformed Git tree-change output")
		}
		path := string(tokens[i])
		i++
		if strings.ContainsAny(status, "RC") {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: unexpected rename/copy status %q for path %q", status, path)
		}
		changes = append(changes, TreeChange{Status: status, Path: path})
	}
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Path == changes[j].Path {
			return changes[i].Status < changes[j].Status
		}
		return changes[i].Path < changes[j].Path
	})
	return changes, nil
}
