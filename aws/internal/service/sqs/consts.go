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

const (
	DeduplicationScopeMessageGroup = "messageGroup"
	DeduplicationScopeQueue        = "queue"
)

func DeduplicationScope_Values() []string {
	return []string{
		DeduplicationScopeMessageGroup,
		DeduplicationScopeQueue,
	}
}

const (
	FifoThroughputLimitPerMessageGroupId = "perMessageGroupId"
	FifoThroughputLimitPerQueue          = "perQueue"
)

func FifoThroughputLimit_Values() []string {
	return []string{
		FifoThroughputLimitPerMessageGroupId,
		FifoThroughputLimitPerQueue,
	}
}
