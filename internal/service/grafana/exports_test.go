// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

// Exports for use in tests only.
var (
	ResourceWorkspaceServiceAccount = newResourceWorkspaceServiceAccount

	FindWorkspaceServiceAccount      = findWorkspaceServiceAccount
	FindWorkspaceServiceAccountToken = findWorkspaceServiceAccountToken
)
