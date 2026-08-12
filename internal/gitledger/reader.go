package gitledger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const DefaultRef = "refs/heads/main"

var ErrNonLinearHistory = errors.New("INTEGRITY_FAILURE: authoritative ledger history is not linear")

type Reader struct {
	gitPath  string
	gitDir   string
	rootInfo os.FileInfo
	ref      string
	timeout  time.Duration
}

type Commit struct {
	ID     string
	Parent string
}

type EventAddition struct {
	Commit string
	Path   string
}

func New(gitDir, ref string) (*Reader, error) {
	if gitDir == "" {
		return nil, fmt.Errorf("ledger git directory is required")
	}
	canonical, err := canonicalLedgerRoot(gitDir)
	if err != nil {
		return nil, err
	}
	rootInfo, err := os.Lstat(canonical)
	if err != nil {
		return nil, fmt.Errorf("pin ledger root identity %q: %w", canonical, err)
	}
	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("INTEGRITY_FAILURE: authoritative ledger Git root is not a directory: %s", canonical)
	}
	if _, err := os.Lstat(filepath.Join(canonical, "HEAD")); err != nil {
		return nil, fmt.Errorf("open ledger %q: %w", canonical, err)
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("GIT_FAILURE: git executable not found: %w", err)
	}
	if ref == "" {
		ref = DefaultRef
	}
	if !strings.HasPrefix(ref, "refs/") || strings.ContainsAny(ref, "\x00\r\n") {
		return nil, fmt.Errorf("invalid authoritative ref %q", ref)
	}
	r := &Reader{gitPath: gitPath, gitDir: canonical, rootInfo: rootInfo, ref: ref, timeout: 60 * time.Second}
	if err := r.checkAuthoritativeRefSafety(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Reader) Ref() string    { return r.ref }
func (r *Reader) GitDir() string { return r.gitDir }

func (r *Reader) Head(ctx context.Context) (string, error) {
	if err := r.checkAuthoritativeRefSafety(); err != nil {
		return "", err
	}
	out, err := r.run(ctx, "rev-parse", "--verify", r.ref+"^{commit}")
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(string(out))
	if !isObjectID(id) {
		return "", fmt.Errorf("INTEGRITY_FAILURE: invalid commit object id %q", id)
	}
	return strings.ToLower(id), nil
}

func (r *Reader) ObjectFormat(ctx context.Context) (string, error) {
	out, err := r.run(ctx, "rev-parse", "--show-object-format")
	if err != nil {
		return "", err
	}
	format := strings.TrimSpace(string(out))
	if format != "sha1" && format != "sha256" {
		return "", fmt.Errorf("INTEGRITY_FAILURE: unsupported Git object format %q", format)
	}
	return format, nil
}

func (r *Reader) IsBare(ctx context.Context) (bool, error) {
	out, err := r.run(ctx, "rev-parse", "--is-bare-repository")
	if err != nil {
		return false, err
	}
	switch strings.TrimSpace(string(out)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("GIT_FAILURE: unexpected --is-bare-repository output")
	}
}

// CheckHistorySafety rejects repository features that can cause Git to present
// an authority history or object namespace different from the repository's own
// stored commit graph and object database.
func (r *Reader) CheckHistorySafety(ctx context.Context) error {
	// Repository-layout checks must run before invoking Git. A commondir file or
	// alternate ref/config store can otherwise make Git inspect a different
	// repository surface than the filesystem metadata Threadkeeper validates.
	if err := r.checkRepositoryLayoutSafety(); err != nil {
		return err
	}
	if err := r.checkAuthoritativeRefSafety(); err != nil {
		return err
	}

	out, err := r.run(ctx, "rev-parse", "--is-shallow-repository")
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(out)) == "true" {
		return fmt.Errorf("INTEGRITY_FAILURE: shallow authoritative ledger histories are forbidden")
	}
	if _, err := os.Stat(filepath.Join(r.gitDir, "shallow")); err == nil {
		return fmt.Errorf("INTEGRITY_FAILURE: shallow metadata exists in authoritative ledger")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect shallow metadata: %w", err)
	}
	if _, err := os.Stat(filepath.Join(r.gitDir, "info", "grafts")); err == nil {
		return fmt.Errorf("INTEGRITY_FAILURE: Git grafts are forbidden in authoritative ledger")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect graft metadata: %w", err)
	}
	for _, rel := range []string{
		filepath.Join("objects", "info", "alternates"),
		filepath.Join("objects", "info", "http-alternates"),
	} {
		full := filepath.Join(r.gitDir, rel)
		if _, err := os.Lstat(full); err == nil {
			return fmt.Errorf("INTEGRITY_FAILURE: Git object alternates are forbidden in authoritative ledger: %s", filepath.ToSlash(rel))
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect Git object alternates %s: %w", filepath.ToSlash(rel), err)
		}
	}
	config, err := os.ReadFile(filepath.Join(r.gitDir, "config"))
	if err != nil {
		return fmt.Errorf("read ledger Git config: %w", err)
	}
	lower := strings.ToLower(string(config))
	if strings.Contains(lower, "[include") {
		return fmt.Errorf("INTEGRITY_FAILURE: Git config includes are forbidden in authoritative ledger")
	}
	if strings.Contains(lower, "[fsck") {
		return fmt.Errorf("INTEGRITY_FAILURE: repository-local fsck overrides are forbidden")
	}
	if strings.Contains(lower, "promisor") || strings.Contains(lower, "partialclone") {
		return fmt.Errorf("INTEGRITY_FAILURE: Git promisor/partial-clone repositories are forbidden in authoritative ledger")
	}
	if strings.Contains(lower, "refstorage") {
		return fmt.Errorf("INTEGRITY_FAILURE: alternate Git ref-storage backends are forbidden; authoritative ledger v1 requires the files backend")
	}
	if strings.Contains(lower, "worktreeconfig") {
		return fmt.Errorf("INTEGRITY_FAILURE: worktree-specific Git configuration is forbidden in authoritative ledger")
	}
	return nil
}

func (r *Reader) FSCK(ctx context.Context) error {
	if _, err := r.run(ctx, "-c", "fsck.skipList="+os.DevNull, "fsck", "--full", "--strict", "--no-dangling"); err != nil {
		return fmt.Errorf("INTEGRITY_FAILURE: git fsck: %w", err)
	}
	return nil
}

func (r *Reader) History(ctx context.Context, head string) ([]Commit, error) {
	if !isObjectID(head) {
		return nil, fmt.Errorf("invalid head object id %q", head)
	}
	out, err := r.run(ctx, "rev-list", "--reverse", "--parents", head)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	commits := make([]Commit, 0, len(lines))
	var previous string
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 2 || len(fields) == 0 {
			return nil, ErrNonLinearHistory
		}
		id := strings.ToLower(fields[0])
		if !isObjectID(id) {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: invalid history object id %q", fields[0])
		}
		parent := ""
		if len(fields) == 2 {
			parent = strings.ToLower(fields[1])
			if i == 0 || parent != previous {
				return nil, ErrNonLinearHistory
			}
		} else if i != 0 {
			return nil, ErrNonLinearHistory
		}
		commits = append(commits, Commit{ID: id, Parent: parent})
		previous = id
	}
	if len(commits) == 0 || commits[len(commits)-1].ID != strings.ToLower(head) {
		return nil, fmt.Errorf("INTEGRITY_FAILURE: history does not terminate at authoritative head")
	}
	return commits, nil
}

func (r *Reader) EventAdditions(ctx context.Context, commit string) ([]EventAddition, error) {
	out, err := r.run(ctx, "diff-tree", "--root", "--no-commit-id", "--name-status", "-r", "-z", commit, "--", "events")
	if err != nil {
		return nil, err
	}
	tokens := bytes.Split(out, []byte{0})
	additions := []EventAddition{}
	for i := 0; i < len(tokens); {
		if len(tokens[i]) == 0 {
			i++
			continue
		}
		status := string(tokens[i])
		i++
		if i >= len(tokens) || len(tokens[i]) == 0 {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: malformed Git name-status output")
		}
		path := string(tokens[i])
		i++
		if status != "A" {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: durable event path %q changed with status %q; event files are immutable", path, status)
		}
		if !strings.HasPrefix(path, "events/") {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: unexpected event tree path %q", path)
		}
		if !strings.HasSuffix(path, ".json") {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: durable event file %q is not JSON", path)
		}
		if err := r.requireRegularBlobAt(ctx, commit, path); err != nil {
			return nil, err
		}
		additions = append(additions, EventAddition{Commit: strings.ToLower(commit), Path: path})
	}
	sort.Slice(additions, func(i, j int) bool { return additions[i].Path < additions[j].Path })
	return additions, nil
}

func (r *Reader) ListJSON(ctx context.Context, commit, prefix string) ([]string, error) {
	if strings.ContainsAny(prefix, "\x00\r\n") {
		return nil, fmt.Errorf("invalid tree prefix")
	}
	out, err := r.run(ctx, "ls-tree", "-r", "-z", "--name-only", commit, "--", prefix)
	if err != nil {
		return nil, err
	}
	parts := bytes.Split(out, []byte{0})
	paths := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		path := string(p)
		if strings.HasSuffix(path, ".json") {
			if err := r.requireRegularBlobAt(ctx, commit, path); err != nil {
				return nil, err
			}
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths, nil
}

func (r *Reader) ReadFile(ctx context.Context, commit, path string) ([]byte, error) {
	if !isObjectID(commit) || path == "" || strings.ContainsAny(path, "\x00\r\n") {
		return nil, fmt.Errorf("invalid Git object/path request")
	}
	out, err := r.run(ctx, "cat-file", "blob", commit+":"+path)
	if err != nil {
		return nil, fmt.Errorf("read %s at %s: %w", path, commit, err)
	}
	return out, nil
}

func (r *Reader) run(parent context.Context, args ...string) ([]byte, error) {
	if err := r.checkRepositoryRootSafety(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(parent, r.timeout)
	defer cancel()
	base := []string{"--no-replace-objects", "--git-dir=" + r.gitDir}
	cmd := exec.CommandContext(ctx, r.gitPath, append(base, args...)...)
	cmd.Env = controlledEnv()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("GIT_FAILURE: command timed out")
		}
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > 800 {
			msg = msg[:800]
		}
		command := "git"
		if len(args) > 0 {
			command += " " + args[0]
		}
		return nil, fmt.Errorf("GIT_FAILURE: %s failed: %w: %s", command, err, msg)
	}
	return stdout.Bytes(), nil
}

func controlledEnv() []string {
	blocked := []string{
		"GIT_DIR=", "GIT_WORK_TREE=", "GIT_INDEX_FILE=", "GIT_OBJECT_DIRECTORY=", "GIT_ALTERNATE_OBJECT_DIRECTORIES=",
		"GIT_REPLACE_REF_BASE=", "GIT_CONFIG=", "GIT_CONFIG_SYSTEM=", "GIT_CONFIG_GLOBAL=", "GIT_CONFIG_NOSYSTEM=", "GIT_CONFIG_COUNT=",
		"GIT_CONFIG_KEY_", "GIT_CONFIG_VALUE_", "GIT_CONFIG_PARAMETERS=", "GIT_COMMON_DIR=", "GIT_NAMESPACE=", "GIT_SHALLOW_FILE=",
		"GIT_GRAFT_FILE=", "GIT_EXEC_PATH=", "GIT_TEMPLATE_DIR=", "GIT_SSH=", "GIT_SSH_COMMAND=", "GIT_ASKPASS=", "SSH_ASKPASS=",
		"GIT_TERMINAL_PROMPT=", "GIT_PAGER=", "PAGER=", "GIT_EDITOR=", "GIT_SEQUENCE_EDITOR=", "GIT_LITERAL_PATHSPECS=",
		"GIT_GLOB_PATHSPECS=", "GIT_NOGLOB_PATHSPECS=", "GIT_ICASE_PATHSPECS=", "GIT_ATTR_NOSYSTEM=", "GIT_NO_LAZY_FETCH=", "LC_ALL=", "LANG=",
	}
	env := make([]string, 0, len(os.Environ())+11)
	for _, item := range os.Environ() {
		blockedItem := false
		for _, prefix := range blocked {
			if strings.HasPrefix(item, prefix) {
				blockedItem = true
				break
			}
		}
		if !blockedItem {
			env = append(env, item)
		}
	}
	return append(env,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_TERMINAL_PROMPT=0",
		"GIT_PAGER=cat",
		"PAGER=cat",
		"GIT_OPTIONAL_LOCKS=0",
		"GIT_NO_REPLACE_OBJECTS=1",
		"GIT_ATTR_NOSYSTEM=1",
		"GIT_NO_LAZY_FETCH=1",
		"LC_ALL=C",
		"LANG=C",
	)
}

func isObjectID(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
