// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

// Exports for use in tests only.
var (
	ResourceAggregateAuthorization      = resourceAggregateAuthorization
	ResourceConfigurationAggregator     = resourceConfigurationAggregator
	ResourceConfigurationRecorder       = resourceConfigurationRecorder
	ResourceConfigurationRecorderStatus = resourceConfigurationRecorderStatus
	ResourceConformancePack             = resourceConformancePack
	ResourceDeliveryChannel             = resourceDeliveryChannel
	ResourceRemediationConfiguration    = resourceRemediationConfiguration

	FindAggregateAuthorizationByTwoPartKey       = findAggregateAuthorizationByTwoPartKey
	FindConfigurationRecorderByName              = findConfigurationRecorderByName
	FindConfigurationRecorderStatusByName        = findConfigurationRecorderStatusByName
	FindConformancePackByName                    = findConformancePackByName
	FindDeliveryChannelByName                    = findDeliveryChannelByName
	FindRemediationConfigurationByConfigRuleName = findRemediationConfigurationByConfigRuleName
)
