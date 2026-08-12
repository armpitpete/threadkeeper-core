package gitledger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// checkRepositoryLayoutSafety rejects repository-local filesystem indirection
// that can make Git read or mutate authority outside r.gitDir. Runtime service
// ownership/permissions remain necessary to close races after these checks.
func (r *Reader) checkRepositoryLayoutSafety() error {
	commonDir := filepath.Join(r.gitDir, "commondir")
	if _, err := os.Lstat(commonDir); err == nil {
		return fmt.Errorf("INTEGRITY_FAILURE: Git common-dir indirection is forbidden in authoritative ledger")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect Git commondir metadata: %w", err)
	}

	for _, rel := range []string{"HEAD", "config", "packed-refs", "objects", "refs"} {
		full := filepath.Join(r.gitDir, filepath.FromSlash(rel))
		info, err := os.Lstat(full)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect Git repository path %s: %w", rel, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("INTEGRITY_FAILURE: symlinked Git repository path is forbidden in authoritative ledger: %s", rel)
		}
	}

	for _, rel := range []string{"objects", "refs"} {
		root := filepath.Join(r.gitDir, rel)
		if err := rejectSymlinksUnder(root, r.gitDir); err != nil {
			return err
		}
	}
	return nil
}

func rejectSymlinksUnder(root, gitDir string) error {
	if _, err := os.Lstat(root); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("inspect Git repository tree %s: %w", filepath.ToSlash(root), err)
	}
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("inspect Git repository tree %s: %w", filepath.ToSlash(path), err)
		}
		if entry.Type()&os.ModeSymlink == 0 {
			return nil
		}
		rel, relErr := filepath.Rel(gitDir, path)
		if relErr != nil {
			rel = path
		}
		return fmt.Errorf("INTEGRITY_FAILURE: symlinked Git repository path is forbidden in authoritative ledger: %s", filepath.ToSlash(rel))
	})
}

// requireRegularBlobAt makes the durable Git tree shape part of validation.
// Name-status alone cannot distinguish a normal file from executable or
// symlink modes because all can appear as an added path with a blob object.
func (r *Reader) requireRegularBlobAt(ctx context.Context, commit, path string) error {
	if !isObjectID(commit) || path == "" || strings.ContainsAny(path, "\x00\r\n") {
		return fmt.Errorf("invalid Git tree entry request")
	}
	out, err := r.run(ctx, "ls-tree", "-z", commit, "--", path)
	if err != nil {
		return err
	}
	entries := bytes.Split(out, []byte{0})
	var entry []byte
	for _, candidate := range entries {
		if len(candidate) == 0 {
			continue
		}
		if entry != nil {
			return fmt.Errorf("INTEGRITY_FAILURE: multiple Git tree entries matched %q", path)
		}
		entry = candidate
	}
	if entry == nil {
		return fmt.Errorf("INTEGRITY_FAILURE: Git tree entry %q is missing at %s", path, commit)
	}
	tab := bytes.IndexByte(entry, '\t')
	if tab <= 0 || tab == len(entry)-1 {
		return fmt.Errorf("INTEGRITY_FAILURE: malformed Git tree entry for %q", path)
	}
	meta := strings.Fields(string(entry[:tab]))
	if len(meta) != 3 {
		return fmt.Errorf("INTEGRITY_FAILURE: malformed Git tree metadata for %q", path)
	}
	entryPath := string(entry[tab+1:])
	if entryPath != path || meta[0] != "100644" || meta[1] != "blob" || !isObjectID(meta[2]) {
		return fmt.Errorf("INTEGRITY_FAILURE: durable JSON path %q must be a 100644 regular blob, got mode=%q type=%q path=%q", path, meta[0], meta[1], entryPath)
	}
	return nil
}
