// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appsync

// Exports for use in tests only.
var (
	ResourceAPI                      = newAPIResource
	ResourceAPICache                 = resourceAPICache
	ResourceAPIKey                   = resourceAPIKey
	ResourceChannelNamespace         = newChannelNamespaceResource
	ResourceDataSource               = resourceDataSource
	ResourceDomainName               = resourceDomainName
	ResourceDomainNameAPIAssociation = resourceDomainNameAPIAssociation
	ResourceFunction                 = resourceFunction
	ResourceGraphQLAPI               = resourceGraphQLAPI
	ResourceResolver                 = resourceResolver
	ResourceType                     = resourceType
	ResourceSourceAPIAssociation     = newSourceAPIAssociationResource

	DefaultAuthorizerResultTTLInSeconds  = defaultAuthorizerResultTTLInSeconds
	FindAPIByID                          = findAPIByID
	FindAPICacheByID                     = findAPICacheByID
	FindAPIKeyByTwoPartKey               = findAPIKeyByTwoPartKey
	FindChannelNamespaceByTwoPartKey     = findChannelNamespaceByTwoPartKey
	FindDataSourceByTwoPartKey           = findDataSourceByTwoPartKey
	FindDomainNameAPIAssociationByID     = findDomainNameAPIAssociationByID
	FindDomainNameByID                   = findDomainNameByID
	FindFunctionByTwoPartKey             = findFunctionByTwoPartKey
	FindGraphQLAPIByID                   = findGraphQLAPIByID
	FindResolverByThreePartKey           = findResolverByThreePartKey
	FindSourceAPIAssociationByTwoPartKey = findSourceAPIAssociationByTwoPartKey
	FindTypeByThreePartKey               = findTypeByThreePartKey
)
