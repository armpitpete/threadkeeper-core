package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestPinnedOpenRootHandleRejectsStableRepositoryReplacementBeforeCAS(t *testing.T) {
	r, head := candidateTestReader(t)
	t.Cleanup(func() { _ = r.Close() })

	event := makeCreateCandidateEvent(t, head, "candidate-root-replacement", "idem-root-replacement", json.RawMessage(`{"enabled":true}`))
	candidate, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/candidate-root-replacement.json",
		Event:        event,
	})
	if err != nil {
		t.Fatal(err)
	}

	root := r.GitDir()
	replacement := root + "-replacement"
	copyRepositoryTree(t, root, replacement)

	// Remove the entire original repository while Reader still holds its live
	// root-directory handle, then install a complete ordinary replacement at the
	// exact same pathname. The replacement already contains both H0 and the
	// prepared unreachable H1, so if identity pinning is fooled the exact
	// update-ref CAS is capable of succeeding. This is the inode-reuse attack
	// from the independent review, not a symlink and not a check-to-exec race.
	if err := os.RemoveAll(root); err != nil {
		// Some platforms refuse removal while a directory handle is open. That is
		// itself stronger than the required invariant, so the attack is not
		// reproducible there.
		t.Skipf("platform prevents stable root replacement while pinned: %v", err)
	}
	if _, err := os.Lstat(root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("original repository root still exists after removal: %v", err)
	}
	if err := os.Rename(replacement, root); err != nil {
		t.Skipf("platform prevents installing stable replacement while pinned: %v", err)
	}

	_, err = AcceptWriteCandidate(context.Background(), r, *candidate)
	if err == nil || !strings.Contains(err.Error(), "filesystem identity changed") {
		t.Fatalf("expected stable same-path replacement rejection, got %v", err)
	}

	// Prove the replacement repository's authority ref stayed at H0. If the
	// historical snapshot-only os.SameFile check were fooled by inode reuse,
	// this would instead become candidate.CandidateCommit.
	got := runGitInDirWithEnv(t, root, nil, nil, "rev-parse", "--verify", "refs/heads/main^{commit}")
	if got != head {
		t.Fatalf("replacement repository authority moved: got %s want %s", got, head)
	}
}

func TestReaderCloseReleasesRootHandleAndFailsClosed(t *testing.T) {
	r, _ := candidateTestReader(t)
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := r.Head(context.Background())
	if err == nil || !strings.Contains(err.Error(), "Reader is closed") {
		t.Fatalf("expected closed Reader to fail closed, got %v", err)
	}
}

func TestSymbolicAuthoritativeRefCannotRedirectCAS(t *testing.T) {
	r, head := candidateTestReader(t)
	t.Cleanup(func() { _ = r.Close() })
	event := makeCreateCandidateEvent(t, head, "candidate-symref", "idem-symref", json.RawMessage(`{"enabled":true}`))
	candidate, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/candidate-symref.json",
		Event:        event,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create a different direct ref at H0, then turn the configured authority
	// ref into an ordinary Git symbolic-ref file that points at it. This is not
	// an OS symlink and was the attack class missed by earlier filesystem checks.
	runGitInDirWithEnv(t, r.GitDir(), nil, nil, "update-ref", "refs/heads/redirect-target", head)
	mainPath := filepath.Join(r.GitDir(), "refs", "heads", "main")
	if err := os.WriteFile(mainPath, []byte("ref: refs/heads/redirect-target\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = AcceptWriteCandidate(context.Background(), r, *candidate)
	if err == nil || !strings.Contains(err.Error(), "symbolic authoritative refs are forbidden") {
		t.Fatalf("expected symbolic authoritative ref rejection, got %v", err)
	}

	// The target ref must remain H0: the failed authority attempt must not have
	// followed the symbolic ref and mutated the redirect target.
	gotTarget := runGitInDirWithEnv(t, r.GitDir(), nil, nil, "rev-parse", "--verify", "refs/heads/redirect-target^{commit}")
	if gotTarget != head {
		t.Fatalf("symbolic ref redirected authority movement: target=%s want=%s", gotTarget, head)
	}

	// A newly opened Reader must reject the same static symbolic authority ref.
	if reopened, err := gitledger.New(r.GitDir(), gitledger.DefaultRef); err == nil {
		_ = reopened.Close()
		t.Fatalf("new Reader accepted symbolic authoritative ref")
	} else if !strings.Contains(err.Error(), "symbolic authoritative refs are forbidden") {
		t.Fatalf("unexpected symbolic-ref rejection: %v", err)
	}
}

func copyRepositoryTree(t *testing.T, src, dst string) {
	t.Helper()
	if err := filepath.WalkDir(src, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("unexpected symlink in authoritative repository fixture")
		}
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if !info.Mode().IsRegular() {
			return errors.New("unexpected non-regular file in authoritative repository fixture")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	}); err != nil {
		t.Fatal(err)
	}
}
