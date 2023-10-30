// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	kinesis_sdkv1 "github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *kinesis_sdkv1.Kinesis) (*kinesis_sdkv1.Kinesis, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		if r.Operation.Name == "CreateStream" {
			if tfawserr.ErrMessageContains(r.Error, kinesis_sdkv1.ErrCodeLimitExceededException, "simultaneously be in CREATING or DELETING") {
				r.Retryable = aws_sdkv1.Bool(true)
			}
		}
		if r.Operation.Name == "CreateStream" || r.Operation.Name == "DeleteStream" {
			if tfawserr.ErrMessageContains(r.Error, kinesis_sdkv1.ErrCodeLimitExceededException, "Rate exceeded for stream") {
				r.Retryable = aws_sdkv1.Bool(true)
			}
		}
	})

	return conn, nil
}
