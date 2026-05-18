// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package xray

// Exports for use in tests only.
var (
	FindEncryptionConfig        = findEncryptionConfig
	FindGroupByARN              = findGroupByARN
	FindResourcePolicyByName    = findResourcePolicyByName
	FindSamplingRuleByName      = findSamplingRuleByName
	FindTraceSegmentDestination = findTraceSegmentDestination

	ResourceEncryptionConfig = resourceEncryptionConfig
	ResourceGroup            = resourceGroup
	ResourceResourcePolicy   = newResourcePolicyResource
	ResourceSamplingRule     = resourceSamplingRule
)
