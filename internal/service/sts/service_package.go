// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (p *servicePackage) withExtraOptions(ctx context.Context, config map[string]any) []func(*sts.Options) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return []func(*sts.Options){
		func(o *sts.Options) {
			if stsRegion := config["sts_region"].(string); stsRegion != "" {
				tflog.Info(ctx, "overriding region", map[string]any{
					"original_region": cfg.Region,
					"override_region": stsRegion,
				})
				o.Region = stsRegion
			}
		},
	}
}
