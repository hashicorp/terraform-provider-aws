// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

// Exports for use in tests only.
var (
	ResourceAccountAssignment                  = resourceAccountAssignment
	ResourceApplication                        = newApplicationResource
	ResourceApplicationAssignment              = newApplicationAssignmentResource
	ResourceApplicationAssignmentConfiguration = newApplicationAssignmentConfigurationResource
	ResourceApplicationAccessScope             = newApplicationAccessScopeResource
	ResourceCustomerManagedPolicyAttachment    = resourceCustomerManagedPolicyAttachment
	ResourceInstanceAccessControlAttributes    = resourceInstanceAccessControlAttributes
	ResourceManagedPolicyAttachment            = resourceManagedPolicyAttachment
	ResourcePermissionsBoundaryAttachment      = resourcePermissionsBoundaryAttachment
	ResourcePermissionSet                      = resourcePermissionSet
	ResourcePermissionSetInlinePolicy          = resourcePermissionSetInlinePolicy
	ResourceTrustedTokenIssuer                 = newTrustedTokenIssuerResource

	FindAccountAssignmentByFivePartKey          = findAccountAssignmentByFivePartKey
	FindApplicationByID                         = findApplicationByID
	FindApplicationAssignmentByID               = findApplicationAssignmentByID
	FindApplicationAssignmentConfigurationByID  = findApplicationAssignmentConfigurationByID
	FindApplicationAccessScopeByID              = findApplicationAccessScopeByID
	FindCustomerManagedPolicyByFourPartKey      = findCustomerManagedPolicyByFourPartKey
	FindInstanceAttributeControlAttributesByARN = findInstanceAttributeControlAttributesByARN
	FindManagedPolicyByThreePartKey             = findManagedPolicyByThreePartKey
	FindPermissionsBoundaryByTwoPartKey         = findPermissionsBoundaryByTwoPartKey
	FindPermissionSetByTwoPartKey               = findPermissionSetByTwoPartKey
	FindPermissionSetInlinePolicyByTwoPartKey   = findPermissionSetInlinePolicyByTwoPartKey
	FindTrustedTokenIssuerByARN                 = findTrustedTokenIssuerByARN
)
