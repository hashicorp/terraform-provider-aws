// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfexec

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
)

type applyConfig struct {
	allowDeferral bool
	backup        string
	destroy       bool
	dirOrPlan     string
	lock          bool

	// LockTimeout must be a string with time unit, e.g. '10s'
	lockTimeout  string
	parallelism  int
	reattachInfo ReattachInfo
	refresh      bool
	refreshOnly  bool
	replaceAddrs []string
	state        string
	stateOut     string
	targets      []string

	// Vars: each var must be supplied as a single string, e.g. 'foo=bar'
	vars     []string
	varFiles []string
}

var defaultApplyOptions = applyConfig{
	destroy:     false,
	lock:        true,
	parallelism: 10,
	refresh:     true,
}

// ApplyOption represents options used in the Apply method.
type ApplyOption interface {
	configureApply(*applyConfig)
}

func (opt *ParallelismOption) configureApply(conf *applyConfig) {
	conf.parallelism = opt.parallelism
}

func (opt *BackupOption) configureApply(conf *applyConfig) {
	conf.backup = opt.path
}

func (opt *TargetOption) configureApply(conf *applyConfig) {
	conf.targets = append(conf.targets, opt.target)
}

func (opt *LockTimeoutOption) configureApply(conf *applyConfig) {
	conf.lockTimeout = opt.timeout
}

func (opt *StateOption) configureApply(conf *applyConfig) {
	conf.state = opt.path
}

func (opt *StateOutOption) configureApply(conf *applyConfig) {
	conf.stateOut = opt.path
}

func (opt *VarFileOption) configureApply(conf *applyConfig) {
	conf.varFiles = append(conf.varFiles, opt.path)
}

func (opt *LockOption) configureApply(conf *applyConfig) {
	conf.lock = opt.lock
}

func (opt *RefreshOption) configureApply(conf *applyConfig) {
	conf.refresh = opt.refresh
}

func (opt *RefreshOnlyOption) configureApply(conf *applyConfig) {
	conf.refreshOnly = opt.refreshOnly
}

func (opt *ReplaceOption) configureApply(conf *applyConfig) {
	conf.replaceAddrs = append(conf.replaceAddrs, opt.address)
}

func (opt *VarOption) configureApply(conf *applyConfig) {
	conf.vars = append(conf.vars, opt.assignment)
}

func (opt *DirOrPlanOption) configureApply(conf *applyConfig) {
	conf.dirOrPlan = opt.path
}

func (opt *ReattachOption) configureApply(conf *applyConfig) {
	conf.reattachInfo = opt.info
}

func (opt *DestroyFlagOption) configureApply(conf *applyConfig) {
	conf.destroy = opt.destroy
}

func (opt *AllowDeferralOption) configureApply(conf *applyConfig) {
	conf.allowDeferral = opt.allowDeferral
}

// Apply represents the terraform apply subcommand.
func (tf *Terraform) Apply(ctx context.Context, opts ...ApplyOption) error {
	cmd, err := tf.applyCmd(ctx, opts...)
	if err != nil {
		return err
	}
	return tf.runTerraformCmd(ctx, cmd)
}

// ApplyJSON represents the terraform apply subcommand with the `-json` flag.
// Using the `-json` flag will result in
// [machine-readable](https://developer.hashicorp.com/terraform/internals/machine-readable-ui)
// JSON being written to the supplied `io.Writer`. ApplyJSON is likely to be
// removed in a future major version in favour of Apply returning JSON by default.
func (tf *Terraform) ApplyJSON(ctx context.Context, w io.Writer, opts ...ApplyOption) error {
	err := tf.compatible(ctx, tf0_15_3, nil)
	if err != nil {
		return fmt.Errorf("terraform apply -json was added in 0.15.3: %w", err)
	}

	tf.SetStdout(w)

	cmd, err := tf.applyJSONCmd(ctx, opts...)
	if err != nil {
		return err
	}

	return tf.runTerraformCmd(ctx, cmd)
}

func (tf *Terraform) applyCmd(ctx context.Context, opts ...ApplyOption) (*exec.Cmd, error) {
	c := defaultApplyOptions

	for _, o := range opts {
		o.configureApply(&c)
	}

	args, err := tf.buildApplyArgs(ctx, c)
	if err != nil {
		return nil, err
	}

	return tf.buildApplyCmd(ctx, c, args)
}

func (tf *Terraform) applyJSONCmd(ctx context.Context, opts ...ApplyOption) (*exec.Cmd, error) {
	c := defaultApplyOptions

	for _, o := range opts {
		o.configureApply(&c)
	}

	args, err := tf.buildApplyArgs(ctx, c)
	if err != nil {
		return nil, err
	}

	args = append(args, "-json")

	return tf.buildApplyCmd(ctx, c, args)
}

func (tf *Terraform) buildApplyArgs(ctx context.Context, c applyConfig) ([]string, error) {
	args := []string{"apply", "-no-color", "-auto-approve", "-input=false"}

	// string opts: only pass if set
	if c.backup != "" {
		args = append(args, "-backup="+c.backup)
	}
	if c.lockTimeout != "" {
		args = append(args, "-lock-timeout="+c.lockTimeout)
	}
	if c.state != "" {
		args = append(args, "-state="+c.state)
	}
	if c.stateOut != "" {
		args = append(args, "-state-out="+c.stateOut)
	}
	for _, vf := range c.varFiles {
		args = append(args, "-var-file="+vf)
	}

	// boolean and numerical opts: always pass
	args = append(args, "-lock="+strconv.FormatBool(c.lock))
	args = append(args, "-parallelism="+fmt.Sprint(c.parallelism))
	args = append(args, "-refresh="+strconv.FormatBool(c.refresh))

	if c.refreshOnly {
		err := tf.compatible(ctx, tf0_15_4, nil)
		if err != nil {
			return nil, fmt.Errorf("refresh-only option was introduced in Terraform 0.15.4: %w", err)
		}
		if !c.refresh {
			return nil, fmt.Errorf("you cannot use refresh=false in refresh-only planning mode")
		}
		args = append(args, "-refresh-only")
	}

	// string slice opts: split into separate args
	if c.replaceAddrs != nil {
		err := tf.compatible(ctx, tf0_15_2, nil)
		if err != nil {
			return nil, fmt.Errorf("replace option was introduced in Terraform 0.15.2: %w", err)
		}
		for _, addr := range c.replaceAddrs {
			args = append(args, "-replace="+addr)
		}
	}
	if c.destroy {
		err := tf.compatible(ctx, tf0_15_2, nil)
		if err != nil {
			return nil, fmt.Errorf("-destroy option was introduced in Terraform 0.15.2: %w", err)
		}
		args = append(args, "-destroy")
	}

	if c.targets != nil {
		for _, ta := range c.targets {
			args = append(args, "-target="+ta)
		}
	}
	if c.vars != nil {
		for _, v := range c.vars {
			args = append(args, "-var", v)
		}
	}

	if c.allowDeferral {
		// Ensure the version is later than 1.9.0
		err := tf.compatible(ctx, tf1_9_0, nil)
		if err != nil {
			return nil, fmt.Errorf("-allow-deferral is an experimental option introduced in Terraform 1.9.0: %w", err)
		}

		// Ensure the version has experiments enabled (alpha or dev builds)
		err = tf.experimentsEnabled(ctx)
		if err != nil {
			return nil, fmt.Errorf("-allow-deferral is only available in experimental Terraform builds: %w", err)
		}

		args = append(args, "-allow-deferral")
	}

	return args, nil
}

func (tf *Terraform) buildApplyCmd(ctx context.Context, c applyConfig, args []string) (*exec.Cmd, error) {
	// string argument: pass if set
	if c.dirOrPlan != "" {
		args = append(args, c.dirOrPlan)
	}

	mergeEnv := map[string]string{}
	if c.reattachInfo != nil {
		reattachStr, err := c.reattachInfo.marshalString()
		if err != nil {
			return nil, err
		}
		mergeEnv[reattachEnvVar] = reattachStr
	}

	return tf.buildTerraformCmd(ctx, mergeEnv, args...), nil
}
