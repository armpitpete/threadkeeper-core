package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestIssue32UnresolvedRefLockContentionReturnsAcceptanceUnknown(t *testing.T) {
	r, head := candidateTestReader(t)
	defer r.Close()
	event := makeCreateCandidateEvent(t, head, "issue32-ref-lock-unknown-response", "idem-issue32-ref-lock-unknown-response", json.RawMessage(`{"enabled":true}`))
	candidate, response, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/issue32-ref-lock-unknown-response.json",
		Event:        event,
	})
	if err != nil {
		t.Fatal(err)
	}
	if candidate == nil || response != nil {
		t.Fatalf("unexpected prepare candidate=%#v response=%#v", candidate, response)
	}

	lockPath := filepath.Join(r.GitDir(), "refs", "heads", "main.lock")
	if err := os.WriteFile(lockPath, []byte(candidate.CandidateCommit+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(lockPath)

	response, err = AcceptWriteCandidate(context.Background(), r, *candidate)
	if err == nil || !errors.Is(err, gitledger.ErrCASOutcomeUnknown) {
		t.Fatalf("unresolved ref-lock contention error = %v, want POST_CAS_RECOVERY_REQUIRED", err)
	}
	if response == nil || response.Status != WriteStatusAcceptanceUnknown {
		t.Fatalf("unresolved ref-lock response = %#v, want acceptance_unknown", response)
	}
	if response.AcceptedCommit != "" {
		t.Fatalf("unknown outcome asserted accepted commit %s", response.AcceptedCommit)
	}
	got, headErr := r.Head(context.Background())
	if headErr != nil {
		t.Fatal(headErr)
	}
	if got != head {
		t.Fatalf("unresolved contention moved authority: got %s want %s", got, head)
	}
}
