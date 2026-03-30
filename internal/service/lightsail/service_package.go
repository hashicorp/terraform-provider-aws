// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
)

func (p *servicePackage) withExtraOptions(ctx context.Context, config map[string]any) []func(*lightsail.Options) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return []func(*lightsail.Options){
		func(o *lightsail.Options) {
			retryables := []retry.IsErrorRetryable{
				retry.IsErrorRetryableFunc(func(err error) aws.Ternary {
					if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Please try again in a few minutes") ||
						strings.Contains(err.Error(), "Please wait for it to complete before trying again") {
						return aws.TrueTernary
					}
					return aws.UnknownTernary
				}),
			}
			// Include go-vcr retryable to prevent generated client retryer from being overridden
			if inContext, ok := conns.FromContext(ctx); ok && inContext.VCREnabled() {
				tflog.Info(ctx, "overriding retry behavior to immediately return VCR errors")
				retryables = append(retryables, vcr.InteractionNotFoundRetryableFunc)
			}

			o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws.RetryerV2), retryables...)
		},
	}
}
