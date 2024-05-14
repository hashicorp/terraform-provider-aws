// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

// Exports for use in tests only.
var (
	ResourceAccount = resourceAccount

	FindAccountByID                        = findAccountByID
	FindDelegatedAdministratorByTwoPartKey = findDelegatedAdministratorByTwoPartKey
	FindOrganizationalUnitByID             = findOrganizationalUnitByID
	FindPolicyByID                         = findPolicyByID
	FindResourcePolicy                     = findResourcePolicy
)
