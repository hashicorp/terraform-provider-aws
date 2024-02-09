// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

// Exports for use in tests only.
var (
	ResourceGroup                 = resourceGroup
	ResourceGroupPolicyAttachment = resourceGroupPolicyAttachment
	ResourceInstanceProfile       = resourceInstanceProfile
	ResourceOpenIDConnectProvider = resourceOpenIDConnectProvider
	ResourcePolicy                = resourcePolicy
	ResourcePolicyAttachment      = resourcePolicyAttachment
	ResourceRolePolicyAttachment  = resourceRolePolicyAttachment
	ResourceSAMLProvider          = resourceSAMLProvider
	ResourceServerCertificate     = resourceServerCertificate
	ResourceServiceLinkedRole     = resourceServiceLinkedRole
	ResourceUser                  = resourceUser
	ResourceUserPolicyAttachment  = resourceUserPolicyAttachment
	ResourceVirtualMFADevice      = resourceVirtualMFADevice

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
	FindUserByName                      = findUserByName
	FindVirtualMFADeviceBySerialNumber  = findVirtualMFADeviceBySerialNumber
)
