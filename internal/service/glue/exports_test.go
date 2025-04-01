// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

// Exports for use in tests only.
var (
	ResourceCatalogDatabase       = resourceCatalogDatabase
	ResourceCatalogTable          = resourceCatalogTable
	ResourceCatalogTableOptimizer = newResourceCatalogTableOptimizer
	ResourceClassifier            = resourceClassifier
	ResourceConnection            = resourceConnection
	ResourceCrawler               = resourceCrawler
	ResourceJob                   = resourceJob

	FindCatalogTableOptimizer  = findCatalogTableOptimizer
	FindClassifierByName       = findClassifierByName
	FindConnectionByTwoPartKey = findConnectionByTwoPartKey
	FindCrawlerByName          = findCrawlerByName
	FindDatabaseByName         = findDatabaseByName
	FindJobByName              = findJobByName
	FindTableByName            = findTableByName
)
