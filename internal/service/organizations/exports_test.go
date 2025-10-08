// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

// Exports for use in tests only.
var (
	ResourceAccount                = resourceAccount
	ResourceDelegatedAdministrator = resourceDelegatedAdministrator
	ResourceOrganization           = resourceOrganization
	ResourceOrganizationalUnit     = resourceOrganizationalUnit
	ResourcePolicy                 = resourcePolicy
	ResourcePolicyAttachment       = resourcePolicyAttachment
	ResourceResourcePolicy         = resourceResourcePolicy

	FindAccountByID                  = findAccountByID
	FindOrganizationalUnitByID       = findOrganizationalUnitByID
	FindPolicyAttachmentByTwoPartKey = findPolicyAttachmentByTwoPartKey
	FindPolicyByID                   = findPolicyByID
	FindResourcePolicy               = findResourcePolicy
)
