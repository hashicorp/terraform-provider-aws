// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
)

func (p *servicePackage) withExtraOptions(ctx context.Context, config map[string]any) []func(*cognitoidentityprovider.Options) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return []func(*cognitoidentityprovider.Options){
		func(o *cognitoidentityprovider.Options) {
			retryables := []retry.IsErrorRetryable{
				retry.IsErrorRetryableFunc(func(err error) aws.Ternary {
					// Cognito Identity Provider's LimitExceededException always indicates a hard,
					// per-account or per-user-pool quota (e.g. groups per user pool, user pool
					// clients per user pool) that will never succeed on retry with the same
					// input. It is distinct from, and should not be confused with, the service's
					// dedicated throttling error, TooManyRequestsException, which is left to the
					// configured Retryer. Without this override, the generated client's default
					// retryer treats LimitExceededException as retryable/throttling (its modeled
					// error code, in the absence of an explicit override, falls back to the
					// exception's shape name, which happens to collide with a fixed list of
					// generically-throttling error codes), causing resource creation to retry with
					// exponential backoff for many minutes before finally surfacing the
					// non-retryable quota error.
					if errs.IsA[*awstypes.LimitExceededException](err) {
						return aws.FalseTernary
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
