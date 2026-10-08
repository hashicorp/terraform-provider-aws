// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package detective

// Exports for use in tests only.
var (
	ResourceGraph                     = resourceGraph
	ResourceInvitationAccepter        = resourceInvitationAccepter
	ResourceMember                    = resourceMember
	ResourceOrganizationAdminAccount  = resourceOrganizationAdminAccount
	ResourceOrganizationConfiguration = resourceOrganizationConfiguration

	FindGraphByARN                          = findGraphByARN
	FindInvitationByGraphARN                = findInvitationByGraphARN
	FindMemberByTwoPartKey                  = findMemberByTwoPartKey
	FindOrganizationAdminAccountByAccountID = findOrganizationAdminAccountByAccountID
	FindOrganizationConfigurationByGraphARN = findOrganizationConfigurationByGraphARN
)
