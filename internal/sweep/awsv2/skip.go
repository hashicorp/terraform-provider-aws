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
	// Example: InvalidAction: InvalidAction: Operation (ListPlatformApplications) is not supported in this region
	if tfawserr.ErrMessageContains(err, "InvalidAction", "is not supported") {
		return true
	}
	// Example (GovCloud): AccessGrantsInstanceNotExistsError: Access Grants Instance does not exist
	if tfawserr.ErrCodeEquals(err, "AccessGrantsInstanceNotExistsError") {
		return true
	}
	// Example (GovCloud): AccessDeniedException: Unable to determine service/operation name to be authorized
	if tfawserr.ErrMessageContains(err, "AccessDeniedException", "Unable to determine service/operation name to be authorized") {
		return true
	}
	// Example (ssmcontacts): ValidationException: Invalid value provided - Account not found for the request
	if tfawserr.ErrMessageContains(err, "ValidationException", "Account not found for the request") {
		return true
	}
	// Example (shield): ResourceNotFoundException: The subscription does not exist
	if tfawserr.ErrMessageContains(err, "ResourceNotFoundException", "The subscription does not exist") {
		return true
	}
	// Example (GovCloud): InvalidParameterValueException: Access Denied to API Version: CORNERSTONE_V1
	if tfawserr.ErrMessageContains(err, "InvalidParameterValueException", "Access Denied to API Version") {
		return true
	}
	// Example (GovCloud): UnknownOperationException: Operation is disabled in this region
	if tfawserr.ErrMessageContains(err, "UnknownOperationException", "Operation is disabled in this region") {
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
	//  Example (ec2): UnsupportedOperation: The functionality you requested is not available in this region
	if tfawserr.ErrMessageContains(err, "UnsupportedOperation", "The functionality you requested is not available in this region") {
		return true
	}

	return false
}
