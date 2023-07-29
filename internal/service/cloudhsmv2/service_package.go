// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudhsmv2

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	cloudhsmv2_sdkv1 "github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *cloudhsmv2_sdkv1.CloudHSMV2) (*cloudhsmv2_sdkv1.CloudHSMV2, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		if tfawserr.ErrMessageContains(r.Error, cloudhsmv2_sdkv1.ErrCodeCloudHsmInternalFailureException, "request was rejected because of an AWS CloudHSM internal failure") {
			r.Retryable = aws_sdkv1.Bool(true)
		}
	})

	return conn, nil
}
