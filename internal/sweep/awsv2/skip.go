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

	return false
}
