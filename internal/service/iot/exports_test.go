// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

// Exports for use in tests only.
var (
	ResourceAuthorizer            = resourceAuthorizer
	ResourceBillingGroup          = resourceBillingGroup
	ResourceCACertificate         = resourceCACertificate
	ResourceCertificate           = resourceCertificate
	ResourceDomainConfiguration   = resourceDomainConfiguration
	ResourceEventConfigurations   = resourceEventConfigurations
	ResourceIndexingConfiguration = resourceIndexingConfiguration
	ResourceLoggingOptions        = resourceLoggingOptions
	ResourcePolicy                = resourcePolicy
	ResourcePolicyAttachment      = resourcePolicyAttachment
	ResourceProvisioningTemplate  = resourceProvisioningTemplate
	ResourceThingType             = resourceThingType

	FindAttachedPolicyByTwoPartKey = findAttachedPolicyByTwoPartKey
	FindAuthorizerByName           = findAuthorizerByName
	FindBillingGroupByName         = findBillingGroupByName
	FindCACertificateByID          = findCACertificateByID
	FindCertificateByID            = findCertificateByID
	FindDomainConfigurationByName  = findDomainConfigurationByName
	FindPolicyByName               = findPolicyByName
	FindProvisioningTemplateByName = findProvisioningTemplateByName
	FindRoleAliasByID              = findRoleAliasByID
	FindTopicRuleDestinationByARN  = findTopicRuleDestinationByARN
	FindThingTypeByName            = findThingTypeByName
	FindThingGroupByName           = findThingGroupByName
	FindThingGroupMembership       = findThingGroupMembership
	FindThingByName                = findThingByName
	FindTopicRuleByName            = findTopicRuleByName
	FindPolicyVersionsByName       = findPolicyVersionsByName
)
