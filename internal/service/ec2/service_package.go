// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	retry_sdkv2 "github.com/aws/aws-sdk-go-v2/aws/retry"
	ec2_sdkv2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	tfawserr_sdkv2 "github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*ec2_sdkv2.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws_sdkv2.Config))

	return ec2_sdkv2.NewFromConfig(cfg,
		ec2_sdkv2.WithEndpointResolverV2(newEndpointResolverSDKv2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *ec2_sdkv2.Options) {
			o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws_sdkv2.RetryerV2), retry_sdkv2.IsErrorRetryableFunc(func(err error) aws_sdkv2.Ternary {
				if tfawserr_sdkv2.ErrMessageContains(err, errCodeInvalidParameterValue, "This call cannot be completed because there are pending VPNs or Virtual Interfaces") { // AttachVpnGateway, DetachVpnGateway
					return aws_sdkv2.TrueTernary
				}

				if tfawserr_sdkv2.ErrCodeEquals(err, errCodeInsufficientInstanceCapacity) { // CreateCapacityReservation, RunInstances
					return aws_sdkv2.TrueTernary
				}

				if tfawserr_sdkv2.ErrMessageContains(err, errCodeOperationNotPermitted, "Endpoint cannot be created while another endpoint is being created") { // CreateClientVpnEndpoint
					return aws_sdkv2.TrueTernary
				}

				if tfawserr_sdkv2.ErrMessageContains(err, errCodeConcurrentMutationLimitExceeded, "Cannot initiate another change for this endpoint at this time") { // CreateClientVpnRoute, DeleteClientVpnRoute
					return aws_sdkv2.TrueTernary
				}

				if tfawserr_sdkv2.ErrMessageContains(err, errCodeVPNConnectionLimitExceeded, "maximum number of mutating objects has been reached") { // CreateVpnConnection
					return aws_sdkv2.TrueTernary
				}

				if tfawserr_sdkv2.ErrMessageContains(err, errCodeVPNGatewayLimitExceeded, "maximum number of mutating objects has been reached") { // CreateVpnGateway
					return aws_sdkv2.TrueTernary
				}

				return aws_sdkv2.UnknownTernary // Delegate to configured Retryer.
			}))
		},
	), nil
}
