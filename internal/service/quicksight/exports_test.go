// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
)

// Exports for use in tests only.
var (
	ResourceAccountSettings      = newAccountSettingsResource
	ResourceAccountSubscription  = resourceAccountSubscription
	ResourceAnalysis             = resourceAnalysis
	ResourceCustomPermissions    = newCustomPermissionsResource
	ResourceDashboard            = resourceDashboard
	ResourceDataSet              = resourceDataSet
	ResourceDataSource           = resourceDataSource
	ResourceFolder               = resourceFolder
	ResourceFolderMembership     = newFolderMembershipResource
	ResourceGroup                = resourceGroup
	ResourceGroupMembership      = resourceGroupMembership
	ResourceIAMPolicyAssignment  = newIAMPolicyAssignmentResource
	ResourceIngestion            = newIngestionResource
	ResourceIPRestriction        = newIPRestrictionResource
	ResourceKeyRegistration      = newKeyRegistrationResource
	ResourceNamespace            = newNamespaceResource
	ResourceRefreshSchedule      = newRefreshScheduleResource
	ResourceRoleCustomPermission = newRoleCustomPermissionResource
	ResourceRoleMembership       = newRoleMembershipResource
	ResourceTemplate             = resourceTemplate
	ResourceTemplateAlias        = newTemplateAliasResource
	ResourceTheme                = resourceTheme
	ResourceUser                 = resourceUser
	ResourceUserCustomPermission = newUserCustomPermissionResource
	ResourceVPCConnection        = newVPCConnectionResource

	DashboardLatestVersion                 = dashboardLatestVersion
	DefaultNamespace                       = quicksightschema.DefaultNamespace
	FindAccountSettingsByID                = findAccountSettingsByID
	FindAccountSubscriptionByID            = findAccountSubscriptionByID
	FindAnalysisByTwoPartKey               = findAnalysisByTwoPartKey
	FindCustomPermissionsByTwoPartKey      = findCustomPermissionsByTwoPartKey
	FindDashboardByThreePartKey            = findDashboardByThreePartKey
	FindDataSetByTwoPartKey                = findDataSetByTwoPartKey
	FindDataSourceByTwoPartKey             = findDataSourceByTwoPartKey
	FindFolderByTwoPartKey                 = findFolderByTwoPartKey
	FindFolderMembershipByFourPartKey      = findFolderMembershipByFourPartKey
	FindGroupByThreePartKey                = findGroupByThreePartKey
	FindGroupMembershipByFourPartKey       = findGroupMembershipByFourPartKey
	FindIAMPolicyAssignmentByThreePartKey  = findIAMPolicyAssignmentByThreePartKey
	FindIngestionByThreePartKey            = findIngestionByThreePartKey
	FindIPRestrictionByID                  = findIPRestrictionByID
	FindKeyRegistrationByID                = findKeyRegistrationByID
	FindNamespaceByTwoPartKey              = findNamespaceByTwoPartKey
	FindRefreshScheduleByThreePartKey      = findRefreshScheduleByThreePartKey
	FindRoleCustomPermissionByThreePartKey = findRoleCustomPermissionByThreePartKey
	FindRoleMembershipByFourPartKey        = findRoleMembershipByFourPartKey
	FindTemplateAliasByThreePartKey        = findTemplateAliasByThreePartKey
	FindTemplateByTwoPartKey               = findTemplateByTwoPartKey
	FindThemeByTwoPartKey                  = findThemeByTwoPartKey
	FindUserByThreePartKey                 = findUserByThreePartKey
	FindUserCustomPermissionByThreePartKey = findUserCustomPermissionByThreePartKey
	FindVPCConnectionByTwoPartKey          = findVPCConnectionByTwoPartKey

	StartAfterDateTimeLayout = startAfterDateTimeLayout
)
