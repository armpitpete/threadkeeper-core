package quarantine

import (
	"os"
	"testing"
	"time"
)

func TestEnsureWaitsForIdenticalInProgressCreator(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	id := "concurrent-identical"
	content := []byte("complete identical candidate bytes")
	prefix := content[:7]

	f, err := s.root.OpenFile(candidateName(id), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(prefix); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		t.Fatal(err)
	}

	writerDone := make(chan error, 1)
	go func() {
		// Leave the short prefix visible long enough for Ensure's first Put to
		// collide and its first ReadID to observe an in-progress file.
		time.Sleep(25 * time.Millisecond)
		_, writeErr := f.Write(content[len(prefix):])
		if writeErr == nil {
			writeErr = f.Sync()
		}
		closeErr := f.Close()
		if writeErr == nil {
			writeErr = closeErr
		}
		writerDone <- writeErr
	}()

	entry, err := s.Ensure(id, content)
	if err != nil {
		t.Fatalf("identical concurrent Ensure misclassified in-progress content: %v", err)
	}
	if err := <-writerDone; err != nil {
		t.Fatal(err)
	}
	want := entryFor(id, content)
	if entry != want {
		t.Fatalf("settled entry = %#v want %#v", entry, want)
	}
	gotEntry, got, err := s.ReadID(id)
	if err != nil {
		t.Fatal(err)
	}
	if gotEntry != want || string(got) != string(content) {
		t.Fatalf("settled quarantine content entry=%#v bytes=%q", gotEntry, got)
	}
}
