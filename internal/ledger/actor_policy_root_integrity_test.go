package ledger

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestReplayRejectsRemovedRootActorPolicy(t *testing.T) {
	work := newWorkRepo(t)
	path := filepath.Join(work, filepath.FromSlash(actorauth.LedgerPolicyPath))
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	commitAll(t, work, "remove root actor policy")
	bare := cloneBare(t, work)
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if _, err := Replay(context.Background(), r); err == nil || !strings.Contains(err.Error(), "AUTH_POLICY_IMMUTABLE") {
		t.Fatalf("removed root actor policy replay = %v", err)
	}
}

func TestReplayRejectsNonRegularRootActorPolicy(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink fixture is POSIX-only")
	}
	work := rawWorkRepo(t)
	writeTestGenesis(t, work, nil)
	policyPath := filepath.Join(work, filepath.FromSlash(actorauth.LedgerPolicyPath))
	policyBytes, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(policyPath); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(work, "actor-policy-payload.json")
	if err := os.WriteFile(payload, policyBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../../../actor-policy-payload.json", policyPath); err != nil {
		t.Fatal(err)
	}
	commitAll(t, work, "symlink root actor policy")
	bare := cloneBare(t, work)
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if _, err := Replay(context.Background(), r); err == nil || !strings.Contains(err.Error(), "100644 regular blob") {
		t.Fatalf("non-regular root actor policy replay = %v", err)
	}
}
