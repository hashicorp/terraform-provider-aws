// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*sts.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return sts.NewFromConfig(cfg,
		sts.WithEndpointResolverV2(newEndpointResolverSDKv2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *sts.Options) {
			if stsRegion := config["sts_region"].(string); stsRegion != "" {
				tflog.Info(ctx, "overriding region", map[string]any{
					"original_region": cfg.Region,
					"override_region": stsRegion,
				})
				o.Region = stsRegion
			}
		}), nil
}
