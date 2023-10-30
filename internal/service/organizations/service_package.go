// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	organizations_sdkv1 "github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *organizations_sdkv1.Organizations) (*organizations_sdkv1.Organizations, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		// Retry on the following error:
		// ConcurrentModificationException: AWS Organizations can't complete your request because it conflicts with another attempt to modify the same entity. Try again later.
		if tfawserr.ErrMessageContains(r.Error, organizations_sdkv1.ErrCodeConcurrentModificationException, "Try again later") {
			r.Retryable = aws_sdkv1.Bool(true)
		}
	})

	return conn, nil
}
