// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package xray

// Exports for use in tests only.
var (
	FindEncryptionConfig   = findEncryptionConfig
	FindGroupByARN         = findGroupByARN
	FindSamplingRuleByName = findSamplingRuleByName

	ResourceEncryptionConfig = resourceEncryptionConfig
	ResourceGroup            = resourceGroup
	ResourceSamplingRule     = resourceSamplingRule
)
