// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	configservice_sdkv1 "github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *configservice_sdkv1.ConfigService) (*configservice_sdkv1.ConfigService, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		// When calling Config Organization Rules API actions immediately
		// after Organization creation, the API can randomly return the
		// OrganizationAccessDeniedException error for a few minutes, even
		// after succeeding a few requests.
		switch r.Operation.Name {
		case "DeleteOrganizationConfigRule", "DescribeOrganizationConfigRules", "DescribeOrganizationConfigRuleStatuses", "PutOrganizationConfigRule":
			if !tfawserr.ErrMessageContains(r.Error, configservice_sdkv1.ErrCodeOrganizationAccessDeniedException, "This action can be only made by AWS Organization's master account.") {
				return
			}

			// We only want to retry briefly as the default max retry count would
			// excessively retry when the error could be legitimate.
			// We currently depend on the DefaultRetryer exponential backoff here.
			// ~10 retries gives a fair backoff of a few seconds.
			if r.RetryCount < 9 {
				r.Retryable = aws_sdkv1.Bool(true)
			} else {
				r.Retryable = aws_sdkv1.Bool(false)
			}
		case "DeleteOrganizationConformancePack", "DescribeOrganizationConformancePacks", "DescribeOrganizationConformancePackStatuses", "PutOrganizationConformancePack":
			if !tfawserr.ErrCodeEquals(r.Error, configservice_sdkv1.ErrCodeOrganizationAccessDeniedException) {
				if r.Operation.Name == "DeleteOrganizationConformancePack" && tfawserr.ErrCodeEquals(r.Error, configservice_sdkv1.ErrCodeResourceInUseException) {
					r.Retryable = aws_sdkv1.Bool(true)
				}
				return
			}

			// We only want to retry briefly as the default max retry count would
			// excessively retry when the error could be legitimate.
			// We currently depend on the DefaultRetryer exponential backoff here.
			// ~10 retries gives a fair backoff of a few seconds.
			if r.RetryCount < 9 {
				r.Retryable = aws_sdkv1.Bool(true)
			} else {
				r.Retryable = aws_sdkv1.Bool(false)
			}
		}
	})

	return conn, nil
}
