package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-version"
)

// installGoVersion installs given version of Go using Go
// according to https://golang.org/doc/manage-install
func (gb *GoBuild) installGoVersion(ctx context.Context, v *version.Version) (Go, error) {
	versionString := v.Core().String()

	// trim 0 patch versions as that's how Go does it :shrug:
	shortVersion := strings.TrimSuffix(versionString, ".0")
	pkgURL := fmt.Sprintf("golang.org/dl/go%s", shortVersion)

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

	cmdName := fmt.Sprintf("go%s", shortVersion)

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
