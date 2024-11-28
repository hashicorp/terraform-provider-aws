// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	r53rcc "github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*r53rcc.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return r53rcc.NewFromConfig(cfg,
		r53rcc.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *r53rcc.Options) {
			// Always override the service region
			switch config["partition"].(string) {
			case endpoints.AwsPartitionID:
				// https://docs.aws.amazon.com/general/latest/gr/r53arc.html Setting default to us-west-2.
				if cfg.Region != endpoints.UsWest2RegionID {
					tflog.Info(ctx, "overriding region", map[string]any{
						"original_region": cfg.Region,
						"override_region": endpoints.UsWest2RegionID,
					})
				}
				o.Region = endpoints.UsWest2RegionID
			}
		},
	), nil
}
