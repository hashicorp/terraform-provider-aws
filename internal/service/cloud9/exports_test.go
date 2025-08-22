// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloud9

// Exports for use in tests only.
var (
	ResourceEnvironmentEC2        = resourceEnvironmentEC2
	ResourceEnvironmentMembership = resourceEnvironmentMembership

	FindEnvironmentByID                   = findEnvironmentByID
	FindEnvironmentMembershipByTwoPartKey = findEnvironmentMembershipByTwoPartKey
)
