package quarantine

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

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
	info, err := os.Lstat(abs)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("QUARANTINE_MISSING: %w", err)
		}
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return nil, fmt.Errorf("QUARANTINE_INVALID: directory must be a real directory")
	}
	root, err := os.OpenRoot(abs)
	if err != nil {
		return nil, err
	}
	return &Store{dir: abs, root: root}, nil
}
