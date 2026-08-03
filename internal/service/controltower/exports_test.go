// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package controltower

// Exports for use in tests only.
var (
	ResourceControl     = resourceControl
	ResourceLandingZone = resourceLandingZone
	ResourceBaseline    = newResourceBaseline

	FindBaselineByID               = findBaselineByID
	FindEnabledControlByTwoPartKey = findEnabledControlByTwoPartKey
	FindLandingZoneByID            = findLandingZoneByID
)
