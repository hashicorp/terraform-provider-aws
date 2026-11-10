// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

// Exports for use in tests only.
var (
	ResourceAccessGrant        = newAccessGrantResource
	ResourceAccessProfile      = newAccessProfileResource
	ResourceDashboard          = newDashboardResource
	ResourceDomain             = newDomainResource
	ResourceOrganizationDomain = newOrganizationDomainResource
	ResourceSpace              = newSpaceResource
	ResourceView               = newViewResource

	FindAccessGrantByID           = findAccessGrantByID
	FindAccessProfileByTwoPartKey = findAccessProfileByTwoPartKey
	FindDashboardByTwoPartKey     = findDashboardByTwoPartKey
	FindDomainByID                = findDomainByID
	FindOrganizationDomainByID    = findOrganizationDomainByID
	FindSpaceByID                 = findSpaceByID
	FindViewByName                = findViewByName
)
