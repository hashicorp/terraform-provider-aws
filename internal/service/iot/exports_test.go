// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

// Exports for use in tests only.
var (
	FindAuthorizerByName           = findAuthorizerByName
	FindBillingGroupByName         = findBillingGroupByName
	FindCACertificateByID          = findCACertificateByID
	FindCertificateByID            = findCertificateByID
	FindDomainConfigurationByName  = findDomainConfigurationByName
	FindPolicyByName               = findPolicyByName
	FindProvisioningTemplateByName = findProvisioningTemplateByName
	FindTopicRuleDestinationByARN  = findTopicRuleDestinationByARN
	FindThingTypeByName            = findThingTypeByName
	FindThingGroupByName           = findThingGroupByName
	FindThingGroupMembership       = findThingGroupMembership
	FindThingByName                = findThingByName
	FindTopicRuleByName            = findTopicRuleByName
	FindPolicyVersionsByName       = findPolicyVersionsByName
)
