// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appautoscaling

import (
	"context"
	"strings"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	applicationautoscaling_sdkv1 "github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *applicationautoscaling_sdkv1.ApplicationAutoScaling) (*applicationautoscaling_sdkv1.ApplicationAutoScaling, error) {
	// Workaround for https://github.com/aws/aws-sdk-go/issues/1472.
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		if !strings.HasPrefix(r.Operation.Name, "Describe") && !strings.HasPrefix(r.Operation.Name, "List") {
			return
		}
		if tfawserr.ErrCodeEquals(r.Error, applicationautoscaling_sdkv1.ErrCodeFailedResourceAccessException) {
			r.Retryable = aws_sdkv1.Bool(true)
		}
	})

	return conn, nil
}
