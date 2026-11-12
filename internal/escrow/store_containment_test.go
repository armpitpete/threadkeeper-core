package escrow

import (
	"os"
	"path/filepath"
	"testing"
)

// An escrow blob must never resolve to bytes outside the store, even when
// external bytes happen to match the digest of the requested snapshot.
func TestStoreRejectsExternalBlobSymlink(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "escrow")
	store, err := OpenStore(storePath)
	if err != nil {
		t.Fatal(err)
	}
	content := []byte("harmless test evidence")
	snapshot, err := store.Put("sample", "v1", "text/plain", content)
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, content, 0o600); err != nil {
		t.Fatal(err)
	}
	blob := filepath.Join(storePath, snapshot.ContentSHA256+".blob")
	if err := os.Remove(blob); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, blob); err != nil {
		t.Skipf("symlink creation unavailable on this host: %v", err)
	}
	if _, err := store.Get(snapshot); err == nil {
		t.Fatal("escrow followed blob symlink outside storage root")
	}
	if _, err := store.Put("sample2", "v2", "text/plain", content); err == nil {
		t.Fatal("escrow accepted existing blob symlink outside storage root")
	}
}
