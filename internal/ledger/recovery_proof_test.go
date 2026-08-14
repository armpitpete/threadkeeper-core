package ledger

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestRecoveryProofSurvivesDestructiveBareRestore(t *testing.T) {
	r, head := candidateTestReader(t)
	event := makeCreateCandidateEvent(t, head, "recovery-event-1", "idem-recovery-1", json.RawMessage(`{"enabled":true}`))
	candidate, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/recovery-event-1.json",
		Event:        event,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AcceptWriteCandidate(context.Background(), r, *candidate); err != nil {
		t.Fatal(err)
	}

	original, err := ProveRecovery(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	if original.EventCount == 0 || original.GovernedRecordCount == 0 {
		t.Fatalf("recovery fixture is not substantive: %#v", original)
	}
	if original.ActorPolicyVersion == "" || original.ActorPolicyRootContentSHA256 == "" {
		t.Fatalf("recovery proof omits actor-policy identity: %#v", original)
	}

	backupRoot := t.TempDir()
	backup := filepath.Join(backupRoot, "secondary-ledger.git")
	runGit(t, "", "clone", "--bare", "--no-hardlinks", r.GitDir(), backup)

	// Destroy the original authority store after the backup exists. The proof
	// below must be reconstructable from the restored copy alone.
	if err := os.RemoveAll(r.GitDir()); err != nil {
		t.Fatal(err)
	}

	restoredDir := filepath.Join(t.TempDir(), "restored-ledger.git")
	runGit(t, "", "clone", "--bare", "--no-hardlinks", backup, restoredDir)
	restoredReader, err := gitledger.New(restoredDir, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := ProveRecovery(context.Background(), restoredReader)
	if err != nil {
		t.Fatal(err)
	}
	if err := CompareRecoveryProofs(*original, *restored); err != nil {
		t.Fatal(err)
	}
}

func TestRecoveryProofDetectsDifferentAuthoritativeState(t *testing.T) {
	r, _ := candidateTestReader(t)
	before, err := ProveRecovery(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}

	head, err := r.Head(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	event := makeCreateCandidateEvent(t, head, "recovery-event-change", "idem-recovery-change", json.RawMessage(`{"enabled":true}`))
	candidate, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: head, EventPath: "events/governance/recovery-event-change.json", Event: event})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AcceptWriteCandidate(context.Background(), r, *candidate); err != nil {
		t.Fatal(err)
	}
	after, err := ProveRecovery(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	if err := CompareRecoveryProofs(*before, *after); err == nil {
		t.Fatal("different authoritative state unexpectedly produced equivalent recovery proof")
	}
}
