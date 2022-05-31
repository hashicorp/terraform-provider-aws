package sns

import "time"

const (
	FIFOTopicNameSuffix = ".fifo"
)

const (
	SubscriptionProtocolApplication = "application"
	SubscriptionProtocolEmail       = "email"
	SubscriptionProtocolEmailJSON   = "email-json"
	SubscriptionProtocolFirehose    = "firehose"
	SubscriptionProtocolHTTP        = "http"
	SubscriptionProtocolHTTPS       = "https"
	SubscriptionProtocolLambda      = "lambda"
	SubscriptionProtocolSMS         = "sms"
	SubscriptionProtocolSQS         = "sqs"
)

func SubscriptionProtocol_Values() []string {
	return []string{
		SubscriptionProtocolApplication,
		SubscriptionProtocolEmail,
		SubscriptionProtocolEmailJSON,
		SubscriptionProtocolFirehose,
		SubscriptionProtocolHTTP,
		SubscriptionProtocolHTTPS,
		SubscriptionProtocolLambda,
		SubscriptionProtocolSMS,
		SubscriptionProtocolSQS,
	}
}

const (
	SubscriptionAttributeNameConfirmationWasAuthenticated = "ConfirmationWasAuthenticated"
	SubscriptionAttributeNameDeliveryPolicy               = "DeliveryPolicy"
	SubscriptionAttributeNameEndpoint                     = "Endpoint"
	SubscriptionAttributeNameFilterPolicy                 = "FilterPolicy"
	SubscriptionAttributeNameOwner                        = "Owner"
	SubscriptionAttributeNamePendingConfirmation          = "PendingConfirmation"
	SubscriptionAttributeNameProtocol                     = "Protocol"
	SubscriptionAttributeNameRawMessageDelivery           = "RawMessageDelivery"
	SubscriptionAttributeNameRedrivePolicy                = "RedrivePolicy"
	SubscriptionAttributeNameSubscriptionARN              = "SubscriptionArn"
	SubscriptionAttributeNameSubscriptionRoleARN          = "SubscriptionRoleArn"
	SubscriptionAttributeNameTopicARN                     = "TopicArn"
)

const (
	TopicAttributeNameApplicationFailureFeedbackRoleARN    = "ApplicationFailureFeedbackRoleArn"
	TopicAttributeNameApplicationSuccessFeedbackRoleARN    = "ApplicationSuccessFeedbackRoleArn"
	TopicAttributeNameApplicationSuccessFeedbackSampleRate = "ApplicationSuccessFeedbackSampleRate"
	TopicAttributeNameContentBasedDeduplication            = "ContentBasedDeduplication"
	TopicAttributeNameDeliveryPolicy                       = "DeliveryPolicy"
	TopicAttributeNameDisplayName                          = "DisplayName"
	TopicAttributeNameFIFOTopic                            = "FifoTopic"
	TopicAttributeNameFirehoseFailureFeedbackRoleARN       = "FirehoseFailureFeedbackRoleArn"
	TopicAttributeNameFirehoseSuccessFeedbackRoleARN       = "FirehoseSuccessFeedbackRoleArn"
	TopicAttributeNameFirehoseSuccessFeedbackSampleRate    = "FirehoseSuccessFeedbackSampleRate"
	TopicAttributeNameHTTPFailureFeedbackRoleARN           = "HTTPFailureFeedbackRoleArn"
	TopicAttributeNameHTTPSuccessFeedbackRoleARN           = "HTTPSuccessFeedbackRoleArn"
	TopicAttributeNameHTTPSuccessFeedbackSampleRate        = "HTTPSuccessFeedbackSampleRate"
	TopicAttributeNameKMSMasterKeyId                       = "KmsMasterKeyId"
	TopicAttributeNameLambdaFailureFeedbackRoleARN         = "LambdaFailureFeedbackRoleArn"
	TopicAttributeNameLambdaSuccessFeedbackRoleARN         = "LambdaSuccessFeedbackRoleArn"
	TopicAttributeNameLambdaSuccessFeedbackSampleRate      = "LambdaSuccessFeedbackSampleRate"
	TopicAttributeNameOwner                                = "Owner"
	TopicAttributeNamePolicy                               = "Policy"
	TopicAttributeNameSQSFailureFeedbackRoleARN            = "SQSFailureFeedbackRoleArn"
	TopicAttributeNameSQSSuccessFeedbackRoleARN            = "SQSSuccessFeedbackRoleArn"
	TopicAttributeNameSQSSuccessFeedbackSampleRate         = "SQSSuccessFeedbackSampleRate"
	TopicAttributeNameTopicARN                             = "TopicArn"
)

const (
	propagationTimeout = 2 * time.Minute
)
