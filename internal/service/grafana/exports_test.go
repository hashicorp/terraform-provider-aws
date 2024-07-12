// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

// Exports for use in tests only.
var (
	ResourceWorkspace                  = resourceWorkspace
	ResourceWorkspaceAPIKey            = resourceWorkspaceAPIKey
	ResourceWorkspaceSAMLConfiguration = resourceWorkspaceSAMLConfiguration
	ResourceWorkspaceServiceAccount    = newResourceWorkspaceServiceAccount

	FindLicensedWorkspaceByID        = findLicensedWorkspaceByID
	FindRoleAssociationsByTwoPartKey = findRoleAssociationsByTwoPartKey
	FindSAMLConfigurationByID        = findSAMLConfigurationByID
	FindWorkspaceByID                = findWorkspaceByID
	FindWorkspaceServiceAccount      = findWorkspaceServiceAccount
	FindWorkspaceServiceAccountToken = findWorkspaceServiceAccountToken
)
