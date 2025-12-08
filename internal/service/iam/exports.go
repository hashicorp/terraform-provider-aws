// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

// Exports for use in other packages.
var (
	ResourceRole = resourceRole

	DeleteServiceLinkedRole     = deleteServiceLinkedRole
	FindRoleByName              = findRoleByName
	PolicyHasValidAWSPrincipals = policyHasValidAWSPrincipals // nosemgrep:ci.aws-in-var-name
)

type (
	IAMPolicyDoc       = iamPolicyDoc
	IAMPolicyStatement = iamPolicyStatement
)
