// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-version"
)

var v1_21 = version.Must(version.NewVersion("1.21"))

// installGoVersion installs given version of Go using Go
// according to https://golang.org/doc/manage-install
func (gb *GoBuild) installGoVersion(ctx context.Context, v *version.Version) (Go, error) {
	goVersion := v.String()

	// trim 0 patch versions as that's how Go does it
	// for versions prior to 1.21
	// See https://github.com/golang/go/issues/62136
	if v.LessThan(v1_21) {
		versionString := v.Core().String()
		goVersion = strings.TrimSuffix(versionString, ".0")
	}
	pkgURL := fmt.Sprintf("golang.org/dl/go%s", goVersion)

	gb.log().Printf("go getting %q", pkgURL)
	cmd := exec.CommandContext(ctx, "go", "get", pkgURL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Go{}, fmt.Errorf("unable to get Go %s: %w\n%s", v, err, out)
	}

	gb.log().Printf("go installing %q", pkgURL)
	cmd = exec.CommandContext(ctx, "go", "install", pkgURL)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return Go{}, fmt.Errorf("unable to install Go %s: %w\n%s", v, err, out)
	}

	cmdName := fmt.Sprintf("go%s", goVersion)

	gb.log().Printf("downloading go %q", v)
	cmd = exec.CommandContext(ctx, cmdName, "download")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return Go{}, fmt.Errorf("unable to download Go %s: %w\n%s", v, err, out)
	}
	gb.log().Printf("download of go %q finished", v)

	cleanupFunc := func(ctx context.Context) {
		cmd = exec.CommandContext(ctx, cmdName, "env", "GOROOT")
		out, err = cmd.CombinedOutput()
		if err != nil {
			return
		}
		rootPath := strings.TrimSpace(string(out))

		// run some extra checks before deleting, just to be sure
		if rootPath != "" && strings.HasSuffix(rootPath, v.String()) {
			os.RemoveAll(rootPath)
		}
	}

	return Go{
		Cmd:         cmdName,
		CleanupFunc: cleanupFunc,
		Version:     v,
	}, nil
}
