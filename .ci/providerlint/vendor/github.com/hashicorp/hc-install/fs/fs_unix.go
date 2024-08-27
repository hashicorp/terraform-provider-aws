// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !windows
// +build !windows

package fs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func lookupDirs(extraDirs []string) []string {
	pathVar := os.Getenv("PATH")
	dirs := filepath.SplitList(pathVar)
	for _, ep := range extraDirs {
		dirs = append(dirs, ep)
	}
	return dirs
}

func findFile(dirs []string, file string, f fileCheckFunc) (string, error) {
	for _, dir := range dirs {
		if dir == "" {
			// Unix shell semantics: path element "" means "."
			dir = "."
		}
		path := filepath.Join(dir, file)
		if err := f(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("%s: %w", file, exec.ErrNotFound)
}

func checkExecutable(file string) error {
	d, err := os.Stat(file)
	if err != nil {
		return err
	}
	if m := d.Mode(); !m.IsDir() && m&0111 != 0 {
		return nil
	}
	return os.ErrPermission
}
