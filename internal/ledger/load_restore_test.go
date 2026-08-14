package ledger

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/loadproof"
)

func TestRestoredCopiesPreserveRecoveryProofUnderLoad(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		t.Skip("reference restore-load envelope requires a process open-handle metric")
	}
	original, _ := candidateTestReader(t)
	defer original.Close()
	baseline, err := ProveRecovery(context.Background(), original)
	if err != nil {
		t.Fatal(err)
	}

	backup := filepath.Join(t.TempDir(), "backup.git")
	runGit(t, "", "clone", "--bare", "--no-hardlinks", original.GitDir(), backup)

	const copies = 4
	readers := make([]*gitledger.Reader, 0, copies)
	for i := 0; i < copies; i++ {
		restored := filepath.Join(t.TempDir(), "restored.git")
		runGit(t, "", "clone", "--bare", "--no-hardlinks", backup, restored)
		r, err := gitledger.New(restored, gitledger.DefaultRef)
		if err != nil {
			t.Fatal(err)
		}
		readers = append(readers, r)
	}
	defer func() {
		for _, r := range readers {
			r.Close()
		}
	}()

	envelope := loadproof.ReferenceEnvelope()
	envelope.Name = "threadkeeper-core-ci-restored-replay-v1"
	envelope.ConcurrentWorkers = copies
	envelope.IterationsPerWorker = 2
	evidence, err := loadproof.Run(context.Background(), envelope, func(ctx context.Context, worker, _ int) error {
		proof, err := ProveRecovery(ctx, readers[worker%len(readers)])
		if err != nil {
			return err
		}
		return CompareRecoveryProofs(*baseline, *proof)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !evidence.Passed {
		t.Fatalf("restored-copy resource evidence did not pass: %#v", evidence)
	}
}
