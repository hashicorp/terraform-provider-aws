// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

// exports used for testing only.
var (
	ResourceDataCellsFilter = newDataCellsFilterResource
	ResourceResourceLFTag   = newResourceLFTagResource
	ResourceOptIn           = newOptInResource

	FindDataCellsFilterByID = findDataCellsFilterByID
	FindResourceLFTagByID   = findResourceLFTagByID
	LFTagParseResourceID    = lfTagParseResourceID
	FindOptInByID           = findOptInByID

	ValidPrincipal = validPrincipal
)
