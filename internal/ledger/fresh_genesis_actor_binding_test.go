package ledger

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestFreshGenesisRejectsMissingActorPolicyBindingBeforeCreation(t *testing.T) {
	seed := freshGenesisSeed(t, nil)
	delete(seed, "config/authority/reducer-bindings/actor-auth-policy-v1.json")
	target := filepath.Join(t.TempDir(), "must-not-exist.git")
	_, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, nil), seed)
	if err == nil || !strings.Contains(err.Error(), "no reducer binding for actor policy record kind") {
		t.Fatalf("missing actor-policy binding = %v", err)
	}
	if _, statErr := os.Lstat(target); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("missing actor-policy binding created target: %v", statErr)
	}
}
