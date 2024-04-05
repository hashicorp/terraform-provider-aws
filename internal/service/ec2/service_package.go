// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	ec2_sdkv1 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *ec2_sdkv1.EC2) (*ec2_sdkv1.EC2, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		switch err := r.Error; r.Operation.Name {
		case "AttachVpnGateway", "DetachVpnGateway":
			if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "This call cannot be completed because there are pending VPNs or Virtual Interfaces") {
				r.Retryable = aws_sdkv1.Bool(true)
			}

		case "CreateClientVpnEndpoint":
			if tfawserr.ErrMessageContains(err, errCodeOperationNotPermitted, "Endpoint cannot be created while another endpoint is being created") {
				r.Retryable = aws_sdkv1.Bool(true)
			}

		case "CreateClientVpnRoute", "DeleteClientVpnRoute":
			if tfawserr.ErrMessageContains(err, errCodeConcurrentMutationLimitExceeded, "Cannot initiate another change for this endpoint at this time") {
				r.Retryable = aws_sdkv1.Bool(true)
			}

		case "CreateVpnConnection":
			if tfawserr.ErrMessageContains(err, errCodeVPNConnectionLimitExceeded, "maximum number of mutating objects has been reached") {
				r.Retryable = aws_sdkv1.Bool(true)
			}

		case "CreateVpnGateway":
			if tfawserr.ErrMessageContains(err, errCodeVPNGatewayLimitExceeded, "maximum number of mutating objects has been reached") {
				r.Retryable = aws_sdkv1.Bool(true)
			}

		case "RunInstances":
			// `InsufficientInstanceCapacity` error has status code 500 and AWS SDK try retry this error by default.
			if tfawserr.ErrCodeEquals(err, errCodeInsufficientInstanceCapacity) {
				r.Retryable = aws_sdkv1.Bool(false)
			}
		}
	})

	return conn, nil
}
