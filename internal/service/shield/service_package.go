// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(_ context.Context, config map[string]any) (*shield.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	// Force "global" services to correct Regions.
	if config["partition"].(string) == names.StandardPartitionID {
		cfg.Region = names.USEast1RegionID
	}

	return shield.NewFromConfig(cfg, func(o *shield.Options) {
		if endpoint := config[names.AttrEndpoint].(string); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	}), nil
}
