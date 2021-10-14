package sqs

const (
	ErrCodeInvalidAction = "InvalidAction"
)

const (
	FIFOQueueNameSuffix = ".fifo"
)

const (
	DefaultQueueDelaySeconds                  = 0
	DefaultQueueKMSDataKeyReusePeriodSeconds  = 300
	DefaultQueueMaximumMessageSize            = 262_144 // 256 KiB.
	DefaultQueueMessageRetentionPeriod        = 345_600 // 4 days.
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
	FIFOThroughputLimitPerMessageGroupID = "perMessageGroupId"
	FIFOThroughputLimitPerQueue          = "perQueue"
)

func FIFOThroughputLimit_Values() []string {
	return []string{
		FIFOThroughputLimitPerMessageGroupID,
		FIFOThroughputLimitPerQueue,
	}
}
