// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscalingplans

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscalingplans"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscalingplans/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	scalingPlanCreatedTimeout = 5 * time.Minute
	scalingPlanDeletedTimeout = 5 * time.Minute
	scalingPlanUpdatedTimeout = 5 * time.Minute
)

func waitScalingPlanCreated(ctx context.Context, conn *autoscalingplans.Client, scalingPlanName string, scalingPlanVersion int) (*awstypes.ScalingPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScalingPlanStatusCodeCreationInProgress),
		Target:  enum.Slice(awstypes.ScalingPlanStatusCodeActive, awstypes.ScalingPlanStatusCodeActiveWithProblems),
		Refresh: statusScalingPlanCode(ctx, conn, scalingPlanName, scalingPlanVersion),
		Timeout: scalingPlanCreatedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ScalingPlan); ok {
		if output.StatusCode == awstypes.ScalingPlanStatusCodeCreationFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitScalingPlanDeleted(ctx context.Context, conn *autoscalingplans.Client, scalingPlanName string, scalingPlanVersion int) (*awstypes.ScalingPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScalingPlanStatusCodeDeletionInProgress),
		Target:  []string{},
		Refresh: statusScalingPlanCode(ctx, conn, scalingPlanName, scalingPlanVersion),
		Timeout: scalingPlanDeletedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ScalingPlan); ok {
		if output.StatusCode == awstypes.ScalingPlanStatusCodeDeletionFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitScalingPlanUpdated(ctx context.Context, conn *autoscalingplans.Client, scalingPlanName string, scalingPlanVersion int) (*awstypes.ScalingPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScalingPlanStatusCodeUpdateInProgress),
		Target:  enum.Slice(awstypes.ScalingPlanStatusCodeActive, awstypes.ScalingPlanStatusCodeActiveWithProblems),
		Refresh: statusScalingPlanCode(ctx, conn, scalingPlanName, scalingPlanVersion),
		Timeout: scalingPlanUpdatedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ScalingPlan); ok {
		if output.StatusCode == awstypes.ScalingPlanStatusCodeUpdateFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}
