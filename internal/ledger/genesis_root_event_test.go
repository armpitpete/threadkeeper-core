package ledger

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestReplayRejectsDurableEventInGenesisRoot(t *testing.T) {
	work := rawWorkRepo(t)
	writeTestGenesis(t, work, nil)
	eventPath := filepath.Join(work, "events", "forbidden.json")
	if err := os.MkdirAll(filepath.Dir(eventPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(eventPath, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	commitAll(t, work, "Genesis plus forbidden event")
	bare := cloneBare(t, work)
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if _, err := Replay(context.Background(), r); err == nil || !strings.Contains(err.Error(), "GENESIS_ROOT_INVALID") {
		t.Fatalf("Genesis root event was accepted: %v", err)
	}
}
