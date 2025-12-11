// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package dataexchange

// Exports for use in tests only.
var (
	ResourceDataSet        = resourceDataSet
	ResourceEventAction    = newEventActionResource
	ResourceRevision       = resourceRevision
	ResourceRevisionAssets = newRevisionAssetsResource

	FindDataSetByID     = findDataSetByID
	FindEventActionByID = findEventActionByID
	FindRevisionByID    = findRevisionByID
)
