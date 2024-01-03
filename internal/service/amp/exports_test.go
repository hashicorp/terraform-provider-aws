// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

// Exports for use in tests only.
var (
	ResourceAlertManagerDefinition = resourceAlertManagerDefinition
	ResourceRuleGroupNamespace     = resourceRuleGroupNamespace
	ResourceScraper                = newResourceScraper
	ResourceWorkspace              = resourceWorkspace

	FindAlertManagerDefinitionByID = findAlertManagerDefinitionByID
	FindRuleGroupNamespaceByARN    = findRuleGroupNamespaceByARN
	FindWorkspaceByID              = findWorkspaceByID
)
