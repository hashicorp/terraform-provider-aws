// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

// Exports for use in tests only.
var (
	ResourceAlias              = resourceAlias
	ResourceCodeSigningConfig  = resourceCodeSigningConfig
	ResourceEventSourceMapping = resourceEventSourceMapping

	FindAliasByTwoPartKey                   = findAliasByTwoPartKey
	FindCodeSigningConfigByARN              = findCodeSigningConfigByARN
	FindEventSourceMappingConfigurationByID = findEventSourceMappingConfigurationByID
)
