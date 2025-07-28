// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

// Exports for use in tests only.
var (
	ResourceGroup           = resourceGroup
	ResourceGroupMembership = resourceGroupMembership
	ResourceUser            = resourceUser

	FindGroupByTwoPartKey           = findGroupByTwoPartKey
	FindGroupMembershipByTwoPartKey = findGroupMembershipByTwoPartKey
	FindUserByTwoPartKey            = findUserByTwoPartKey
)
