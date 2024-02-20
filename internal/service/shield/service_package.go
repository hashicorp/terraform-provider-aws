// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	shield_sdkv2 "github.com/aws/aws-sdk-go-v2/service/shield"
	endpoints_sdkv1 "github.com/aws/aws-sdk-go/aws/endpoints"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(_ context.Context, config map[string]any) (*shield_sdkv2.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws_sdkv2.Config))

	// Force "global" services to correct Regions.
	if config["partition"].(string) == endpoints_sdkv1.AwsPartitionID {
		cfg.Region = endpoints_sdkv1.UsEast1RegionID
	}

	return shield_sdkv2.NewFromConfig(cfg, func(o *shield_sdkv2.Options) {
		if endpoint := config["endpoint"].(string); endpoint != "" {
			o.BaseEndpoint = aws_sdkv2.String(endpoint)
		}
	}), nil
}
