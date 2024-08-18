// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

// Exports for use in tests only.
var (
	ResourceBotAssociation            = resourceBotAssociation
	ResourceContactFlow               = resourceContactFlow
	ResourceContactFlowModule         = resourceContactFlowModule
	ResourceHoursOfOperation          = resourceHoursOfOperation
	ResourceLambdaFunctionAssociation = resourceLambdaFunctionAssociation
	ResourcePhoneNumber               = resourcePhoneNumber
	ResourceUser                      = resourceUser
	ResourceUserHierarchyGroup        = resourceUserHierarchyGroup
	ResourceVocabulary                = resourceVocabulary

	FindBotAssociationByThreePartKey          = findBotAssociationByThreePartKey
	FindContactFlowByTwoPartKey               = findContactFlowByTwoPartKey
	FindContactFlowModuleByTwoPartKey         = findContactFlowModuleByTwoPartKey
	FindHoursOfOperationByTwoPartKey          = findHoursOfOperationByTwoPartKey
	FindLambdaFunctionAssociationByTwoPartKey = findLambdaFunctionAssociationByTwoPartKey
	FindPhoneNumberByID                       = findPhoneNumberByID
	FindUserByTwoPartKey                      = findUserByTwoPartKey
	FindUserHierarchyGroupByTwoPartKey        = findUserHierarchyGroupByTwoPartKey
	FindVocabularyByTwoPartKey                = findVocabularyByTwoPartKey
)
