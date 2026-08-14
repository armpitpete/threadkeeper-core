package quarantine

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Store struct {
	dir  string
	root *os.Root
}

type Entry struct {
	ID            string `json:"id"`
	ContentSHA256 string `json:"content_sha256"`
	Size          int64  `json:"size"`
}

func Open(dir string) (*Store, error) {
	if dir == "" {
		return nil, fmt.Errorf("QUARANTINE_INVALID: directory is required")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if info, err := os.Lstat(abs); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return nil, fmt.Errorf("QUARANTINE_INVALID: directory must be a real directory")
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o700); err != nil {
		return nil, err
	}
	if err := os.Chmod(abs, 0o700); err != nil {
		return nil, err
	}
	return openPinned(abs)
}

// OpenExisting pins an already-created quarantine directory without creating
// filesystem state. Acceptance uses this fail-closed read path.
func OpenExisting(dir string) (*Store, error) {
	if dir == "" {
		return nil, fmt.Errorf("QUARANTINE_INVALID: directory is required")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if _, err := os.Lstat(abs); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("QUARANTINE_MISSING: %w", err)
		}
		return nil, err
	}
	return openPinned(abs)
}

func openPinned(abs string) (*Store, error) {
	before, err := os.Lstat(abs)
	if err != nil {
		return nil, err
	}
	if before.Mode()&os.ModeSymlink != 0 || !before.IsDir() {
		return nil, fmt.Errorf("QUARANTINE_INVALID: directory must be a real directory")
	}
	root, err := os.OpenRoot(abs)
	if err != nil {
		return nil, err
	}
	opened, err := root.Stat(".")
	if err != nil {
		root.Close()
		return nil, err
	}
	after, err := os.Lstat(abs)
	if err != nil {
		root.Close()
		return nil, err
	}
	if after.Mode()&os.ModeSymlink != 0 || !after.IsDir() || !os.SameFile(before, opened) || !os.SameFile(opened, after) {
		root.Close()
		return nil, fmt.Errorf("QUARANTINE_INTEGRITY_FAILURE: quarantine root changed while opening")
	}
	return &Store{dir: abs, root: root}, nil
}

func (s *Store) Close() error {
	if s == nil || s.root == nil {
		return nil
	}
	return s.root.Close()
}

func (s *Store) Put(id string, content []byte) (Entry, error) {
	if !validID(id) {
		return Entry{}, fmt.Errorf("QUARANTINE_INVALID: unsafe candidate id %q", id)
	}
	name := candidateName(id)
	f, err := s.root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return Entry{}, err
	}
	cleanup := true
	defer func() {
		_ = f.Close()
		if cleanup {
			_ = s.root.Remove(name)
		}
	}()
	if _, err := f.Write(content); err != nil {
		return Entry{}, err
	}
	if err := f.Sync(); err != nil {
		return Entry{}, err
	}
	if err := f.Close(); err != nil {
		return Entry{}, err
	}
	cleanup = false
	return entryFor(id, content), nil
}

// Ensure is idempotent for identical candidate bytes. Reusing a candidate
// identity for different bytes fails closed.
func (s *Store) Ensure(id string, content []byte) (Entry, error) {
	entry, err := s.Put(id, content)
	if err == nil {
		return entry, nil
	}
	if !errors.Is(err, fs.ErrExist) {
		return Entry{}, err
	}
	existing, got, readErr := s.ReadID(id)
	if readErr != nil {
		return Entry{}, readErr
	}
	if !bytes.Equal(got, content) {
		return Entry{}, fmt.Errorf("QUARANTINE_CONFLICT: candidate id %q already contains different bytes", id)
	}
	return existing, nil
}

func (s *Store) Read(entry Entry) ([]byte, error) {
	actual, content, err := s.ReadID(entry.ID)
	if err != nil {
		return nil, err
	}
	if actual.Size != entry.Size || actual.ContentSHA256 != entry.ContentSHA256 {
		return nil, fmt.Errorf("QUARANTINE_INTEGRITY_FAILURE: candidate bytes changed")
	}
	return content, nil
}

func (s *Store) ReadID(id string) (Entry, []byte, error) {
	if !validID(id) {
		return Entry{}, nil, fmt.Errorf("QUARANTINE_INVALID: unsafe candidate id")
	}
	name := candidateName(id)
	info, err := s.root.Lstat(name)
	if err != nil {
		return Entry{}, nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return Entry{}, nil, fmt.Errorf("QUARANTINE_INTEGRITY_FAILURE: candidate is not a regular file")
	}
	content, err := s.root.ReadFile(name)
	if err != nil {
		return Entry{}, nil, err
	}
	return entryFor(id, content), content, nil
}

func (s *Store) Remove(id string) error {
	if !validID(id) {
		return fmt.Errorf("QUARANTINE_INVALID: unsafe candidate id")
	}
	name := candidateName(id)
	info, err := s.root.Lstat(name)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("QUARANTINE_INTEGRITY_FAILURE: candidate is not a regular file")
	}
	return s.root.Remove(name)
}

// PruneBefore removes only well-formed regular candidate files whose mtime is
// at or before cutoff. Suspicious candidate-shaped filesystem entries fail
// closed rather than being followed or silently ignored.
func (s *Store) PruneBefore(cutoff time.Time) (int, error) {
	dir, err := s.root.Open(".")
	if err != nil {
		return 0, err
	}
	entries, err := dir.ReadDir(-1)
	closeErr := dir.Close()
	if err != nil {
		return 0, err
	}
	if closeErr != nil {
		return 0, closeErr
	}
	removed := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".candidate") {
			continue
		}
		id := strings.TrimSuffix(name, ".candidate")
		if !validID(id) {
			return removed, fmt.Errorf("QUARANTINE_INTEGRITY_FAILURE: unsafe candidate filename %q", name)
		}
		info, err := s.root.Lstat(name)
		if err != nil {
			return removed, err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return removed, fmt.Errorf("QUARANTINE_INTEGRITY_FAILURE: candidate %q is not a regular file", id)
		}
		if info.ModTime().After(cutoff) {
			continue
		}
		if err := s.root.Remove(name); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

func candidateName(id string) string { return id + ".candidate" }

func entryFor(id string, content []byte) Entry {
	sum := sha256.Sum256(content)
	return Entry{ID: id, ContentSHA256: hex.EncodeToString(sum[:]), Size: int64(len(content))}
}

func validID(id string) bool {
	if id == "" {
		return false
	}
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}
