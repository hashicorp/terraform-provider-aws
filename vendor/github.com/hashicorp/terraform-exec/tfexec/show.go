package tfexec

import (
	"context"
	"fmt"
	"os/exec"

	tfjson "github.com/hashicorp/terraform-json"
)

type showConfig struct {
	reattachInfo ReattachInfo
}

var defaultShowOptions = showConfig{}

type ShowOption interface {
	configureShow(*showConfig)
}

func (opt *ReattachOption) configureShow(conf *showConfig) {
	conf.reattachInfo = opt.info
}

// Show reads the default state path and outputs the state.
// To read a state or plan file, ShowState or ShowPlan must be used instead.
func (tf *Terraform) Show(ctx context.Context, opts ...ShowOption) (*tfjson.State, error) {
	err := tf.compatible(ctx, tf0_12_0, nil)
	if err != nil {
		return nil, fmt.Errorf("terraform show -json was added in 0.12.0: %w", err)
	}

	c := defaultShowOptions

	for _, o := range opts {
		o.configureShow(&c)
	}

	mergeEnv := map[string]string{}
	if c.reattachInfo != nil {
		reattachStr, err := c.reattachInfo.marshalString()
		if err != nil {
			return nil, err
		}
		mergeEnv[reattachEnvVar] = reattachStr
	}

	showCmd := tf.showCmd(ctx, mergeEnv)

	var ret tfjson.State
	err = tf.runTerraformCmdJSON(showCmd, &ret)
	if err != nil {
		return nil, err
	}

	err = ret.Validate()
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

// ShowStateFile reads a given state file and outputs the state.
func (tf *Terraform) ShowStateFile(ctx context.Context, statePath string, opts ...ShowOption) (*tfjson.State, error) {
	err := tf.compatible(ctx, tf0_12_0, nil)
	if err != nil {
		return nil, fmt.Errorf("terraform show -json was added in 0.12.0: %w", err)
	}

	if statePath == "" {
		return nil, fmt.Errorf("statePath cannot be blank: use Show() if not passing statePath")
	}

	c := defaultShowOptions

	for _, o := range opts {
		o.configureShow(&c)
	}

	mergeEnv := map[string]string{}
	if c.reattachInfo != nil {
		reattachStr, err := c.reattachInfo.marshalString()
		if err != nil {
			return nil, err
		}
		mergeEnv[reattachEnvVar] = reattachStr
	}

	showCmd := tf.showCmd(ctx, mergeEnv, statePath)

	var ret tfjson.State
	err = tf.runTerraformCmdJSON(showCmd, &ret)
	if err != nil {
		return nil, err
	}

	err = ret.Validate()
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

// ShowPlanFile reads a given plan file and outputs the plan.
func (tf *Terraform) ShowPlanFile(ctx context.Context, planPath string, opts ...ShowOption) (*tfjson.Plan, error) {
	err := tf.compatible(ctx, tf0_12_0, nil)
	if err != nil {
		return nil, fmt.Errorf("terraform show -json was added in 0.12.0: %w", err)
	}

	if planPath == "" {
		return nil, fmt.Errorf("planPath cannot be blank: use Show() if not passing planPath")
	}

	c := defaultShowOptions

	for _, o := range opts {
		o.configureShow(&c)
	}

	mergeEnv := map[string]string{}
	if c.reattachInfo != nil {
		reattachStr, err := c.reattachInfo.marshalString()
		if err != nil {
			return nil, err
		}
		mergeEnv[reattachEnvVar] = reattachStr
	}

	showCmd := tf.showCmd(ctx, mergeEnv, planPath)

	var ret tfjson.Plan
	err = tf.runTerraformCmdJSON(showCmd, &ret)
	if err != nil {
		return nil, err
	}

	err = ret.Validate()
	if err != nil {
		return nil, err
	}

	return &ret, nil

}

func (tf *Terraform) showCmd(ctx context.Context, mergeEnv map[string]string, args ...string) *exec.Cmd {
	allArgs := []string{"show", "-json", "-no-color"}
	allArgs = append(allArgs, args...)

	return tf.buildTerraformCmd(ctx, mergeEnv, allArgs...)
}
