package waiter

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sqs/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
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
)

func QueueDeleted(conn *sqs.SQS, url string) error {
	err := resource.Retry(QueueDeletedTimeout, func() *resource.RetryError {
		var err error

		_, err = finder.QueueAttributesByURL(conn, url)

		if tfresource.NotFound(err) {
			return nil
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return resource.RetryableError(fmt.Errorf("SQS Queue (%s) still exists", url))
	})

	if tfresource.TimedOut(err) {
		_, err = finder.QueueAttributesByURL(conn, url)

		if tfresource.NotFound(err) {
			return nil
		}
	}

	if err != nil {
		return err
	}

	return nil
}
