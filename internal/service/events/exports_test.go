// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

// Exports for use in tests only.
var (
	ResourceAPIDestination = resourceAPIDestination
	ResourceArchive        = resourceArchive
	ResourceBus            = resourceBus
	ResourceBusPolicy      = resourceBusPolicy
	ResourceConnection     = resourceConnection

	FindAPIDestinationByName = findAPIDestinationByName
	FindArchiveByName        = findArchiveByName
	FindConnectionByName     = findConnectionByName
	FindEventBusByName       = findEventBusByName
	FindEventBusPolicyByName = findEventBusPolicyByName
)
