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
	ResourceEventApi                 = newEventApiResource
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
	FindEventApiByID                     = findEventApiByID
	FindFunctionByTwoPartKey             = findFunctionByTwoPartKey
	FindGraphQLAPIByID                   = findGraphQLAPIByID
	FindResolverByThreePartKey           = findResolverByThreePartKey
	FindSourceAPIAssociationByTwoPartKey = findSourceAPIAssociationByTwoPartKey
	FindTypeByThreePartKey               = findTypeByThreePartKey
)
