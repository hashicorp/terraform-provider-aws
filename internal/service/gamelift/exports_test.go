// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

// Exports for use in tests only.
var (
	ResourceAlias            = resourceAlias
	ResourceBuild            = resourceBuild
	ResourceFleet            = resourceFleet
	ResourceGameServerGroup  = resourceGameServerGroup
	ResourceGameSessionQueue = resourceGameSessionQueue
	ResourceScript           = resourceScript

	DiffPortSettings           = diffPortSettings
	FindAliasByID              = findAliasByID
	FindBuildByID              = findBuildByID
	FindFleetByID              = findFleetByID
	FindGameServerGroupByName  = findGameServerGroupByName
	FindGameSessionQueueByName = findGameSessionQueueByName
	FindScriptByID             = findScriptByID
)
