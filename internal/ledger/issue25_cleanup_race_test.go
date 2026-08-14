package ledger

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestIssue25WinnerCleanupRecoversConcurrentExactRequest(t *testing.T) {
	r, head := candidateTestReader(t)
	defer r.Close()

	event := makeCreateCandidateEvent(t, head, "issue25-cleanup-race", "idem-issue25-cleanup-race", json.RawMessage(`{"enabled":true}`))
	candidate, response, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/issue25-cleanup-race.json",
		Event:        event,
	})
	if err != nil {
		t.Fatal(err)
	}
	if candidate == nil || response != nil {
		t.Fatalf("unexpected prepare result candidate=%#v response=%#v", candidate, response)
	}

	qPath := filepath.Join(r.CandidateQuarantineDir(), candidate.Quarantine.ID+".candidate")
	if _, err := os.Stat(qPath); err != nil {
		t.Fatal(err)
	}

	beforeRead := make(chan struct{})
	releaseRead := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(releaseRead) })

	type result struct {
		response *WriteResponse
		err      error
	}
	bDone := make(chan result, 1)
	go func() {
		resp, err := acceptWriteCandidate(context.Background(), r, *candidate, &acceptWriteHooks{
			beforeQuarantineRead: func() {
				close(beforeRead)
				<-releaseRead
			},
		})
		bDone <- result{response: resp, err: err}
	}()

	// B has already replayed H0, found no accepted K, passed preflight, opened
	// the quarantine store, and is now paused immediately before Read(Q).
	<-beforeRead

	// A wins the exact same request, makes H1 authoritative, verifies it, and
	// removes Q as normal successful-acceptance cleanup.
	aResponse, err := AcceptWriteCandidate(context.Background(), r, *candidate)
	if err != nil {
		t.Fatal(err)
	}
	if aResponse == nil || aResponse.Status != WriteStatusAccepted {
		t.Fatalf("winner response = %#v, want accepted", aResponse)
	}
	if aResponse.AcceptedCommit != candidate.CandidateCommit {
		t.Fatalf("winner accepted commit = %s want %s", aResponse.AcceptedCommit, candidate.CandidateCommit)
	}
	if _, err := os.Stat(qPath); !os.IsNotExist(err) {
		t.Fatalf("winner did not remove bound quarantine entry: %v", err)
	}

	// B now observes the missing Q. It must fresh-replay authority and recover
	// the durable exact request as already_accepted rather than CANDIDATE_INVALID.
	releaseOnce.Do(func() { close(releaseRead) })
	b := <-bDone
	if b.err != nil {
		t.Fatalf("identical concurrent request failed after winner cleanup: %v", b.err)
	}
	if b.response == nil || b.response.Status != WriteStatusAlreadyAccepted {
		t.Fatalf("loser response = %#v, want already_accepted", b.response)
	}
	if b.response.AcceptedCommit != candidate.CandidateCommit {
		t.Fatalf("loser accepted commit = %s want %s", b.response.AcceptedCommit, candidate.CandidateCommit)
	}
	if b.response.LedgerCommit != aResponse.LedgerCommit {
		t.Fatalf("loser ledger commit = %s want winner snapshot %s", b.response.LedgerCommit, aResponse.LedgerCommit)
	}

	// Restart-style recovery must remain reconstructable solely from durable
	// authority after the quarantine entry is gone.
	restarted, err := gitledger.New(r.GitDir(), gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	reprepared, retry, err := PrepareWriteCandidate(context.Background(), restarted, CandidateRequest{
		ExpectedHead: head,
		EventPath:    candidate.EventPath,
		Event:        event,
	})
	if err != nil {
		t.Fatal(err)
	}
	if reprepared != nil || retry == nil || retry.Status != WriteStatusAlreadyAccepted {
		t.Fatalf("restart retry candidate=%#v response=%#v", reprepared, retry)
	}
	if retry.AcceptedCommit != candidate.CandidateCommit {
		t.Fatalf("restart retry accepted commit = %s want %s", retry.AcceptedCommit, candidate.CandidateCommit)
	}
}
