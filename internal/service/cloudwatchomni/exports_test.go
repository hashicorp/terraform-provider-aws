// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

// Exports for use in tests only.
var (
	ResourceAccessGrant = newAccessGrantResource
	ResourceDomain      = newDomainResource
	ResourceSpace       = newSpaceResource

	FindAccessGrantByID = findAccessGrantByID
	FindDomainByID      = findDomainByID
	FindSpaceByID       = findSpaceByID
)
