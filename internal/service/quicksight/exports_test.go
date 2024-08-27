// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

// Exports for use in tests only.
var (
	ResourceAccountSubscription = resourceAccountSubscription
	ResourceFolderMembership    = newResourceFolderMembership
	ResourceGroup               = resourceGroup
	ResourceGroupMembership     = resourceGroupMembership
	ResourceIAMPolicyAssignment = newResourceIAMPolicyAssignment
	ResourceIngestion           = newResourceIngestion
	ResourceNamespace           = newResourceNamespace
	ResourceRefreshSchedule     = newResourceRefreshSchedule
	ResourceTemplateAlias       = newResourceTemplateAlias
	ResourceUser                = resourceUser
	ResourceVPCConnection       = newVPCConnectionResource

	DefaultGroupNamespace            = defaultGroupNamespace
	DefaultUserNamespace             = defaultUserNamespace
	FindAccountSubscriptionByID      = findAccountSubscriptionByID
	FindGroupByThreePartKey          = findGroupByThreePartKey
	FindGroupMembershipByFourPartKey = findGroupMembershipByFourPartKey
	FindUserByThreePartKey           = findUserByThreePartKey
	FindVPCConnectionByTwoPartKey    = findVPCConnectionByTwoPartKey
)
