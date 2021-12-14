package sqs

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
	queueAttributePropagationTimeout = 2 * time.Minute

	// If you delete a queue, you must wait at least 60 seconds before creating a queue with the same name.
	// ReferenceL https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_CreateQueue.html
	queueCreatedTimeout = 70 * time.Second
	queueReadTimeout    = 20 * time.Second
	queueDeletedTimeout = 3 * time.Minute
	queueTagsTimeout    = 60 * time.Second

	queuePolicyReadTimeout = 20 * time.Second

	queueStateExists = "exists"

	queuePolicyStateNotEqual = "notequal"
	queuePolicyStateEqual    = "equal"
)

func waitQueueAttributesPropagated(conn *sqs.SQS, url string, expected map[string]string) error {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{queuePolicyStateNotEqual},
		Target:                    []string{queuePolicyStateEqual},
		Refresh:                   statusQueueAttributeState(conn, url, expected),
		Timeout:                   queueAttributePropagationTimeout,
		ContinuousTargetOccurence: 6,               // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
		MinTimeout:                5 * time.Second, // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
		NotFoundChecks:            10,              // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitQueueDeleted(conn *sqs.SQS, url string) error {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{queueStateExists},
		Target:                    []string{},
		Refresh:                   statusQueueState(conn, url),
		Timeout:                   queueDeletedTimeout,
		ContinuousTargetOccurence: 15,              // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
		MinTimeout:                3 * time.Second, // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
		NotFoundChecks:            5,               // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
	}

	_, err := stateConf.WaitForState()

	return err
}
