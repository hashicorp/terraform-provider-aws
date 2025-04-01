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
	ResourceJob                   = resourceJob

	FindCatalogTableOptimizer  = findCatalogTableOptimizer
	FindClassifierByName       = findClassifierByName
	FindConnectionByTwoPartKey = findConnectionByTwoPartKey
	FindDatabaseByName         = findDatabaseByName
	FindJobByName              = findJobByName
	FindTableByName            = findTableByName
)
