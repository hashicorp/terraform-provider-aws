// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

// Exports for use in tests only.
var (
	ResourceAccessKey             = resourceAccessKey
	ResourceAccountAlias          = resourceAccountAlias
	ResourceAccountPasswordPolicy = resourceAccountPasswordPolicy
	ResourceGroup                 = resourceGroup
	// ResourceGroupMembership       = resourceGroupMembership
	ResourceGroupPolicy                   = resourceGroupPolicy
	ResourceGroupPolicyAttachment         = resourceGroupPolicyAttachment
	ResourceInstanceProfile               = resourceInstanceProfile
	ResourceOpenIDConnectProvider         = resourceOpenIDConnectProvider
	ResourceOrganizationsFeatures         = newOrganizationsFeaturesResource
	ResourceOutboundWebIdentityFederation = newOutboundWebIdentityFederationResource
	ResourcePolicy                        = resourcePolicy
	ResourcePolicyAttachment              = resourcePolicyAttachment
	ResourceRolePolicy                    = resourceRolePolicy
	ResourceRolePolicyAttachment          = resourceRolePolicyAttachment
	ResourceSAMLProvider                  = resourceSAMLProvider
	ResourceServerCertificate             = resourceServerCertificate
	ResourceServiceLinkedRole             = resourceServiceLinkedRole
	ResourceServiceSpecificCredential     = resourceServiceSpecificCredential
	ResourceSigningCertificate            = resourceSigningCertificate
	ResourceUser                          = resourceUser
	ResourceUserGroupMembership           = resourceUserGroupMembership
	ResourceUserLoginProfile              = resourceUserLoginProfile
	ResourceUserPolicy                    = resourceUserPolicy
	ResourceUserPolicyAttachment          = resourceUserPolicyAttachment
	ResourceUserSSHKey                    = resourceUserSSHKey
	ResourceVirtualMFADevice              = resourceVirtualMFADevice

	FindAccessKeyByTwoPartKey                   = findAccessKeyByTwoPartKey
	FindAccountAlias                            = findAccountAlias
	FindAccountPasswordPolicy                   = findAccountPasswordPolicy
	FindAttachedGroupPolicies                   = findAttachedGroupPolicies
	FindAttachedGroupPolicyByTwoPartKey         = findAttachedGroupPolicyByTwoPartKey
	FindAttachedRolePolicies                    = findAttachedRolePolicies
	FindAttachedRolePolicyByTwoPartKey          = findAttachedRolePolicyByTwoPartKey
	FindAttachedUserPolicies                    = findAttachedUserPolicies
	FindAttachedUserPolicyByTwoPartKey          = findAttachedUserPolicyByTwoPartKey
	FindEntitiesForPolicyByARN                  = findEntitiesForPolicyByARN
	FindGroupByName                             = findGroupByName
	FindGroupPoliciesByName                     = findGroupPoliciesByName
	FindGroupPolicyAttachmentsByName            = findGroupPolicyAttachmentsByName
	FindGroupPolicyByTwoPartKey                 = findGroupPolicyByTwoPartKey
	FindInstanceProfileByName                   = findInstanceProfileByName
	FindOpenIDConnectProviderByARN              = findOpenIDConnectProviderByARN
	FindOrganizationsFeatures                   = findOrganizationsFeatures
	FindOutboundWebIdentityFederation           = findOutboundWebIdentityFederation
	FindPolicyByARN                             = findPolicyByARN
	FindRolePolicyByTwoPartKey                  = findRolePolicyByTwoPartKey
	FindRolePoliciesByName                      = findRolePoliciesByName
	FindRolePolicyAttachmentsByName             = findRolePolicyAttachmentsByName
	FindSAMLProviderByARN                       = findSAMLProviderByARN
	FindServerCertificateByName                 = findServerCertificateByName
	FindServiceSpecificCredentialByThreePartKey = findServiceSpecificCredentialByThreePartKey
	FindSigningCertificateByTwoPartKey          = findSigningCertificateByTwoPartKey
	FindSSHPublicKeyByThreePartKey              = findSSHPublicKeyByThreePartKey
	FindUserByName                              = findUserByName
	FindUserPoliciesByName                      = findUserPoliciesByName
	FindUserPolicyAttachmentsByName             = findUserPolicyAttachmentsByName
	FindUserPolicyByTwoPartKey                  = findUserPolicyByTwoPartKey
	FindVirtualMFADeviceBySerialNumber          = findVirtualMFADeviceBySerialNumber

	AttachPolicyToUser                = attachPolicyToUser
	CheckPwdPolicy                    = checkPwdPolicy
	GeneratePassword                  = generatePassword
	IsValidPolicyAWSPrincipal         = isValidPolicyAWSPrincipal // nosemgrep:ci.aws-in-var-name
	ListGroupsForUserPages            = listGroupsForUserPages
	RoleNameSessionFromARN            = roleNameSessionFromARN
	RolePolicyParseID                 = rolePolicyParseID
	ServiceLinkedRoleParseResourceID  = serviceLinkedRoleParseResourceID
	SESSMTPPasswordFromSecretKeySigV4 = sesSMTPPasswordFromSecretKeySigV4
)

type (
	IAMPolicyStatementConditionSet = iamPolicyStatementConditionSet
)
