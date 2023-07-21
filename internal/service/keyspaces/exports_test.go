// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package keyspaces

// Exports for use in tests only.
var (
	ResourceKeyspace = resourceKeyspace
	ResourceTable    = resourceTable

	FindKeyspaceByName    = findKeyspaceByName
	FindTableByTwoPartKey = findTableByTwoPartKey

	TableParseResourceID = tableParseResourceID
)
