// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

// Exports for use in tests only.
var (
	ResourceAPICache   = resourceAPICache
	ResourceAPIKey     = resourceAPIKey
	ResourceDataSource = resourceDataSource

	FindAPICacheByID           = findAPICacheByID
	FindAPIKeyByTwoPartKey     = findAPIKeyByTwoPartKey
	FindDataSourceByTwoPartKey = findDataSourceByTwoPartKey
)
