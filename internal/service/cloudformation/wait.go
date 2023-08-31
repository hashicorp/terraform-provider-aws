// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	ChangeSetCreatedTimeout = 5 * time.Minute
)

func WaitChangeSetCreated(ctx context.Context, conn *cloudformation.CloudFormation, stackID, changeSetName string) (*cloudformation.DescribeChangeSetOutput, error) {
	stateConf := retry.StateChangeConf{
		Pending: []string{cloudformation.ChangeSetStatusCreateInProgress, cloudformation.ChangeSetStatusCreatePending},
		Target:  []string{cloudformation.ChangeSetStatusCreateComplete},
		Timeout: ChangeSetCreatedTimeout,
		Refresh: StatusChangeSet(ctx, conn, stackID, changeSetName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudformation.DescribeChangeSetOutput); ok {
		if status := aws.StringValue(output.Status); status == cloudformation.ChangeSetStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

const (
	// Default maximum amount of time to wait for a StackSetInstance to be Created
	StackSetInstanceCreatedDefaultTimeout = 30 * time.Minute

	// Default maximum amount of time to wait for a StackSetInstance to be Updated
	StackSetInstanceUpdatedDefaultTimeout = 30 * time.Minute

	// Default maximum amount of time to wait for a StackSetInstance to be Deleted
	StackSetInstanceDeletedDefaultTimeout = 30 * time.Minute

	stackSetOperationDelay = 5 * time.Second
)

func WaitStackSetOperationSucceeded(ctx context.Context, conn *cloudformation.CloudFormation, stackSetName, operationID, callAs string, timeout time.Duration) (*cloudformation.StackSetOperation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{cloudformation.StackSetOperationStatusRunning, cloudformation.StackSetOperationStatusQueued},
		Target:  []string{cloudformation.StackSetOperationStatusSucceeded},
		Refresh: StatusStackSetOperation(ctx, conn, stackSetName, operationID, callAs),
		Timeout: timeout,
		Delay:   stackSetOperationDelay,
	}

	outputRaw, waitErr := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudformation.StackSetOperation); ok {
		if status := aws.StringValue(output.Status); status == cloudformation.StackSetOperationStatusFailed {
			input := &cloudformation.ListStackSetOperationResultsInput{
				OperationId:  aws.String(operationID),
				StackSetName: aws.String(stackSetName),
				CallAs:       aws.String(callAs),
			}
			var summaries []*cloudformation.StackSetOperationResultSummary

			listErr := conn.ListStackSetOperationResultsPagesWithContext(ctx, input, func(page *cloudformation.ListStackSetOperationResultsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				summaries = append(summaries, page.Summaries...)

				return !lastPage
			})

			if listErr == nil {
				tfresource.SetLastError(waitErr, fmt.Errorf("Operation (%s) Results: %w", operationID, StackSetOperationError(summaries)))
			} else {
				tfresource.SetLastError(waitErr, fmt.Errorf("listing CloudFormation Stack Set (%s) Operation (%s) results: %w", stackSetName, operationID, listErr))
			}
		}

		return output, waitErr
	}

	return nil, waitErr
}

const (
	TypeRegistrationTimeout = 5 * time.Minute
)

func WaitTypeRegistrationProgressStatusComplete(ctx context.Context, conn *cloudformation.CloudFormation, registrationToken string) (*cloudformation.DescribeTypeRegistrationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{cloudformation.RegistrationStatusInProgress},
		Target:  []string{cloudformation.RegistrationStatusComplete},
		Refresh: StatusTypeRegistrationProgress(ctx, conn, registrationToken),
		Timeout: TypeRegistrationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudformation.DescribeTypeRegistrationOutput); ok {
		return output, err
	}

	return nil, err
}
