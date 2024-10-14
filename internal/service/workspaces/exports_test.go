// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

// Exports for use in tests only.
var (
	ResourceConnectionAlias = newConnectionAliasResource
	ResourceDirectory       = resourceDirectory

	FindConnectionAliasByID = findConnectionAliasByID
	FindDirectoryByID       = findDirectoryByID
)
