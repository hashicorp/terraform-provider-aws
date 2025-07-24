// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

// Exports for use in tests only.
var (
	ResourceAPICache                 = resourceAPICache
	ResourceAPIKey                   = resourceAPIKey
	ResourceDataSource               = resourceDataSource
	ResourceDomainName               = resourceDomainName
	ResourceDomainNameAPIAssociation = resourceDomainNameAPIAssociation
	ResourceAPI                      = newAPIResource
	ResourceFunction                 = resourceFunction
	ResourceGraphQLAPI               = resourceGraphQLAPI
	ResourceResolver                 = resourceResolver
	ResourceType                     = resourceType
	ResourceSourceAPIAssociation     = newSourceAPIAssociationResource

	DefaultAuthorizerResultTTLInSeconds  = defaultAuthorizerResultTTLInSeconds
	FindAPICacheByID                     = findAPICacheByID
	FindAPIKeyByTwoPartKey               = findAPIKeyByTwoPartKey
	FindDataSourceByTwoPartKey           = findDataSourceByTwoPartKey
	FindDomainNameAPIAssociationByID     = findDomainNameAPIAssociationByID
	FindDomainNameByID                   = findDomainNameByID
	FindAPIByID                          = findAPIByID
	FindFunctionByTwoPartKey             = findFunctionByTwoPartKey
	FindGraphQLAPIByID                   = findGraphQLAPIByID
	FindResolverByThreePartKey           = findResolverByThreePartKey
	FindSourceAPIAssociationByTwoPartKey = findSourceAPIAssociationByTwoPartKey
	FindTypeByThreePartKey               = findTypeByThreePartKey
)
