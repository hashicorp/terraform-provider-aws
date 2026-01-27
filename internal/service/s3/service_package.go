// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/middleware"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
)

func (p *servicePackage) withExtraOptions(ctx context.Context, config map[string]any) []func(*s3.Options) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return []func(*s3.Options){
		func(o *s3.Options) {
			// Inject original region for S3-compatible storage (Ceph, MinIO, etc.)
			// This allows non-standard region strings when custom endpoints are configured
			if originalRegion, ok := config["s3_original_region"].(string); ok && originalRegion != "" {
				tflog.Debug(ctx, "Injecting original region for S3-compatible endpoint", map[string]any{
					"original_region": originalRegion,
				})
				o.APIOptions = append(o.APIOptions, func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"InjectOriginalS3Region",
							func(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								ctx = awsmiddleware.SetSigningRegion(ctx, originalRegion)
								return next.HandleFinalize(ctx, in)
							},
						),
						middleware.Before,
					)
				})
			}
		},
		func(o *s3.Options) {
			switch region, s3USEast1RegionalEndpoint := o.Region, config["s3_us_east_1_regional_endpoint"].(string) == "regional"; region {
			case endpoints.UsEast1RegionID:
				if !s3USEast1RegionalEndpoint {
					// Maintain the AWS SDK for Go v1 default of using the global endpoint in us-east-1.
					// See https://github.com/hashicorp/terraform-provider-aws/issues/33028.
					overrideRegion := endpoints.AwsGlobalRegionID
					tflog.Info(ctx, "overriding region", map[string]any{
						"original_region": cfg.Region,
						"override_region": overrideRegion,
					})
					o.Region = overrideRegion
				}
			}
		},
		func(o *s3.Options) {
			o.UsePathStyle = config["s3_use_path_style"].(bool)
		},
		func(o *s3.Options) {
			retryables := []retry.IsErrorRetryable{
				retry.IsErrorRetryableFunc(func(err error) aws.Ternary {
					if tfawserr.ErrMessageContains(err, errCodeOperationAborted, "A conflicting conditional operation is currently in progress against this resource. Please try again.") {
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
