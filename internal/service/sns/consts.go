// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"time"

	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	fifoTopicNameSuffix = ".fifo"
)

const (
	platformApplicationAttributeNameAppleCertificateExpiryDate = "AppleCertificateExpiryDate"
	platformApplicationAttributeNameApplePlatformBundleID      = "ApplePlatformBundleID"
	platformApplicationAttributeNameApplePlatformTeamID        = "ApplePlatformTeamID"
	platformApplicationAttributeNameEventDeliveryFailure       = "EventDeliveryFailure"
	platformApplicationAttributeNameEventEndpointCreated       = "EventEndpointCreated"
	platformApplicationAttributeNameEventEndpointDeleted       = "EventEndpointDeleted"
	platformApplicationAttributeNameEventEndpointUpdated       = "EventEndpointUpdated"
	platformApplicationAttributeNameFailureFeedbackRoleARN     = "FailureFeedbackRoleArn"
	platformApplicationAttributeNamePlatformCredential         = "PlatformCredential"
	platformApplicationAttributeNamePlatformPrincipal          = "PlatformPrincipal"
	platformApplicationAttributeNameSuccessFeedbackRoleARN     = "SuccessFeedbackRoleArn"
	platformApplicationAttributeNameSuccessFeedbackSampleRate  = "SuccessFeedbackSampleRate"
)

/*
const (
	platfomAPNS        = "APNS"
	platfomAPNSSandbox = "APNS_SANDBOX"
	platfomGCM         = "GCM"
)
*/

const (
	subscriptionProtocolApplication = "application"
	subscriptionProtocolEmail       = names.AttrEmail
	subscriptionProtocolEmailJSON   = "email-json"
	subscriptionProtocolFirehose    = "firehose"
	subscriptionProtocolHTTP        = "http"
	subscriptionProtocolHTTPS       = "https"
	subscriptionProtocolLambda      = "lambda"
	subscriptionProtocolSMS         = "sms"
	subscriptionProtocolSQS         = "sqs"
)

func subscriptionProtocol_Values() []string {
	return []string{
		subscriptionProtocolApplication,
		names.AttrEmail,
		subscriptionProtocolEmailJSON,
		subscriptionProtocolFirehose,
		subscriptionProtocolHTTP,
		subscriptionProtocolHTTPS,
		subscriptionProtocolLambda,
		subscriptionProtocolSMS,
		subscriptionProtocolSQS,
	}
}

const (
	subscriptionAttributeNameConfirmationWasAuthenticated = "ConfirmationWasAuthenticated"
	subscriptionAttributeNameDeliveryPolicy               = "DeliveryPolicy"
	subscriptionAttributeNameEndpoint                     = "Endpoint"
	subscriptionAttributeNameFilterPolicy                 = "FilterPolicy"
	subscriptionAttributeNameFilterPolicyScope            = "FilterPolicyScope"
	subscriptionAttributeNameOwner                        = "Owner"
	subscriptionAttributeNamePendingConfirmation          = "PendingConfirmation"
	subscriptionAttributeNameProtocol                     = "Protocol"
	subscriptionAttributeNameRawMessageDelivery           = "RawMessageDelivery"
	subscriptionAttributeNameRedrivePolicy                = "RedrivePolicy"
	subscriptionAttributeNameReplayPolicy                 = "ReplayPolicy"
	subscriptionAttributeNameSubscriptionARN              = "SubscriptionArn"
	subscriptionAttributeNameSubscriptionRoleARN          = "SubscriptionRoleArn"
	subscriptionAttributeNameTopicARN                     = "TopicArn"
)

const (
	topicAttributeNameApplicationFailureFeedbackRoleARN    = "ApplicationFailureFeedbackRoleArn"
	topicAttributeNameApplicationSuccessFeedbackRoleARN    = "ApplicationSuccessFeedbackRoleArn"
	topicAttributeNameApplicationSuccessFeedbackSampleRate = "ApplicationSuccessFeedbackSampleRate"
	topicAttributeNameArchivePolicy                        = "ArchivePolicy"
	topicAttributeNameBeginningArchiveTime                 = "BeginningArchiveTime"
	topicAttributeNameContentBasedDeduplication            = "ContentBasedDeduplication"
	topicAttributeNameDeliveryPolicy                       = "DeliveryPolicy"
	topicAttributeNameDisplayName                          = "DisplayName"
	topicAttributeNameFIFOTopic                            = "FifoTopic"
	topicAttributeNameFirehoseFailureFeedbackRoleARN       = "FirehoseFailureFeedbackRoleArn"
	topicAttributeNameFirehoseSuccessFeedbackRoleARN       = "FirehoseSuccessFeedbackRoleArn"
	topicAttributeNameFirehoseSuccessFeedbackSampleRate    = "FirehoseSuccessFeedbackSampleRate"
	topicAttributeNameHTTPFailureFeedbackRoleARN           = "HTTPFailureFeedbackRoleArn"
	topicAttributeNameHTTPSuccessFeedbackRoleARN           = "HTTPSuccessFeedbackRoleArn"
	topicAttributeNameHTTPSuccessFeedbackSampleRate        = "HTTPSuccessFeedbackSampleRate"
	topicAttributeNameKMSMasterKeyId                       = "KmsMasterKeyId"
	topicAttributeNameLambdaFailureFeedbackRoleARN         = "LambdaFailureFeedbackRoleArn"
	topicAttributeNameLambdaSuccessFeedbackRoleARN         = "LambdaSuccessFeedbackRoleArn"
	topicAttributeNameLambdaSuccessFeedbackSampleRate      = "LambdaSuccessFeedbackSampleRate"
	topicAttributeNameOwner                                = "Owner"
	topicAttributeNamePolicy                               = "Policy"
	topicAttributeNameSignatureVersion                     = "SignatureVersion"
	topicAttributeNameSQSFailureFeedbackRoleARN            = "SQSFailureFeedbackRoleArn"
	topicAttributeNameSQSSuccessFeedbackRoleARN            = "SQSSuccessFeedbackRoleArn"
	topicAttributeNameSQSSuccessFeedbackSampleRate         = "SQSSuccessFeedbackSampleRate"
	topicAttributeNameTopicARN                             = "TopicArn"
	topicAttributeNameTracingConfig                        = "TracingConfig"
)

const (
	propagationTimeout = 2 * time.Minute
)

const (
	subscriptionFilterPolicyScopeMessageAttributes = "MessageAttributes"
	subscriptionFilterPolicyScopeMessageBody       = "MessageBody"
)

func subscriptionFilterPolicyScope_Values() []string {
	return []string{
		subscriptionFilterPolicyScopeMessageAttributes,
		subscriptionFilterPolicyScopeMessageBody,
	}
}

const (
	topicTracingConfigActive      = "Active"
	topicTracingConfigPassThrough = "PassThrough"
)

func topicTracingConfig_Values() []string {
	return []string{
		topicTracingConfigActive,
		topicTracingConfigPassThrough,
	}
}
