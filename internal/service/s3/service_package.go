// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func (p *servicePackage) withExtraOptions(ctx context.Context, config map[string]any) []func(*s3.Options) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return []func(*s3.Options){
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
			o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws.RetryerV2), retry.IsErrorRetryableFunc(func(err error) aws.Ternary {
				if tfawserr.ErrMessageContains(err, errCodeOperationAborted, "A conflicting conditional operation is currently in progress against this resource. Please try again.") {
					return aws.TrueTernary
				}
				return aws.UnknownTernary // Delegate to configured Retryer.
			}))
		},
	}
}
