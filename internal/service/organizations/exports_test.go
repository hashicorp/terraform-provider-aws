// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations

// Exports for use in tests only.
var (
	ResourceAccount                = resourceAccount
	ResourceAWSServiceAccess       = newAWSServiceAccessResource // nosemgrep:ci.aws-in-var-name
	ResourceDelegatedAdministrator = resourceDelegatedAdministrator
	ResourceOrganization           = resourceOrganization
	ResourceOrganizationalUnit     = resourceOrganizationalUnit
	ResourcePolicy                 = resourcePolicy
	ResourcePolicyAttachment       = resourcePolicyAttachment
	ResourceResourcePolicy         = resourceResourcePolicy
	ResourceTag                    = resourceTag

	FindAccountByID                        = findAccountByID
	FindAWSServiceAccessByServicePrincipal = findAWSServiceAccessByServicePrincipal // nosemgrep:ci.aws-in-var-name
	FindOrganizationalUnitByID             = findOrganizationalUnitByID
	FindPolicyAttachmentByTwoPartKey       = findPolicyAttachmentByTwoPartKey
	FindPolicyByID                         = findPolicyByID
	FindResourcePolicy                     = findResourcePolicy
	FindTag                                = findTag
)
