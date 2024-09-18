// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

// Exports for use in tests only.
var (
	ResourceIdentityNotificationTopic = resourceIdentityNotificationTopic
	ResourceIdentityPolicy            = resourceIdentityPolicy
	ResourceReceiptFilter             = resourceReceiptFilter
	ResourceReceiptRule               = resourceReceiptRule
	ResourceReceiptRuleSet            = resourceReceiptRuleSet
	ResourceTemplate                  = resourceTemplate

	FindIdentityNotificationAttributesByIdentity = findIdentityNotificationAttributesByIdentity
	FindIdentityPolicyByTwoPartKey               = findIdentityPolicyByTwoPartKey
	FindReceiptFilterByName                      = findReceiptFilterByName
	FindReceiptRuleByTwoPartKey                  = findReceiptRuleByTwoPartKey
	FindReceiptRuleSetByName                     = findReceiptRuleSetByName
	FindTemplateByName                           = findTemplateByName
)
