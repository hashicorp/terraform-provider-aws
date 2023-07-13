// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	cloudformation_sdkv1 "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *cloudformation_sdkv1.CloudFormation) (*cloudformation_sdkv1.CloudFormation, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		if tfawserr.ErrMessageContains(r.Error, cloudformation_sdkv1.ErrCodeOperationInProgressException, "Another Operation on StackSet") {
			r.Retryable = aws_sdkv1.Bool(true)
		}
	})

	return conn, nil
}
