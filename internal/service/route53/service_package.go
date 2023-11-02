// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*route53.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return route53.NewFromConfig(cfg, func(o *route53.Options) {
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
		} else {
			switch config["partition"].(string) {
			case names.StandardPartitionID:
				// https://docs.aws.amazon.com/general/latest/gr/r53.html Setting default to us-east-1.
				o.Region = names.USEast1RegionID
			case names.ChinaPartitionID:
				// The AWS Go SDK is missing endpoint information for Route 53 in the AWS China partition.
				// This can likely be removed in the future.
				if aws.ToString(o.BaseEndpoint) == "" {
					o.BaseEndpoint = aws.String("https://api.route53.cn")
				}
				o.Region = names.CNNorthwest1RegionID
			case names.USGovCloudPartitionID:
				o.Region = names.USGovWest1RegionID
			}
		}
	}), nil
}
