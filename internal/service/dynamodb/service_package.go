// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	dynamodb_sdkv1 "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *dynamodb_sdkv1.DynamoDB) (*dynamodb_sdkv1.DynamoDB, error) {
	// See https://github.com/aws/aws-sdk-go/pull/1276.
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		if r.Operation.Name != "PutItem" && r.Operation.Name != "UpdateItem" && r.Operation.Name != "DeleteItem" {
			return
		}
		if tfawserr.ErrMessageContains(r.Error, dynamodb_sdkv1.ErrCodeLimitExceededException, "Subscriber limit exceeded:") {
			r.Retryable = aws_sdkv1.Bool(true)
		}
	})

	return conn, nil
}
