// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2

// Exports for use in tests only.
var (
	ResourceDelegatedAdminAccount     = resourceDelegatedAdminAccount
	ResourceMemberAssociation         = resourceMemberAssociation
	ResourceOrganizationConfiguration = resourceOrganizationConfiguration
	ResourceFilter                    = newResourceFilter

	FindDelegatedAdminAccountByID = findDelegatedAdminAccountByID
	FindMemberByAccountID         = findMemberByAccountID
	FindOrganizationConfiguration = findOrganizationConfiguration
	FindFilterByARN               = findFilterByARN

	EnablerID      = enablerID
	ParseEnablerID = parseEnablerID
)
