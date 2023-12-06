// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sts

import (
	"context"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	sts_sdkv2 "github.com/aws/aws-sdk-go-v2/service/sts"
	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	sts_sdkv1 "github.com/aws/aws-sdk-go/service/sts"
)

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, config map[string]any) (*sts_sdkv1.STS, error) {
	sess := config["session"].(*session_sdkv1.Session)
	cfg := &aws_sdkv1.Config{Endpoint: aws_sdkv1.String(config["endpoint"].(string))}

	if stsRegion := config["sts_region"].(string); stsRegion != "" {
		cfg.Region = aws_sdkv1.String(stsRegion)
	}

	return sts_sdkv1.New(sess.Copy(cfg)), nil
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*sts_sdkv2.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws_sdkv2.Config))

	return sts_sdkv2.NewFromConfig(cfg, func(o *sts_sdkv2.Options) {
		if endpoint := config["endpoint"].(string); endpoint != "" {
			o.BaseEndpoint = aws_sdkv2.String(endpoint)
		} else if stsRegion := config["sts_region"].(string); stsRegion != "" {
			o.Region = stsRegion
		}
	}), nil
}
