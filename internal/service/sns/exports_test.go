// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sns

// Exports for use in tests only.
var (
	ResourcePlatformApplication       = resourcePlatformApplication
	ResourceTopic                     = resourceTopic
	ResourceTopicDataProtectionPolicy = resourceTopicDataProtectionPolicy
	ResourceTopicPolicy               = resourceTopicPolicy
	ResourceTopicSubscription         = resourceTopicSubscription

	FindDataProtectionPolicyByARN                  = findDataProtectionPolicyByARN
	FindPlatformApplicationAttributesByARN         = findPlatformApplicationAttributesByARN
	FindSubscriptionAttributesByARN                = findSubscriptionAttributesByARN
	FindTopicAttributesByARN                       = findTopicAttributesByARN
	FindTopicAttributesWithValidAWSPrincipalsByARN = findTopicAttributesWithValidAWSPrincipalsByARN // nosemgrep:ci.aws-in-var-name

	FIFOTopicNameSuffix                = fifoTopicNameSuffix
	ParsePlatformApplicationResourceID = parsePlatformApplicationResourceID
	TopicAttributeNameDeliveryPolicy   = topicAttributeNameDeliveryPolicy
	TopicAttributeNamePolicy           = topicAttributeNamePolicy

	SubscriptionProtocolApplication = subscriptionProtocolApplication
	SubscriptionProtocolHTTP        = subscriptionProtocolHTTP
	SubscriptionProtocolHTTPS       = subscriptionProtocolHTTPS
	SubscriptionProtocolEmail       = subscriptionProtocolEmail
	SubscriptionProtocolEmailJSON   = subscriptionProtocolEmailJSON
	WaitForConfirmation             = waitForConfirmation
)
