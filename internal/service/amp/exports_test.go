// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

// Exports for use in tests only.
var (
	ResourceAlertManagerDefinition = resourceAlertManagerDefinition
	ResourceRuleGroupNamespace     = resourceRuleGroupNamespace
	ResourceScraper                = newScraperResource
	ResourceWorkspace              = resourceWorkspace

	FindAlertManagerDefinitionByID = findAlertManagerDefinitionByID
	FindRuleGroupNamespaceByARN    = findRuleGroupNamespaceByARN
	FindScraperByID                = findScraperByID
	FindWorkspaceByID              = findWorkspaceByID
)
