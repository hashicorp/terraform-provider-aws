// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

// Exports for use in tests only.
var (
	ResourceAggregateAuthorization      = resourceAggregateAuthorization
	ResourceConfigRule                  = resourceConfigRule
	ResourceConfigurationAggregator     = resourceConfigurationAggregator
	ResourceConfigurationRecorder       = resourceConfigurationRecorder
	ResourceConfigurationRecorderStatus = resourceConfigurationRecorderStatus
	ResourceConformancePack             = resourceConformancePack
	ResourceDeliveryChannel             = resourceDeliveryChannel
	ResourceOrganizationConformancePack = resourceOrganizationConformancePack
	ResourceOrganizationManagedRule     = resourceOrganizationManagedRule
	ResourceRemediationConfiguration    = resourceRemediationConfiguration

	FindAggregateAuthorizationByTwoPartKey       = findAggregateAuthorizationByTwoPartKey
	FindConfigRuleByName                         = findConfigRuleByName
	FindConfigurationRecorderByName              = findConfigurationRecorderByName
	FindConfigurationRecorderStatusByName        = findConfigurationRecorderStatusByName
	FindConformancePackByName                    = findConformancePackByName
	FindDeliveryChannelByName                    = findDeliveryChannelByName
	FindRemediationConfigurationByConfigRuleName = findRemediationConfigurationByConfigRuleName
)
