// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

// Exports for use in tests only.
var (
	ResourceCatalogDatabase       = resourceCatalogDatabase
	ResourceCatalogTableOptimizer = newResourceCatalogTableOptimizer
	ResourceJob                   = resourceJob

	FindCatalogTableOptimizer = findCatalogTableOptimizer
	FindDatabaseByName        = findDatabaseByName
	FindJobByName             = findJobByName
)
