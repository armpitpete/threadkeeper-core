package quarantine

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ensureRetryTimeout     = time.Second
	ensureRetryInterval    = 5 * time.Millisecond
	publicationTempPrefix  = ".publish-"
	publicationTempSuffix  = ".tmp"
	publicationNonceBytes  = 16
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

// putHooks is deterministic test instrumentation for the private publication
// lifecycle. Production callers always pass nil.
type putHooks struct {
	beforeSync func()
	syncFile   func(*os.File) error
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
	s, err := openPinned(abs)
	if err != nil {
		return nil, err
	}
	if err := s.root.Chmod(".", 0o700); err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
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

// Put publishes one candidate under id without ever exposing creator-owned
// partial bytes at the final capability name. The complete content is written,
// synced and closed in a private file first. A root-relative hard link then
// atomically publishes the final name without replacing an existing file.
func (s *Store) Put(id string, content []byte) (Entry, error) {
	return s.put(id, content, nil)
}

func (s *Store) put(id string, content []byte, hooks *putHooks) (Entry, error) {
	if !validID(id) {
		return Entry{}, fmt.Errorf("QUARANTINE_INVALID: unsafe candidate id %q", id)
	}
	tempName, err := newPublicationTempName()
	if err != nil {
		return Entry{}, err
	}
	f, err := s.root.OpenFile(tempName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return Entry{}, err
	}
	closed := false
	defer func() {
		if !closed {
			_ = f.Close()
		}
		// This invocation owns only its private file. Once the final hard link
		// exists, cleanup here can never remove another publisher's capability.
		_ = s.root.Remove(tempName)
	}()

	n, err := f.Write(content)
	if err != nil {
		return Entry{}, err
	}
	if n != len(content) {
		return Entry{}, io.ErrShortWrite
	}
	if hooks != nil && hooks.beforeSync != nil {
		hooks.beforeSync()
	}
	if hooks != nil && hooks.syncFile != nil {
		err = hooks.syncFile(f)
	} else {
		err = f.Sync()
	}
	if err != nil {
		return Entry{}, err
	}
	if err := f.Close(); err != nil {
		return Entry{}, err
	}
	closed = true

	name := candidateName(id)
	if err := s.root.Link(tempName, name); err != nil {
		return Entry{}, err
	}

	// Link succeeded only after the private file was fully synced and closed.
	// Verify the published name before reporting success. The temporary hard link
	// is then removed by the deferred cleanup, leaving the final capability.
	published, got, err := s.ReadID(id)
	if err != nil {
		return Entry{}, fmt.Errorf("QUARANTINE_INTEGRITY_FAILURE: verify published candidate: %w", err)
	}
	if !bytes.Equal(got, content) {
		return Entry{}, fmt.Errorf("QUARANTINE_INTEGRITY_FAILURE: published candidate bytes changed")
	}
	return published, nil
}

// Ensure is idempotent for identical completed candidate bytes. A final
// capability name is created only after a private write has synced and closed,
// so an existing final file is never treated as an in-progress publication.
// Reusing a candidate identity for different bytes fails closed.
func (s *Store) Ensure(id string, content []byte) (Entry, error) {
	return s.ensure(id, content, nil)
}

func (s *Store) ensure(id string, content []byte, hooks *putHooks) (Entry, error) {
	deadline := time.Now().Add(ensureRetryTimeout)
	for {
		entry, err := s.put(id, content, hooks)
		if err == nil {
			return entry, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return Entry{}, err
		}

		existing, got, readErr := s.ReadID(id)
		if readErr != nil {
			if errors.Is(readErr, fs.ErrNotExist) && time.Now().Before(deadline) {
				// A completed final capability can disappear only because another
				// lifecycle operation removed it. Retry private no-overwrite
				// publication; higher-level snapshot reconciliation decides whether
				// that material became obsolete through durable acceptance.
				time.Sleep(ensureRetryInterval)
				continue
			}
			return Entry{}, readErr
		}
		if bytes.Equal(got, content) {
			return existing, nil
		}
		return Entry{}, fmt.Errorf("QUARANTINE_CONFLICT: candidate id %q already contains different bytes", id)
	}
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

// PruneBefore removes well-formed regular candidate files and abandoned private
// publication files whose mtime is at or before cutoff. Suspicious candidate or
// publication-shaped filesystem entries fail closed rather than being followed.
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
		if strings.HasPrefix(name, publicationTempPrefix) {
			if !validPublicationTempName(name) {
				return removed, fmt.Errorf("QUARANTINE_INTEGRITY_FAILURE: unsafe publication temp filename %q", name)
			}
			info, err := s.root.Lstat(name)
			if err != nil {
				return removed, err
			}
			if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
				return removed, fmt.Errorf("QUARANTINE_INTEGRITY_FAILURE: publication temp %q is not a regular file", name)
			}
			if info.ModTime().After(cutoff) {
				continue
			}
			if err := s.root.Remove(name); err != nil {
				return removed, err
			}
			removed++
			continue
		}
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

func newPublicationTempName() (string, error) {
	var nonce [publicationNonceBytes]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", fmt.Errorf("QUARANTINE_INVALID: generate publication temp name: %w", err)
	}
	return publicationTempPrefix + hex.EncodeToString(nonce[:]) + publicationTempSuffix, nil
}

func validPublicationTempName(name string) bool {
	if !strings.HasPrefix(name, publicationTempPrefix) || !strings.HasSuffix(name, publicationTempSuffix) {
		return false
	}
	hexPart := strings.TrimSuffix(strings.TrimPrefix(name, publicationTempPrefix), publicationTempSuffix)
	if len(hexPart) != publicationNonceBytes*2 {
		return false
	}
	_, err := hex.DecodeString(hexPart)
	return err == nil
}

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
