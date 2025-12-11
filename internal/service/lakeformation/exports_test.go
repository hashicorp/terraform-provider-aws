// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package lakeformation

// exports used for testing only.
var (
	ResourceDataCellsFilter             = newDataCellsFilterResource
	ResourceLFTagExpression             = newLFTagExpressionResource
	ResourceResourceLFTag               = newResourceLFTagResource
	ResourceOptIn                       = newOptInResource
	ResourceIdentityCenterConfiguration = newResourceIdentityCenterConfiguration

	FindDataCellsFilterByID             = findDataCellsFilterByID
	FindLFTagExpression                 = findLFTagExpression
	LFTagParseResourceID                = lfTagParseResourceID
	FindOptInByID                       = findOptInByID
	FindIdentityCenterConfigurationByID = findIdentityCenterConfigurationByID

	ValidPrincipal = validPrincipal

	IncludePrincipalIdentifierInList = includePrincipalIdentifierInList

	FilterCatalogPermissions          = filterCatalogPermissions
	FilterDataCellsFilter             = filterDataCellsFilter
	FilterDataLocationPermissions     = filterDataLocationPermissions
	FilterDatabasePermissions         = filterDatabasePermissions
	FilterLFTagPermissions            = filterLFTagPermissions
	FilterLFTagPolicyPermissions      = filterLFTagPolicyPermissions
	FilterTablePermissions            = filterTablePermissions
	FilterTableWithColumnsPermissions = filterTableWithColumnsPermissions
)
