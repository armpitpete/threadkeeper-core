package ledger

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestFreshGenesisRejectsActorPolicyForWrongLedgerBeforeCreation(t *testing.T) {
	target := filepath.Join(t.TempDir(), "must-not-exist.git")
	seed := freshGenesisSeed(t, nil)
	seed[actorauth.LedgerPolicyPath] = makeTestActorPolicy(t, "ledger:other", testPolicyV1)
	_, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, nil), seed)
	if err == nil || !strings.Contains(err.Error(), "FRESH_GENESIS_INVALID") || !strings.Contains(err.Error(), "ledger_id") {
		t.Fatalf("wrong-ledger root actor policy = %v", err)
	}
	if _, statErr := os.Lstat(target); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("wrong-ledger root actor policy created target: %v", statErr)
	}
}

func TestFreshGenesisRejectsActorPolicyForWrongPolicyVersionBeforeCreation(t *testing.T) {
	target := filepath.Join(t.TempDir(), "must-not-exist.git")
	seed := freshGenesisSeed(t, nil)
	seed[actorauth.LedgerPolicyPath] = makeTestActorPolicy(t, "ledger:fresh-test", "authority-policy:other:v1")
	_, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, nil), seed)
	if err == nil || !strings.Contains(err.Error(), "FRESH_GENESIS_INVALID") || !strings.Contains(err.Error(), "authority_policy_version") {
		t.Fatalf("wrong-version root actor policy = %v", err)
	}
	if _, statErr := os.Lstat(target); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("wrong-version root actor policy created target: %v", statErr)
	}
}
