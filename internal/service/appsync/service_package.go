// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
)

func (p *servicePackage) withExtraOptions(ctx context.Context, config map[string]any) []func(*appsync.Options) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return []func(*appsync.Options){
		func(o *appsync.Options) {
			retryables := []retry.IsErrorRetryable{
				retry.IsErrorRetryableFunc(func(err error) aws.Ternary {
					if errs.IsAErrorMessageContains[*awstypes.ConcurrentModificationException](err, "a GraphQL API creation is already in progress") {
						return aws.TrueTernary
					}
					return aws.UnknownTernary // Delegate to configured Retryer.
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
