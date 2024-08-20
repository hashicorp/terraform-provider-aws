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
	ResourceModel               = resourceModel
	ResourceRoute               = resourceRoute
	ResourceRouteResponse       = resourceRouteResponse
	ResourceStage               = resourceStage
	ResourceVPCLink             = resourceVPCLink

	FindAPIByID                           = findAPIByID
	FindAPIMappingByTwoPartKey            = findAPIMappingByTwoPartKey
	FindAuthorizerByTwoPartKey            = findAuthorizerByTwoPartKey
	FindDeploymentByTwoPartKey            = findDeploymentByTwoPartKey
	FindDomainName                        = findDomainName
	FindIntegrationByTwoPartKey           = findIntegrationByTwoPartKey
	FindIntegrationResponseByThreePartKey = findIntegrationResponseByThreePartKey
	FindModelByTwoPartKey                 = findModelByTwoPartKey
	FindRouteByTwoPartKey                 = findRouteByTwoPartKey
	FindRouteResponseByThreePartKey       = findRouteResponseByThreePartKey
	FindStageByTwoPartKey                 = findStageByTwoPartKey
	FindVPCLinkByID                       = findVPCLinkByID
)
