// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

// Exports for use in tests only.
var (
	ResourceAccount                        = resourceAccount
	ResourceActionTarget                   = resourceActionTarget
	ResourceAutomationRule                 = newAutomationRuleResource
	ResourceConfigurationPolicy            = resourceConfigurationPolicy
	ResourceConfigurationPolicyAssociation = resourceConfigurationPolicyAssociation
	ResourceOrganizationConfiguration      = resourceOrganizationConfiguration

	AccountHubARN                          = accountHubARN
	FindActionTargetByARN                  = findActionTargetByARN
	FindAutomationRuleByARN                = findAutomationRuleByARN
	FindConfigurationPolicyAssociationByID = findConfigurationPolicyAssociationByID
	FindConfigurationPolicyByID            = findConfigurationPolicyByID
	FindHubByARN                           = findHubByARN
	FindOrganizationConfiguration          = findOrganizationConfiguration
)
