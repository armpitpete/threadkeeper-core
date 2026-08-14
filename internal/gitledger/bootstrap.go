package gitledger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var ErrLedgerAlreadyExists = errors.New("LEDGER_ALREADY_EXISTS")

// InitializeBareRoot creates a new dedicated bare repository with exactly one
// root commit assembled from rootFiles and establishes ref once. It never opens,
// adopts or overwrites an existing path. Failure may leave explicitly incomplete
// create-only residue at gitDir; callers must never interpret that residue as a
// successful ledger and must verify the result through New + normal replay.
func InitializeBareRoot(ctx context.Context, gitDir, ref string, rootFiles map[string][]byte) (string, error) {
	if gitDir == "" {
		return "", fmt.Errorf("ledger git directory is required")
	}
	if ref == "" {
		ref = DefaultRef
	}
	if !strings.HasPrefix(ref, "refs/") || strings.ContainsAny(ref, "\x00\r\n") {
		return "", fmt.Errorf("invalid authoritative ref %q", ref)
	}
	if len(rootFiles) == 0 {
		return "", fmt.Errorf("FRESH_GENESIS_INVALID: root file set is empty")
	}
	paths := make([]string, 0, len(rootFiles))
	for p := range rootFiles {
		if err := validateBootstrapPath(p); err != nil {
			return "", err
		}
		paths = append(paths, p)
	}
	sort.Strings(paths)

	gitPath, err := exec.LookPath("git")
	if err != nil {
		return "", fmt.Errorf("GIT_FAILURE: git executable not found: %w", err)
	}
	abs, err := filepath.Abs(filepath.Clean(gitDir))
	if err != nil {
		return "", fmt.Errorf("resolve fresh ledger path: %w", err)
	}

	// Prove the existing parent path is already inside the same no-symlink,
	// canonical filesystem boundary required by Reader before creating anything.
	// The later live Reader pin still closes replacement/reuse after creation;
	// production service ownership closes races between these operations.
	parent := filepath.Dir(abs)
	canonicalParent, err := canonicalLedgerRoot(parent)
	if err != nil {
		return "", fmt.Errorf("FRESH_GENESIS_INVALID: unsafe target parent: %w", err)
	}
	if canonicalParent != parent {
		return "", fmt.Errorf("FRESH_GENESIS_INVALID: target parent changed canonical identity: got %s want %s", canonicalParent, parent)
	}

	if err := os.Mkdir(abs, 0o700); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return "", fmt.Errorf("%w: target %q already exists", ErrLedgerAlreadyExists, abs)
		}
		return "", fmt.Errorf("create fresh ledger root %q: %w", abs, err)
	}

	if _, err := runBootstrapGit(ctx, gitPath, nil, nil, "init", "--bare", "--initial-branch=main", "--object-format=sha1", abs); err != nil {
		return "", fmt.Errorf("initialize fresh bare ledger: %w", err)
	}

	indexDir, err := os.MkdirTemp("", "threadkeeper-genesis-index-*")
	if err != nil {
		return "", fmt.Errorf("create private Genesis index directory: %w", err)
	}
	defer os.RemoveAll(indexDir)
	indexPath := filepath.Join(indexDir, "index")
	hooksDir, err := os.MkdirTemp("", "threadkeeper-genesis-no-hooks-*")
	if err != nil {
		return "", fmt.Errorf("create empty Genesis hooks directory: %w", err)
	}
	defer os.RemoveAll(hooksDir)
	extra := []string{"GIT_INDEX_FILE=" + indexPath}

	if _, err := runBootstrapRepoGit(ctx, gitPath, abs, hooksDir, nil, extra, "read-tree", "--empty"); err != nil {
		return "", err
	}
	for _, p := range paths {
		blobOut, err := runBootstrapRepoGit(ctx, gitPath, abs, hooksDir, rootFiles[p], nil, "hash-object", "-w", "--stdin")
		if err != nil {
			return "", err
		}
		blob := strings.TrimSpace(string(blobOut))
		if !isObjectID(blob) {
			return "", fmt.Errorf("GIT_FAILURE: invalid bootstrap blob id %q", blob)
		}
		if _, err := runBootstrapRepoGit(ctx, gitPath, abs, hooksDir, nil, extra, "update-index", "--add", "--cacheinfo", "100644", blob, p); err != nil {
			return "", err
		}
	}
	treeOut, err := runBootstrapRepoGit(ctx, gitPath, abs, hooksDir, nil, extra, "write-tree")
	if err != nil {
		return "", err
	}
	tree := strings.TrimSpace(string(treeOut))
	if !isObjectID(tree) {
		return "", fmt.Errorf("GIT_FAILURE: invalid Genesis tree id %q", tree)
	}

	commitOut, err := runBootstrapRepoGit(ctx, gitPath, abs, hooksDir, []byte("Threadkeeper Genesis\n"), []string{
		"GIT_AUTHOR_NAME=Threadkeeper Core Genesis",
		"GIT_AUTHOR_EMAIL=threadkeeper-core@localhost",
		"GIT_AUTHOR_DATE=2000-01-01T00:00:00Z",
		"GIT_COMMITTER_NAME=Threadkeeper Core Genesis",
		"GIT_COMMITTER_EMAIL=threadkeeper-core@localhost",
		"GIT_COMMITTER_DATE=2000-01-01T00:00:00Z",
	}, "commit-tree", tree)
	if err != nil {
		return "", err
	}
	commit := strings.ToLower(strings.TrimSpace(string(commitOut)))
	if !isObjectID(commit) {
		return "", fmt.Errorf("GIT_FAILURE: invalid Genesis commit id %q", commit)
	}

	// v1 bootstrap explicitly creates SHA-1 repositories, so a zero SHA-1 old
	// value means the authoritative ref must not already exist.
	if _, err := runBootstrapRepoGit(ctx, gitPath, abs, hooksDir, nil, nil, "update-ref", "--no-deref", ref, commit, strings.Repeat("0", 40)); err != nil {
		return "", fmt.Errorf("establish fresh authoritative ref: %w", err)
	}

	r, err := New(abs, ref)
	if err != nil {
		return "", fmt.Errorf("reopen fresh ledger through hardened reader: %w", err)
	}
	defer r.Close()
	if err := r.CheckHistorySafety(ctx); err != nil {
		return "", fmt.Errorf("fresh ledger safety verification: %w", err)
	}
	head, err := r.Head(ctx)
	if err != nil {
		return "", fmt.Errorf("resolve fresh authoritative head: %w", err)
	}
	if head != commit {
		return "", fmt.Errorf("FRESH_GENESIS_INTEGRITY_FAILURE: authoritative head %s differs from created root %s", head, commit)
	}
	return commit, nil
}

func validateBootstrapPath(p string) error {
	if p == "" || strings.HasPrefix(p, "-") || strings.Contains(p, "\\") || strings.ContainsAny(p, "\x00\r\n") || path.IsAbs(p) || path.Clean(p) != p {
		return fmt.Errorf("FRESH_GENESIS_INVALID: unsafe root path %q", p)
	}
	for _, part := range strings.Split(p, "/") {
		if part == "" || part == "." || part == ".." {
			return fmt.Errorf("FRESH_GENESIS_INVALID: unsafe root path %q", p)
		}
	}
	return nil
}

func runBootstrapGit(parent context.Context, gitPath string, stdin []byte, extraEnv []string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, gitPath, args...)
	cmd.Env = append(controlledEnv(), extraEnv...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("GIT_FAILURE: bootstrap command timed out")
		}
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > 800 {
			msg = msg[:800]
		}
		return nil, fmt.Errorf("GIT_FAILURE: git %s failed: %w: %s", strings.Join(args, " "), err, msg)
	}
	return stdout.Bytes(), nil
}

func runBootstrapRepoGit(parent context.Context, gitPath, gitDir, hooksDir string, stdin []byte, extraEnv []string, args ...string) ([]byte, error) {
	base := []string{"--no-replace-objects", "--git-dir=" + gitDir, "-c", "core.hooksPath=" + hooksDir, "-c", "commit.gpgSign=false"}
	return runBootstrapGit(parent, gitPath, stdin, extraEnv, append(base, args...)...)
}
