// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

// Exports for use in tests only.
var (
	ResourceLicenseAssociation = resourceLicenseAssociation
	ResourceWorkspace          = resourceWorkspace

	FindLicensedWorkspaceByID = findLicensedWorkspaceByID
	FindWorkspaceByID         = findWorkspaceByID
)
