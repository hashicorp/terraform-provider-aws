// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

// Exports for use in tests only.
var (
	ResourceLicenseAssociation = resourceLicenseAssociation
	ResourceRoleAssociation    = resourceRoleAssociation
	ResourceWorkspace          = resourceWorkspace
	ResourceWorkspaceAPIKey    = resourceWorkspaceAPIKey

	FindLicensedWorkspaceByID        = findLicensedWorkspaceByID
	FindRoleAssociationsByTwoPartKey = findRoleAssociationsByTwoPartKey
	FindWorkspaceByID                = findWorkspaceByID
)
