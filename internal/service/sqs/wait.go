// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

	queueAttributeReadTimeout = 20 * time.Second

	queueStateExists = "exists"

	queueAttributeStateNotEqual = "notequal"
	queueAttributeStateEqual    = "equal"
)

func waitQueueAttributesPropagated(ctx context.Context, conn *sqs.SQS, url string, expected map[string]string) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{queueAttributeStateNotEqual},
		Target:                    []string{queueAttributeStateEqual},
		Refresh:                   statusQueueAttributeState(ctx, conn, url, expected),
		Timeout:                   queueAttributePropagationTimeout,
		ContinuousTargetOccurence: 6,               // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
		MinTimeout:                5 * time.Second, // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
		NotFoundChecks:            10,              // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitQueueDeleted(ctx context.Context, conn *sqs.SQS, url string) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{queueStateExists},
		Target:                    []string{},
		Refresh:                   statusQueueState(ctx, conn, url),
		Timeout:                   queueDeletedTimeout,
		ContinuousTargetOccurence: 15,              // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
		MinTimeout:                3 * time.Second, // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
		NotFoundChecks:            5,               // set to accommodate GovCloud, commercial, China, etc. - avoid lowering
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
