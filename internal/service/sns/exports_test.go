// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns

// Exports for use in tests only.
var (
	ResourcePlatformApplication       = resourcePlatformApplication
	ResourceTopic                     = resourceTopic
	ResourceTopicDataProtectionPolicy = resourceTopicDataProtectionPolicy
	ResourceTopicPolicy               = resourceTopicPolicy
	ResourceTopicSubscription         = resourceTopicSubscription

	FindPlatformApplicationAttributesByARN         = findPlatformApplicationAttributesByARN
	FindSubscriptionAttributesByARN                = findSubscriptionAttributesByARN
	FindTopicAttributesByARN                       = findTopicAttributesByARN
	FindTopicAttributesWithValidAWSPrincipalsByARN = findTopicAttributesWithValidAWSPrincipalsByARN

	FIFOTopicNameSuffix                = fifoTopicNameSuffix
	ParsePlatformApplicationResourceID = parsePlatformApplicationResourceID
	TopicAttributeNameDeliveryPolicy   = topicAttributeNameDeliveryPolicy
	TopicAttributeNamePolicy           = topicAttributeNamePolicy
)
