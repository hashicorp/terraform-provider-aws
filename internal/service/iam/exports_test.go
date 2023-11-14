// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

// Exports for use in tests only.
var (
	ResourceGroupPolicyAttachment = resourceGroupPolicyAttachment

	FindAttachedGroupPolicies           = findAttachedGroupPolicies
	FindAttachedGroupPolicyByTwoPartKey = findAttachedGroupPolicyByTwoPartKey
)
