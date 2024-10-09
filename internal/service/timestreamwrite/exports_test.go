// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite

// Exports for use in tests only.
var (
	ResourceDatabase = resourceDatabase
	ResourceTable    = resourceTable

	FindDatabaseByName    = findDatabaseByName
	FindTableByTwoPartKey = findTableByTwoPartKey

	TableParseResourceID = tableParseResourceID
)
