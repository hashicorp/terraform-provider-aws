// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

// Exports for use in tests only.
var (
	ResourceCatalogDatabase       = resourceCatalogDatabase
	ResourceCatalogTable          = resourceCatalogTable
	ResourceCatalogTableOptimizer = newResourceCatalogTableOptimizer
	ResourceClassifier            = resourceClassifier
	ResourceJob                   = resourceJob

	FindCatalogTableOptimizer = findCatalogTableOptimizer
	FindClassifierByName      = findClassifierByName
	FindDatabaseByName        = findDatabaseByName
	FindJobByName             = findJobByName
	FindTableByName           = findTableByName
)
