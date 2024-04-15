// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

// Exports for use in tests only.
var (
	ResourceAPIDestination = resourceAPIDestination
	ResourceArchive        = resourceArchive
	ResourceBus            = resourceBus

	FindAPIDestinationByName = findAPIDestinationByName
	FindArchiveByName        = findArchiveByName
	FindEventBusByName       = findEventBusByName
)
