package quarantine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestQuarantineLifecycle(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "q")); if err != nil { t.Fatal(err) }
	entry, err := s.Put("candidate-1", []byte("sensitive proposal")); if err != nil { t.Fatal(err) }
	got, err := s.Read(entry); if err != nil { t.Fatal(err) }
	if string(got) != "sensitive proposal" { t.Fatalf("unexpected content %q", got) }
	if err := s.Remove(entry.ID); err != nil { t.Fatal(err) }
	_, err = s.Read(entry); if !errors.Is(err, os.ErrNotExist) { t.Fatalf("expected removed candidate, got %v", err) }
}
