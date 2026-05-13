// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub

// Exports for use in tests only.
var (
	ResourceAccount                        = resourceAccount
	ResourceAccountV2                      = newAccountV2Resource
	ResourceAggregatorV2                   = newAggregatorV2Resource
	ResourceActionTarget                   = resourceActionTarget
	ResourceAutomationRule                 = newAutomationRuleResource
	ResourceConfigurationPolicy            = resourceConfigurationPolicy
	ResourceConfigurationPolicyAssociation = resourceConfigurationPolicyAssociation
	ResourceConnectorV2                    = newConnectorV2Resource
	ResourceFindingAggregator              = resourceFindingAggregator
	ResourceInsight                        = resourceInsight
	ResourceInviteAccepter                 = resourceInviteAccepter
	ResourceMember                         = resourceMember
	ResourceOrganizationAdminAccount       = resourceOrganizationAdminAccount
	ResourceOrganizationConfiguration      = resourceOrganizationConfiguration
	ResourceProductSubscription            = resourceProductSubscription
	ResourceStandardsControl               = resourceStandardsControl
	ResourceStandardsControlAssociation    = newStandardsControlAssociationResource
	ResourceStandardsSubscription          = resourceStandardsSubscription

	AccountHubARN                                 = accountHubARN
	FindAccountV2                                 = findAccountV2
	FindActionTargetByARN                         = findActionTargetByARN
	FindAdminAccountByID                          = findAdminAccountByID
	FindAggregatorV2ByARN                         = findAggregatorV2ByARN
	FindAutomationRuleByARN                       = findAutomationRuleByARN
	FindConfigurationPolicyAssociationByID        = findConfigurationPolicyAssociationByID
	FindConfigurationPolicyByID                   = findConfigurationPolicyByID
	FindConnectorV2ByID                           = findConnectorV2ByID
	FindFindingAggregatorByARN                    = findFindingAggregatorByARN
	FindHubByARN                                  = findHubByARN
	FindInsightByARN                              = findInsightByARN
	FindMasterAccount                             = findMasterAccount
	FindMemberByAccountID                         = findMemberByAccountID
	FindOrganizationConfiguration                 = findOrganizationConfiguration
	FindProductSubscriptionByARN                  = findProductSubscriptionByARN
	FindStandardsControlAssociationByTwoPartKey   = findStandardsControlAssociationByTwoPartKey
	FindStandardsControlByTwoPartKey              = findStandardsControlByTwoPartKey
	FindStandardsSubscriptionByARN                = findStandardsSubscriptionByARN
	StandardsControlARNToStandardsSubscriptionARN = standardsControlARNToStandardsSubscriptionARN
)
