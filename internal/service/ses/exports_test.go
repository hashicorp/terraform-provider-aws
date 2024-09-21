// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

// Exports for use in tests only.
var (
	ResourceActiveReceiptRuleSet       = resourceActiveReceiptRuleSet
	ResourceConfigurationSet           = resourceConfigurationSet
	ResourceDomainDKIM                 = resourceDomainDKIM
	ResourceDomainIdentity             = resourceDomainIdentity
	ResourceDomainIdentityVerification = resourceDomainIdentityVerification
	ResourceDomainMailFrom             = resourceDomainMailFrom
	ResourceEmailIdentity              = resourceEmailIdentity
	ResourceEventDestination           = resourceEventDestination
	ResourceIdentityNotificationTopic  = resourceIdentityNotificationTopic
	ResourceIdentityPolicy             = resourceIdentityPolicy
	ResourceReceiptFilter              = resourceReceiptFilter
	ResourceReceiptRule                = resourceReceiptRule
	ResourceReceiptRuleSet             = resourceReceiptRuleSet
	ResourceTemplate                   = resourceTemplate

	FindActiveReceiptRuleSet                       = findActiveReceiptRuleSet
	FindConfigurationSetByName                     = findConfigurationSetByName
	FindEventDestinationByTwoPartKey               = findEventDestinationByTwoPartKey
	FindIdentityDKIMAttributesByIdentity           = findIdentityDKIMAttributesByIdentity
	FindIdentityMailFromDomainAttributesByIdentity = findIdentityMailFromDomainAttributesByIdentity
	FindIdentityNotificationAttributesByIdentity   = findIdentityNotificationAttributesByIdentity
	FindIdentityVerificationAttributesByIdentity   = findIdentityVerificationAttributesByIdentity
	FindIdentityPolicyByTwoPartKey                 = findIdentityPolicyByTwoPartKey
	FindReceiptFilterByName                        = findReceiptFilterByName
	FindReceiptRuleByTwoPartKey                    = findReceiptRuleByTwoPartKey
	FindReceiptRuleSetByName                       = findReceiptRuleSetByName
	FindTemplateByName                             = findTemplateByName
)
