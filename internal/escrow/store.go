package escrow

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Store records the identity of the original escrow directory.
// Each operation opens a rooted handle, verifies that identity and keeps
// the handle open until its reads or writes have completed.
type Store struct {
	dir string
	rootIdentity os.FileInfo
}

func OpenStore(dir string) (*Store, error) {
	if dir == "" {
		return nil, fmt.Errorf("ESCROW_STORE_INVALID: directory required")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o700); err != nil {
		return nil, err
	}
	before, err := os.Lstat(abs)
	if err != nil {
		return nil, err
	}
	if before.Mode()&os.ModeSymlink != 0 || !before.IsDir() {
		return nil, fmt.Errorf("ESCROW_STORE_INVALID: root must be a real directory")
	}
	root, err := os.OpenRoot(abs)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	opened, err := root.Stat(".")
	if err != nil {
		return nil, err
	}
	after, err := os.Lstat(abs)
	if err != nil {
		return nil, err
	}
	if after.Mode()&os.ModeSymlink != 0 || !os.SameFile(before, opened) || !os.SameFile(opened, after) {
		return nil, fmt.Errorf("ESCROW_STORE_INVALID: root changed while opening")
	}
	if err := root.Chmod(".", 0o700); err != nil {
		return nil, err
	}
	return &Store{dir: abs, rootIdentity: opened}, nil
}

// openVerifiedRoot detects replacement of the accepted store directory.
// The returned handle pins the directory while the calling operation runs.
func (s *Store) openVerifiedRoot() (*os.Root, error) {
	root, err := os.OpenRoot(s.dir)
	if err != nil {
		return nil, err
	}
	current, err := root.Stat(".")
	if err != nil {
		root.Close()
		return nil, err
	}
	if !os.SameFile(s.rootIdentity, current) {
		root.Close()
		return nil, fmt.Errorf("ESCROW_STORE_INVALID: configured root identity changed")
	}
	return root, nil
}

func (s *Store) Put(sourceID, versionID, mediaType string, content []byte) (Snapshot, error) {
	if sourceID == "" || versionID == "" {
		return Snapshot{}, fmt.Errorf("ESCROW_INVALID: source/version identity required")
	}
	root, err := s.openVerifiedRoot()
	if err != nil {
		return Snapshot{}, err
	}
	defer root.Close()

	digest := HashContent(content)
	name := digest + ".blob"
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		existing, readErr := readBlobFromRoot(root, name)
		if readErr != nil {
			return Snapshot{}, readErr
		}
		if HashContent(existing) != digest {
			return Snapshot{}, fmt.Errorf("ESCROW_INTEGRITY_FAILURE: existing blob does not match filename digest")
		}
		return Snapshot{
			SourceID: sourceID, VersionID: versionID, ContentSHA256: digest,
			Size: int64(len(content)), MediaType: mediaType,
		}, nil
	}
	if err != nil {
		return Snapshot{}, err
	}
	if _, err := file.Write(content); err != nil {
		file.Close()
		root.Remove(name)
		return Snapshot{}, err
	}
	if err := file.Close(); err != nil {
		root.Remove(name)
		return Snapshot{}, err
	}
	return Snapshot{
		SourceID: sourceID, VersionID: versionID, ContentSHA256: digest,
		Size: int64(len(content)), MediaType: mediaType,
	}, nil
}

func (s *Store) Get(snapshot Snapshot) ([]byte, error) {
	root, err := s.openVerifiedRoot()
	if err != nil {
		return nil, err
	}
	defer root.Close()
	content, err := readBlobFromRoot(root, snapshot.ContentSHA256+".blob")
	if err != nil {
		return nil, err
	}
	if err := VerifyContent(snapshot, content); err != nil {
		return nil, err
	}
	return content, nil
}

// readBlobFromRoot accepts only regular files within the already pinned root.
func readBlobFromRoot(root *os.Root, name string) ([]byte, error) {
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("ESCROW_STORE_INVALID: blob must be a regular file")
	}
	return io.ReadAll(file)
}
