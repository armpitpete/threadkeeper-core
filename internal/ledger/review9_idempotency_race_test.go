package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestStaleCASRecoveryClassifiesConflictingAcceptedIdempotency(t *testing.T) {
	r, head := candidateTestReader(t)
	const key = "idem-race-conflict"

	winnerEvent := makeCreateCandidateEventForTarget(t, head, "candidate-winner", key, testTargetID, json.RawMessage(`{"enabled":true}`))
	loserEvent := makeCreateCandidateEventForTarget(t, head, "candidate-loser", key, "setting:other", json.RawMessage(`{"enabled":false}`))

	winner, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/candidate-winner.json",
		Event:        winnerEvent,
	})
	if err != nil {
		t.Fatal(err)
	}
	loser, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/candidate-loser.json",
		Event:        loserEvent,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := AcceptWriteCandidate(context.Background(), r, *winner); err != nil {
		t.Fatal(err)
	}

	response, err := recoverCandidateAfterStaleCAS(context.Background(), r, *loser)
	if err == nil || !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("expected post-race idempotency conflict, response=%#v err=%v", response, err)
	}

	got, err := r.Head(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != winner.CandidateCommit {
		t.Fatalf("conflict classification moved authority: got %s want %s", got, winner.CandidateCommit)
	}
}

func TestStaleCASRecoveryReturnsExactDurableRetry(t *testing.T) {
	r, head := candidateTestReader(t)
	event := makeCreateCandidateEvent(t, head, "candidate-retry-race", "idem-race-retry", json.RawMessage(`{"enabled":true}`))
	candidate, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/candidate-retry-race.json",
		Event:        event,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AcceptWriteCandidate(context.Background(), r, *candidate); err != nil {
		t.Fatal(err)
	}

	response, err := recoverCandidateAfterStaleCAS(context.Background(), r, *candidate)
	if err != nil {
		t.Fatal(err)
	}
	if response == nil || response.Status != WriteStatusAlreadyAccepted {
		t.Fatalf("expected durable retry response, got %#v", response)
	}
	if response.AcceptedCommit != candidate.CandidateCommit || response.LedgerCommit != candidate.CandidateCommit {
		t.Fatalf("unexpected durable retry identity: %#v", response)
	}
}
