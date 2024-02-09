// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

// Exports for use in tests only.
var (
	ResourceGroupPolicyAttachment = resourceGroupPolicyAttachment
	ResourceInstanceProfile       = resourceInstanceProfile
	ResourceOpenIDConnectProvider = resourceOpenIDConnectProvider
	ResourcePolicyAttachment      = resourcePolicyAttachment
	ResourceRolePolicyAttachment  = resourceRolePolicyAttachment
	ResourceUserPolicyAttachment  = resourceUserPolicyAttachment

	FindAttachedGroupPolicies           = findAttachedGroupPolicies
	FindAttachedGroupPolicyByTwoPartKey = findAttachedGroupPolicyByTwoPartKey
	FindAttachedRolePolicies            = findAttachedRolePolicies
	FindAttachedRolePolicyByTwoPartKey  = findAttachedRolePolicyByTwoPartKey
	FindAttachedUserPolicies            = findAttachedUserPolicies
	FindAttachedUserPolicyByTwoPartKey  = findAttachedUserPolicyByTwoPartKey
	FindEntitiesForPolicyByARN          = findEntitiesForPolicyByARN
	FindPolicyByARN                     = findPolicyByARN
)
