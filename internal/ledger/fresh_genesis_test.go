package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/genesis"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestInitializeFreshGenesisCreatesReplayableRoot(t *testing.T) {
	target := filepath.Join(t.TempDir(), "production-ledger.git")
	raw := freshGenesisFixture(t, nil)
	evidence, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if evidence.ProjectID != "project:fresh-test" || evidence.LedgerID != "ledger:fresh-test" {
		t.Fatalf("unexpected evidence identity: %#v", evidence)
	}
	if evidence.GenesisCommit == "" || evidence.GenesisCommit != evidence.LedgerCommit {
		t.Fatalf("fresh ledger is not exactly one Genesis root: %#v", evidence)
	}
	if evidence.StoragePath == "" || evidence.GitObjectFormat != "sha1" {
		t.Fatalf("incomplete bootstrap evidence: %#v", evidence)
	}

	restarted, err := gitledger.New(target, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	manifest, err := Replay(context.Background(), restarted)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.GenesisCommit != evidence.GenesisCommit || manifest.GenesisRoot.LedgerID != evidence.LedgerID || manifest.HistoryCommitCount != 1 {
		t.Fatalf("restart changed Genesis identity: %#v", manifest)
	}
}

func TestInitializeFreshGenesisRefusesExistingTargetWithoutOverwrite(t *testing.T) {
	target := filepath.Join(t.TempDir(), "existing")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(target, "sentinel")
	if err := os.WriteFile(sentinel, []byte("preserve me"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, nil), nil)
	if err == nil || !errors.Is(err, gitledger.ErrLedgerAlreadyExists) {
		t.Fatalf("existing target = %v, want LEDGER_ALREADY_EXISTS", err)
	}
	got, readErr := os.ReadFile(sentinel)
	if readErr != nil || string(got) != "preserve me" {
		t.Fatalf("existing target was modified: bytes=%q err=%v", got, readErr)
	}
}

func TestInitializeFreshGenesisRejectsInvalidInputBeforeCreatingTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "must-not-exist")
	_, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, []byte(`{"project_id":"incomplete"}`), nil)
	if err == nil || !strings.Contains(err.Error(), "FRESH_GENESIS_INVALID") {
		t.Fatalf("invalid Genesis = %v", err)
	}
	if _, statErr := os.Lstat(target); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("invalid Genesis created target: %v", statErr)
	}
}

func TestInitializeFreshGenesisRejectsUnsafeSeedBeforeCreatingTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "must-not-exist")
	_, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, nil), map[string][]byte{
		"events/not-authorised.json": []byte(`{}`),
	})
	if err == nil || !strings.Contains(err.Error(), "outside initial schema/reducer-binding namespaces") {
		t.Fatalf("unsafe seed = %v", err)
	}
	if _, statErr := os.Lstat(target); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("unsafe seed created target: %v", statErr)
	}
}

func TestFreshGenesisInitialSchemaSetMustMatchRoot(t *testing.T) {
	target := filepath.Join(t.TempDir(), "schema-mismatch.git")
	_, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, nil), map[string][]byte{
		"config/schemas/event/test-v1.json": testSchema,
	})
	if err == nil || !strings.Contains(err.Error(), "GENESIS_SCHEMA_MISMATCH") {
		t.Fatalf("schema mismatch = %v", err)
	}

	// The create-only attempt may leave residue, but it must remain unusable
	// through the normal authoritative replay path.
	r, openErr := gitledger.New(target, gitledger.DefaultRef)
	if openErr != nil {
		t.Fatal(openErr)
	}
	defer r.Close()
	if _, replayErr := Replay(context.Background(), r); replayErr == nil || !strings.Contains(replayErr.Error(), "GENESIS_SCHEMA_MISMATCH") {
		t.Fatalf("partial bootstrap residue became valid: %v", replayErr)
	}
}

func TestFreshGenesisAcceptsDeclaredInitialSchemaRoot(t *testing.T) {
	target := filepath.Join(t.TempDir(), "schema-root.git")
	evidence, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, []string{testSchemaID}), map[string][]byte{
		"config/schemas/event/test-v1.json": testSchema,
	})
	if err != nil {
		t.Fatal(err)
	}
	if evidence.InitialSchemaCount != 1 || evidence.InitialBindingCount != 0 {
		t.Fatalf("unexpected initial config evidence: %#v", evidence)
	}
}

func TestReplayRejectsGenesisAddedAfterRoot(t *testing.T) {
	work := rawWorkRepo(t)
	runGit(t, work, "commit", "--allow-empty", "-m", "pre-Genesis root")
	writeTestGenesis(t, work, nil)
	commitAll(t, work, "late Genesis")
	bare := cloneBare(t, work)
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if _, err := Replay(context.Background(), r); err == nil || !strings.Contains(err.Error(), "root commit") {
		t.Fatalf("late Genesis was accepted: %v", err)
	}
}

func TestReplayRejectsGenesisMutationAfterRoot(t *testing.T) {
	work := newWorkRepo(t)
	path := filepath.Join(work, filepath.FromSlash(genesis.LedgerPath))
	mutated := freshGenesisFixture(t, nil)
	if err := os.WriteFile(path, mutated, 0o644); err != nil {
		t.Fatal(err)
	}
	commitAll(t, work, "mutate Genesis")
	bare := cloneBare(t, work)
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if _, err := Replay(context.Background(), r); err == nil || !strings.Contains(err.Error(), "GENESIS_IMMUTABLE") {
		t.Fatalf("mutated Genesis was accepted: %v", err)
	}
}

func TestReplayRejectsNonRegularRootGenesis(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink fixture is POSIX-only")
	}
	work := rawWorkRepo(t)
	payload := filepath.Join(work, "genesis-payload.json")
	if err := os.WriteFile(payload, freshGenesisFixture(t, nil), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(work, filepath.FromSlash(genesis.LedgerPath))
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../../genesis-payload.json", link); err != nil {
		t.Fatal(err)
	}
	commitAll(t, work, "symlink Genesis")
	bare := cloneBare(t, work)
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if _, err := Replay(context.Background(), r); err == nil || !strings.Contains(err.Error(), "expected regular 100644 blob") {
		t.Fatalf("non-regular Genesis was accepted: %v", err)
	}
}

func freshGenesisFixture(t *testing.T, initialSchemas []string) []byte {
	t.Helper()
	if initialSchemas == nil {
		initialSchemas = []string{}
	}
	raw, err := json.Marshal(map[string]any{
		"project_id":               "project:fresh-test",
		"ledger_id":                "ledger:fresh-test",
		"created_at":               "2026-08-14T11:00:00Z",
		"initial_authority_policy": testPolicyV1,
		"initial_schemas":          initialSchemas,
		"initial_authorities":      []string{"owner:test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	completed, _, err := digest.Complete(raw)
	if err != nil {
		t.Fatal(err)
	}
	return completed
}

func rawWorkRepo(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "raw-work")
	runGit(t, "", "init", "-b", "main", dir)
	runGit(t, dir, "config", "user.name", "Threadkeeper Test")
	runGit(t, dir, "config", "user.email", "threadkeeper-test@example.invalid")
	return dir
}
