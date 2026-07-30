// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package amp

// Exports for use in tests only.
var (
	ResourceAlertManagerDefinition    = resourceAlertManagerDefinition
	ResourceAnomalyDetector           = newAnomalyDetectorResource
	ResourceQueryLoggingConfiguration = newQueryLoggingConfigurationResource
	ResourceRuleGroupNamespace        = resourceRuleGroupNamespace
	ResourceScraper                   = newScraperResource
	ResourceWorkspace                 = resourceWorkspace
	ResourceResourcePolicy            = newResourcePolicyResource

	FindAlertManagerDefinitionByID    = findAlertManagerDefinitionByID
	FindAnomalyDetectorByID           = findAnomalyDetectorByID
	FindQueryLoggingConfigurationByID = findQueryLoggingConfigurationByID
	FindResourcePolicyByWorkspaceID   = findResourcePolicyByWorkspaceID
	FindRuleGroupNamespaceByARN       = findRuleGroupNamespaceByARN
	FindScraperByID                   = findScraperByID
	FindWorkspaceByID                 = findWorkspaceByID
	FindWorkspaceConfigurationByID    = findWorkspaceConfigurationByID
)
