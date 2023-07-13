// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sts

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	sts_sdkv1 "github.com/aws/aws-sdk-go/service/sts"
)

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, m map[string]any) (*sts_sdkv1.STS, error) {
	sess := m["session"].(*session_sdkv1.Session)
	config := &aws_sdkv1.Config{Endpoint: aws_sdkv1.String(m["endpoint"].(string))}

	if stsRegion := m["sts_region"].(string); stsRegion != "" {
		config.Region = aws_sdkv1.String(stsRegion)
	}

	return sts_sdkv1.New(sess.Copy(config)), nil
}
