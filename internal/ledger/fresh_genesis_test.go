package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
	"github.com/armpitpete/threadkeeper-core/internal/contracts"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/genesis"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/reducer"
	contractschemas "github.com/armpitpete/threadkeeper-core/schemas"
)

func TestInitializeFreshGenesisCreatesReplayableRoot(t *testing.T) {
	target := filepath.Join(t.TempDir(), "production-ledger.git")
	raw := freshGenesisFixture(t, nil)
	evidence, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, raw, freshGenesisSeed(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	if evidence.ProjectID != "project:fresh-test" || evidence.LedgerID != "ledger:fresh-test" || evidence.ActorPolicyContentSHA256 == "" {
		t.Fatalf("unexpected evidence identity: %#v", evidence)
	}
	if evidence.GenesisCommit == "" || evidence.GenesisCommit != evidence.LedgerCommit {
		t.Fatalf("fresh ledger is not exactly one Genesis root: %#v", evidence)
	}
	if evidence.StoragePath == "" || evidence.GitObjectFormat != "sha1" || evidence.InitialSchemaCount != 2 || evidence.InitialBindingCount != 1 {
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
	policy, err := LoadCurrentActorPolicy(context.Background(), restarted)
	if err != nil || policy.PolicyContentSHA != evidence.ActorPolicyContentSHA256 || policy.SourceEventID != "" {
		t.Fatalf("restart changed actor policy identity: snapshot=%#v err=%v", policy, err)
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
	_, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, nil), freshGenesisSeed(t, nil))
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
	_, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, []byte(`{"project_id":"incomplete"}`), freshGenesisSeed(t, nil))
	if err == nil || !strings.Contains(err.Error(), "FRESH_GENESIS_INVALID") {
		t.Fatalf("invalid Genesis = %v", err)
	}
	if _, statErr := os.Lstat(target); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("invalid Genesis created target: %v", statErr)
	}
}

func TestInitializeFreshGenesisRejectsMissingActorPolicyBeforeCreatingTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "must-not-exist")
	_, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, nil), nil)
	if err == nil || !strings.Contains(err.Error(), "must supply authoritative actor policy") {
		t.Fatalf("missing actor policy = %v", err)
	}
	if _, statErr := os.Lstat(target); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("missing actor policy created target: %v", statErr)
	}
}

func TestInitializeFreshGenesisRejectsUnsafeSeedBeforeCreatingTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "must-not-exist")
	_, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, nil), freshGenesisSeed(t, map[string][]byte{
		"events/not-authorised.json": []byte(`{}`),
	}))
	if err == nil || !strings.Contains(err.Error(), "outside initial schema/reducer-binding/actor-policy namespaces") {
		t.Fatalf("unsafe seed = %v", err)
	}
	if _, statErr := os.Lstat(target); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("unsafe seed created target: %v", statErr)
	}
}

func TestFreshGenesisInitialSchemaSetMustMatchRoot(t *testing.T) {
	target := filepath.Join(t.TempDir(), "schema-mismatch.git")
	_, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, nil), freshGenesisSeed(t, map[string][]byte{
		"config/schemas/event/test-v1.json": testSchema,
	}))
	if err == nil || !strings.Contains(err.Error(), "GENESIS_SCHEMA_MISMATCH") {
		t.Fatalf("schema mismatch = %v", err)
	}

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
	evidence, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, []string{testSchemaID}), freshGenesisSeed(t, map[string][]byte{
		"config/schemas/event/test-v1.json": testSchema,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if evidence.InitialSchemaCount != 3 || evidence.InitialBindingCount != 1 {
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
	if _, err := Replay(context.Background(), r); err == nil || !strings.Contains(err.Error(), "100644 regular blob") {
		t.Fatalf("non-regular Genesis was accepted: %v", err)
	}
}

func freshGenesisFixture(t *testing.T, additionalSchemas []string) []byte {
	t.Helper()
	set := map[string]struct{}{
		contracts.ExclusiveRecordEventSchemaV1: {},
		contracts.ReducerBindingSchemaV1:      {},
	}
	for _, id := range additionalSchemas {
		set[id] = struct{}{}
	}
	initialSchemas := make([]string, 0, len(set))
	for id := range set {
		initialSchemas = append(initialSchemas, id)
	}
	sort.Strings(initialSchemas)
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

func freshGenesisSeed(t *testing.T, extra map[string][]byte) map[string][]byte {
	t.Helper()
	bindingCandidate, err := json.Marshal(map[string]any{
		"schema_version":           contracts.ReducerBindingSchemaV1,
		"binding_id":               "binding:actor-auth-policy:v1",
		"record_kind":              actorauth.PolicyRecordKind,
		"state_model":              reducer.ModelExclusiveV1,
		"event_schema":             contracts.ExclusiveRecordEventSchemaV1,
		"authority_policy_version": testPolicyV1,
	})
	if err != nil {
		t.Fatal(err)
	}
	binding, _, err := digest.Complete(bindingCandidate)
	if err != nil {
		t.Fatal(err)
	}
	seed := map[string][]byte{
		actorauth.LedgerPolicyPath:                                  writeTestActorPolicy(t, t.TempDir()),
		"config/schemas/exclusive-record-event-v1.json":           contractschemas.ExclusiveGovernedRecordEventV1,
		"config/schemas/reducer-binding-v1.json":                  contractschemas.ReducerBindingV1,
		"config/authority/reducer-bindings/actor-auth-policy-v1.json": binding,
	}
	for path, raw := range extra {
		seed[path] = raw
	}
	return seed
}

func rawWorkRepo(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "raw-work")
	runGit(t, "", "init", "-b", "main", dir)
	runGit(t, dir, "config", "user.name", "Threadkeeper Test")
	runGit(t, dir, "config", "user.email", "threadkeeper-test@example.invalid")
	return dir
}
