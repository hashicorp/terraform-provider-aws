// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	chime_sdkv1 "github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *chime_sdkv1.Chime) (*chime_sdkv1.Chime, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		// When calling CreateVoiceConnector across multiple resources,
		// the API can randomly return a BadRequestException without explanation
		if r.Operation.Name == "CreateVoiceConnector" {
			if tfawserr.ErrMessageContains(r.Error, chime_sdkv1.ErrCodeBadRequestException, "Service received a bad request") {
				r.Retryable = aws_sdkv1.Bool(true)
			}
		}
	})

	return conn, nil
}
