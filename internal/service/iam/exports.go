// Copyright IBM Corp. 2014, 2026
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
	IAMPolicyDoc                   = iamPolicyDoc
	IAMPolicyStatement             = iamPolicyStatement
	IAMPolicyStatementPrincipal    = iamPolicyStatementPrincipal
	IAMPolicyStatementPrincipalSet = iamPolicyStatementPrincipalSet
)
