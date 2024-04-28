// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

// Exports for use in tests only.
var (
	ResourceAlias                     = resourceAlias
	ResourceCodeSigningConfig         = resourceCodeSigningConfig
	ResourceEventSourceMapping        = resourceEventSourceMapping
	ResourceFunctionEventInvokeConfig = resourceFunctionEventInvokeConfig

	FindAliasByTwoPartKey                     = findAliasByTwoPartKey
	FindCodeSigningConfigByARN                = findCodeSigningConfigByARN
	FindEventSourceMappingByID                = findEventSourceMappingByID
	FindFunctionEventInvokeConfigByTwoPartKey = findFunctionEventInvokeConfigByTwoPartKey
	FunctionEventInvokeConfigParseResourceID  = functionEventInvokeConfigParseResourceID
	SignerServiceIsAvailable                  = signerServiceIsAvailable
)
