// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

// Exports for use in tests only.
var (
	ResourceAccountSubscription = resourceAccountSubscription
	ResourceAnalysis            = resourceAnalysis
	ResourceDashboard           = resourceDashboard
	ResourceDataSet             = resourceDataSet
	ResourceDataSource          = resourceDataSource
	ResourceFolder              = resourceFolder
	ResourceFolderMembership    = newFolderMembershipResource
	ResourceGroup               = resourceGroup
	ResourceGroupMembership     = resourceGroupMembership
	ResourceIAMPolicyAssignment = newIAMPolicyAssignmentResource
	ResourceIngestion           = newIngestionResource
	ResourceNamespace           = newNamespaceResource
	ResourceRefreshSchedule     = newRefreshScheduleResource
	ResourceRoleMembership      = newResourceRoleMembership
	ResourceTemplate            = resourceTemplate
	ResourceTemplateAlias       = newTemplateAliasResource
	ResourceTheme               = resourceTheme
	ResourceUser                = resourceUser
	ResourceVPCConnection       = newVPCConnectionResource

	DashboardLatestVersion                = dashboardLatestVersion
	DefaultGroupNamespace                 = defaultGroupNamespace
	DefaultIAMPolicyAssignmentNamespace   = defaultIAMPolicyAssignmentNamespace
	DefaultUserNamespace                  = defaultUserNamespace
	FindAccountSubscriptionByID           = findAccountSubscriptionByID
	FindAnalysisByTwoPartKey              = findAnalysisByTwoPartKey
	FindDashboardByThreePartKey           = findDashboardByThreePartKey
	FindDataSetByTwoPartKey               = findDataSetByTwoPartKey
	FindDataSourceByTwoPartKey            = findDataSourceByTwoPartKey
	FindFolderByTwoPartKey                = findFolderByTwoPartKey
	FindFolderMembershipByFourPartKey     = findFolderMembershipByFourPartKey
	FindGroupByThreePartKey               = findGroupByThreePartKey
	FindGroupMembershipByFourPartKey      = findGroupMembershipByFourPartKey
	FindIAMPolicyAssignmentByThreePartKey = findIAMPolicyAssignmentByThreePartKey
	FindIngestionByThreePartKey           = findIngestionByThreePartKey
	FindNamespaceByTwoPartKey             = findNamespaceByTwoPartKey
	FindRefreshScheduleByThreePartKey     = findRefreshScheduleByThreePartKey
	FindRoleMembershipByMultiPartKey      = findRoleMembershipByMultiPartKey
	FindTemplateAliasByThreePartKey       = findTemplateAliasByThreePartKey
	FindTemplateByTwoPartKey              = findTemplateByTwoPartKey
	FindThemeByTwoPartKey                 = findThemeByTwoPartKey
	FindUserByThreePartKey                = findUserByThreePartKey
	FindVPCConnectionByTwoPartKey         = findVPCConnectionByTwoPartKey

	StartAfterDateTimeLayout = startAfterDateTimeLayout
)
