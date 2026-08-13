package ledger

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestReplayRejectsRepositoryLocalObjectAlternates(t *testing.T) {
	for _, name := range []string{"alternates", "http-alternates"} {
		t.Run(name, func(t *testing.T) {
			work := newWorkRepo(t)
			writeSchema(t, work)
			writeEvent(t, work, "events/decisions/001.json", "event-1", "decision.accepted", "target-a")
			commitAll(t, work, "accept event")
			bare := cloneBare(t, work)

			alternateObjects := filepath.Join(t.TempDir(), "objects")
			if err := os.MkdirAll(alternateObjects, 0o755); err != nil {
				t.Fatal(err)
			}
			info := filepath.Join(bare, "objects", "info")
			if err := os.MkdirAll(info, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(info, name), []byte(alternateObjects+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}

			r, err := gitledger.New(bare, gitledger.DefaultRef)
			if err != nil {
				t.Fatal(err)
			}
			_, err = Replay(context.Background(), r)
			if err == nil || !strings.Contains(err.Error(), "object alternates") {
				t.Fatalf("expected repository-local alternates rejection, got %v", err)
			}
		})
	}
}

func TestAcceptRejectsRepositoryLocalAlternatesBeforeRefMovement(t *testing.T) {
	r, head := candidateTestReader(t)
	event := makeCreateCandidateEvent(t, head, "candidate-alternate", "idem-alternate", json.RawMessage(`{"enabled":true}`))
	candidate, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/candidate-alternate.json",
		Event:        event,
	})
	if err != nil {
		t.Fatal(err)
	}

	alternateObjects := filepath.Join(t.TempDir(), "objects")
	if err := os.MkdirAll(alternateObjects, 0o755); err != nil {
		t.Fatal(err)
	}
	info := filepath.Join(r.GitDir(), "objects", "info")
	if err := os.MkdirAll(info, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(info, "alternates"), []byte(alternateObjects+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = AcceptWriteCandidate(context.Background(), r, *candidate)
	if err == nil || !strings.Contains(err.Error(), "object alternates") {
		t.Fatalf("expected acceptance to reject repository-local alternates, got %v", err)
	}
	if err := os.Remove(filepath.Join(info, "alternates")); err != nil {
		t.Fatal(err)
	}
	got, err := r.Head(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != head {
		t.Fatalf("alternates rejection moved authority: got %s want %s", got, head)
	}
}

func TestIdempotencyLookupIsBoundToExactReplaySnapshot(t *testing.T) {
	r, head := candidateTestReader(t)
	event := makeCreateCandidateEvent(t, head, "candidate-snapshot", "idem-snapshot", json.RawMessage(`{"enabled":true}`))
	req := CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/candidate-snapshot.json",
		Event:        event,
	}
	candidate, _, err := PrepareWriteCandidate(context.Background(), r, req)
	if err != nil {
		t.Fatal(err)
	}

	// H0 is the captured replay snapshot. Accept the event after that snapshot,
	// exactly matching the race that previously let an H1 event be paired with
	// LedgerCommit=H0.
	if err := r.CompareAndSwap(context.Background(), head, candidate.CandidateCommit); err != nil {
		t.Fatal(err)
	}

	atOldSnapshot, err := findAcceptedIdempotencyAt(context.Background(), r, head, candidate.IdempotencyKey)
	if err != nil {
		t.Fatal(err)
	}
	if atOldSnapshot != nil {
		t.Fatalf("H0 snapshot observed event accepted only at H1: %#v", atOldSnapshot)
	}
	atNewSnapshot, err := findAcceptedIdempotencyAt(context.Background(), r, candidate.CandidateCommit, candidate.IdempotencyKey)
	if err != nil {
		t.Fatal(err)
	}
	if atNewSnapshot == nil || atNewSnapshot.Entry.AcceptedCommit != candidate.CandidateCommit {
		t.Fatalf("H1 snapshot did not recover acceptance: %#v", atNewSnapshot)
	}

	_, response, err := PrepareWriteCandidate(context.Background(), r, req)
	if err != nil {
		t.Fatal(err)
	}
	if response == nil || response.Status != WriteStatusAlreadyAccepted {
		t.Fatalf("retry did not return durable acceptance: %#v", response)
	}
	if response.AcceptedCommit != candidate.CandidateCommit || response.LedgerCommit != candidate.CandidateCommit {
		t.Fatalf("snapshot-inconsistent retry response: %#v", response)
	}
}
