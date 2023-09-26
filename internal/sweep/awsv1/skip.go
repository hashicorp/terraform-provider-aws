// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsv1

import (
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// Check sweeper API call error for reasons to skip sweeping
// These include missing API endpoints and unsupported API calls
func SkipSweepError(err error) bool {
	// Ignore missing API endpoints for AWS SDK for Go v1
	if tfawserr.ErrMessageContains(err, "RequestError", "send request failed") {
		return true
	}
	// Ignore unsupported API calls
	if tfawserr.ErrCodeEquals(err, "UnsupportedOperation") {
		return true
	}
	// Ignore more unsupported API calls
	// InvalidParameterValue: Use of cache security groups is not permitted in this API version for your account.
	if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "not permitted in this API version for your account") {
		return true
	}
	// InvalidParameterValue: Access Denied to API Version: APIGlobalDatabases
	if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Access Denied to API Version") {
		return true
	}
	// GovCloud has endpoints that respond with (no message provided):
	// AccessDeniedException:
	// Since acceptance test sweepers are best effort and this response is very common,
	// we allow bypassing this error globally instead of individual test sweeper fixes.
	if tfawserr.ErrCodeEquals(err, "AccessDeniedException") {
		return true
	}
	// Example: BadRequestException: vpc link not supported for region us-gov-west-1
	if tfawserr.ErrMessageContains(err, "BadRequestException", "not supported") {
		return true
	}
	// Example: InvalidAction: InvalidAction: Operation (ListPlatformApplications) is not supported in this region
	if tfawserr.ErrMessageContains(err, "InvalidAction", "is not supported") {
		return true
	}
	// Example: InvalidAction: The action DescribeTransitGatewayAttachments is not valid for this web service
	if tfawserr.ErrMessageContains(err, "InvalidAction", "is not valid") {
		return true
	}
	// For example from GovCloud SES.SetActiveReceiptRuleSet.
	if tfawserr.ErrMessageContains(err, "InvalidAction", "Unavailable Operation") {
		return true
	}
	// For example from us-west-2 Route53 key signing key
	if tfawserr.ErrMessageContains(err, "InvalidKeySigningKeyStatus", "cannot be deleted because") {
		return true
	}
	// For example from us-west-2 Route53 zone
	if tfawserr.ErrMessageContains(err, "KeySigningKeyInParentDSRecord", "Due to DNS lookup failure") {
		return true
	}
	// For example from us-gov-west-1 EventBridge archive
	if tfawserr.ErrMessageContains(err, "UnknownOperationException", "Operation is disabled in this region") {
		return true
	}
	// For example from us-east-1 SageMaker
	if tfawserr.ErrMessageContains(err, "UnknownOperationException", "The requested operation is not supported in the called region") {
		return true
	}
	// For example from us-west-2 ECR public repository
	if tfawserr.ErrMessageContains(err, "UnsupportedCommandException", "command is only supported in") {
		return true
	}
	// For example from us-west-1 EMR studio
	if tfawserr.ErrMessageContains(err, "ValidationException", "Account is not whitelisted to use this feature") {
		return true
	}

	return false
}
