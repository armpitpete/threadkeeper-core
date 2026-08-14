package ledger

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/contracts"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/reducer"
	contractschemas "github.com/armpitpete/threadkeeper-core/schemas"
)

func TestFreshGenesisRejectsSymlinkedTargetParentBeforeCreation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink parent fixture is POSIX-only")
	}
	root := t.TempDir()
	realParent := filepath.Join(root, "real-parent")
	if err := os.Mkdir(realParent, 0o700); err != nil {
		t.Fatal(err)
	}
	aliasParent := filepath.Join(root, "alias-parent")
	if err := os.Symlink(realParent, aliasParent); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(aliasParent, "ledger.git")
	_, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, nil), nil)
	if err == nil || !strings.Contains(err.Error(), "symlinked Git repository root or ancestor") {
		t.Fatalf("symlinked bootstrap parent = %v", err)
	}
	if _, statErr := os.Lstat(filepath.Join(realParent, "ledger.git")); !os.IsNotExist(statErr) {
		t.Fatalf("unsafe parent was followed before rejection: %v", statErr)
	}
}

func TestFreshGenesisRejectsRootBindingPolicyDifferentFromGenesis(t *testing.T) {
	bindingCandidate, err := json.Marshal(map[string]any{
		"schema_version":           contracts.ReducerBindingSchemaV1,
		"binding_id":               "binding:fresh-policy-mismatch:v1",
		"record_kind":              testRecordKind,
		"state_model":              reducer.ModelExclusiveV1,
		"event_schema":             contracts.ExclusiveRecordEventSchemaV1,
		"authority_policy_version": "authority-policy:v2",
	})
	if err != nil {
		t.Fatal(err)
	}
	binding, _, err := digest.Complete(bindingCandidate)
	if err != nil {
		t.Fatal(err)
	}
	initialSchemas := []string{contracts.ExclusiveRecordEventSchemaV1, contracts.ReducerBindingSchemaV1}
	target := filepath.Join(t.TempDir(), "policy-mismatch.git")
	_, err = InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, initialSchemas), map[string][]byte{
		"config/schemas/exclusive-record-event-v1.json": contractschemas.ExclusiveGovernedRecordEventV1,
		"config/schemas/reducer-binding-v1.json":        contractschemas.ReducerBindingV1,
		"config/authority/reducer-bindings/fresh-policy-mismatch-v1.json": binding,
	})
	if err == nil || !strings.Contains(err.Error(), "GENESIS_POLICY_MISMATCH") {
		t.Fatalf("root policy mismatch = %v", err)
	}
}

func TestFreshGenesisSupportsExplicitDirectAuthoritativeRef(t *testing.T) {
	const ref = "refs/heads/authority"
	target := filepath.Join(t.TempDir(), "custom-ref.git")
	evidence, err := InitializeFreshGenesis(context.Background(), target, ref, freshGenesisFixture(t, nil), nil)
	if err != nil {
		t.Fatal(err)
	}
	if evidence.AuthoritativeRef != ref {
		t.Fatalf("authoritative ref = %q want %q", evidence.AuthoritativeRef, ref)
	}
	r, err := gitledger.New(target, ref)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	manifest, err := Replay(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.GenesisCommit != evidence.GenesisCommit || manifest.LedgerCommit != evidence.GenesisCommit {
		t.Fatalf("custom ref changed root identity: manifest=%#v evidence=%#v", manifest, evidence)
	}
}
