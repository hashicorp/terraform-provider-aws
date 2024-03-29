// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package controltower

// Exports for use in tests only.
var (
	ResourceControl     = resourceControl
	ResourceLandingZone = resourceLandingZone

	FindEnabledControlByTwoPartKey = findEnabledControlByTwoPartKey
	FindLandingZoneByID            = findLandingZoneByID
)
