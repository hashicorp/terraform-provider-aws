// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

// Exports for use in tests only.
var (
	ResourceWorkspace                    = resourceWorkspace
	ResourceWorkspaceAPIKey              = resourceWorkspaceAPIKey
	ResourceWorkspaceSAMLConfiguration   = resourceWorkspaceSAMLConfiguration
	ResourceWorkspaceServiceAccount      = newWorkspaceServiceAccountResource
	ResourceWorkspaceServiceAccountToken = newWorkspaceServiceAccountTokenResource

	FindLicensedWorkspaceByID                      = findLicensedWorkspaceByID
	FindRoleAssociationsByTwoPartKey               = findRoleAssociationsByTwoPartKey
	FindSAMLConfigurationByID                      = findSAMLConfigurationByID
	FindWorkspaceByID                              = findWorkspaceByID
	FindWorkspaceServiceAccountByTwoPartKey        = findWorkspaceServiceAccountByTwoPartKey
	FindWorkspaceServiceAccountTokenByThreePartKey = findWorkspaceServiceAccountTokenByThreePartKey
)
