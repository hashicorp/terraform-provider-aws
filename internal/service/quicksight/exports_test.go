// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

// Exports for use in tests only.
var (
	ResourceAccountSubscription = resourceAccountSubscription
	ResourceFolderMembership    = newResourceFolderMembership
	ResourceGroup               = resourceGroup
	ResourceGroupMembership     = resourceGroupMembership
	ResourceIAMPolicyAssignment = newIAMPolicyAssignmentResource
	ResourceIngestion           = newResourceIngestion
	ResourceNamespace           = newNamespaceResource
	ResourceRefreshSchedule     = newResourceRefreshSchedule
	ResourceTemplateAlias       = newResourceTemplateAlias
	ResourceUser                = resourceUser
	ResourceVPCConnection       = newVPCConnectionResource

	DefaultGroupNamespace                 = defaultGroupNamespace
	DefaultIAMPolicyAssignmentNamespace   = defaultIAMPolicyAssignmentNamespace
	DefaultUserNamespace                  = defaultUserNamespace
	FindAccountSubscriptionByID           = findAccountSubscriptionByID
	FindGroupByThreePartKey               = findGroupByThreePartKey
	FindGroupMembershipByFourPartKey      = findGroupMembershipByFourPartKey
	FindIAMPolicyAssignmentByThreePartKey = findIAMPolicyAssignmentByThreePartKey
	FindNamespaceByTwoPartKey             = findNamespaceByTwoPartKey
	FindUserByThreePartKey                = findUserByThreePartKey
	FindVPCConnectionByTwoPartKey         = findVPCConnectionByTwoPartKey
)
