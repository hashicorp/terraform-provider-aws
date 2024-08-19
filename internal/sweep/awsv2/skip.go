// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsv2

import (
	"net"

	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// Check sweeper API call error for reasons to skip sweeping
// These include missing API endpoints and unsupported API calls
func SkipSweepError(err error) bool {
	// Ignore missing API endpoints
	if dnsErr, ok := errs.As[*net.DNSError](err); ok {
		return dnsErr.IsNotFound
	}

	// Example (GovCloud): AccessDeniedException: Feature is not accessible
	if tfawserr.ErrMessageContains(err, "AccessDeniedException", "Feature is not accessible") {
		return true
	}
	// Example (GovCloud): AccessDeniedException: Unable to determine service/operation name to be authorized
	if tfawserr.ErrMessageContains(err, "AccessDeniedException", "Unable to determine service/operation name to be authorized") {
		return true
	}
	// Example (GovCloud): AccessGrantsInstanceNotExistsError: Access Grants Instance does not exist
	if tfawserr.ErrCodeEquals(err, "AccessGrantsInstanceNotExistsError") {
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
	// Example (lightsail): InvalidInputException: Distribution-related APIs are only available in the us-east-1 Region
	if tfawserr.ErrMessageContains(err, "InvalidInputException", "Distribution-related APIs are only available in the us-east-1 Region") {
		return true
	}
	// Example (lightsail): InvalidInputException: Domain-related APIs are only available in the us-east-1 Region
	if tfawserr.ErrMessageContains(err, "InvalidInputException", "Domain-related APIs are only available in the us-east-1 Region") {
		return true
	}
	// For example from us-west-2 Route53 key signing key
	if tfawserr.ErrMessageContains(err, "InvalidKeySigningKeyStatus", "cannot be deleted because") {
		return true
	}
	// InvalidParameterValue: Access Denied to API Version: APIGlobalDatabases
	if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Access Denied to API Version") {
		return true
	}
	// Ignore more unsupported API calls
	// InvalidParameterValue: Use of cache security groups is not permitted in this API version for your account.
	if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "not permitted in this API version for your account") {
		return true
	}
	// Example (GovCloud): InvalidParameterValueException: Access Denied to API Version: CORNERSTONE_V1
	if tfawserr.ErrMessageContains(err, "InvalidParameterValueException", "Access Denied to API Version") {
		return true
	}
	// For example from us-west-2 Route53 zone
	if tfawserr.ErrMessageContains(err, "KeySigningKeyInParentDSRecord", "Due to DNS lookup failure") {
		return true
	}
	// Example (shield): ResourceNotFoundException: The subscription does not exist
	if tfawserr.ErrMessageContains(err, "ResourceNotFoundException", "The subscription does not exist") {
		return true
	}
	// For example from us-gov-east-1 IoT domain configuration
	if tfawserr.ErrMessageContains(err, "UnauthorizedException", "API is not available in") {
		return true
	}
	// Example (GovCloud): UnknownOperationException: Operation is disabled in this region
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
	//  Example (ec2): UnsupportedOperation: The functionality you requested is not available in this region
	if tfawserr.ErrMessageContains(err, "UnsupportedOperation", "The functionality you requested is not available in this region") {
		return true
	}
	// For example from us-west-1 EMR studio
	if tfawserr.ErrMessageContains(err, "ValidationException", "Account is not whitelisted to use this feature") {
		return true
	}
	// Example (ssmcontacts): ValidationException: Invalid value provided - Account not found for the request
	if tfawserr.ErrMessageContains(err, "ValidationException", "Account not found for the request") {
		return true
	}
	// For example from us-west-2 SageMaker device fleet
	if tfawserr.ErrMessageContains(err, "ValidationException", "We are retiring Amazon Sagemaker Edge") {
		return true
	}

	return false
}
