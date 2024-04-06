// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

// Exports for use in tests only.
var (
	ResourceAPI                 = resourceAPI
	ResourceAPIMapping          = resourceAPIMapping
	ResourceAuthorizer          = resourceAuthorizer
	ResourceDeployment          = resourceDeployment
	ResourceDomainName          = resourceDomainName
	ResourceIntegration         = resourceIntegration
	ResourceIntegrationResponse = resourceIntegrationResponse

	FindAPIByID                           = findAPIByID
	FindAPIMappingByTwoPartKey            = findAPIMappingByTwoPartKey
	FindAuthorizerByTwoPartKey            = findAuthorizerByTwoPartKey
	FindDeploymentByTwoPartKey            = findDeploymentByTwoPartKey
	FindDomainName                        = findDomainName
	FindIntegrationByTwoPartKey           = findIntegrationByTwoPartKey
	FindIntegrationResponseByThreePartKey = findIntegrationResponseByThreePartKey
)
