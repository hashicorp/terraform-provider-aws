// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

// Exports for use in tests only.
var (
	ResourceBotAssociation            = resourceBotAssociation
	ResourceContactFlow               = resourceContactFlow
	ResourceContactFlowModule         = resourceContactFlowModule
	ResourceHoursOfOperation          = resourceHoursOfOperation
	ResourceInstance                  = resourceInstance
	ResourceInstanceStorageConfig     = resourceInstanceStorageConfig
	ResourceLambdaFunctionAssociation = resourceLambdaFunctionAssociation
	ResourcePhoneNumber               = resourcePhoneNumber
	ResourceQueue                     = resourceQueue
	ResourceQuickConnect              = resourceQuickConnect
	ResourceRoutingProfile            = resourceRoutingProfile
	ResourceSecurityProfile           = resourceSecurityProfile
	ResourceUser                      = resourceUser
	ResourceUserHierarchyGroup        = resourceUserHierarchyGroup
	ResourceUserHierarchyStructure    = resourceUserHierarchyStructure
	ResourceVocabulary                = resourceVocabulary

	FindBotAssociationByThreePartKey          = findBotAssociationByThreePartKey
	FindContactFlowByTwoPartKey               = findContactFlowByTwoPartKey
	FindContactFlowModuleByTwoPartKey         = findContactFlowModuleByTwoPartKey
	FindHoursOfOperationByTwoPartKey          = findHoursOfOperationByTwoPartKey
	FindInstanceByID                          = findInstanceByID
	FindInstanceStorageConfigByThreePartKey   = findInstanceStorageConfigByThreePartKey
	FindLambdaFunctionAssociationByTwoPartKey = findLambdaFunctionAssociationByTwoPartKey
	FindPhoneNumberByID                       = findPhoneNumberByID
	FindQueueByTwoPartKey                     = findQueueByTwoPartKey
	FindQuickConnectByTwoPartKey              = findQuickConnectByTwoPartKey
	FindRoutingProfileByTwoPartKey            = findRoutingProfileByTwoPartKey
	FindSecurityProfileByTwoPartKey           = findSecurityProfileByTwoPartKey
	FindUserByTwoPartKey                      = findUserByTwoPartKey
	FindUserHierarchyGroupByTwoPartKey        = findUserHierarchyGroupByTwoPartKey
	FindUserHierarchyStructureByID            = findUserHierarchyStructureByID
	FindVocabularyByTwoPartKey                = findVocabularyByTwoPartKey
)
