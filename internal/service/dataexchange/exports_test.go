// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

// Exports for use in tests only.
var (
	ResourceEventAction    = newEventActionResource
	ResourceRevisionAssets = newResourceRevisionAssets

	FindEventActionByID = findEventActionByID
	FindRevisionByID    = findRevisionByID
)
