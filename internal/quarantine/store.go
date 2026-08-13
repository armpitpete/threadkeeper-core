package quarantine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct { dir string }

type Entry struct {
	ID            string `json:"id"`
	ContentSHA256 string `json:"content_sha256"`
	Size          int64  `json:"size"`
}

func Open(dir string) (*Store, error) {
	if dir == "" { return nil, fmt.Errorf("QUARANTINE_INVALID: directory is required") }
	if info, err := os.Lstat(dir); err == nil && info.Mode()&os.ModeSymlink != 0 { return nil, fmt.Errorf("QUARANTINE_INVALID: directory must not be a symlink") }
	if err := os.MkdirAll(dir, 0o700); err != nil { return nil, err }
	if err := os.Chmod(dir, 0o700); err != nil { return nil, err }
	abs, err := filepath.Abs(dir); if err != nil { return nil, err }
	return &Store{dir: abs}, nil
}

func (s *Store) Put(id string, content []byte) (Entry, error) {
	if !validID(id) { return Entry{}, fmt.Errorf("QUARANTINE_INVALID: unsafe candidate id %q", id) }
	path := filepath.Join(s.dir, id+".candidate")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil { return Entry{}, err }
	if _, err := f.Write(content); err != nil { f.Close(); os.Remove(path); return Entry{}, err }
	if err := f.Close(); err != nil { os.Remove(path); return Entry{}, err }
	sum := sha256.Sum256(content)
	return Entry{ID: id, ContentSHA256: hex.EncodeToString(sum[:]), Size: int64(len(content))}, nil
}

func (s *Store) Read(entry Entry) ([]byte, error) {
	if !validID(entry.ID) { return nil, fmt.Errorf("QUARANTINE_INVALID: unsafe candidate id") }
	content, err := os.ReadFile(filepath.Join(s.dir, entry.ID+".candidate")); if err != nil { return nil, err }
	sum := sha256.Sum256(content)
	if int64(len(content)) != entry.Size || hex.EncodeToString(sum[:]) != entry.ContentSHA256 { return nil, fmt.Errorf("QUARANTINE_INTEGRITY_FAILURE: candidate bytes changed") }
	return content, nil
}

func (s *Store) Remove(id string) error {
	if !validID(id) { return fmt.Errorf("QUARANTINE_INVALID: unsafe candidate id") }
	return os.Remove(filepath.Join(s.dir, id+".candidate"))
}

func validID(id string) bool {
	if id == "" { return false }
	for _, c := range id { if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') { return false } }
	return true
}
