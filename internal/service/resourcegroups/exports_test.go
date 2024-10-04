// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcegroups

// Exports for use in tests only.
var (
	ResourceGroup    = resourceGroup
	ResourceResource = resourceResource

	FindGroupByName          = findGroupByName
	FindResourceByTwoPartKey = findResourceByTwoPartKey
)
