package sns

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
	SubscriptionAttributeNameDeliveryPolicy      = "DeliveryPolicy"
	SubscriptionAttributeNameFilterPolicy        = "FilterPolicy"
	SubscriptionAttributeNameProtocol            = "Protocol"
	SubscriptionAttributeNameRawMessageDelivery  = "RawMessageDelivery"
	SubscriptionAttributeNameRedrivePolicy       = "RedrivePolicy"
	SubscriptionAttributeNameSubscriptionRoleArn = "SubscriptionRoleArn"
)

const (
	TopicAttributeNameApplicationFailureFeedbackRoleArn    = "ApplicationFailureFeedbackRoleArn"
	TopicAttributeNameApplicationSuccessFeedbackRoleArn    = "ApplicationSuccessFeedbackRoleArn"
	TopicAttributeNameApplicationSuccessFeedbackSampleRate = "ApplicationSuccessFeedbackSampleRate"
	TopicAttributeNameContentBasedDeduplication            = "ContentBasedDeduplication"
	TopicAttributeNameDeliveryPolicy                       = "DeliveryPolicy"
	TopicAttributeNameDisplayName                          = "DisplayName"
	TopicAttributeNameFifoTopic                            = "FifoTopic"
	TopicAttributeNameFirehoseFailureFeedbackRoleArn       = "FirehoseFailureFeedbackRoleArn"
	TopicAttributeNameFirehoseSuccessFeedbackRoleArn       = "FirehoseSuccessFeedbackRoleArn"
	TopicAttributeNameFirehoseSuccessFeedbackSampleRate    = "FirehoseSuccessFeedbackSampleRate"
	TopicAttributeNameHTTPFailureFeedbackRoleArn           = "HTTPFailureFeedbackRoleArn"
	TopicAttributeNameHTTPSuccessFeedbackRoleArn           = "HTTPSuccessFeedbackRoleArn"
	TopicAttributeNameHTTPSuccessFeedbackSampleRate        = "HTTPSuccessFeedbackSampleRate"
	TopicAttributeNameKmsMasterKeyId                       = "KmsMasterKeyId"
	TopicAttributeNameLambdaFailureFeedbackRoleArn         = "LambdaFailureFeedbackRoleArn"
	TopicAttributeNameLambdaSuccessFeedbackRoleArn         = "LambdaSuccessFeedbackRoleArn"
	TopicAttributeNameLambdaSuccessFeedbackSampleRate      = "LambdaSuccessFeedbackSampleRate"
	TopicAttributeNameOwner                                = "Owner"
	TopicAttributeNamePolicy                               = "Policy"
	TopicAttributeNameSQSFailureFeedbackRoleArn            = "SQSFailureFeedbackRoleArn"
	TopicAttributeNameSQSSuccessFeedbackRoleArn            = "SQSSuccessFeedbackRoleArn"
	TopicAttributeNameSQSSuccessFeedbackSampleRate         = "SQSSuccessFeedbackSampleRate"
	TopicAttributeNameTopicArn                             = "TopicArn"
)
