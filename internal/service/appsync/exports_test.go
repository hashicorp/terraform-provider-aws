// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

// Exports for use in tests only.
var (
	ResourceAPICache   = resourceAPICache
	ResourceAPIKey     = resourceAPIKey
	ResourceDataSource = resourceDataSource
	ResourceDomainName = resourceDomainName

	FindAPICacheByID           = findAPICacheByID
	FindAPIKeyByTwoPartKey     = findAPIKeyByTwoPartKey
	FindDataSourceByTwoPartKey = findDataSourceByTwoPartKey
	FindDomainNameByID         = findDomainNameByID
)
