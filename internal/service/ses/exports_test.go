// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

// Exports for use in tests only.
var (
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
