package ledger

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestRetryAfterCASCompletesQuarantineCleanup(t *testing.T) {
	fixture, event := preparedQuarantineCandidate(t)
	qPath := filepath.Join(fixture.reader.CandidateQuarantineDir(), fixture.candidate.Quarantine.ID+".candidate")
	if _, err := os.Stat(qPath); err != nil {
		t.Fatal(err)
	}

	// Simulate a crash after the authority-changing CAS but before the normal
	// post-acceptance verification/cleanup path can return.
	if err := fixture.reader.CompareAndSwap(context.Background(), fixture.candidate.ExpectedHead, fixture.candidate.CandidateCommit); err != nil {
		t.Fatal(err)
	}

	restarted, err := gitledger.New(fixture.reader.GitDir(), gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	candidate, response, err := PrepareWriteCandidate(context.Background(), restarted, CandidateRequest{
		ExpectedHead: fixture.head,
		EventPath:    fixture.candidate.EventPath,
		Event:        event,
	})
	if err != nil {
		t.Fatal(err)
	}
	if candidate != nil || response == nil || response.Status != WriteStatusAlreadyAccepted {
		t.Fatalf("unexpected retry result candidate=%#v response=%#v", candidate, response)
	}
	if _, err := os.Stat(qPath); !os.IsNotExist(err) {
		t.Fatalf("retry did not clean accepted quarantine entry: %v", err)
	}
}
