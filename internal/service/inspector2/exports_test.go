// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2

// Exports for use in tests only.
var (
	ResourceDelegatedAdminAccount     = resourceDelegatedAdminAccount
	ResourceFilter                    = newFilterResource
	ResourceMemberAssociation         = resourceMemberAssociation
	ResourceOrganizationConfiguration = resourceOrganizationConfiguration

	FindDelegatedAdminAccountByID = findDelegatedAdminAccountByID
	FindFilterByARN               = findFilterByARN
	FindMemberByAccountID         = findMemberByAccountID
	FindOrganizationConfiguration = findOrganizationConfiguration

	EnablerID      = enablerID
	ParseEnablerID = parseEnablerID
)
