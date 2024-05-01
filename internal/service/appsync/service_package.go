// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	appsync_sdkv1 "github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *appsync_sdkv1.AppSync) (*appsync_sdkv1.AppSync, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		if r.Operation.Name == "CreateGraphqlApi" {
			if tfawserr.ErrMessageContains(r.Error, appsync_sdkv1.ErrCodeConcurrentModificationException, "a GraphQL API creation is already in progress") {
				r.Retryable = aws_sdkv1.Bool(true)
			}
		}
	})

	return conn, nil
}
