// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/globalaccelerator"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewConn returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*globalaccelerator.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return globalaccelerator.NewFromConfig(cfg, func(o *globalaccelerator.Options) {
		if endpoint := config[names.AttrEndpoint].(string); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		} else if config["partition"].(string) == names.StandardPartitionID {
			// Global Accelerator endpoint is only available in AWS Commercial us-west-2 Region.
			o.Region = names.USWest2RegionID
		}
	}), nil
}
