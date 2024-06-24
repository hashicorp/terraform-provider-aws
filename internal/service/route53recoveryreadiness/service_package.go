// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoveryreadiness

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, config map[string]any) (*route53recoveryreadiness.Route53RecoveryReadiness, error) {
	sess := config[names.AttrSession].(*session.Session)

	cfg := aws.Config{}

	if endpoint := config[names.AttrEndpoint].(string); endpoint != "" {
		tflog.Debug(ctx, "setting endpoint", map[string]any{
			"tf_aws.endpoint": endpoint,
		})
		cfg.Endpoint = aws.String(endpoint)
	} else {
		cfg.EndpointResolver = newEndpointResolverSDKv1(ctx)
	}

	// Force "global" services to correct Regions.
	if config["partition"].(string) == endpoints.AwsPartitionID {
		if aws.StringValue(cfg.Region) != endpoints.UsWest2RegionID {
			tflog.Info(ctx, "overriding region", map[string]any{
				"original_region": aws.StringValue(cfg.Region),
				"override_region": endpoints.UsWest2RegionID,
			})
			cfg.Region = aws.String(endpoints.UsWest2RegionID)
		}
	}

	return route53recoveryreadiness.New(sess.Copy(&cfg)), nil
}
