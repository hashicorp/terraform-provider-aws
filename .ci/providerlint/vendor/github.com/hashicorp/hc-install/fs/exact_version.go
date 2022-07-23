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

// ExactVersion finds the first executable binary of the product name
// which matches the Version within system $PATH and any declared ExtraPaths
// (which are *appended* to any directories in $PATH)
type ExactVersion struct {
	Product    product.Product
	Version    *version.Version
	ExtraPaths []string
	Timeout    time.Duration

	logger *log.Logger
}

func (*ExactVersion) IsSourceImpl() src.InstallSrcSigil {
	return src.InstallSrcSigil{}
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
	if !validators.IsBinaryNameValid(ev.Product.BinaryName()) {
		return fmt.Errorf("invalid binary name: %q", ev.Product.BinaryName())
	}
	if ev.Version == nil {
		return fmt.Errorf("undeclared version")
	}
	if ev.Product.GetVersion == nil {
		return fmt.Errorf("undeclared version getter")
	}
	return nil
}

func (ev *ExactVersion) Find(ctx context.Context) (string, error) {
	timeout := defaultTimeout
	if ev.Timeout > 0 {
		timeout = ev.Timeout
	}
	ctx, cancelFunc := context.WithTimeout(ctx, timeout)
	defer cancelFunc()

	execPath, err := findFile(lookupDirs(ev.ExtraPaths), ev.Product.BinaryName(), func(file string) error {
		err := checkExecutable(file)
		if err != nil {
			return err
		}

		v, err := ev.Product.GetVersion(ctx, file)
		if err != nil {
			return err
		}

		if !ev.Version.Equal(v) {
			return fmt.Errorf("version (%s) doesn't match %s", v, ev.Version)
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
