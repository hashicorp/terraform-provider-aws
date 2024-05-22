// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	retry_sdkv2 "github.com/aws/aws-sdk-go-v2/aws/retry"
	ec2_sdkv2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	ec2_sdkv1 "github.com/aws/aws-sdk-go/service/ec2"
	tfawserr_sdkv1 "github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	tfawserr_sdkv2 "github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, config map[string]any) (*ec2_sdkv1.EC2, error) {
	sess := config[names.AttrSession].(*session_sdkv1.Session)

	return ec2_sdkv1.New(sess.Copy(&aws_sdkv1.Config{Endpoint: aws_sdkv1.String(config[names.AttrEndpoint].(string))})), nil
}

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *ec2_sdkv1.EC2) (*ec2_sdkv1.EC2, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		switch err := r.Error; r.Operation.Name {
		case "RunInstances":
			// `InsufficientInstanceCapacity` error has status code 500 and AWS SDK try retry this error by default.
			if tfawserr_sdkv1.ErrCodeEquals(err, errCodeInsufficientInstanceCapacity) {
				r.Retryable = aws_sdkv1.Bool(false)
			}
		}
	})

	return conn, nil
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*ec2_sdkv2.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws_sdkv2.Config))

	return ec2_sdkv2.NewFromConfig(cfg, func(o *ec2_sdkv2.Options) {
		if endpoint := config[names.AttrEndpoint].(string); endpoint != "" {
			o.BaseEndpoint = aws_sdkv2.String(endpoint)
		}

		o.Retryer = conns.AddIsErrorRetryables(cfg.Retryer().(aws_sdkv2.RetryerV2), retry_sdkv2.IsErrorRetryableFunc(func(err error) aws_sdkv2.Ternary {
			if tfawserr_sdkv2.ErrMessageContains(err, errCodeInvalidParameterValue, "This call cannot be completed because there are pending VPNs or Virtual Interfaces") { // AttachVpnGateway, DetachVpnGateway
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
	}), nil
}
