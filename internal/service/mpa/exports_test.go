// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mpa

// Exports for use in tests only.
var (
	ResourceApprovalTeam   = newApprovalTeamResource
	ResourceIdentitySource = newIdentitySourceResource

	FindApprovalTeamByARN  = findApprovalTeamByARN
	FindIdentitySourceByID = findIdentitySourceByID
)
