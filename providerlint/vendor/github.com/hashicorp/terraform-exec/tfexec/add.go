package tfexec

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type addConfig struct {
	fromState       bool
	out             string
	includeOptional bool
	provider        string
	reattachInfo    ReattachInfo
}

var defaultAddOptions = addConfig{}

type AddOption interface {
	configureAdd(*addConfig)
}

func (opt *FromStateOption) configureAdd(conf *addConfig) {
	conf.fromState = opt.fromState
}

func (opt *OutOption) configureAdd(conf *addConfig) {
	conf.out = opt.path
}

func (opt *IncludeOptionalOption) configureAdd(conf *addConfig) {
	conf.includeOptional = opt.includeOptional
}

func (opt *ProviderOption) configureAdd(conf *addConfig) {
	conf.provider = opt.provider
}

func (opt *ReattachOption) configureAdd(conf *addConfig) {
	conf.reattachInfo = opt.info
}

// Add represents the `terraform add` subcommand (added in 1.1.0).
//
// Note that this function signature and behaviour is subject
// to breaking changes including removal of that function
// until final 1.1.0 Terraform version (with this command) is released.
func (tf *Terraform) Add(ctx context.Context, address string, opts ...AddOption) (string, error) {
	cmd, err := tf.addCmd(ctx, address, opts...)
	if err != nil {
		return "", err
	}

	var outBuf strings.Builder
	cmd.Stdout = mergeWriters(cmd.Stdout, &outBuf)

	if err := tf.runTerraformCmd(ctx, cmd); err != nil {
		return "", err
	}

	return outBuf.String(), nil
}

func (tf *Terraform) addCmd(ctx context.Context, address string, opts ...AddOption) (*exec.Cmd, error) {
	err := tf.compatible(ctx, tf1_1_0, nil)
	if err != nil {
		return nil, fmt.Errorf("terraform add was added in 1.1.0: %w", err)
	}

	c := defaultAddOptions

	for _, o := range opts {
		o.configureAdd(&c)
	}

	args := []string{"add"}

	args = append(args, "-from-state="+strconv.FormatBool(c.fromState))
	if c.out != "" {
		args = append(args, "-out="+c.out)
	}
	args = append(args, "-optional="+strconv.FormatBool(c.includeOptional))
	if c.provider != "" {
		args = append(args, "-provider="+c.provider)
	}

	args = append(args, address)

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
