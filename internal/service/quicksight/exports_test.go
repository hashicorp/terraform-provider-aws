// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

// Exports for use in tests only.
var (
	ResourceAccountSubscription = resourceAccountSubscription
	ResourceAnalysis            = resourceAnalysis
	ResourceDashboard           = resourceDashboard
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

	DashboardLatestVersion                = dashboardLatestVersion
	DefaultGroupNamespace                 = defaultGroupNamespace
	DefaultIAMPolicyAssignmentNamespace   = defaultIAMPolicyAssignmentNamespace
	DefaultUserNamespace                  = defaultUserNamespace
	FindAccountSubscriptionByID           = findAccountSubscriptionByID
	FindAnalysisByTwoPartKey              = findAnalysisByTwoPartKey
	FindDashboardByThreePartKey           = findDashboardByThreePartKey
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
