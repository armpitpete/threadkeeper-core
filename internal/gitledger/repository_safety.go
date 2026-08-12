package gitledger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// canonicalLedgerRoot establishes the v1 filesystem boundary for a ledger.
// The supplied Git directory and every ancestor used to reach it must be real
// directories, not symlinks. After that check EvalSymlinks is used as a second
// proof that the path resolves to itself; the returned path is the only path
// retained by Reader and subsequently passed to Git.
func canonicalLedgerRoot(gitDir string) (string, error) {
	abs, err := filepath.Abs(gitDir)
	if err != nil {
		return "", fmt.Errorf("resolve ledger path: %w", err)
	}
	abs = filepath.Clean(abs)
	if err := rejectSymlinkedPathComponents(abs); err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolve canonical ledger path %q: %w", abs, err)
	}
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return "", fmt.Errorf("resolve canonical ledger path %q: %w", abs, err)
	}
	resolved = filepath.Clean(resolved)
	if resolved != abs {
		return "", fmt.Errorf("INTEGRITY_FAILURE: symlinked Git repository root or ancestor is forbidden in authoritative ledger: %s resolves to %s", abs, resolved)
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return "", fmt.Errorf("inspect ledger root %q: %w", abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("INTEGRITY_FAILURE: authoritative ledger Git root is not a directory: %s", abs)
	}
	return resolved, nil
}

// rejectSymlinkedPathComponents checks the final ledger root and every
// filesystem ancestor. Lstat is deliberate: Stat would follow the exact
// indirection this boundary must reject.
func rejectSymlinkedPathComponents(path string) error {
	current := filepath.Clean(path)
	for {
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("inspect authoritative ledger path component %q: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("INTEGRITY_FAILURE: symlinked Git repository root or ancestor is forbidden in authoritative ledger: %s", current)
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return nil
}

// checkRepositoryRootSafety repeats the root/ancestor proof before Git is
// invoked and additionally requires the root to be the same filesystem object
// that Reader pinned during construction. This rejects an ordinary directory
// replacement at the same pathname, not merely symlink substitution.
// Runtime service ownership/permissions are still required to prevent an
// untrusted local process racing replacement between this check and exec.
func (r *Reader) checkRepositoryRootSafety() error {
	canonical, err := canonicalLedgerRoot(r.gitDir)
	if err != nil {
		return err
	}
	if canonical != r.gitDir {
		return fmt.Errorf("INTEGRITY_FAILURE: authoritative ledger root changed after Reader construction: got %s want %s", canonical, r.gitDir)
	}
	currentInfo, err := os.Lstat(r.gitDir)
	if err != nil {
		return fmt.Errorf("inspect current authoritative ledger root identity %q: %w", r.gitDir, err)
	}
	if r.rootInfo == nil || !os.SameFile(r.rootInfo, currentInfo) {
		return fmt.Errorf("INTEGRITY_FAILURE: authoritative ledger filesystem identity changed after Reader construction: %s", r.gitDir)
	}
	return nil
}

// checkAuthoritativeRefSafety requires the configured authority ref itself to
// be direct. Git symbolic refs are ordinary files whose "ref: ..." content
// causes update-ref to dereference to another ref by default; that indirection
// is not an accepted authority mechanism in v1.
func (r *Reader) checkAuthoritativeRefSafety() error {
	if err := r.checkRepositoryRootSafety(); err != nil {
		return err
	}
	if !strings.HasPrefix(r.ref, "refs/") || path.Clean(r.ref) != r.ref || strings.Contains(r.ref, "\\") || strings.ContainsAny(r.ref, "\x00\r\n") {
		return fmt.Errorf("INTEGRITY_FAILURE: invalid authoritative ref path %q", r.ref)
	}
	for _, part := range strings.Split(r.ref, "/") {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("INTEGRITY_FAILURE: invalid authoritative ref path %q", r.ref)
		}
	}
	full := filepath.Join(r.gitDir, filepath.FromSlash(r.ref))
	info, err := os.Lstat(full)
	if errors.Is(err, os.ErrNotExist) {
		// Under the v1 files backend a missing loose ref may be represented as a
		// direct packed ref. Symbolic refs require a loose ref file, so absence
		// here is not itself an integrity failure; Head() will still require the
		// configured ref to resolve to a commit.
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect authoritative ref %s: %w", r.ref, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("INTEGRITY_FAILURE: authoritative ref must be a regular direct ref: %s", r.ref)
	}
	raw, err := os.ReadFile(full)
	if err != nil {
		return fmt.Errorf("read authoritative ref %s: %w", r.ref, err)
	}
	value := strings.TrimSpace(string(raw))
	if strings.HasPrefix(value, "ref:") {
		return fmt.Errorf("INTEGRITY_FAILURE: symbolic authoritative refs are forbidden in authoritative ledger: %s -> %s", r.ref, strings.TrimSpace(strings.TrimPrefix(value, "ref:")))
	}
	if !isObjectID(value) {
		return fmt.Errorf("INTEGRITY_FAILURE: malformed loose authoritative ref %s", r.ref)
	}
	return nil
}

// checkRepositoryLayoutSafety rejects repository-local filesystem indirection
// that can make Git read or mutate authority outside r.gitDir. Runtime service
// ownership/permissions remain necessary to close races after these checks.
func (r *Reader) checkRepositoryLayoutSafety() error {
	if err := r.checkRepositoryRootSafety(); err != nil {
		return err
	}

	commonDir := filepath.Join(r.gitDir, "commondir")
	if _, err := os.Lstat(commonDir); err == nil {
		return fmt.Errorf("INTEGRITY_FAILURE: Git common-dir indirection is forbidden in authoritative ledger")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect Git commondir metadata: %w", err)
	}

	// Threadkeeper v1 deliberately supports only the classic files ref backend
	// and one repository config file. Reftable and config.worktree introduce
	// additional ref/config stores that are unnecessary for the authority
	// ledger and would otherwise expand the filesystem trust boundary.
	for rel, reason := range map[string]string{
		"config.worktree": "worktree-specific Git configuration is forbidden in authoritative ledger",
		"reftable":        "Git reftable ref storage is unsupported in authoritative ledger v1",
	} {
		full := filepath.Join(r.gitDir, rel)
		if _, err := os.Lstat(full); err == nil {
			return fmt.Errorf("INTEGRITY_FAILURE: %s: %s", reason, rel)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect Git repository path %s: %w", rel, err)
		}
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
