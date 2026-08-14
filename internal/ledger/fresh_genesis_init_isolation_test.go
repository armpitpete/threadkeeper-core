package ledger

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestFreshGenesisRejectsInvalidRefBeforeTargetCreation(t *testing.T) {
	target := filepath.Join(t.TempDir(), "must-not-exist.git")
	_, err := InitializeFreshGenesis(context.Background(), target, "refs/heads/../redirect", freshGenesisFixture(t, nil), nil)
	if err == nil || !strings.Contains(err.Error(), "invalid authoritative ref") {
		t.Fatalf("invalid ref = %v", err)
	}
	if _, statErr := os.Lstat(target); !os.IsNotExist(statErr) {
		t.Fatalf("invalid ref created target: %v", statErr)
	}
}

func TestFreshGenesisInitImportsNoGitTemplateHooks(t *testing.T) {
	target := filepath.Join(t.TempDir(), "template-isolated.git")
	if _, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, nil), nil); err != nil {
		t.Fatal(err)
	}
	hooksPath := filepath.Join(target, "hooks")
	entries, err := os.ReadDir(hooksPath)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Fatalf("fresh ledger imported Git template hook material: %v", names)
	}
}
