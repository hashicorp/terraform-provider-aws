// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

// exports used for testing only.
var (
	ResourceDataCellsFilter = newResourceDataCellsFilter
	ResourceResourceLFTag   = newResourceResourceLFTag
	ResourceOptIn           = newResourceOptIn

	FindDataCellsFilterByID = findDataCellsFilterByID
	FindResourceLFTagByID   = findResourceLFTagByID
	LFTagParseResourceID    = lfTagParseResourceID
	FindOptInByID           = findOptInByID

	ValidPrincipal = validPrincipal
)
