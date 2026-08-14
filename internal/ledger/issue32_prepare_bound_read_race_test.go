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

func TestIssue32PrepareBoundReadCleanupRaceRecoversAlreadyAccepted(t *testing.T) {
	r, head := candidateTestReader(t)
	defer r.Close()

	event := makeCreateCandidateEvent(t, head, "issue32-bound-read-race", "idem-issue32-bound-read-race", json.RawMessage(`{"enabled":true}`))
	req := CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/issue32-bound-read-race.json",
		Event:        event,
	}
	winner, response, err := PrepareWriteCandidate(context.Background(), r, req)
	if err != nil {
		t.Fatal(err)
	}
	if winner == nil || response != nil {
		t.Fatalf("unexpected initial prepare candidate=%#v response=%#v", winner, response)
	}
	boundPath := filepath.Join(r.CandidateQuarantineDir(), winner.Quarantine.ID+".candidate")
	if _, err := os.Stat(boundPath); err != nil {
		t.Fatal(err)
	}

	afterEnsure := make(chan struct{})
	releaseRead := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(releaseRead) })

	type prepareResult struct {
		candidate *WriteCandidate
		response  *WriteResponse
		err       error
	}
	loserDone := make(chan prepareResult, 1)
	go func() {
		candidate, response, err := prepareWriteCandidate(context.Background(), r, req, &prepareWriteHooks{
			afterBoundEnsure: func() {
				close(afterEnsure)
				<-releaseRead
			},
		})
		loserDone <- prepareResult{candidate: candidate, response: response, err: err}
	}()

	<-afterEnsure

	accepted, err := AcceptWriteCandidate(context.Background(), r, *winner)
	if err != nil {
		t.Fatal(err)
	}
	if accepted == nil || accepted.Status != WriteStatusAccepted {
		t.Fatalf("winner response = %#v, want accepted", accepted)
	}
	if accepted.AcceptedCommit != winner.CandidateCommit {
		t.Fatalf("winner accepted commit = %s want %s", accepted.AcceptedCommit, winner.CandidateCommit)
	}
	if _, err := os.Stat(boundPath); !os.IsNotExist(err) {
		t.Fatalf("winner did not remove bound quarantine entry: %v", err)
	}

	releaseOnce.Do(func() { close(releaseRead) })
	loser := <-loserDone
	if loser.err != nil {
		t.Fatalf("racing Prepare returned ordinary failure after durable acceptance: %v", loser.err)
	}
	if loser.candidate != nil {
		t.Fatalf("racing Prepare returned stale candidate %#v", loser.candidate)
	}
	if loser.response == nil || loser.response.Status != WriteStatusAlreadyAccepted {
		t.Fatalf("racing Prepare response = %#v, want already_accepted", loser.response)
	}
	if loser.response.AcceptedCommit != winner.CandidateCommit {
		t.Fatalf("racing Prepare accepted commit = %s want %s", loser.response.AcceptedCommit, winner.CandidateCommit)
	}

	stageMatches, err := filepath.Glob(filepath.Join(r.CandidateQuarantineDir(), quarantineStagePrefix+"*.candidate"))
	if err != nil {
		t.Fatal(err)
	}
	if len(stageMatches) != 0 {
		t.Fatalf("racing Prepare left staging material: %v", stageMatches)
	}
	if _, err := os.Stat(boundPath); !os.IsNotExist(err) {
		t.Fatalf("racing Prepare resurrected accepted bound quarantine: %v", err)
	}

	restarted, err := gitledger.New(r.GitDir(), gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	reprepared, retry, err := PrepareWriteCandidate(context.Background(), restarted, req)
	if err != nil {
		t.Fatal(err)
	}
	if reprepared != nil || retry == nil || retry.Status != WriteStatusAlreadyAccepted {
		t.Fatalf("restart retry candidate=%#v response=%#v", reprepared, retry)
	}
	if retry.AcceptedCommit != winner.CandidateCommit {
		t.Fatalf("restart accepted commit = %s want %s", retry.AcceptedCommit, winner.CandidateCommit)
	}
}
