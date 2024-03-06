// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

// Exports for use in tests only.
var (
	ResourceServerlessCache = newServerlessCacheResource
	ResourceSubnetGroup     = resourceSubnetGroup

	FindCacheSubnetGroupByName = findCacheSubnetGroupByName
)
