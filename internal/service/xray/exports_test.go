// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package xray

// Exports for use in tests only.
var (
	FindEncryptionConfig     = findEncryptionConfig
	FindGroupByARN           = findGroupByARN
	FindIndexingRuleByName   = findIndexingRuleByName
	FindResourcePolicyByName = findResourcePolicyByName
	FindSamplingRuleByName   = findSamplingRuleByName

	ResourceEncryptionConfig = resourceEncryptionConfig
	ResourceGroup            = resourceGroup
	ResourceIndexingRule     = newIndexingRuleResource
	ResourceResourcePolicy   = newResourcePolicyResource
	ResourceSamplingRule     = resourceSamplingRule
)
