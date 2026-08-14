package gitledger

import (
	"bytes"
	"context"
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

	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(signalPath); err == nil {
			break
		}
		if time.Now().After(deadline) {
			cancel()
			t.Fatal("timed out waiting for real update-ref to complete")
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()

	if err := <-done; err != nil {
		t.Fatalf("CAS was actually accepted but caller cancellation was reported as failure: %v", err)
	}
	got, err := r.Head(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != candidate.Commit {
		t.Fatalf("recovered authoritative head = %s, want accepted candidate %s", got, candidate.Commit)
	}
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
