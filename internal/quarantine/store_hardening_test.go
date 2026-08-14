package quarantine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenExistingFailsClosedWhenMissing(t *testing.T) {
	_, err := OpenExisting(filepath.Join(t.TempDir(), "missing"))
	if err == nil || !strings.Contains(err.Error(), "QUARANTINE_MISSING") {
		t.Fatalf("expected missing quarantine failure, got %v", err)
	}
}

func TestEnsureIsIdempotentOnlyForIdenticalBytes(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "q"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	first, err := s.Ensure("candidate-1", []byte("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Ensure("candidate-1", []byte("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("idempotent entry changed: %#v %#v", first, second)
	}
	if _, err := s.Ensure("candidate-1", []byte("beta")); err == nil || !strings.Contains(err.Error(), "QUARANTINE_CONFLICT") {
		t.Fatalf("expected conflicting bytes to fail closed, got %v", err)
	}
}

func TestReadRejectsSymlinkCandidate(t *testing.T) {
	parent := t.TempDir()
	qdir := filepath.Join(parent, "q")
	s, err := Open(qdir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	target := filepath.Join(qdir, "target")
	if err := os.WriteFile(target, []byte("payload"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(qdir, "candidate-1.candidate")
	if err := os.Symlink("target", link); err != nil {
		if errors.Is(err, os.ErrPermission) {
			t.Skipf("symlink creation unavailable: %v", err)
		}
		t.Fatal(err)
	}
	_, _, err = s.ReadID("candidate-1")
	if err == nil || !strings.Contains(err.Error(), "QUARANTINE_INTEGRITY_FAILURE") {
		t.Fatalf("expected symlink candidate rejection, got %v", err)
	}
}
