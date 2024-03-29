// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package releases

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/internal/pubkey"
	rjson "github.com/hashicorp/hc-install/internal/releasesjson"
	isrc "github.com/hashicorp/hc-install/internal/src"
	"github.com/hashicorp/hc-install/internal/validators"
	"github.com/hashicorp/hc-install/product"
)

// ExactVersion installs the given Version of product
// to OS temp directory, or to InstallDir (if not empty)
type ExactVersion struct {
	Product    product.Product
	Version    *version.Version
	InstallDir string
	Timeout    time.Duration

	// Enterprise indicates installation of enterprise version (leave nil for Community editions)
	Enterprise *EnterpriseOptions

	SkipChecksumVerification bool

	// ArmoredPublicKey is a public PGP key in ASCII/armor format to use
	// instead of built-in pubkey to verify signature of downloaded checksums
	ArmoredPublicKey string

	apiBaseURL    string
	logger        *log.Logger
	pathsToRemove []string
}

func (*ExactVersion) IsSourceImpl() isrc.InstallSrcSigil {
	return isrc.InstallSrcSigil{}
}

func (ev *ExactVersion) SetLogger(logger *log.Logger) {
	ev.logger = logger
}

func (ev *ExactVersion) log() *log.Logger {
	if ev.logger == nil {
		return discardLogger
	}
	return ev.logger
}

func (ev *ExactVersion) Validate() error {
	if !validators.IsProductNameValid(ev.Product.Name) {
		return fmt.Errorf("invalid product name: %q", ev.Product.Name)
	}

	if !validators.IsBinaryNameValid(ev.Product.BinaryName()) {
		return fmt.Errorf("invalid binary name: %q", ev.Product.BinaryName())
	}

	if ev.Version == nil {
		return fmt.Errorf("unknown version")
	}

	if err := validateEnterpriseOptions(ev.Enterprise); err != nil {
		return err
	}

	return nil
}

func (ev *ExactVersion) Install(ctx context.Context) (string, error) {
	timeout := defaultInstallTimeout
	if ev.Timeout > 0 {
		timeout = ev.Timeout
	}
	ctx, cancelFunc := context.WithTimeout(ctx, timeout)
	defer cancelFunc()

	if ev.pathsToRemove == nil {
		ev.pathsToRemove = make([]string, 0)
	}

	dstDir := ev.InstallDir
	if dstDir == "" {
		var err error
		dirName := fmt.Sprintf("%s_*", ev.Product.Name)
		dstDir, err = ioutil.TempDir("", dirName)
		if err != nil {
			return "", err
		}
		ev.pathsToRemove = append(ev.pathsToRemove, dstDir)
		ev.log().Printf("created new temp dir at %s", dstDir)
	}
	ev.log().Printf("will install into dir at %s", dstDir)

	rels := rjson.NewReleases()
	if ev.apiBaseURL != "" {
		rels.BaseURL = ev.apiBaseURL
	}
	rels.SetLogger(ev.log())
	installVersion := ev.Version
	if ev.Enterprise != nil {
		installVersion = versionWithMetadata(installVersion, enterpriseVersionMetadata(ev.Enterprise))
	}
	pv, err := rels.GetProductVersion(ctx, ev.Product.Name, installVersion)
	if err != nil {
		return "", err
	}

	d := &rjson.Downloader{
		Logger:           ev.log(),
		VerifyChecksum:   !ev.SkipChecksumVerification,
		ArmoredPublicKey: pubkey.DefaultPublicKey,
		BaseURL:          rels.BaseURL,
	}
	if ev.ArmoredPublicKey != "" {
		d.ArmoredPublicKey = ev.ArmoredPublicKey
	}
	if ev.apiBaseURL != "" {
		d.BaseURL = ev.apiBaseURL
	}

	licenseDir := ""
	if ev.Enterprise != nil {
		licenseDir = ev.Enterprise.LicenseDir
	}
	zipFilePath, err := d.DownloadAndUnpack(ctx, pv, dstDir, licenseDir)
	if zipFilePath != "" {
		ev.pathsToRemove = append(ev.pathsToRemove, zipFilePath)
	}
	if err != nil {
		return "", err
	}

	execPath := filepath.Join(dstDir, ev.Product.BinaryName())

	ev.pathsToRemove = append(ev.pathsToRemove, execPath)

	ev.log().Printf("changing perms of %s", execPath)
	err = os.Chmod(execPath, 0o700)
	if err != nil {
		return "", err
	}

	return execPath, nil
}

func (ev *ExactVersion) Remove(ctx context.Context) error {
	if ev.pathsToRemove != nil {
		for _, path := range ev.pathsToRemove {
			err := os.RemoveAll(path)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// versionWithMetadata returns a new version by combining the given version with the given metadata
func versionWithMetadata(v *version.Version, metadata string) *version.Version {
	if v == nil {
		return nil
	}

	if metadata == "" {
		return v
	}

	v2, err := version.NewVersion(fmt.Sprintf("%s+%s", v.Core(), metadata))
	if err != nil {
		return nil
	}

	return v2
}
