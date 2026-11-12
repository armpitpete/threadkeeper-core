package sourceadapter

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/escrow"
)

// The accepted source directory must not silently change identity if its
// parent is writable. A symlink replacing the original root is outside it.
func TestFileAdapterRejectsRootReplacement(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "trusted")
	external := filepath.Join(base, "external")
	if err := os.Mkdir(root, 0o700); err != nil { t.Fatal(err) }
	if err := os.Mkdir(external, 0o700); err != nil { t.Fatal(err) }
	if err := os.WriteFile(filepath.Join(root, "evidence.txt"), []byte("trusted"), 0o600); err != nil { t.Fatal(err) }
	outsideBytes := []byte("external content matching the requested digest")
	if err := os.WriteFile(filepath.Join(external, "evidence.txt"), outsideBytes, 0o600); err != nil { t.Fatal(err) }

	adapter, err := NewFileAdapter(root)
	if err != nil { t.Fatal(err) }
	if err := os.Rename(root, root+"-moved"); err != nil { t.Fatal(err) }
	if err := os.Symlink(external, root); err != nil {
		t.Skipf("directory symlinks unavailable: %v", err)
	}

	_, err = adapter.Fetch(context.Background(), Request{
		SourceID: "sample",
		RelativePath: "evidence.txt",
		ExpectedSHA256: escrow.HashContent(outsideBytes),
	})
	if err == nil {
		t.Fatal("SOURCE_ROOT_REPLACED: accepted bytes from a replacement of the original trusted root")
	}
}

func TestFileAdapterRejectsDirectoryReplacement(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "trusted")
	replacement := filepath.Join(base, "replacement")
	if err := os.Mkdir(root, 0o700); err != nil { t.Fatal(err) }
	if err := os.Mkdir(replacement, 0o700); err != nil { t.Fatal(err) }
	if err := os.WriteFile(filepath.Join(root, "evidence.txt"), []byte("trusted"), 0o600); err != nil { t.Fatal(err) }
	outside := []byte("bytes from a different directory")
	if err := os.WriteFile(filepath.Join(replacement, "evidence.txt"), outside, 0o600); err != nil { t.Fatal(err) }
	adapter, err := NewFileAdapter(root)
	if err != nil { t.Fatal(err) }

	if err := os.Rename(root, root+"-moved"); err != nil { t.Fatal(err) }
	if err := os.Rename(replacement, root); err != nil { t.Fatal(err) }
	if _, err := adapter.Fetch(context.Background(), Request{
		SourceID: "sample", RelativePath: "evidence.txt", ExpectedSHA256: escrow.HashContent(outside),
	}); err == nil {
		t.Fatal("SOURCE_ROOT_REPLACED: accepted matching bytes from a different root directory")
	}
}
