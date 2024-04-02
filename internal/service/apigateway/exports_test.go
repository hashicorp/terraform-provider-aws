// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

// Exports for use in tests only.
var (
	ResourceAccount         = resourceAccount
	ResourceAPIKey          = resourceAPIKey
	ResourceAuthorizer      = resourceAuthorizer
	ResourceBasePathMapping = resourceBasePathMapping

	FindAPIKeyByID                  = findAPIKeyByID
	FindAuthorizerByTwoPartKey      = findAuthorizerByTwoPartKey
	FindBasePathMappingByTwoPartKey = findBasePathMappingByTwoPartKey
)
