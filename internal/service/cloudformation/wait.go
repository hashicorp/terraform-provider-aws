// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	ChangeSetCreatedTimeout = 5 * time.Minute
)

func WaitChangeSetCreated(ctx context.Context, conn *cloudformation.Client, stackID, changeSetName string) (*cloudformation.DescribeChangeSetOutput, error) {
	stateConf := retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ChangeSetStatusCreateInProgress, awstypes.ChangeSetStatusCreatePending),
		Target:  enum.Slice(awstypes.ChangeSetStatusCreateComplete),
		Timeout: ChangeSetCreatedTimeout,
		Refresh: StatusChangeSet(ctx, conn, stackID, changeSetName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudformation.DescribeChangeSetOutput); ok {
		if output.Status == awstypes.ChangeSetStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func WaitStackSetCreated(ctx context.Context, conn *cloudformation.CloudFormation, name, callAs string, timeout time.Duration) (*cloudformation.StackSet, error) {
	stateConf := retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{cloudformation.StackSetStatusActive},
		Timeout: ChangeSetCreatedTimeout,
		Refresh: StatusStackSet(ctx, conn, name, callAs),
		Delay:   15 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudformation.StackSet); ok {
		if status := aws.StringValue(output.Status); status == cloudformation.ChangeSetStatusFailed {
			tfresource.SetLastError(err, fmt.Errorf("describing CloudFormation Stack Set (%s) results: %w", name, err))
		}

		return output, err
	}

	return nil, err
}

const (
	stackSetOperationDelay = 5 * time.Second
)

func WaitStackSetOperationSucceeded(ctx context.Context, conn *cloudformation.Client, stackSetName, operationID, callAs string, timeout time.Duration) (*awstypes.StackSetOperation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StackSetOperationStatusRunning, awstypes.StackSetOperationStatusQueued),
		Target:  enum.Slice(awstypes.StackSetOperationStatusSucceeded),
		Refresh: StatusStackSetOperation(ctx, conn, stackSetName, operationID, callAs),
		Timeout: timeout,
		Delay:   stackSetOperationDelay,
	}

	outputRaw, waitErr := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.StackSetOperation); ok {
		if output.Status == awstypes.StackSetOperationStatusFailed {
			input := &cloudformation.ListStackSetOperationResultsInput{
				OperationId:  aws.String(operationID),
				StackSetName: aws.String(stackSetName),
				CallAs:       awstypes.CallAs(callAs),
			}
			var summaries []awstypes.StackSetOperationResultSummary

			pages := cloudformation.NewListStackSetOperationResultsPaginator(conn, input)

			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err == nil {
					tfresource.SetLastError(waitErr, fmt.Errorf("Operation (%s) Results: %w", operationID, stackSetOperationError(summaries)))
				} else {
					tfresource.SetLastError(waitErr, fmt.Errorf("listing CloudFormation Stack Set (%s) Operation (%s) results: %w", stackSetName, operationID, err))
				}

				summaries = append(summaries, page.Summaries...)
			}
		}

		return output, waitErr
	}

	return nil, waitErr
}

const (
	TypeRegistrationTimeout = 5 * time.Minute
)

func WaitTypeRegistrationProgressStatusComplete(ctx context.Context, conn *cloudformation.Client, registrationToken string) (*cloudformation.DescribeTypeRegistrationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RegistrationStatusInProgress),
		Target:  enum.Slice(awstypes.RegistrationStatusComplete),
		Refresh: StatusTypeRegistrationProgress(ctx, conn, registrationToken),
		Timeout: TypeRegistrationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudformation.DescribeTypeRegistrationOutput); ok {
		return output, err
	}

	return nil, err
}
