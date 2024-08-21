// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoveryreadiness

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53recoveryreadiness"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*route53recoveryreadiness.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return route53recoveryreadiness.NewFromConfig(cfg,
		route53recoveryreadiness.WithEndpointResolverV2(newEndpointResolverSDKv2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *route53recoveryreadiness.Options) {
			// Always override the service region
			switch config["partition"].(string) {
			case names.StandardPartitionID:
				// https://docs.aws.amazon.com/general/latest/gr/r53arc.html Setting default to us-west-2.
				if cfg.Region != names.USWest2RegionID {
					tflog.Info(ctx, "overriding region", map[string]any{
						"original_region": cfg.Region,
						"override_region": names.USWest2RegionID,
					})
				}
				o.Region = names.USWest2RegionID
			}
		},
	), nil
}
