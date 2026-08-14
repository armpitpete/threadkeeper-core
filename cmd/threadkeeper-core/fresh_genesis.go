package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/ledger"
)

func freshGenesisInitCommand(args []string) error {
	if len(args) < 3 || len(args) > 4 {
		return fmt.Errorf("expected ledger git directory, Genesis file, bootstrap seed root and optional authoritative ref")
	}
	rawGenesis, err := os.ReadFile(args[1])
	if err != nil {
		return err
	}
	seedFiles, err := collectFreshGenesisSeedFiles(args[2])
	if err != nil {
		return err
	}
	ref := gitledger.DefaultRef
	if len(args) == 4 {
		ref = args[3]
	}
	evidence, err := ledger.InitializeFreshGenesis(context.Background(), args[0], ref, rawGenesis, seedFiles)
	if err != nil {
		return err
	}
	return writeIndentedJSON(evidence)
}

func collectFreshGenesisSeedFiles(root string) (map[string][]byte, error) {
	info, err := os.Lstat(root)
	if err != nil {
		return nil, fmt.Errorf("inspect bootstrap seed root: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return nil, fmt.Errorf("bootstrap seed root must be a real directory, not a symlink or non-directory")
	}
	files := map[string][]byte{}
	err = filepath.WalkDir(root, func(full string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, full)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("bootstrap seed entry %q is a symlink", filepath.ToSlash(rel))
		}
		if entry.IsDir() {
			return nil
		}
		entryInfo, err := entry.Info()
		if err != nil {
			return err
		}
		if !entryInfo.Mode().IsRegular() {
			return fmt.Errorf("bootstrap seed entry %q is not a regular file", filepath.ToSlash(rel))
		}
		raw, err := os.ReadFile(full)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = raw
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
