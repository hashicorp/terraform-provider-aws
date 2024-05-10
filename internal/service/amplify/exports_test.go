// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify

// Exports for use in tests only.
var (
	ResourceApp                = resourceApp
	ResourceBackendEnvironment = resourceBackendEnvironment
	ResourceBranch             = resourceBranch
	ResourceDomainAssociation  = resourceDomainAssociation
	ResourceWebhook            = resourceWebhook

	FindAppByID                        = findAppByID
	FindBackendEnvironmentByTwoPartKey = findBackendEnvironmentByTwoPartKey
	FindBranchByTwoPartKey             = findBranchByTwoPartKey
	FindDomainAssociationByTwoPartKey  = findDomainAssociationByTwoPartKey
	FindWebhookByID                    = findWebhookByID
)
