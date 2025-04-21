// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

// Exports for use in tests only.
var (
	ResourceAPIKey                      = resourceAPIKey
	ResourceAuthorizer                  = resourceAuthorizer
	ResourceBasePathMapping             = resourceBasePathMapping
	ResourceClientCertificate           = resourceClientCertificate
	ResourceDeployment                  = resourceDeployment
	ResourceDocumentationPart           = resourceDocumentationPart
	ResourceDocumentationVersion        = resourceDocumentationVersion
	ResourceDomainName                  = resourceDomainName
	ResourceDomainNameAccessAssociation = newDomainNameAccessAssociationResource
	ResourceGatewayResponse             = resourceGatewayResponse
	ResourceIntegration                 = resourceIntegration
	ResourceIntegrationResponse         = resourceIntegrationResponse
	ResourceMethod                      = resourceMethod
	ResourceMethodResponse              = resourceMethodResponse
	ResourceMethodSettings              = resourceMethodSettings
	ResourceModel                       = resourceModel
	ResourceRequestValidator            = resourceRequestValidator
	ResourceResource                    = resourceResource
	ResourceRestAPI                     = resourceRestAPI
	ResourceRestAPIPolicy               = resourceRestAPIPolicy
	ResourceRestAPIPut                  = newResourceRestAPIPut
	ResourceStage                       = resourceStage
	ResourceUsagePlan                   = resourceUsagePlan
	ResourceUsagePlanKey                = resourceUsagePlanKey
	ResourceVPCLink                     = resourceVPCLink

	DefaultAuthorizerTTL                 = defaultAuthorizerTTL
	FindAPIKeyByID                       = findAPIKeyByID
	FindAccount                          = findAccount
	FindAuthorizerByTwoPartKey           = findAuthorizerByTwoPartKey
	FindBasePathMappingByThreePartKey    = findBasePathMappingByThreePartKey
	FindClientCertificateByID            = findClientCertificateByID
	FindDeploymentByTwoPartKey           = findDeploymentByTwoPartKey
	FindDocumentationPartByTwoPartKey    = findDocumentationPartByTwoPartKey
	FindDocumentationVersionByTwoPartKey = findDocumentationVersionByTwoPartKey
	FindDomainNameAccessAssociationByARN = findDomainNameAccessAssociationByARN
	FindDomainNameByTwoPartKey           = findDomainNameByTwoPartKey
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
