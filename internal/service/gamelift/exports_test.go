// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package gamelift

// Exports for use in tests only.
var (
	ResourceAlias                    = resourceAlias
	ResourceBuild                    = resourceBuild
	ResourceContainerFleet           = resourceContainerFleet
	ResourceContainerGroupDefinition = resourceContainerGroupDefinition
	ResourceFleet                    = resourceFleet
	ResourceGameServerGroup          = resourceGameServerGroup
	ResourceGameSessionQueue         = resourceGameSessionQueue
	ResourceScript                   = resourceScript

	DiffPortSettings             = diffPortSettings
	FindAliasByID                = findAliasByID
	FindBuildByID                = findBuildByID
	FindContainerFleetByID       = findContainerFleetByID
	FindContainerGroupDefinition = findContainerGroupDefinition
	FindFleetByID                = findFleetByID
	FindGameServerGroupByName    = findGameServerGroupByName
	FindGameSessionQueueByName   = findGameSessionQueueByName
	FindScriptByID               = findScriptByID
)
