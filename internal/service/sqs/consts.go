// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs

import "time"

const (
	propagationTimeout = 1 * time.Minute
)

const (
	fifoQueueNameSuffix = ".fifo"
)

const (
	defaultQueueDelaySeconds                  = 0
	defaultQueueKMSDataKeyReusePeriodSeconds  = 300
	defaultQueueMaximumMessageSize            = 262_144 // 256 KiB.
	defaultQueueMessageRetentionPeriod        = 345_600 // 4 days.
	defaultQueueReceiveMessageWaitTimeSeconds = 0
	defaultQueueVisibilityTimeout             = 30
)

const (
	deduplicationScopeMessageGroup = "messageGroup"
	deduplicationScopeQueue        = "queue"
)

func deduplicationScope_Values() []string {
	return []string{
		deduplicationScopeMessageGroup,
		deduplicationScopeQueue,
	}
}

const (
	fifoThroughputLimitPerMessageGroupID = "perMessageGroupId"
	fifoThroughputLimitPerQueue          = "perQueue"
)

func fifoThroughputLimit_Values() []string {
	return []string{
		fifoThroughputLimitPerMessageGroupID,
		fifoThroughputLimitPerQueue,
	}
}

const (
	errCodeQueueDoesNotExist     = "AWS.SimpleQueueService.NonExistentQueue"
	errCodeQueueDeletedRecently  = "AWS.SimpleQueueService.QueueDeletedRecently"
	errCodeInvalidAttributeValue = "InvalidAttributeValue"
)
