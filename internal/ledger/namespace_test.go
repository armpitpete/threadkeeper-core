package ledger

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestReplayAcceptsRecognizedCoreV1Namespace(t *testing.T) {
	work := newWorkRepo(t)
	writeSchema(t, work)
	writeEvent(t, work, "events/decisions/001.json", "namespace-event-1", "decision.accepted", "target-a")
	commitAll(t, work, "add recognized schema and event")

	bare := cloneBare(t, work)
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if _, err := Replay(context.Background(), r); err != nil {
		t.Fatalf("recognized Core v1 ledger rejected: %v", err)
	}
}

func TestReplayRejectsUnknownCoreV1LedgerPaths(t *testing.T) {
	cases := []struct {
		name string
		path string
		raw  []byte
	}{
		{"root_text", "stray.txt", []byte("stray")},
		{"root_json", "stray.json", []byte(`{}`)},
		{"future_config_namespace", "config/sources/source.json", []byte(`{}`)},
		{"actor_policy_extra", "config/authority/actor-policy/extra.json", []byte(`{}`)},
		{"schema_non_json", "config/schemas/stray.txt", []byte("stray")},
		{"reducer_binding_non_json", "config/authority/reducer-bindings/stray.txt", []byte("stray")},
		{"event_non_json", "events/governance/stray.txt", []byte("stray")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			work := newWorkRepo(t)
			writeRawLedgerPath(t, work, tc.path, tc.raw)
			commitAll(t, work, "add unrecognized ledger path")
			requireNamespaceReplayFailure(t, cloneBare(t, work), tc.path)
		})
	}
}

func TestReplayRejectsMixedValidEventAndUnknownPath(t *testing.T) {
	work := newWorkRepo(t)
	writeSchema(t, work)
	writeEvent(t, work, "events/decisions/001.json", "namespace-event-mixed", "decision.accepted", "target-a")
	writeRawLedgerPath(t, work, "stray.txt", []byte("stray"))
	commitAll(t, work, "add valid event plus stray path")

	requireNamespaceReplayFailure(t, cloneBare(t, work), "stray.txt")
}

func TestReplayRejectsPollutedGenesisCommit(t *testing.T) {
	work := rawWorkRepo(t)
	writeTestGenesis(t, work, nil)
	writeRawLedgerPath(t, work, "stray.txt", []byte("stray"))
	commitAll(t, work, "polluted Genesis root")

	requireNamespaceReplayFailure(t, cloneBare(t, work), "stray.txt")
}

func TestReplayRejectsRenameFromKnownToUnknownNamespace(t *testing.T) {
	work := newWorkRepo(t)
	writeSchema(t, work)
	commitAll(t, work, "add accepted schema")

	from := filepath.Join(work, "config", "schemas", "event", "test-v1.json")
	to := filepath.Join(work, "config", "sources", "future.json")
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(from, to); err != nil {
		t.Fatal(err)
	}
	commitAll(t, work, "rename known schema into unknown namespace")

	requireNamespaceReplayFailure(t, cloneBare(t, work), "config/sources/future.json")
}

func TestReplayRejectsUnknownPathEvenIfLaterDeleted(t *testing.T) {
	work := newWorkRepo(t)
	writeRawLedgerPath(t, work, "stray.txt", []byte("stray"))
	commitAll(t, work, "add stray path")
	if err := os.Remove(filepath.Join(work, "stray.txt")); err != nil {
		t.Fatal(err)
	}
	commitAll(t, work, "delete stray path")

	requireNamespaceReplayFailure(t, cloneBare(t, work), "stray.txt")
}

func writeRawLedgerPath(t *testing.T, work, rel string, raw []byte) {
	t.Helper()
	path := filepath.Join(work, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

func requireNamespaceReplayFailure(t *testing.T, bare, path string) {
	t.Helper()
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	_, err = Replay(context.Background(), r)
	if err == nil ||
		!strings.Contains(err.Error(), "INTEGRITY_FAILURE") ||
		!strings.Contains(err.Error(), path) {
		t.Fatalf("unrecognized path %q did not fail closed: %v", path, err)
	}
}