// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

// Exports for use in tests only.
var (
	ResourceAccount                        = resourceAccount
	ResourceAutomationRule                 = newAutomationRuleResource
	ResourceConfigurationPolicy            = resourceConfigurationPolicy
	ResourceConfigurationPolicyAssociation = resourceConfigurationPolicyAssociation
	ResourceOrganizationConfiguration      = resourceOrganizationConfiguration

	AccountHubARN                          = accountHubARN
	FindAutomationRuleByARN                = findAutomationRuleByARN
	FindConfigurationPolicyAssociationByID = findConfigurationPolicyAssociationByID
	FindConfigurationPolicyByID            = findConfigurationPolicyByID
	FindHubByARN                           = findHubByARN
	FindOrganizationConfiguration          = findOrganizationConfiguration
)
