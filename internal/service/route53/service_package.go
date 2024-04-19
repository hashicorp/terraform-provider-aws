// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	endpoints_sdkv1 "github.com/aws/aws-sdk-go/aws/endpoints"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	route53_sdkv1 "github.com/aws/aws-sdk-go/service/route53"
)

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, m map[string]any) (*route53_sdkv1.Route53, error) {
	sess := m["session"].(*session_sdkv1.Session)
	config := &aws_sdkv1.Config{Endpoint: aws_sdkv1.String(m["endpoint"].(string))}

	// Force "global" services to correct Regions.
	switch m["partition"].(string) {
	case endpoints_sdkv1.AwsPartitionID:
		config.Region = aws_sdkv1.String(endpoints_sdkv1.UsWest2RegionID)
	case endpoints_sdkv1.AwsCnPartitionID:
		// The AWS Go SDK is missing endpoint information for Route 53 in the AWS China partition.
		// This can likely be removed in the future.
		if aws_sdkv1.StringValue(config.Endpoint) == "" {
			config.Endpoint = aws_sdkv1.String("https://api.route53.cn")
		}
		config.Region = aws_sdkv1.String(endpoints_sdkv1.CnNorthwest1RegionID)
	case endpoints_sdkv1.AwsUsGovPartitionID:
		config.Region = aws_sdkv1.String(endpoints_sdkv1.UsGovWest1RegionID)
	}

	return route53_sdkv1.New(sess.Copy(config)), nil
}
