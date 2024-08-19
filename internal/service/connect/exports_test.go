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
	ResourceSecurityProfile           = resourceSecurityProfile
	ResourceUser                      = resourceUser
	ResourceUserHierarchyGroup        = resourceUserHierarchyGroup
	ResourceUserHierarchyStructure    = resourceUserHierarchyStructure
	ResourceVocabulary                = resourceVocabulary

	FindBotAssociationByThreePartKey          = findBotAssociationByThreePartKey
	FindContactFlowByTwoPartKey               = findContactFlowByTwoPartKey
	FindContactFlowModuleByTwoPartKey         = findContactFlowModuleByTwoPartKey
	FindHoursOfOperationByTwoPartKey          = findHoursOfOperationByTwoPartKey
	FindLambdaFunctionAssociationByTwoPartKey = findLambdaFunctionAssociationByTwoPartKey
	FindPhoneNumberByID                       = findPhoneNumberByID
	FindSecurityProfileByTwoPartKey           = findSecurityProfileByTwoPartKey
	FindUserByTwoPartKey                      = findUserByTwoPartKey
	FindUserHierarchyGroupByTwoPartKey        = findUserHierarchyGroupByTwoPartKey
	FindUserHierarchyStructureByID            = findUserHierarchyStructureByID
	FindVocabularyByTwoPartKey                = findVocabularyByTwoPartKey
)
