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
	ResourceThingGroup            = resourceThingGroup
	ResourceThingGroupMembership  = resourceThingGroupMembership
	ResourceThingType             = resourceThingType

	FindAttachedPolicyByTwoPartKey       = findAttachedPolicyByTwoPartKey
	FindAuthorizerByName                 = findAuthorizerByName
	FindBillingGroupByName               = findBillingGroupByName
	FindCACertificateByID                = findCACertificateByID
	FindCertificateByID                  = findCertificateByID
	FindDomainConfigurationByName        = findDomainConfigurationByName
	FindPolicyByName                     = findPolicyByName
	FindPolicyVersionsByName             = findPolicyVersionsByName
	FindProvisioningTemplateByName       = findProvisioningTemplateByName
	FindRoleAliasByID                    = findRoleAliasByID
	FindTopicRuleDestinationByARN        = findTopicRuleDestinationByARN
	FindThingTypeByName                  = findThingTypeByName
	FindThingGroupByName                 = findThingGroupByName
	FindThingGroupMembershipByTwoPartKey = findThingGroupMembershipByTwoPartKey
	FindThingByName                      = findThingByName
	FindTopicRuleByName                  = findTopicRuleByName
)
