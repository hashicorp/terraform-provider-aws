// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package build

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"golang.org/x/mod/modfile"
)

var discardLogger = log.New(ioutil.Discard, "", 0)

// GoBuild represents a Go builder (to run "go build")
type GoBuild struct {
	Version         *version.Version
	DetectVendoring bool

	pathToRemove string
	logger       *log.Logger
}

func (gb *GoBuild) SetLogger(logger *log.Logger) {
	gb.logger = logger
}

func (gb *GoBuild) log() *log.Logger {
	if gb.logger == nil {
		return discardLogger
	}
	return gb.logger
}

// Build runs "go build" within a given repo to produce binaryName in targetDir
func (gb *GoBuild) Build(ctx context.Context, repoDir, targetDir, binaryName string) (string, error) {
	reqGo, err := gb.ensureRequiredGoVersion(ctx, repoDir)
	if err != nil {
		return "", err
	}
	defer reqGo.CleanupFunc(ctx)

	if reqGo.Version == nil {
		gb.logger.Println("building using default available Go")
	} else {
		gb.logger.Printf("building using Go %s", reqGo.Version)
	}

	// `go build` would download dependencies as a side effect, but we attempt
	// to do it early in a separate step, such that we can easily distinguish
	// network failures from build failures.
	//
	// Note, that `go mod download` was introduced in Go 1.11
	// See https://github.com/golang/go/commit/9f4ea6c2
	minGoVersion := version.Must(version.NewVersion("1.11"))
	if reqGo.Version.GreaterThanOrEqual(minGoVersion) {
		downloadArgs := []string{"mod", "download"}
		gb.log().Printf("executing %s %q in %q", reqGo.Cmd, downloadArgs, repoDir)
		cmd := exec.CommandContext(ctx, reqGo.Cmd, downloadArgs...)
		cmd.Dir = repoDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("unable to download dependencies: %w\n%s", err, out)
		}
	}

	buildArgs := []string{"build", "-o", filepath.Join(targetDir, binaryName)}

	if gb.DetectVendoring {
		vendorDir := filepath.Join(repoDir, "vendor")
		if fi, err := os.Stat(vendorDir); err == nil && fi.IsDir() {
			buildArgs = append(buildArgs, "-mod", "vendor")
		}
	}

	gb.log().Printf("executing %s %q in %q", reqGo.Cmd, buildArgs, repoDir)
	cmd := exec.CommandContext(ctx, reqGo.Cmd, buildArgs...)
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("unable to build: %w\n%s", err, out)
	}

	binPath := filepath.Join(targetDir, binaryName)

	gb.pathToRemove = binPath

	return binPath, nil
}

func (gb *GoBuild) Remove(ctx context.Context) error {
	return os.RemoveAll(gb.pathToRemove)
}

type Go struct {
	Cmd         string
	CleanupFunc CleanupFunc
	Version     *version.Version
}

func (gb *GoBuild) ensureRequiredGoVersion(ctx context.Context, repoDir string) (Go, error) {
	cmdName := "go"
	noopCleanupFunc := func(context.Context) {}

	var installedVersion *version.Version

	if gb.Version != nil {
		gb.logger.Printf("attempting to satisfy explicit requirement for Go %s", gb.Version)
		goVersion, err := GetGoVersion(ctx)
		if err != nil {
			return Go{
				Cmd:         cmdName,
				CleanupFunc: noopCleanupFunc,
			}, err
		}

		if !goVersion.GreaterThanOrEqual(gb.Version) {
			// found incompatible version, try downloading the desired one
			return gb.installGoVersion(ctx, gb.Version)
		}
		installedVersion = goVersion
	}

	if requiredVersion, ok := guessRequiredGoVersion(repoDir); ok {
		gb.logger.Printf("attempting to satisfy guessed Go requirement %s", requiredVersion)
		goVersion, err := GetGoVersion(ctx)
		if err != nil {
			return Go{
				Cmd:         cmdName,
				CleanupFunc: noopCleanupFunc,
			}, err
		}

		if !goVersion.GreaterThanOrEqual(requiredVersion) {
			// found incompatible version, try downloading the desired one
			return gb.installGoVersion(ctx, requiredVersion)
		}
		installedVersion = goVersion
	} else {
		gb.logger.Println("unable to guess Go requirement")
	}

	return Go{
		Cmd:         cmdName,
		CleanupFunc: noopCleanupFunc,
		Version:     installedVersion,
	}, nil
}

// CleanupFunc represents a function to be called once Go is no longer needed
// e.g. to remove any version installed temporarily per requirements
type CleanupFunc func(context.Context)

func guessRequiredGoVersion(repoDir string) (*version.Version, bool) {
	goEnvFile := filepath.Join(repoDir, ".go-version")
	if fi, err := os.Stat(goEnvFile); err == nil && !fi.IsDir() {
		b, err := ioutil.ReadFile(goEnvFile)
		if err != nil {
			return nil, false
		}
		requiredVersion, err := version.NewVersion(string(bytes.TrimSpace(b)))
		if err != nil {
			return nil, false
		}
		return requiredVersion, true
	}

	goModFile := filepath.Join(repoDir, "go.mod")
	if fi, err := os.Stat(goModFile); err == nil && !fi.IsDir() {
		b, err := ioutil.ReadFile(goModFile)
		if err != nil {
			return nil, false
		}
		f, err := modfile.ParseLax(fi.Name(), b, nil)
		if err != nil {
			return nil, false
		}
		if f.Go == nil {
			return nil, false
		}
		requiredVersion, err := version.NewVersion(f.Go.Version)
		if err != nil {
			return nil, false
		}
		return requiredVersion, true
	}

	return nil, false
}
