// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	wafv2_sdkv1 "github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *wafv2_sdkv1.WAFV2) (*wafv2_sdkv1.WAFV2, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		if tfawserr.ErrMessageContains(r.Error, wafv2_sdkv1.ErrCodeWAFInternalErrorException, "Retry your request") {
			r.Retryable = aws_sdkv1.Bool(true)
		}

		if tfawserr.ErrMessageContains(r.Error, wafv2_sdkv1.ErrCodeWAFServiceLinkedRoleErrorException, "Retry") {
			r.Retryable = aws_sdkv1.Bool(true)
		}

		if r.Operation.Name == "CreateIPSet" || r.Operation.Name == "CreateRegexPatternSet" ||
			r.Operation.Name == "CreateRuleGroup" || r.Operation.Name == "CreateWebACL" {
			// WAFv2 supports tag on create which can result in the below error codes according to the documentation
			if tfawserr.ErrMessageContains(r.Error, wafv2_sdkv1.ErrCodeWAFTagOperationException, "Retry your request") {
				r.Retryable = aws_sdkv1.Bool(true)
			}
			if tfawserr.ErrMessageContains(r.Error, wafv2_sdkv1.ErrCodeWAFTagOperationInternalErrorException, "Retry your request") {
				r.Retryable = aws_sdkv1.Bool(true)
			}
		}
	})

	return conn, nil
}
