package ledger

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestNewRejectsSymlinkedGitDirectoryRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink fixture requires Unix-like symlink semantics")
	}
	r, _ := candidateTestReader(t)
	alias := filepath.Join(t.TempDir(), "link.git")
	if err := os.Symlink(r.GitDir(), alias); err != nil {
		t.Fatal(err)
	}

	_, err := gitledger.New(alias, gitledger.DefaultRef)
	if err == nil || !strings.Contains(err.Error(), "symlinked Git repository root or ancestor") {
		t.Fatalf("expected symlinked Git-directory root rejection, got %v", err)
	}
}

func TestNewRejectsSymlinkedGitDirectoryAncestor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink fixture requires Unix-like symlink semantics")
	}
	r, _ := candidateTestReader(t)
	realParent := filepath.Dir(r.GitDir())
	aliasParent := filepath.Join(t.TempDir(), "alias-parent")
	if err := os.Symlink(realParent, aliasParent); err != nil {
		t.Fatal(err)
	}
	aliasRoot := filepath.Join(aliasParent, filepath.Base(r.GitDir()))

	_, err := gitledger.New(aliasRoot, gitledger.DefaultRef)
	if err == nil || !strings.Contains(err.Error(), "symlinked Git repository root or ancestor") {
		t.Fatalf("expected symlinked Git-directory ancestor rejection, got %v", err)
	}
}

func TestExistingReaderRejectsGitRootReplacedBySymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink fixture requires Unix-like symlink semantics")
	}
	r, _ := candidateTestReader(t)
	original := r.GitDir()
	moved := original + ".moved"
	if err := os.Rename(original, moved); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(moved, original); err != nil {
		t.Fatal(err)
	}

	_, err := r.Head(context.Background())
	if err == nil || !strings.Contains(err.Error(), "symlinked Git repository root or ancestor") {
		t.Fatalf("expected per-invocation root replacement rejection, got %v", err)
	}
}
