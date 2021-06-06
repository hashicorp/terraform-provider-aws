package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for SQS queue attribute changes to propagate
	// This timeout should not be increased without strong consideration
	// as this will negatively impact user experience when configurations
	// have incorrect references or permissions.
	// Reference: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_SetQueueAttributes.html
	QueueAttributePropagationTimeout = 1 * time.Minute

	// If you delete a queue, you must wait at least 60 seconds before creating a queue with the same name.
	// ReferenceL https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_CreateQueue.html
	QueueCreatedTimeout = 70 * time.Second

	QueueDeletedTimeout = 15 * time.Second

	queueStateExists = "exists"
)

func QueueDeleted(conn *sqs.SQS, url string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{queueStateExists},
		Target:  []string{},
		Refresh: QueueState(conn, url),
		Timeout: QueueDeletedTimeout,

		ContinuousTargetOccurence: 3,
	}

	_, err := stateConf.WaitForState()

	return err
}
