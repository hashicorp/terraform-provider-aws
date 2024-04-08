// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

// Exports for use in tests only.
var (
	ResourceAccount              = resourceAccount
	ResourceAPIKey               = resourceAPIKey
	ResourceAuthorizer           = resourceAuthorizer
	ResourceBasePathMapping      = resourceBasePathMapping
	ResourceClientCertificate    = resourceClientCertificate
	ResourceDeployment           = resourceDeployment
	ResourceDocumentationPart    = resourceDocumentationPart
	ResourceDocumentationVersion = resourceDocumentationVersion
	ResourceDomainName           = resourceDomainName
	ResourceGatewayResponse      = resourceGatewayResponse
	ResourceIntegration          = resourceIntegration
	ResourceIntegrationResponse  = resourceIntegrationResponse
	ResourceMethod               = resourceMethod
	ResourceMethodResponse       = resourceMethodResponse
	ResourceMethodSettings       = resourceMethodSettings
	ResourceModel                = resourceModel
	ResourceRequestValidator     = resourceRequestValidator
	ResourceResource             = resourceResource
	ResourceRestAPI              = resourceRestAPI
	ResourceRestAPIPolicy        = resourceRestAPIPolicy
	ResourceStage                = resourceStage
	ResourceUsagePlan            = resourceUsagePlan
	ResourceUsagePlanKey         = resourceUsagePlanKey
	ResourceVPCLink              = resourceVPCLink

	DefaultAuthorizerTTL                 = defaultAuthorizerTTL
	FindAPIKeyByID                       = findAPIKeyByID
	FindAuthorizerByTwoPartKey           = findAuthorizerByTwoPartKey
	FindBasePathMappingByTwoPartKey      = findBasePathMappingByTwoPartKey
	FindClientCertificateByID            = findClientCertificateByID
	FindDeploymentByTwoPartKey           = findDeploymentByTwoPartKey
	FindDocumentationPartByTwoPartKey    = findDocumentationPartByTwoPartKey
	FindDocumentationVersionByTwoPartKey = findDocumentationVersionByTwoPartKey
	FindDomainByName                     = findDomainByName
	FindGatewayResponseByTwoPartKey      = findGatewayResponseByTwoPartKey
	FindIntegrationByThreePartKey        = findIntegrationByThreePartKey
	FindIntegrationResponseByFourPartKey = findIntegrationResponseByFourPartKey
	FindMethodByThreePartKey             = findMethodByThreePartKey
	FindMethodResponseByFourPartKey      = findMethodResponseByFourPartKey
	FindMethodSettingsByThreePartKey     = findMethodSettingsByThreePartKey
	FindModelByTwoPartKey                = findModelByTwoPartKey
	FindRequestValidatorByTwoPartKey     = findRequestValidatorByTwoPartKey
	FindResourceByTwoPartKey             = findResourceByTwoPartKey
	FindRestAPIByID                      = findRestAPIByID
	FindStageByTwoPartKey                = findStageByTwoPartKey
	FindUsagePlanByID                    = findUsagePlanByID
	FindUsagePlanKeyByTwoPartKey         = findUsagePlanKeyByTwoPartKey
	FindVPCLinkByID                      = findVPCLinkByID
)
