// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package inspector2

// Exports for use in tests only.
var (
	ResourceConfiguration             = resourceConfiguration
	ResourceDelegatedAdminAccount     = resourceDelegatedAdminAccount
	ResourceFilter                    = newFilterResource
	ResourceMemberAssociation         = resourceMemberAssociation
	ResourceOrganizationConfiguration = resourceOrganizationConfiguration

	FindConfiguration             = findConfiguration
	FindDelegatedAdminAccountByID = findDelegatedAdminAccountByID
	FindFilterByARN               = findFilterByARN
	FindMemberByAccountID         = findMemberByAccountID
	FindOrganizationConfiguration = findOrganizationConfiguration

	EnablerID      = enablerID
	ParseEnablerID = parseEnablerID
)
