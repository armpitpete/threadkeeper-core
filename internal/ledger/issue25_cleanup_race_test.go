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

	<-beforeRead

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

func TestIssue25PrepareRacingAcceptanceReconcilesAlreadyAccepted(t *testing.T) {
	r, head := candidateTestReader(t)
	defer r.Close()

	event := makeCreateCandidateEvent(t, head, "issue25-prepare-race", "idem-issue25-prepare-race", json.RawMessage(`{"enabled":true}`))
	req := CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/issue25-prepare-race.json",
		Event:        event,
	}
	winnerCandidate, response, err := PrepareWriteCandidate(context.Background(), r, req)
	if err != nil {
		t.Fatal(err)
	}
	if winnerCandidate == nil || response != nil {
		t.Fatalf("unexpected initial prepare candidate=%#v response=%#v", winnerCandidate, response)
	}

	qPath := filepath.Join(r.CandidateQuarantineDir(), winnerCandidate.Quarantine.ID+".candidate")
	if _, err := os.Stat(qPath); err != nil {
		t.Fatal(err)
	}

	beforeFinalCheck := make(chan struct{})
	releaseFinalCheck := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(releaseFinalCheck) })

	type prepareResult struct {
		candidate *WriteCandidate
		response  *WriteResponse
		err       error
	}
	prepareDone := make(chan prepareResult, 1)
	go func() {
		candidate, response, err := prepareWriteCandidate(context.Background(), r, req, &prepareWriteHooks{
			beforeFinalSnapshotCheck: func() {
				close(beforeFinalCheck)
				<-releaseFinalCheck
			},
		})
		prepareDone <- prepareResult{candidate: candidate, response: response, err: err}
	}()

	<-beforeFinalCheck

	accepted, err := AcceptWriteCandidate(context.Background(), r, *winnerCandidate)
	if err != nil {
		t.Fatal(err)
	}
	if accepted == nil || accepted.Status != WriteStatusAccepted {
		t.Fatalf("winner response = %#v, want accepted", accepted)
	}
	if accepted.AcceptedCommit != winnerCandidate.CandidateCommit {
		t.Fatalf("winner accepted commit = %s want %s", accepted.AcceptedCommit, winnerCandidate.CandidateCommit)
	}
	if _, err := os.Stat(qPath); !os.IsNotExist(err) {
		t.Fatalf("winner did not remove bound quarantine entry: %v", err)
	}

	releaseOnce.Do(func() { close(releaseFinalCheck) })
	prepared := <-prepareDone
	if prepared.err != nil {
		t.Fatalf("racing Prepare failed instead of recovering durable acceptance: %v", prepared.err)
	}
	if prepared.candidate != nil {
		t.Fatalf("racing Prepare returned stale candidate %#v", prepared.candidate)
	}
	if prepared.response == nil || prepared.response.Status != WriteStatusAlreadyAccepted {
		t.Fatalf("racing Prepare response = %#v, want already_accepted", prepared.response)
	}
	if prepared.response.AcceptedCommit != winnerCandidate.CandidateCommit {
		t.Fatalf("racing Prepare accepted commit = %s want %s", prepared.response.AcceptedCommit, winnerCandidate.CandidateCommit)
	}
	if _, err := os.Stat(qPath); !os.IsNotExist(err) {
		t.Fatalf("racing Prepare resurrected or retained accepted quarantine entry: %v", err)
	}
}

func TestIssue25StagedPrepareRacingAcceptanceCleansStageAndRecovers(t *testing.T) {
	r, head := candidateTestReader(t)
	defer r.Close()

	event := makeCreateCandidateEvent(t, head, "issue25-stage-race", "idem-issue25-stage-race", json.RawMessage(`{"enabled":true}`))
	req := CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/issue25-stage-race.json",
		Event:        event,
	}
	winnerCandidate, response, err := PrepareWriteCandidate(context.Background(), r, req)
	if err != nil {
		t.Fatal(err)
	}
	if winnerCandidate == nil || response != nil {
		t.Fatalf("unexpected initial prepare candidate=%#v response=%#v", winnerCandidate, response)
	}

	boundPath := filepath.Join(r.CandidateQuarantineDir(), winnerCandidate.Quarantine.ID+".candidate")
	beforeEventCommit := make(chan struct{})
	releaseEventCommit := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(releaseEventCommit) })

	type prepareResult struct {
		candidate *WriteCandidate
		response  *WriteResponse
		err       error
	}
	prepareDone := make(chan prepareResult, 1)
	go func() {
		candidate, response, err := prepareWriteCandidate(context.Background(), r, req, &prepareWriteHooks{
			beforeEventCommit: func() {
				close(beforeEventCommit)
				<-releaseEventCommit
			},
		})
		prepareDone <- prepareResult{candidate: candidate, response: response, err: err}
	}()

	<-beforeEventCommit
	stageMatches, err := filepath.Glob(filepath.Join(r.CandidateQuarantineDir(), quarantineStagePrefix+"*.candidate"))
	if err != nil {
		t.Fatal(err)
	}
	if len(stageMatches) != 1 {
		t.Fatalf("racing Prepare staging entries = %v, want exactly one private stage", stageMatches)
	}
	stagePath := stageMatches[0]

	accepted, err := AcceptWriteCandidate(context.Background(), r, *winnerCandidate)
	if err != nil {
		t.Fatal(err)
	}
	if accepted == nil || accepted.Status != WriteStatusAccepted {
		t.Fatalf("winner response = %#v, want accepted", accepted)
	}
	if _, err := os.Stat(boundPath); !os.IsNotExist(err) {
		t.Fatalf("winner did not clean bound quarantine entry: %v", err)
	}

	releaseOnce.Do(func() { close(releaseEventCommit) })
	prepared := <-prepareDone
	if prepared.err != nil {
		t.Fatalf("staged racing Prepare failed instead of recovering acceptance: %v", prepared.err)
	}
	if prepared.candidate != nil {
		t.Fatalf("staged racing Prepare returned stale candidate %#v", prepared.candidate)
	}
	if prepared.response == nil || prepared.response.Status != WriteStatusAlreadyAccepted {
		t.Fatalf("staged racing Prepare response = %#v, want already_accepted", prepared.response)
	}
	if prepared.response.AcceptedCommit != winnerCandidate.CandidateCommit {
		t.Fatalf("staged racing Prepare accepted commit = %s want %s", prepared.response.AcceptedCommit, winnerCandidate.CandidateCommit)
	}
	if _, err := os.Stat(stagePath); !os.IsNotExist(err) {
		t.Fatalf("staged racing Prepare left abandoned staging material: %v", err)
	}
	if _, err := os.Stat(boundPath); !os.IsNotExist(err) {
		t.Fatalf("staged racing Prepare resurrected accepted bound quarantine: %v", err)
	}
}
