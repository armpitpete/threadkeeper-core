package escrow

import (
	"os"
	"path/filepath"
	"testing"
)

// Root path substitution must not redirect trusted Get or Put operations
// to a replacement directory, even with content-addressed matching bytes.
func TestEscrowRejectsRootReplacement(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "escrow")
	external := filepath.Join(base, "external")
	store, err := OpenStore(root)
	if err != nil { t.Fatal(err) }
	payload := []byte("harmless matching content")
	snapshot, err := store.Put("source:1", "v1", "text/plain", payload)
	if err != nil { t.Fatal(err) }
	if err := os.Mkdir(external, 0o700); err != nil { t.Fatal(err) }
	if err := os.WriteFile(filepath.Join(external, snapshot.ContentSHA256+".blob"), payload, 0o600); err != nil { t.Fatal(err) }
	if err := os.Rename(root, root+"-moved"); err != nil { t.Fatal(err) }
	if err := os.Symlink(external, root); err != nil {
		t.Skipf("directory symlinks unavailable: %v", err)
	}
	if _, err := store.Get(snapshot); err == nil {
		t.Fatal("ESCROW_ROOT_REPLACED: Get accepted bytes from replacement directory")
	}
	if _, err := store.Put("source:2", "v2", "text/plain", payload); err == nil {
		t.Fatal("ESCROW_ROOT_REPLACED: Put accepted a blob in replacement directory")
	}
}

func TestEscrowRejectsDirectoryReplacement(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "escrow")
	replacement := filepath.Join(base, "replacement")
	store, err := OpenStore(root)
	if err != nil { t.Fatal(err) }
	content := []byte("matching bytes in different directory")
	snapshot, err := store.Put("source:1", "v1", "text/plain", content)
	if err != nil { t.Fatal(err) }
	if err := os.Mkdir(replacement, 0o700); err != nil { t.Fatal(err) }
	if err := os.WriteFile(filepath.Join(replacement, snapshot.ContentSHA256+".blob"), content, 0o600); err != nil { t.Fatal(err) }
	if err := os.Rename(root, root+"-moved"); err != nil { t.Fatal(err) }
	if err := os.Rename(replacement, root); err != nil { t.Fatal(err) }
	if _, err := store.Get(snapshot); err == nil {
		t.Fatal("ESCROW_ROOT_REPLACED: Get accepted a different root directory")
	}
	if _, err := store.Put("source:2", "v2", "text/plain", content); err == nil {
		t.Fatal("ESCROW_ROOT_REPLACED: Put accepted a different root directory")
	}
}
