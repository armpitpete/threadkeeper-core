package ledger

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestReplayIgnoresAmbientGitNamespaceAndConfigInjection(t *testing.T) {
	work := newWorkRepo(t)
	writeSchema(t, work)
	writeEvent(t, work, "events/decisions/001.json", "event-1", "decision.accepted", "target-a")
	commitAll(t, work, "accept event")
	bare := cloneBare(t, work)

	t.Setenv("GIT_NAMESPACE", "hostile")
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "core.repositoryformatversion")
	t.Setenv("GIT_CONFIG_VALUE_0", "999")

	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := Replay(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.EventCount != 1 {
		t.Fatalf("event count = %d, want 1", manifest.EventCount)
	}
}

func TestReplayRejectsShallowAuthoritativeHistory(t *testing.T) {
	work := newWorkRepo(t)
	writeSchema(t, work)
	writeEvent(t, work, "events/decisions/001.json", "event-1", "decision.accepted", "target-a")
	commitAll(t, work, "accept event")
	bare := cloneBare(t, work)
	head := runGit(t, bare, "rev-parse", "refs/heads/main")
	if err := os.WriteFile(filepath.Join(bare, "shallow"), []byte(head+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Replay(context.Background(), r)
	if err == nil || !strings.Contains(err.Error(), "shallow") {
		t.Fatalf("expected shallow-history rejection, got %v", err)
	}
}

func TestReplayRejectsGitGrafts(t *testing.T) {
	work := newWorkRepo(t)
	writeSchema(t, work)
	writeEvent(t, work, "events/decisions/001.json", "event-1", "decision.accepted", "target-a")
	commitAll(t, work, "accept event")
	bare := cloneBare(t, work)
	if err := os.MkdirAll(filepath.Join(bare, "info"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bare, "info", "grafts"), []byte("tamper\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Replay(context.Background(), r)
	if err == nil || !strings.Contains(err.Error(), "grafts") {
		t.Fatalf("expected graft rejection, got %v", err)
	}
}

func TestReplayRejectsLocalGitConfigIncludes(t *testing.T) {
	work := newWorkRepo(t)
	writeSchema(t, work)
	writeEvent(t, work, "events/decisions/001.json", "event-1", "decision.accepted", "target-a")
	commitAll(t, work, "accept event")
	bare := cloneBare(t, work)
	configPath := filepath.Join(bare, "config")
	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("\n[include]\n\tpath = /tmp/threadkeeper-hostile-config\n"); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Replay(context.Background(), r)
	if err == nil || !strings.Contains(err.Error(), "config includes") {
		t.Fatalf("expected config-include rejection, got %v", err)
	}
}

func TestReplayRejectsDuplicateLogicalEventID(t *testing.T) {
	work := newWorkRepo(t)
	writeSchema(t, work)
	writeEvent(t, work, "events/decisions/001.json", "event-1", "decision.accepted", "target-a")
	commitAll(t, work, "accept event 1")
	writeEvent(t, work, "events/governance/002.json", "event-1", "governance.recorded", "target-b")
	commitAll(t, work, "duplicate logical event id")
	bare := cloneBare(t, work)

	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Replay(context.Background(), r)
	if err == nil || !strings.Contains(err.Error(), "duplicate logical event_id") {
		t.Fatalf("expected duplicate event-id rejection, got %v", err)
	}
}

func TestReplayRejectsNonJSONDurableEventFile(t *testing.T) {
	work := newWorkRepo(t)
	writeSchema(t, work)
	path := filepath.Join(work, "events/governance/note.txt")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not a durable JSON event"), 0o644); err != nil {
		t.Fatal(err)
	}
	commitAll(t, work, "add invalid event file")
	bare := cloneBare(t, work)

	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Replay(context.Background(), r)
	if err == nil || !strings.Contains(err.Error(), "not JSON") {
		t.Fatalf("expected non-JSON event rejection, got %v", err)
	}
}
