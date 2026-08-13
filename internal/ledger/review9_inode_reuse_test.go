package ledger

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strings"
	"testing"
)

// TestPinnedRootLifetimePreventsInodeReuseCollision is the focused regression
// for the independent CW-39 failure at d21641da76495e463ef85f176a4c9dd5fc50ecc4.
//
// The failed implementation retained only an os.FileInfo snapshot. Once the
// original directory was deleted, its inode could be recycled for a different
// ordinary directory at the same pathname and os.SameFile could then accept
// the replacement. The repaired Reader keeps the original directory handle
// open for its lifetime, so that filesystem object remains live and its
// identity cannot be recycled while the Reader can still authorize work.
func TestPinnedRootLifetimePreventsInodeReuseCollision(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows commonly prevents removing a directory while its pinned handle is open")
	}

	r, _ := candidateTestReader(t)
	t.Cleanup(func() { _ = r.Close() })

	root := r.GitDir()
	originalInfo, err := os.Lstat(root)
	if err != nil {
		t.Fatal(err)
	}

	replacement := root + "-inode-reuse-replacement"
	copyRepositoryTree(t, root, replacement)

	if err := os.RemoveAll(root); err != nil {
		t.Skipf("platform/filesystem prevents removing pinned repository root: %v", err)
	}
	if _, err := os.Lstat(root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("original repository root still exists after removal: %v", err)
	}
	if err := os.Rename(replacement, root); err != nil {
		t.Skipf("platform/filesystem prevents installing stable replacement: %v", err)
	}

	replacementInfo, err := os.Lstat(root)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(originalInfo, replacementInfo) {
		t.Fatal("replacement reused the original repository filesystem identity while Reader kept the original root handle live")
	}

	_, err = r.Head(context.Background())
	if err == nil || !strings.Contains(err.Error(), "filesystem identity changed") {
		t.Fatalf("expected pinned-root replacement rejection before Git invocation, got %v", err)
	}
}
