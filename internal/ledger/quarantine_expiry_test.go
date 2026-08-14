package ledger

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestExpiredQuarantineCandidateCannotBeAccepted(t *testing.T) {
	fixture, _ := preparedQuarantineCandidate(t)
	qPath := filepath.Join(fixture.reader.CandidateQuarantineDir(), fixture.candidate.Quarantine.ID+".candidate")
	expired := time.Now().UTC().Add(-gitledger.CandidateQuarantineRetention - time.Hour)
	if err := os.Chtimes(qPath, expired, expired); err != nil {
		t.Fatal(err)
	}

	if _, err := AcceptWriteCandidate(context.Background(), fixture.reader, *fixture.candidate); err == nil || !strings.Contains(err.Error(), "CANDIDATE_INVALID") {
		t.Fatalf("expected expired quarantine rejection, got %v", err)
	}
	if _, err := os.Stat(qPath); !os.IsNotExist(err) {
		t.Fatalf("expired quarantine entry was not pruned: %v", err)
	}
	assertHeadUnchanged(t, fixture)
}
