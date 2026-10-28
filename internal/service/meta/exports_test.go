// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package meta

// Exports for use in tests only.
var (
	FindRegionByEC2Endpoint = findRegionByEC2Endpoint
	FindRegionByName        = findRegionByName
)

func RegionInfoID(r *regionInfo) string          { return r.id }
func RegionInfoDescription(r *regionInfo) string { return r.description }
func RegionInfoInferred(r *regionInfo) bool      { return r.inferred }
