// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fs

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/errors"
	"github.com/hashicorp/hc-install/internal/src"
	"github.com/hashicorp/hc-install/internal/validators"
	"github.com/hashicorp/hc-install/product"
)

// Version finds the first executable binary of the product name
// which matches the version constraint within system $PATH and any declared ExtraPaths
// (which are *appended* to any directories in $PATH)
type Version struct {
	Product     product.Product
	Constraints version.Constraints
	ExtraPaths  []string
	Timeout     time.Duration

	logger *log.Logger
}

func (*Version) IsSourceImpl() src.InstallSrcSigil {
	return src.InstallSrcSigil{}
}

func (v *Version) SetLogger(logger *log.Logger) {
	v.logger = logger
}

func (v *Version) log() *log.Logger {
	if v.logger == nil {
		return discardLogger
	}
	return v.logger
}

func (v *Version) Validate() error {
	if !validators.IsBinaryNameValid(v.Product.BinaryName()) {
		return fmt.Errorf("invalid binary name: %q", v.Product.BinaryName())
	}
	if len(v.Constraints) == 0 {
		return fmt.Errorf("undeclared version constraints")
	}
	if v.Product.GetVersion == nil {
		return fmt.Errorf("undeclared version getter")
	}
	return nil
}

func (v *Version) Find(ctx context.Context) (string, error) {
	timeout := defaultTimeout
	if v.Timeout > 0 {
		timeout = v.Timeout
	}
	ctx, cancelFunc := context.WithTimeout(ctx, timeout)
	defer cancelFunc()

	execPath, err := findFile(lookupDirs(v.ExtraPaths), v.Product.BinaryName(), func(file string) error {
		err := checkExecutable(file)
		if err != nil {
			return err
		}

		ver, err := v.Product.GetVersion(ctx, file)
		if err != nil {
			return err
		}

		for _, vc := range v.Constraints {
			if !vc.Check(ver) {
				return fmt.Errorf("version (%s) doesn't meet constraints %s", ver, vc.String())
			}
		}

		return nil
	})
	if err != nil {
		return "", errors.SkippableErr(err)
	}

	if !filepath.IsAbs(execPath) {
		var err error
		execPath, err = filepath.Abs(execPath)
		if err != nil {
			return "", errors.SkippableErr(err)
		}
	}

	return execPath, nil
}
