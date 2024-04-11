// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

// Exports for use in tests only.
var (
	ResourceAccessKey = resourceAccessKey
	// ResourceAccountAlias          = resourceAccountAlias
	ResourceAccountPasswordPolicy = resourceAccountPasswordPolicy
	ResourceGroup                 = resourceGroup
	// ResourceGroupMembership       = resourceGroupMembership
	ResourceGroupPolicy               = resourceGroupPolicy
	ResourceGroupPolicyAttachment     = resourceGroupPolicyAttachment
	ResourceInstanceProfile           = resourceInstanceProfile
	ResourceOpenIDConnectProvider     = resourceOpenIDConnectProvider
	ResourcePolicy                    = resourcePolicy
	ResourcePolicyAttachment          = resourcePolicyAttachment
	ResourceRolePolicy                = resourceRolePolicy
	ResourceRolePolicyAttachment      = resourceRolePolicyAttachment
	ResourceSAMLProvider              = resourceSAMLProvider
	ResourceServerCertificate         = resourceServerCertificate
	ResourceServiceLinkedRole         = resourceServiceLinkedRole
	ResourceServiceSpecificCredential = resourceServiceSpecificCredential
	ResourceSigningCertificate        = resourceSigningCertificate
	ResourceUser                      = resourceUser
	ResourceUserGroupMembership       = resourceUserGroupMembership
	ResourceUserLoginProfile          = resourceUserLoginProfile
	ResourceUserPolicy                = resourceUserPolicy
	ResourceUserPolicyAttachment      = resourceUserPolicyAttachment
	ResourceUserSSHKey                = resourceUserSSHKey
	ResourceVirtualMFADevice          = resourceVirtualMFADevice

	FindAccessKeyByTwoPartKey           = findAccessKeyByTwoPartKey
	FindAccountPasswordPolicy           = findAccountPasswordPolicy
	FindAttachedGroupPolicies           = findAttachedGroupPolicies
	FindAttachedGroupPolicyByTwoPartKey = findAttachedGroupPolicyByTwoPartKey
	FindAttachedRolePolicies            = findAttachedRolePolicies
	FindAttachedRolePolicyByTwoPartKey  = findAttachedRolePolicyByTwoPartKey
	FindAttachedUserPolicies            = findAttachedUserPolicies
	FindAttachedUserPolicyByTwoPartKey  = findAttachedUserPolicyByTwoPartKey
	FindEntitiesForPolicyByARN          = findEntitiesForPolicyByARN
	FindGroupByName                     = findGroupByName
	FindInstanceProfileByName           = findInstanceProfileByName
	FindOpenIDConnectProviderByARN      = findOpenIDConnectProviderByARN
	FindPolicyByARN                     = findPolicyByARN
	FindSAMLProviderByARN               = findSAMLProviderByARN
	FindServerCertificateByName         = findServerCertificateByName
	FindSSHPublicKeyByThreePartKey      = findSSHPublicKeyByThreePartKey
	FindUserByName                      = findUserByName
	FindVirtualMFADeviceBySerialNumber  = findVirtualMFADeviceBySerialNumber
	SESSMTPPasswordFromSecretKeySigV4   = sesSMTPPasswordFromSecretKeySigV4
)
