// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func (p *servicePackage) withExtraOptions(_ context.Context, config map[string]any) []func(*ec2.Options) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return []func(*ec2.Options){
		func(o *ec2.Options) {
			o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws.RetryerV2), retry.IsErrorRetryableFunc(func(err error) aws.Ternary {
				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "This call cannot be completed because there are pending VPNs or Virtual Interfaces") { // AttachVpnGateway, DetachVpnGateway
					return aws.TrueTernary
				}

				if tfawserr.ErrCodeEquals(err, errCodeInsufficientInstanceCapacity) { // CreateCapacityReservation, RunInstances
					return aws.TrueTernary
				}

				if tfawserr.ErrMessageContains(err, errCodeOperationNotPermitted, "Endpoint cannot be created while another endpoint is being created") { // CreateClientVpnEndpoint
					return aws.TrueTernary
				}

				if tfawserr.ErrMessageContains(err, errCodeConcurrentMutationLimitExceeded, "Cannot initiate another change for this endpoint at this time") { // CreateClientVpnRoute, DeleteClientVpnRoute
					return aws.TrueTernary
				}

				if tfawserr.ErrMessageContains(err, errCodeVPNConnectionLimitExceeded, "maximum number of mutating objects has been reached") { // CreateVpnConnection
					return aws.TrueTernary
				}

				if tfawserr.ErrMessageContains(err, errCodeVPNGatewayLimitExceeded, "maximum number of mutating objects has been reached") { // CreateVpnGateway
					return aws.TrueTernary
				}

				return aws.UnknownTernary // Delegate to configured Retryer.
			}))
		},
	}
}
