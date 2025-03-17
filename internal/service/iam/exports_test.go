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
	ResourceOrganizationsFeatures     = newOrganizationsFeaturesResource
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
	FindGroupPoliciesByName             = findGroupPoliciesByName
	FindGroupPolicyAttachmentsByName    = findGroupPolicyAttachmentsByName
	FindInstanceProfileByName           = findInstanceProfileByName
	FindOpenIDConnectProviderByARN      = findOpenIDConnectProviderByARN
	FindOrganizationsFeatures           = findOrganizationsFeatures
	FindPolicyByARN                     = findPolicyByARN
	FindRolePoliciesByName              = findRolePoliciesByName
	FindRolePolicyAttachmentsByName     = findRolePolicyAttachmentsByName
	FindSAMLProviderByARN               = findSAMLProviderByARN
	FindServerCertificateByName         = findServerCertificateByName
	FindSSHPublicKeyByThreePartKey      = findSSHPublicKeyByThreePartKey
	FindUserByName                      = findUserByName
	FindUserPoliciesByName              = findUserPoliciesByName
	FindUserPolicyAttachmentsByName     = findUserPolicyAttachmentsByName
	FindVirtualMFADeviceBySerialNumber  = findVirtualMFADeviceBySerialNumber
	SESSMTPPasswordFromSecretKeySigV4   = sesSMTPPasswordFromSecretKeySigV4
)
