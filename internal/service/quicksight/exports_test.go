// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

// Exports for use in tests only.
var (
	ResourceAccountSubscription = resourceAccountSubscription
	ResourceFolderMembership    = newResourceFolderMembership
	ResourceIAMPolicyAssignment = newResourceIAMPolicyAssignment
	ResourceIngestion           = newResourceIngestion
	ResourceNamespace           = newResourceNamespace
	ResourceRefreshSchedule     = newResourceRefreshSchedule
	ResourceTemplateAlias       = newResourceTemplateAlias
	ResourceUser                = resourceUser
	ResourceVPCConnection       = newVPCConnectionResource

	FindAccountSubscriptionByID   = findAccountSubscriptionByID
	FindUserByThreePartKey        = findUserByThreePartKey
	FindVPCConnectionByTwoPartKey = findVPCConnectionByTwoPartKey
)
