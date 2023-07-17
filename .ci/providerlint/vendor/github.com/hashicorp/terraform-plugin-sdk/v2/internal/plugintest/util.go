// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package plugintest

import (
	"fmt"
	"os"
	"path/filepath"
)

func symlinkFile(src string, dest string) error {
	err := os.Symlink(src, dest)

	if err != nil {
		return fmt.Errorf("unable to symlink %q to %q: %w", src, dest, err)
	}

	srcInfo, err := os.Stat(src)

	if err != nil {
		return fmt.Errorf("unable to stat %q: %w", src, err)
	}

	err = os.Chmod(dest, srcInfo.Mode())

	if err != nil {
		return fmt.Errorf("unable to set %q permissions: %w", dest, err)
	}

	return nil
}

// symlinkDirectoriesOnly finds only the first-level child directories in srcDir
// and symlinks them into destDir.
// Unlike symlinkDir, this is done non-recursively in order to limit the number
// of file descriptors used.
func symlinkDirectoriesOnly(srcDir string, destDir string) error {
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return fmt.Errorf("unable to stat source directory %q: %w", srcDir, err)
	}

	err = os.MkdirAll(destDir, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("unable to make destination directory %q: %w", destDir, err)
	}

	dirEntries, err := os.ReadDir(srcDir)

	if err != nil {
		return fmt.Errorf("unable to read source directory %q: %w", srcDir, err)
	}

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			continue
		}

		srcPath := filepath.Join(srcDir, dirEntry.Name())
		destPath := filepath.Join(destDir, dirEntry.Name())
		err := symlinkFile(srcPath, destPath)

		if err != nil {
			return fmt.Errorf("unable to symlink directory %q to %q: %w", srcPath, destPath, err)
		}
	}

	return nil
}
