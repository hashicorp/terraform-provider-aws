// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package costoptimizationhub

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costoptimizationhub"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*costoptimizationhub.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return costoptimizationhub.NewFromConfig(cfg,
		costoptimizationhub.WithEndpointResolverV2(newEndpointResolverSDKv2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *costoptimizationhub.Options) {
			if config["partition"].(string) == names.StandardPartitionID {
				// Cost Optimization Hub endpoint is available only in us-east-1 Region.
				if cfg.Region != names.USEast1RegionID {
					tflog.Info(ctx, "overriding region", map[string]any{
						"original_region": cfg.Region,
						"override_region": names.USEast1RegionID,
					})
					o.Region = names.USEast1RegionID
				}
			}
		},
	), nil
}
