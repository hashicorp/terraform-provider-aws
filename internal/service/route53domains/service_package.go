// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53domains

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*route53domains.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return route53domains.NewFromConfig(cfg, func(o *route53domains.Options) {
		if endpoint := config[names.AttrEndpoint].(string); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		} else if config["partition"].(string) == names.StandardPartitionID {
			// Route 53 Domains is only available in AWS Commercial us-east-1 Region.
			o.Region = names.USEast1RegionID
		}
	}), nil
}
