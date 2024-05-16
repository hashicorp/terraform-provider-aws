// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	endpoints_sdkv1 "github.com/aws/aws-sdk-go/aws/endpoints"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	route53recoverycontrolconfig_sdkv1 "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
)

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, m map[string]any) (*route53recoverycontrolconfig_sdkv1.Route53RecoveryControlConfig, error) {
	sess := m["session"].(*session_sdkv1.Session)
	config := &aws_sdkv1.Config{Endpoint: aws_sdkv1.String(m["endpoint"].(string))}

	// Force "global" services to correct Regions.
	if m["partition"].(string) == endpoints_sdkv1.AwsPartitionID {
		config.Region = aws_sdkv1.String(endpoints_sdkv1.UsWest2RegionID)
	}

	return route53recoverycontrolconfig_sdkv1.New(sess.Copy(config)), nil
}
