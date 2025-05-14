// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

// Exports for use in tests only.
var (
	ResourceDataSet        = resourceDataSet
	ResourceEventAction    = newEventActionResource
	ResourceRevisionAssets = newResourceRevisionAssets

	FindDataSetByID     = findDataSetByID
	FindEventActionByID = findEventActionByID
	FindRevisionByID    = findRevisionByID
)
