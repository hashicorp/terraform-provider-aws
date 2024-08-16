// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

// Exports for use in tests only.
var (
	ResourceBotAssociation    = resourceBotAssociation
	ResourceContactFlow       = resourceContactFlow
	ResourceContactFlowModule = resourceContactFlowModule
	ResourceHoursOfOperation  = resourceHoursOfOperation

	FindBotAssociationByThreePartKey  = findBotAssociationByThreePartKey
	FindContactFlowByTwoPartKey       = findContactFlowByTwoPartKey
	FindContactFlowModuleByTwoPartKey = findContactFlowModuleByTwoPartKey
	FindHoursOfOperationByTwoPartKey  = findHoursOfOperationByTwoPartKey
)
