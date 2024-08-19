// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfexec

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

type testConfig struct {
	testsDirectory string
}

var defaultTestOptions = testConfig{}

type TestOption interface {
	configureTest(*testConfig)
}

func (opt *TestsDirectoryOption) configureTest(conf *testConfig) {
	conf.testsDirectory = opt.testsDirectory
}

// Test represents the terraform test -json subcommand.
//
// The given io.Writer, if specified, will receive
// [machine-readable](https://developer.hashicorp.com/terraform/internals/machine-readable-ui)
// JSON from Terraform including test results.
func (tf *Terraform) Test(ctx context.Context, w io.Writer, opts ...TestOption) error {
	err := tf.compatible(ctx, tf1_6_0, nil)

	if err != nil {
		return fmt.Errorf("terraform test was added in 1.6.0: %w", err)
	}

	tf.SetStdout(w)

	testCmd := tf.testCmd(ctx)

	err = tf.runTerraformCmd(ctx, testCmd)

	if err != nil {
		return err
	}

	return nil
}

func (tf *Terraform) testCmd(ctx context.Context, opts ...TestOption) *exec.Cmd {
	c := defaultTestOptions

	for _, o := range opts {
		o.configureTest(&c)
	}

	args := []string{"test", "-json"}

	if c.testsDirectory != "" {
		args = append(args, "-tests-directory="+c.testsDirectory)
	}

	return tf.buildTerraformCmd(ctx, nil, args...)
}
