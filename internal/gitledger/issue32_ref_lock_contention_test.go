package gitledger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestIssue32RefLockContentionSettlesToRecoveredAcceptance(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("reference-transaction hook fixture is exercised by the Linux conformance lane")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	gitDir, r, head := issue21CASReader(t, realGit)
	defer r.Close()
	candidate, err := r.PrepareEventCommit(context.Background(), head, "events/issue32/ref-lock.json", []byte("{}"), "issue32-ref-lock")
	if err != nil {
		t.Fatal(err)
	}

	// Make the losing update-ref fail quickly while the legitimate winning Git
	// process is held in the prepared reference-transaction state with the ref
	// lock acquired and H0 still authoritative.
	runIssue21Git(t, realGit, nil, nil, "--git-dir="+gitDir, "config", "core.filesRefLockTimeout", "50")
	hooksDir := t.TempDir()
	preparedSignal := filepath.Join(t.TempDir(), "winner-prepared")
	releaseWinner := filepath.Join(t.TempDir(), "release-winner")
	hookPath := filepath.Join(hooksDir, "reference-transaction")
	hook := fmt.Sprintf(`#!/bin/sh
SIGNAL=%s
RELEASE=%s
if [ "$1" = "prepared" ]; then
  : > "$SIGNAL"
  while [ ! -f "$RELEASE" ]; do
    sleep 0.01
  done
fi
exit 0
`, issue21ShellQuote(preparedSignal), issue21ShellQuote(releaseWinner))
	if err := os.WriteFile(hookPath, []byte(hook), 0o700); err != nil {
		t.Fatal(err)
	}

	winnerDone := make(chan error, 1)
	go func() {
		cmd := exec.Command(realGit, "--git-dir="+gitDir, "-c", "core.hooksPath="+hooksDir, "update-ref", "--no-deref", DefaultRef, candidate.Commit, head)
		out, err := cmd.CombinedOutput()
		if err != nil {
			winnerDone <- fmt.Errorf("winner update-ref failed: %w: %s", err, out)
			return
		}
		winnerDone <- nil
	}()
	waitForIssue32Path(t, preparedSignal)

	initialRecovery := make(chan recoveredCASState, 1)
	releaseLoserRecovery := make(chan struct{})
	loserDone := make(chan error, 1)
	go func() {
		loserDone <- r.compareAndSwap(context.Background(), head, candidate.Commit, &compareAndSwapHooks{
			afterInitialUpdateErrorRecovery: func(state recoveredCASState) {
				initialRecovery <- state
				<-releaseLoserRecovery
			},
		})
	}()

	state := <-initialRecovery
	if state.Head != head || state.CandidateInHistory {
		t.Fatalf("loser initial recovery = %#v, want H0 with H1 absent", state)
	}

	if err := os.WriteFile(releaseWinner, []byte("release\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := <-winnerDone; err != nil {
		t.Fatal(err)
	}
	close(releaseLoserRecovery)

	loserErr := <-loserDone
	if loserErr == nil || !errors.Is(loserErr, ErrCASAcceptanceRecovered) {
		t.Fatalf("losing identical CAS did not recover winner acceptance: %v", loserErr)
	}
	got, err := r.Head(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != candidate.Commit {
		t.Fatalf("authoritative head = %s want recovered H1 %s", got, candidate.Commit)
	}
}

func TestIssue32UnresolvedRefLockContentionIsExplicitUnknown(t *testing.T) {
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	gitDir, r, head := issue21CASReader(t, realGit)
	defer r.Close()
	candidate, err := r.PrepareEventCommit(context.Background(), head, "events/issue32/ref-lock-unknown.json", []byte("{}"), "issue32-ref-lock-unknown")
	if err != nil {
		t.Fatal(err)
	}
	runIssue21Git(t, realGit, nil, nil, "--git-dir="+gitDir, "config", "core.filesRefLockTimeout", "50")

	lockPath := filepath.Join(gitDir, "refs", "heads", "main.lock")
	if err := os.WriteFile(lockPath, []byte(candidate.Commit+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(lockPath)

	err = r.CompareAndSwap(context.Background(), head, candidate.Commit)
	if err == nil || !errors.Is(err, ErrCASOutcomeUnknown) {
		t.Fatalf("unresolved ref-lock contention = %v, want explicit POST_CAS_RECOVERY_REQUIRED", err)
	}
	got, headErr := r.Head(context.Background())
	if headErr != nil {
		t.Fatal(headErr)
	}
	if got != head {
		t.Fatalf("unresolved contention moved authority: got %s want %s", got, head)
	}
}

func waitForIssue32Path(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", path)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
