// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*sts.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return sts.NewFromConfig(cfg, func(o *sts.Options) {
		if endpoint := config["endpoint"].(string); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}

		if stsRegion := config["sts_region"].(string); stsRegion != "" {
			o.Region = stsRegion
		}
	}), nil
}
