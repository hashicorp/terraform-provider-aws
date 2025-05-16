// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

// Exports for use in tests only.
var (
	ResourceAccount                = resourceAccount
	ResourceAccountParent          = newResourceAccountParent
	ResourceDelegatedAdministrator = resourceDelegatedAdministrator
	ResourceOrganization           = resourceOrganization
	ResourceOrganizationalUnit     = resourceOrganizationalUnit
	ResourcePolicy                 = resourcePolicy
	ResourcePolicyAttachment       = resourcePolicyAttachment
	ResourceResourcePolicy         = resourceResourcePolicy

	FindAccountByID                  = findAccountByID
	FindOrganizationalUnitByID       = findOrganizationalUnitByID
	FindParentAccountID              = findParentAccountID
	FindPolicyAttachmentByTwoPartKey = findPolicyAttachmentByTwoPartKey
	FindPolicyByID                   = findPolicyByID
	FindResourcePolicy               = findResourcePolicy
)
