// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

// Exports for use in tests only.
var (
	ResourceConnectionAlias = newConnectionAliasResource
	ResourceDirectory       = resourceDirectory
	ResourceIPGroup         = resourceIPGroup
	ResourceWorkspace       = resourceWorkspace

	FindConnectionAliasByID = findConnectionAliasByID
	FindDirectoryByID       = findDirectoryByID
	FindIPGroupByID         = findIPGroupByID
	FindWorkspaceByID       = findWorkspaceByID
)
