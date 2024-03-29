// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package plugintest

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/logging"
)

// AutoInitProviderHelper is the main entrypoint for testing provider plugins
// using this package. It is intended to be called during TestMain to prepare
// for provider testing.
//
// AutoInitProviderHelper will discover the location of a current Terraform CLI
// executable to test against, detect whether a prior version of the plugin is
// available for upgrade tests, and then will return an object containing the
// results of that initialization which can then be stored in a global variable
// for use in other tests.
func AutoInitProviderHelper(ctx context.Context, sourceDir string) *Helper {
	helper, err := AutoInitHelper(ctx, sourceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot run Terraform provider tests: %s\n", err)
		os.Exit(1)
	}
	return helper
}

// Helper is intended as a per-package singleton created in TestMain which
// other tests in a package can use to create Terraform execution contexts
type Helper struct {
	baseDir string

	// sourceDir is the dir containing the provider source code, needed
	// for tests that use fixture files.
	sourceDir     string
	terraformExec string

	// execTempDir is created during DiscoverConfig to store any downloaded
	// binaries
	execTempDir string
}

// AutoInitHelper uses the auto-discovery behavior of DiscoverConfig to prepare
// a configuration and then calls InitHelper with it. This is a convenient
// way to get the standard init behavior based on environment variables, and
// callers should use this unless they have an unusual requirement that calls
// for constructing a config in a different way.
func AutoInitHelper(ctx context.Context, sourceDir string) (*Helper, error) {
	config, err := DiscoverConfig(ctx, sourceDir)
	if err != nil {
		return nil, err
	}

	return InitHelper(ctx, config)
}

// InitHelper prepares a testing helper with the given configuration.
//
// For most callers it is sufficient to call AutoInitHelper instead, which
// will construct a configuration automatically based on certain environment
// variables.
//
// If this function returns an error then it may have left some temporary files
// behind in the system's temporary directory. There is currently no way to
// automatically clean those up.
func InitHelper(ctx context.Context, config *Config) (*Helper, error) {
	tempDir := os.Getenv(EnvTfAccTempDir)
	baseDir, err := os.MkdirTemp(tempDir, "plugintest")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory for test helper: %s", err)
	}

	return &Helper{
		baseDir:       baseDir,
		sourceDir:     config.SourceDir,
		terraformExec: config.TerraformExec,
		execTempDir:   config.execTempDir,
	}, nil
}

// Close cleans up temporary files and directories created to support this
// helper, returning an error if any of the cleanup fails.
//
// Call this before returning from TestMain to minimize the amount of detritus
// left behind in the filesystem after the tests complete.
func (h *Helper) Close() error {
	if h.execTempDir != "" {
		err := os.RemoveAll(h.execTempDir)
		if err != nil {
			return err
		}
	}
	return os.RemoveAll(h.baseDir)
}

// NewWorkingDir creates a new working directory for use in the implementation
// of a single test, returning a WorkingDir object representing that directory.
//
// If the working directory object is not itself closed by the time the test
// program exits, the Close method on the helper itself will attempt to
// delete it.
func (h *Helper) NewWorkingDir(ctx context.Context, t TestControl) (*WorkingDir, error) {
	dir, err := os.MkdirTemp(h.baseDir, "work")
	if err != nil {
		return nil, err
	}

	ctx = logging.TestWorkingDirectoryContext(ctx, dir)

	// symlink the provider source files into the config directory
	// e.g. testdata
	logging.HelperResourceTrace(ctx, "Symlinking source directories to work directory")
	err = symlinkDirectoriesOnly(h.sourceDir, dir)
	if err != nil {
		return nil, err
	}

	tf, err := tfexec.NewTerraform(dir, h.terraformExec)

	if err != nil {
		return nil, fmt.Errorf("unable to create terraform-exec instance: %w", err)
	}

	err = tf.SetDisablePluginTLS(true)

	if err != nil {
		return nil, fmt.Errorf("unable to disable terraform-exec plugin TLS: %w", err)
	}

	err = tf.SetSkipProviderVerify(true) // Only required for Terraform CLI 0.12.x

	var mismatch *tfexec.ErrVersionMismatch
	if err != nil && !errors.As(err, &mismatch) {
		return nil, fmt.Errorf("unable to disable terraform-exec provider verification: %w", err)
	}

	tfAccLog := os.Getenv(EnvTfAccLog)
	tfAccLogPath := os.Getenv(EnvTfAccLogPath)
	tfLogCore := os.Getenv(EnvTfLogCore)
	tfLogPathMask := os.Getenv(EnvTfLogPathMask)
	tfLogProvider := os.Getenv(EnvTfLogProvider)

	if tfAccLog != "" && tfLogCore != "" {
		err = fmt.Errorf(
			"Invalid environment variable configuration. Cannot set both TF_ACC_LOG and TF_LOG_CORE. " +
				"Use TF_LOG_CORE and TF_LOG_PROVIDER to separately control the Terraform CLI logging subsystems. " +
				"To control the Go standard library log package for the provider under test, use TF_LOG.",
		)
		logging.HelperResourceError(ctx, err.Error())
		return nil, err
	}

	if tfAccLog != "" {
		logging.HelperResourceTrace(
			ctx,
			fmt.Sprintf("Setting terraform-exec log level via %s environment variable, if Terraform CLI is version 0.15 or later", EnvTfAccLog),
			map[string]interface{}{logging.KeyTestTerraformLogLevel: tfAccLog},
		)

		err := tf.SetLog(tfAccLog)

		if err != nil {
			if !errors.As(err, new(*tfexec.ErrVersionMismatch)) {
				logging.HelperResourceError(
					ctx,
					"Unable to set terraform-exec log level",
					map[string]interface{}{logging.KeyError: err.Error()},
				)
				return nil, fmt.Errorf("unable to set terraform-exec log level (%s): %w", tfAccLog, err)
			}

			logging.HelperResourceWarn(
				ctx,
				fmt.Sprintf("Unable to set terraform-exec log level via %s environment variable, as Terraform CLI is version 0.14 or earlier. It will default to TRACE.", EnvTfAccLog),
				map[string]interface{}{logging.KeyTestTerraformLogLevel: "TRACE"},
			)
		}
	}

	if tfLogCore != "" {
		logging.HelperResourceTrace(
			ctx,
			fmt.Sprintf("Setting terraform-exec core log level via %s environment variable, if Terraform CLI is version 0.15 or later", EnvTfLogCore),
			map[string]interface{}{
				logging.KeyTestTerraformLogCoreLevel: tfLogCore,
			},
		)

		err := tf.SetLogCore(tfLogCore)

		if err != nil {
			logging.HelperResourceError(
				ctx,
				"Unable to set terraform-exec core log level",
				map[string]interface{}{logging.KeyError: err.Error()},
			)
			return nil, fmt.Errorf("unable to set terraform-exec core log level (%s): %w", tfLogCore, err)
		}
	}

	if tfLogProvider != "" {
		logging.HelperResourceTrace(
			ctx,
			fmt.Sprintf("Setting terraform-exec provider log level via %s environment variable, if Terraform CLI is version 0.15 or later", EnvTfLogProvider),
			map[string]interface{}{
				logging.KeyTestTerraformLogCoreLevel: tfLogProvider,
			},
		)

		err := tf.SetLogProvider(tfLogProvider)

		if err != nil {
			logging.HelperResourceError(
				ctx,
				"Unable to set terraform-exec provider log level",
				map[string]interface{}{logging.KeyError: err.Error()},
			)
			return nil, fmt.Errorf("unable to set terraform-exec provider log level (%s): %w", tfLogProvider, err)
		}
	}

	var logPath, logPathEnvVar string

	if tfAccLogPath != "" {
		logPath = tfAccLogPath
		logPathEnvVar = EnvTfAccLogPath
	}

	// Similar to helper/logging.LogOutput() and
	// terraform-plugin-log/tfsdklog.RegisterTestSink(), the TF_LOG_PATH_MASK
	// environment variable should take precedence over TF_ACC_LOG_PATH.
	if tfLogPathMask != "" {
		// Escape special characters which may appear if we have subtests
		testName := strings.Replace(t.Name(), "/", "__", -1)
		logPath = fmt.Sprintf(tfLogPathMask, testName)
		logPathEnvVar = EnvTfLogPathMask
	}

	if logPath != "" {
		logging.HelperResourceTrace(
			ctx,
			fmt.Sprintf("Setting terraform-exec log path via %s environment variable", logPathEnvVar),
			map[string]interface{}{logging.KeyTestTerraformLogPath: logPath},
		)

		if err := tf.SetLogPath(logPath); err != nil {
			return nil, fmt.Errorf("unable to set terraform-exec log path (%s): %w", logPath, err)
		}
	}

	return &WorkingDir{
		h:             h,
		tf:            tf,
		baseDir:       dir,
		terraformExec: h.terraformExec,
	}, nil
}

// RequireNewWorkingDir is a variant of NewWorkingDir that takes a TestControl
// object and will immediately fail the running test if the creation of the
// working directory fails.
func (h *Helper) RequireNewWorkingDir(ctx context.Context, t TestControl) *WorkingDir {
	t.Helper()

	wd, err := h.NewWorkingDir(ctx, t)
	if err != nil {
		t := testingT{t}
		t.Fatalf("failed to create new working directory: %s", err)
		return nil
	}
	return wd
}

// WorkingDirectory returns the working directory being used when running tests.
func (h *Helper) WorkingDirectory() string {
	return h.baseDir
}

// TerraformExecPath returns the location of the Terraform CLI executable that
// should be used when running tests.
func (h *Helper) TerraformExecPath() string {
	return h.terraformExec
}
