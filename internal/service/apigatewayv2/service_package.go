// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	apigatewayv2_sdkv1 "github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *apigatewayv2_sdkv1.ApiGatewayV2) (*apigatewayv2_sdkv1.ApiGatewayV2, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		// Many operations can return an error such as:
		//   ConflictException: Unable to complete operation due to concurrent modification. Please try again later.
		// Handle them all globally for the service client.
		if tfawserr.ErrMessageContains(r.Error, apigatewayv2_sdkv1.ErrCodeConflictException, "try again later") {
			r.Retryable = aws_sdkv1.Bool(true)
		}
	})

	return conn, nil
}
