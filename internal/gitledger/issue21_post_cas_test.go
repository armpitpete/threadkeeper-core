package gitledger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestIssue21CompareAndSwapRecoversAfterCallerCancellationPostUpdateRef(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell wrapper is only used by the Linux conformance lane")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	_, r, head := issue21CASReader(t, realGit)
	defer r.Close()
	candidate, err := r.PrepareEventCommit(context.Background(), head, "events/issue21/event.json", []byte("{}"), "issue21-event")
	if err != nil {
		t.Fatal(err)
	}

	signalPath := filepath.Join(t.TempDir(), "update-ref-complete")
	wrapper := filepath.Join(t.TempDir(), "git-wrapper.sh")
	script := fmt.Sprintf(`#!/bin/sh
REAL_GIT=%s
SIGNAL=%s
for arg in "$@"; do
  if [ "$arg" = "update-ref" ]; then
    "$REAL_GIT" "$@"
    rc=$?
    if [ $rc -eq 0 ]; then
      : > "$SIGNAL"
      sleep 30
    fi
    exit $rc
  fi
done
exec "$REAL_GIT" "$@"
`, issue21ShellQuote(realGit), issue21ShellQuote(signalPath))
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	r.gitPath = wrapper

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- r.CompareAndSwap(ctx, head, candidate.Commit)
	}()

	waitForIssue21Signal(t, signalPath, cancel)
	cancel()

	err = <-done
	if err == nil || !errors.Is(err, ErrCASAcceptanceRecovered) {
		t.Fatalf("CAS was accepted under caller cancellation but did not return explicit recovered-acceptance condition: %v", err)
	}
	got, err := r.Head(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != candidate.Commit {
		t.Fatalf("recovered authoritative head = %s, want accepted candidate %s", got, candidate.Commit)
	}
}

func TestIssue21CancellationAfterH1ThenH2DoesNotMisreportOriginalCASAsStale(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell wrapper is only used by the Linux conformance lane")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	gitDir, r, head := issue21CASReader(t, realGit)
	defer r.Close()
	candidate, err := r.PrepareEventCommit(context.Background(), head, "events/issue21/first.json", []byte("{}"), "issue21-first")
	if err != nil {
		t.Fatal(err)
	}

	// Pre-create a later linear descendant. The wrapper will move H1->H2 only
	// after the real first H0->H1 update-ref succeeds, simulating another valid
	// writer completing before the cancelled caller's recovery reads the ref.
	h2 := runIssue21Git(t, realGit, []string{
		"GIT_AUTHOR_NAME=Threadkeeper Later Writer",
		"GIT_AUTHOR_EMAIL=later-writer@example.invalid",
		"GIT_AUTHOR_DATE=2002-02-02T02:02:02Z",
		"GIT_COMMITTER_NAME=Threadkeeper Later Writer",
		"GIT_COMMITTER_EMAIL=later-writer@example.invalid",
		"GIT_COMMITTER_DATE=2002-02-02T02:02:02Z",
	}, []byte("later authoritative descendant\n"), "--git-dir="+gitDir, "commit-tree", candidate.Tree, "-p", candidate.Commit)

	signalPath := filepath.Join(t.TempDir(), "h2-complete")
	wrapper := filepath.Join(t.TempDir(), "git-wrapper-descendant.sh")
	script := fmt.Sprintf(`#!/bin/sh
REAL_GIT=%s
GIT_DIR=%s
REF=%s
H1=%s
H2=%s
SIGNAL=%s
for arg in "$@"; do
  if [ "$arg" = "update-ref" ]; then
    "$REAL_GIT" "$@"
    rc=$?
    if [ $rc -eq 0 ]; then
      "$REAL_GIT" --git-dir="$GIT_DIR" update-ref --no-deref "$REF" "$H2" "$H1" || exit $?
      : > "$SIGNAL"
      sleep 30
    fi
    exit $rc
  fi
done
exec "$REAL_GIT" "$@"
`, issue21ShellQuote(realGit), issue21ShellQuote(gitDir), issue21ShellQuote(DefaultRef), issue21ShellQuote(candidate.Commit), issue21ShellQuote(h2), issue21ShellQuote(signalPath))
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	r.gitPath = wrapper

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- r.CompareAndSwap(ctx, head, candidate.Commit)
	}()

	waitForIssue21Signal(t, signalPath, cancel)
	cancel()
	err = <-done
	if err == nil || !errors.Is(err, ErrCASAcceptanceRecovered) {
		t.Fatalf("H1 is in authoritative history but recovery did not classify the ambiguous invocation explicitly: %v", err)
	}
	got, err := r.Head(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != h2 {
		t.Fatalf("recovered head = %s, want later descendant %s", got, h2)
	}
}

func TestIssue21SuccessfulCASPostSafetyFailureIsExplicitRecoveryCondition(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell wrapper is only used by the Linux conformance lane")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	gitDir, r, head := issue21CASReader(t, realGit)
	defer r.Close()
	candidate, err := r.PrepareEventCommit(context.Background(), head, "events/issue21/post-check.json", []byte("{}"), "issue21-post-check")
	if err != nil {
		t.Fatal(err)
	}

	alternates := filepath.Join(gitDir, "objects", "info", "alternates")
	wrapper := filepath.Join(t.TempDir(), "git-wrapper-post-check.sh")
	script := fmt.Sprintf(`#!/bin/sh
REAL_GIT=%s
ALTERNATES=%s
for arg in "$@"; do
  if [ "$arg" = "update-ref" ]; then
    "$REAL_GIT" "$@"
    rc=$?
    if [ $rc -eq 0 ]; then
      printf '/definitely/not/a/threadkeeper/object/store\n' > "$ALTERNATES"
    fi
    exit $rc
  fi
done
exec "$REAL_GIT" "$@"
`, issue21ShellQuote(realGit), issue21ShellQuote(alternates))
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	r.gitPath = wrapper

	err = r.CompareAndSwap(context.Background(), head, candidate.Commit)
	if err == nil || !errors.Is(err, ErrPostCASVerification) {
		t.Fatalf("successful CAS followed by safety failure must be explicit post-CAS verification failure, got %v", err)
	}

	// The test deliberately poisoned repository metadata only after update-ref.
	// Remove that poison and use the real Git executable solely to prove that H1
	// was in fact made authoritative despite the verification error.
	if removeErr := os.Remove(alternates); removeErr != nil {
		t.Fatal(removeErr)
	}
	r.gitPath = realGit
	got, headErr := r.Head(context.Background())
	if headErr != nil {
		t.Fatal(headErr)
	}
	if got != candidate.Commit {
		t.Fatalf("post-CAS verification failure hid authoritative move: got %s want %s", got, candidate.Commit)
	}
}

func waitForIssue21Signal(t *testing.T, signalPath string, cancel context.CancelFunc) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(signalPath); err == nil {
			return
		}
		if time.Now().After(deadline) {
			cancel()
			t.Fatal("timed out waiting for post-update-ref signal")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func issue21CASReader(t *testing.T, realGit string) (string, *Reader, string) {
	t.Helper()
	gitDir := filepath.Join(t.TempDir(), "ledger.git")
	runIssue21Git(t, realGit, nil, nil, "init", "--bare", gitDir)
	emptyTree := runIssue21Git(t, realGit, nil, nil, "--git-dir="+gitDir, "mktree")
	head := runIssue21Git(t, realGit, []string{
		"GIT_AUTHOR_NAME=Threadkeeper Test",
		"GIT_AUTHOR_EMAIL=threadkeeper-test@example.invalid",
		"GIT_AUTHOR_DATE=2000-01-01T00:00:00Z",
		"GIT_COMMITTER_NAME=Threadkeeper Test",
		"GIT_COMMITTER_EMAIL=threadkeeper-test@example.invalid",
		"GIT_COMMITTER_DATE=2000-01-01T00:00:00Z",
	}, []byte("root\n"), "--git-dir="+gitDir, "commit-tree", emptyTree)
	runIssue21Git(t, realGit, nil, nil, "--git-dir="+gitDir, "update-ref", DefaultRef, head)
	r, err := New(gitDir, DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	return gitDir, r, head
}

func runIssue21Git(t *testing.T, gitPath string, extraEnv []string, stdin []byte, args ...string) string {
	t.Helper()
	cmd := exec.Command(gitPath, args...)
	cmd.Env = append(os.Environ(), extraEnv...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v: %s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func issue21ShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
