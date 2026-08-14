package ledger

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAlreadyAcceptedRetryCannotDeleteAnotherCandidateQuarantine(t *testing.T) {
	r, head := candidateTestReader(t)
	firstEvent := makeCreateCandidateEvent(t, head, "cleanup-first", "idem-cleanup-first", json.RawMessage(`{"enabled":true}`))
	first, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/cleanup-first.json",
		Event:        firstEvent,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AcceptWriteCandidate(context.Background(), r, *first); err != nil {
		t.Fatal(err)
	}

	secondEvent := makeCreateCandidateEventForTarget(t, first.CandidateCommit, "cleanup-second", "idem-cleanup-second", "setting:cleanup-second", json.RawMessage(`{"enabled":false}`))
	second, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: first.CandidateCommit,
		EventPath:    "events/governance/cleanup-second.json",
		Event:        secondEvent,
	})
	if err != nil {
		t.Fatal(err)
	}
	secondPath := filepath.Join(r.CandidateQuarantineDir(), second.Quarantine.ID+".candidate")
	if _, err := os.Stat(secondPath); err != nil {
		t.Fatal(err)
	}

	forgedRetry := *first
	forgedRetry.Quarantine = second.Quarantine
	response, err := AcceptWriteCandidate(context.Background(), r, forgedRetry)
	if err != nil {
		t.Fatal(err)
	}
	if response == nil || response.Status != WriteStatusAlreadyAccepted {
		t.Fatalf("unexpected retry response %#v", response)
	}
	if _, err := os.Stat(secondPath); err != nil {
		t.Fatalf("accepted retry deleted another pending candidate quarantine: %v", err)
	}
}
