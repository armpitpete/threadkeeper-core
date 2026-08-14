package ledger

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func preparedQuarantineCandidate(t *testing.T) (*gatedCandidateFixture, []byte) {
	t.Helper()
	r, head := candidateTestReader(t)
	event := makeCreateCandidateEvent(t, head, "quarantine-candidate", "idem-quarantine-candidate", json.RawMessage(`{"enabled":true}`))
	candidate, accepted, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/quarantine-candidate.json",
		Event:        event,
	})
	if err != nil {
		t.Fatal(err)
	}
	if accepted != nil || candidate == nil {
		t.Fatalf("unexpected prepare result candidate=%#v accepted=%#v", candidate, accepted)
	}
	return &gatedCandidateFixture{reader: r, head: head, candidate: candidate}, event
}

type gatedCandidateFixture struct {
	reader    interface {
		Head(context.Context) (string, error)
		ReadFile(context.Context, string, string) ([]byte, error)
		CandidateQuarantineDir() string
	}
	head      string
	candidate *WriteCandidate
}

func TestPrepareMaterialisesExactBytesThroughQuarantine(t *testing.T) {
	fixture, event := preparedQuarantineCandidate(t)
	qPath := filepath.Join(fixture.reader.CandidateQuarantineDir(), fixture.candidate.Quarantine.ID+".candidate")
	quarantined, err := os.ReadFile(qPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(quarantined) != string(event) {
		t.Fatal("quarantined bytes differ from request")
	}
	stored, err := fixture.reader.ReadFile(context.Background(), fixture.candidate.CandidateCommit, fixture.candidate.EventPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(stored) != string(quarantined) {
		t.Fatal("Git candidate bytes differ from quarantine")
	}
	got, err := fixture.reader.Head(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != fixture.head {
		t.Fatalf("prepare moved authority: got %s want %s", got, fixture.head)
	}
}

func TestAcceptanceFailsClosedWhenQuarantineEntryMissing(t *testing.T) {
	fixture, _ := preparedQuarantineCandidate(t)
	qPath := filepath.Join(fixture.reader.CandidateQuarantineDir(), fixture.candidate.Quarantine.ID+".candidate")
	if err := os.Remove(qPath); err != nil {
		t.Fatal(err)
	}
	if _, err := AcceptWriteCandidate(context.Background(), fixture.reader.(*gitledger.Reader), *fixture.candidate); err == nil || !strings.Contains(err.Error(), "CANDIDATE_INVALID") {
		t.Fatalf("expected missing quarantine rejection, got %v", err)
	}
	assertHeadUnchanged(t, fixture)
}

func TestAcceptanceFailsClosedWhenQuarantineBytesChange(t *testing.T) {
	fixture, _ := preparedQuarantineCandidate(t)
	qPath := filepath.Join(fixture.reader.CandidateQuarantineDir(), fixture.candidate.Quarantine.ID+".candidate")
	if err := os.WriteFile(qPath, []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := AcceptWriteCandidate(context.Background(), fixture.reader.(*gitledger.Reader), *fixture.candidate); err == nil || !strings.Contains(err.Error(), "CANDIDATE_INVALID") {
		t.Fatalf("expected tampered quarantine rejection, got %v", err)
	}
	assertHeadUnchanged(t, fixture)
}

func TestAcceptanceRejectsSubstitutedQuarantineHandle(t *testing.T) {
	fixture, _ := preparedQuarantineCandidate(t)
	forged := *fixture.candidate
	forged.Quarantine.ID += "x"
	if _, err := AcceptWriteCandidate(context.Background(), fixture.reader.(*gitledger.Reader), forged); err == nil || !strings.Contains(err.Error(), "CANDIDATE_INVALID") {
		t.Fatalf("expected substituted quarantine handle rejection, got %v", err)
	}
	assertHeadUnchanged(t, fixture)
}

func TestAcceptedCandidateQuarantineIsRemoved(t *testing.T) {
	fixture, _ := preparedQuarantineCandidate(t)
	qPath := filepath.Join(fixture.reader.CandidateQuarantineDir(), fixture.candidate.Quarantine.ID+".candidate")
	if _, err := os.Stat(qPath); err != nil {
		t.Fatal(err)
	}
	response, err := AcceptWriteCandidate(context.Background(), fixture.reader.(*gitledger.Reader), *fixture.candidate)
	if err != nil {
		t.Fatal(err)
	}
	if response == nil || response.Status != WriteStatusAccepted {
		t.Fatalf("unexpected response %#v", response)
	}
	if _, err := os.Stat(qPath); !os.IsNotExist(err) {
		t.Fatalf("accepted quarantine entry still exists: %v", err)
	}
}

func assertHeadUnchanged(t *testing.T, fixture *gatedCandidateFixture) {
	t.Helper()
	got, err := fixture.reader.Head(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != fixture.head {
		t.Fatalf("failed acceptance moved authority: got %s want %s", got, fixture.head)
	}
}
