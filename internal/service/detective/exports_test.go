// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package detective

// Exports for use in tests only.
var (
	ResourceGraph  = resourceGraph
	ResourceMember = resourceMember

	FindGraphByARN                          = findGraphByARN
	FindMemberByTwoPartKey                  = findMemberByTwoPartKey
	FindOrganizationAdminAccountByAccountID = findOrganizationAdminAccountByAccountID
)
