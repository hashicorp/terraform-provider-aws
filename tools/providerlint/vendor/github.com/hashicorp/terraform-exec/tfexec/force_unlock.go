package tfexec

import (
	"context"
	"os/exec"
)

type forceUnlockConfig struct {
	dir string
}

var defaultForceUnlockOptions = forceUnlockConfig{}

type ForceUnlockOption interface {
	configureForceUnlock(*forceUnlockConfig)
}

func (opt *DirOption) configureForceUnlock(conf *forceUnlockConfig) {
	conf.dir = opt.path
}

// ForceUnlock represents the `terraform force-unlock` command
func (tf *Terraform) ForceUnlock(ctx context.Context, lockID string, opts ...ForceUnlockOption) error {
	unlockCmd := tf.forceUnlockCmd(ctx, lockID, opts...)

	if err := tf.runTerraformCmd(ctx, unlockCmd); err != nil {
		return err
	}

	return nil
}

func (tf *Terraform) forceUnlockCmd(ctx context.Context, lockID string, opts ...ForceUnlockOption) *exec.Cmd {
	c := defaultForceUnlockOptions

	for _, o := range opts {
		o.configureForceUnlock(&c)
	}
	args := []string{"force-unlock", "-force"}

	// positional arguments
	args = append(args, lockID)

	// optional positional arguments
	if c.dir != "" {
		args = append(args, c.dir)
	}

	return tf.buildTerraformCmd(ctx, nil, args...)
}
