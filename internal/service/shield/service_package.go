// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	endpoints_sdkv1 "github.com/aws/aws-sdk-go/aws/endpoints"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	shield_sdkv1 "github.com/aws/aws-sdk-go/service/shield"
)

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, m map[string]any) (*shield_sdkv1.Shield, error) {
	sess := m["session"].(*session_sdkv1.Session)
	config := &aws_sdkv1.Config{Endpoint: aws_sdkv1.String(m["endpoint"].(string))}

	// Force "global" services to correct Regions.
	if m["partition"].(string) == endpoints_sdkv1.AwsPartitionID {
		config.Region = aws_sdkv1.String(endpoints_sdkv1.UsEast1RegionID)
	}

	return shield_sdkv1.New(sess.Copy(config)), nil
}
