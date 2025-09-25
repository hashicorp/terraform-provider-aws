// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

// exports used for testing only.
var (
	ResourceDataCellsFilter = newDataCellsFilterResource
	ResourceLFTagExpression = newLFTagExpressionResource
	ResourceResourceLFTag   = newResourceLFTagResource
	ResourceOptIn           = newOptInResource

	FindDataCellsFilterByID = findDataCellsFilterByID
	FindLFTagExpression     = findLFTagExpression
	LFTagParseResourceID    = lfTagParseResourceID
	FindOptInByID           = findOptInByID

	ValidPrincipal = validPrincipal
)
