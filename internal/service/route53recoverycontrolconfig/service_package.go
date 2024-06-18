// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	endpoints_sdkv1 "github.com/aws/aws-sdk-go/aws/endpoints"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	route53recoverycontrolconfig_sdkv1 "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, config map[string]any) (*route53recoverycontrolconfig_sdkv1.Route53RecoveryControlConfig, error) {
	sess := config[names.AttrSession].(*session_sdkv1.Session)

	cfg := aws_sdkv1.Config{}

	if endpoint := config[names.AttrEndpoint].(string); endpoint != "" {
		tflog.Debug(ctx, "setting endpoint", map[string]any{
			"tf_aws.endpoint": endpoint,
		})
		cfg.Endpoint = aws_sdkv1.String(endpoint)

		if sess.Config.UseFIPSEndpoint == endpoints_sdkv1.FIPSEndpointStateEnabled {
			tflog.Debug(ctx, "endpoint set, ignoring UseFIPSEndpoint setting")
			cfg.UseFIPSEndpoint = endpoints_sdkv1.FIPSEndpointStateDisabled
		}
	}

	// Force "global" services to correct Regions.
	if config["partition"].(string) == endpoints_sdkv1.AwsPartitionID {
		cfg.Region = aws_sdkv1.String(endpoints_sdkv1.UsWest2RegionID)
	}

	return route53recoverycontrolconfig_sdkv1.New(sess.Copy(&cfg)), nil
}
