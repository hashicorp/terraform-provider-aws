// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

// Exports for use in tests only.
var (
	ResourceAccountSubscription = resourceAccountSubscription
	ResourceAnalysis            = resourceAnalysis
	ResourceFolderMembership    = newFolderMembershipResource
	ResourceGroup               = resourceGroup
	ResourceGroupMembership     = resourceGroupMembership
	ResourceIAMPolicyAssignment = newIAMPolicyAssignmentResource
	ResourceIngestion           = newIngestionResource
	ResourceNamespace           = newNamespaceResource
	ResourceRefreshSchedule     = newRefreshScheduleResource
	ResourceTemplateAlias       = newTemplateAliasResource
	ResourceUser                = resourceUser
	ResourceVPCConnection       = newVPCConnectionResource

	DefaultGroupNamespace                 = defaultGroupNamespace
	DefaultIAMPolicyAssignmentNamespace   = defaultIAMPolicyAssignmentNamespace
	DefaultUserNamespace                  = defaultUserNamespace
	FindAccountSubscriptionByID           = findAccountSubscriptionByID
	FindAnalysisByTwoPartKey              = findAnalysisByTwoPartKey
	FindFolderMembershipByFourPartKey     = findFolderMembershipByFourPartKey
	FindGroupByThreePartKey               = findGroupByThreePartKey
	FindGroupMembershipByFourPartKey      = findGroupMembershipByFourPartKey
	FindIAMPolicyAssignmentByThreePartKey = findIAMPolicyAssignmentByThreePartKey
	FindIngestionByThreePartKey           = findIngestionByThreePartKey
	FindNamespaceByTwoPartKey             = findNamespaceByTwoPartKey
	FindRefreshScheduleByThreePartKey     = findRefreshScheduleByThreePartKey
	FindTemplateAliasByThreePartKey       = findTemplateAliasByThreePartKey
	FindUserByThreePartKey                = findUserByThreePartKey
	FindVPCConnectionByTwoPartKey         = findVPCConnectionByTwoPartKey
)
