// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package io

import (
	"os"

	homedir "github.com/mitchellh/go-homedir"
)

// ReadFileContents reads the contents of a file into memory.
// Usually a call to this function is protected by an exclusive lock (per resource type)
// to prevent memory exhaustion (e.g. `conns.GlobalMutexKV.Lock`).
// See https://github.com/hashicorp/terraform/issues/9364.
func ReadFileContents(v string) ([]byte, error) {
	filename, err := homedir.Expand(v)
	if err != nil {
		return nil, err
	}

	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return fileContent, nil
}
