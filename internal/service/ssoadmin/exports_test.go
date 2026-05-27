// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

// Exports for use in tests only.
var (
	ResourceAccountAssignment                         = resourceAccountAssignment
	ResourceApplication                               = newApplicationResource
	ResourceApplicationAccessScope                    = newApplicationAccessScopeResource
	ResourceApplicationAssignment                     = newApplicationAssignmentResource
	ResourceApplicationAssignmentConfiguration        = newApplicationAssignmentConfigurationResource
	ResourceCustomerManagedPolicyAttachment           = resourceCustomerManagedPolicyAttachment
	ResourceCustomerManagedPolicyAttachmentsExclusive = newCustomerManagedPolicyAttachmentsExclusiveResource
	ResourceInstanceAccessControlAttributes           = resourceInstanceAccessControlAttributes
	ResourceManagedPolicyAttachment                   = resourceManagedPolicyAttachment
	ResourceManagedPolicyAttachmentsExclusive         = newManagedPolicyAttachmentsExclusiveResource
	ResourcePermissionsBoundaryAttachment             = resourcePermissionsBoundaryAttachment
	ResourcePermissionSet                             = resourcePermissionSet
	ResourcePermissionSetInlinePolicy                 = resourcePermissionSetInlinePolicy
	ResourceTrustedTokenIssuer                        = newTrustedTokenIssuerResource

	FindAccountAssignmentByFivePartKey               = findAccountAssignmentByFivePartKey
	FindApplicationAccessScopeByID                   = findApplicationAccessScopeByID
	FindApplicationAssignmentByID                    = findApplicationAssignmentByID
	FindApplicationAssignmentConfigurationByID       = findApplicationAssignmentConfigurationByID
	FindApplicationByID                              = findApplicationByID
	FindCustomerManagedPolicyAttachmentsByTwoPartKey = findCustomerManagedPolicyAttachmentsByTwoPartKey
	FindCustomerManagedPolicyByFourPartKey           = findCustomerManagedPolicyByFourPartKey
	FindInstanceAttributeControlAttributesByARN      = findInstanceAttributeControlAttributesByARN
	FindManagedPolicyAttachmentsByTwoPartKey         = findManagedPolicyAttachmentsByTwoPartKey
	FindManagedPolicyByThreePartKey                  = findManagedPolicyByThreePartKey
	FindPermissionsBoundaryByTwoPartKey              = findPermissionsBoundaryByTwoPartKey
	FindPermissionSetByTwoPartKey                    = findPermissionSetByTwoPartKey
	FindPermissionSetInlinePolicyByTwoPartKey        = findPermissionSetInlinePolicyByTwoPartKey
	FindTrustedTokenIssuerByARN                      = findTrustedTokenIssuerByARN
)
