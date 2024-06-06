// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chatbot

import (
	"context"
	"log"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chatbot"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*chatbot.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return chatbot.NewFromConfig(cfg, func(o *chatbot.Options) {
		if endpoint := config[names.AttrEndpoint].(string); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)

			if o.EndpointOptions.UseFIPSEndpoint == aws.FIPSEndpointStateEnabled {
				// The SDK doesn't allow setting a custom non-FIPS endpoint *and* enabling UseFIPSEndpoint.
				// However there are a few cases where this is necessary; some services don't have FIPS endpoints,
				// and for some services (e.g. CloudFront) the SDK generates the wrong fips endpoint.
				// While forcing this to disabled may result in the end-user not using a FIPS endpoint as specified
				// by setting UseFIPSEndpoint=true, the user also explicitly changed the endpoint, so
				// here we need to assume the user knows what they're doing.
				log.Printf("[WARN] UseFIPSEndpoint is enabled but a custom endpoint (%s) is configured, ignoring UseFIPSEndpoint.", endpoint)
				o.EndpointOptions.UseFIPSEndpoint = aws.FIPSEndpointStateDisabled
			}
		} else if config["partition"].(string) == names.StandardPartitionID {
			// Chatbot endpoint is available only in the 4 regions us-east-2, us-west-2, eu-west-1, and ap-southeast-1.
			// If the region from the context is one of those four, then use that region. If not default to us-west-2
			if slices.Contains([]string{names.USEast2RegionID, names.USWest2RegionID, names.EUWest1RegionID, names.APSoutheast1RegionID}, cfg.Region) {
				o.Region = cfg.Region
			} else {
				o.Region = names.USWest2RegionID
			}
		}
	}), nil
}
