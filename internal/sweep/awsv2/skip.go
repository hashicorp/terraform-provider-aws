// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsv2

import (
	"net"

	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// Check sweeper API call error for reasons to skip sweeping
// These include missing API endpoints and unsupported API calls
func SkipSweepError(err error) bool {
	// Ignore missing API endpoints
	if dnsErr, ok := errs.As[*net.DNSError](err); ok {
		return dnsErr.IsNotFound
	}
	return false
}
