// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package athena

// Exports for use in tests only.
var (
	DeleteSystemTables                 = deleteSystemTables
	FindCapacityReservationByName      = findCapacityReservationByName
	FindDataCatalogByName              = findDataCatalogByName
	FindDatabaseByName                 = findDatabaseByName
	FindNamedQueryByID                 = findNamedQueryByID
	FindPreparedStatementByTwoPartKey  = findPreparedStatementByTwoPartKey
	FindSystemTableConfigurationByName = findSystemTableConfigurationByName
	FindSystemTables                   = findSystemTables
	FindWorkGroupByName                = findWorkGroupByName
	QueryExecutionResult               = queryExecutionResult
	ResetSystemTables                  = resetSystemTables

	ResourceCapacityReservation      = newCapacityReservationResource
	ResourceDataCatalog              = resourceDataCatalog
	ResourceDatabase                 = resourceDatabase
	ResourceNamedQuery               = resourceNamedQuery
	ResourcePreparedStatement        = resourcePreparedStatement
	ResourceSystemTableConfiguration = newSystemTableConfigurationResource
	ResourceWorkGroup                = resourceWorkGroup
)
