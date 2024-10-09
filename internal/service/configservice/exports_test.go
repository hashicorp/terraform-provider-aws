// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

// Exports for use in tests only.
var (
	ResourceAggregateAuthorization       = resourceAggregateAuthorization
	ResourceConfigRule                   = resourceConfigRule
	ResourceConfigurationAggregator      = resourceConfigurationAggregator
	ResourceConfigurationRecorder        = resourceConfigurationRecorder
	ResourceConformancePack              = resourceConformancePack
	ResourceDeliveryChannel              = resourceDeliveryChannel
	ResourceOrganizationConformancePack  = resourceOrganizationConformancePack
	ResourceOrganizationCustomPolicyRule = resourceOrganizationCustomPolicyRule
	ResourceOrganizationCustomRule       = resourceOrganizationCustomRule
	ResourceOrganizationManagedRule      = resourceOrganizationManagedRule
	ResourceRemediationConfiguration     = resourceRemediationConfiguration
	ResourceRetentionConfiguration       = newRetentionConfigurationResource

	FindAggregateAuthorizationByTwoPartKey       = findAggregateAuthorizationByTwoPartKey
	FindConfigRuleByName                         = findConfigRuleByName
	FindConfigurationAggregatorByName            = findConfigurationAggregatorByName
	FindConfigurationRecorderByName              = findConfigurationRecorderByName
	FindConfigurationRecorderStatusByName        = findConfigurationRecorderStatusByName
	FindConformancePackByName                    = findConformancePackByName
	FindDeliveryChannelByName                    = findDeliveryChannelByName
	FindOrganizationConformancePackByName        = findOrganizationConformancePackByName
	FindOrganizationCustomPolicyRuleByName       = findOrganizationCustomPolicyRuleByName
	FindOrganizationCustomRuleByName             = findOrganizationCustomRuleByName
	FindOrganizationManagedRuleByName            = findOrganizationManagedRuleByName
	FindRemediationConfigurationByConfigRuleName = findRemediationConfigurationByConfigRuleName
	FindRetentionConfigurationByName             = findRetentionConfigurationByName
)
