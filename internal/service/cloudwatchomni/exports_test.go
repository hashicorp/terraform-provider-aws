// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

// Exports for use in tests only.
var (
	ResourceAccessGrant   = newAccessGrantResource
	ResourceAccessProfile = newAccessProfileResource
	ResourceDomain        = newDomainResource
	ResourceSpace         = newSpaceResource
	ResourceView          = newViewResource

	FindAccessGrantByID           = findAccessGrantByID
	FindAccessProfileByTwoPartKey = findAccessProfileByTwoPartKey
	FindDomainByID                = findDomainByID
	FindSpaceByID                 = findSpaceByID
	FindViewByName                = findViewByName
)
