// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acmpca

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (p *servicePackage) withExtraOptions(ctx context.Context, config map[string]any) []func(*acmpca.Options) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	// No one would reasonably set retries to this value manually, so assume it's the provider default
	if cfg.RetryMaxAttempts != 25 {
		return nil
	}

	tflog.Debug(ctx, "Overriding provider default retry attempts with AWS SDK for Go v2 default", map[string]any{
		"retry_max_attempts": retry.DefaultMaxAttempts,
	})

	return []func(*acmpca.Options){
		func(o *acmpca.Options) {
			o.RetryMaxAttempts = retry.DefaultMaxAttempts
		},
	}
}
