package sqs

const (
	FifoQueueNameSuffix = ".fifo"
)

const (
	DefaultQueueDelaySeconds                  = 0
	DefaultQueueMaximumMessageSize            = 262144
	DefaultQueueMessageRetentionPeriod        = 345600
	DefaultQueueReceiveMessageWaitTimeSeconds = 0
	DefaultQueueVisibilityTimeout             = 30
)
