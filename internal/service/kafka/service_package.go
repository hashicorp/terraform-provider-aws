// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	kafka_sdkv1 "github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *kafka_sdkv1.Kafka) (*kafka_sdkv1.Kafka, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		if tfawserr.ErrMessageContains(r.Error, kafka_sdkv1.ErrCodeTooManyRequestsException, "Too Many Requests") {
			r.Retryable = aws_sdkv1.Bool(true)
		}
	})

	return conn, nil
}
