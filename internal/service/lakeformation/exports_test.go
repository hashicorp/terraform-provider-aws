// Copyright (c) HashiCorp, Inc.
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

	FilterPermissions = filterPermissions

	IncludePrincipalIdentifierInList = includePrincipalIdentifierInList

	FilterCatalogPermissions = filterCatalogPermissions
	FilterDataCellsFilter    = filterDataCellsFilter
)
