// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena

// Exports for use in tests only.
var (
	FindDataCatalogByName             = findDataCatalogByName
	FindDatabaseByName                = findDatabaseByName
	FindNamedQueryByID                = findNamedQueryByID
	FindPreparedStatementByTwoPartKey = findPreparedStatementByTwoPartKey
	FindWorkGroupByName               = findWorkGroupByName
	QueryExecutionResult              = queryExecutionResult

	ResourceDataCatalog       = resourceDataCatalog
	ResourceDatabase          = resourceDatabase
	ResourceNamedQuery        = resourceNamedQuery
	ResourcePreparedStatement = resourcePreparedStatement
	ResourceWorkGroup         = resourceWorkGroup
)
