// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chatbot

import (
	"context"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chatbot"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*chatbot.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return chatbot.NewFromConfig(cfg,
		chatbot.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *chatbot.Options) {
			if config["partition"].(string) == endpoints.AwsPartitionID {
				// Chatbot endpoint is available only in the 4 regions us-east-2, us-west-2, eu-west-1, and ap-southeast-1.
				// If the region from the context is one of those four, then use that region. If not default to us-west-2
				if slices.Contains([]string{endpoints.UsEast2RegionID, endpoints.UsWest2RegionID, endpoints.EuWest1RegionID, endpoints.ApSoutheast1RegionID}, cfg.Region) {
					o.Region = cfg.Region
				} else {
					tflog.Info(ctx, "overriding region", map[string]any{
						"original_region": cfg.Region,
						"override_region": endpoints.UsWest2RegionID,
					})
					o.Region = endpoints.UsWest2RegionID
				}
			}
		},
	), nil
}
