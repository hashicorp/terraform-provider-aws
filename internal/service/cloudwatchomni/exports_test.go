// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

// Exports for use in tests only.
var (
	ResourceDashboard = newDashboardResource
	ResourceDomain    = newDomainResource
	ResourceSpace     = newSpaceResource

	FindDashboardByTwoPartKey = findDashboardByTwoPartKey
	FindDomainByID            = findDomainByID
	FindSpaceByID             = findSpaceByID
)
