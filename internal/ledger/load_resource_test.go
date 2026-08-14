package ledger

import (
	"context"
	"runtime"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/loadproof"
)

func TestReferenceReadLoadEnvelopePreservesRecoveryProof(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		t.Skip("reference envelope requires a process open-handle metric")
	}
	r, _ := candidateTestReader(t)
	defer r.Close()

	evidence, err := ProveReadLoad(context.Background(), r, loadproof.ReferenceEnvelope())
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Recovery.LedgerCommit == "" || evidence.Recovery.ReplaySHA256 == "" || evidence.Recovery.GovernedRecordsSHA256 == "" {
		t.Fatalf("incomplete recovery proof: %#v", evidence.Recovery)
	}
	if !evidence.Resources.Passed {
		t.Fatalf("resource evidence did not pass: %#v", evidence.Resources)
	}
}
