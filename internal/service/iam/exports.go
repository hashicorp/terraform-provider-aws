// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

// Exports for use in other packages.
var (
	ResourceRole = resourceRole

	DeleteServiceLinkedRole = deleteServiceLinkedRole
	FindRoleByName          = findRoleByName
	ListGroupsForUserPages  = listGroupsForUserPages
	AttachPolicyToUser      = attachPolicyToUser
)
