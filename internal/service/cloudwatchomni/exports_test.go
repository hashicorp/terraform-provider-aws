// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

// Exports for use in tests only.
var (
	ResourceAccessGrant   = newAccessGrantResource
	ResourceAccessProfile = newAccessProfileResource
	ResourceDashboard     = newDashboardResource
	ResourceDomain        = newDomainResource
	ResourceSpace         = newSpaceResource
	ResourceView          = newViewResource

	FindAccessGrantByID           = findAccessGrantByID
	FindAccessProfileByTwoPartKey = findAccessProfileByTwoPartKey
	FindDashboardByTwoPartKey     = findDashboardByTwoPartKey
	FindDomainByID                = findDomainByID
	FindSpaceByID                 = findSpaceByID
	FindViewByName                = findViewByName
)
