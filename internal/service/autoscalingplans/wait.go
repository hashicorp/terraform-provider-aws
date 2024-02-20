// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscalingplans

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	scalingPlanCreatedTimeout = 5 * time.Minute
	scalingPlanDeletedTimeout = 5 * time.Minute
	scalingPlanUpdatedTimeout = 5 * time.Minute
)

func waitScalingPlanCreated(ctx context.Context, conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeCreationInProgress},
		Target:  []string{autoscalingplans.ScalingPlanStatusCodeActive, autoscalingplans.ScalingPlanStatusCodeActiveWithProblems},
		Refresh: statusScalingPlanCode(ctx, conn, scalingPlanName, scalingPlanVersion),
		Timeout: scalingPlanCreatedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*autoscalingplans.ScalingPlan); ok {
		if statusCode := aws.StringValue(output.StatusCode); statusCode == autoscalingplans.ScalingPlanStatusCodeCreationFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitScalingPlanDeleted(ctx context.Context, conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeDeletionInProgress},
		Target:  []string{},
		Refresh: statusScalingPlanCode(ctx, conn, scalingPlanName, scalingPlanVersion),
		Timeout: scalingPlanDeletedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*autoscalingplans.ScalingPlan); ok {
		if statusCode := aws.StringValue(output.StatusCode); statusCode == autoscalingplans.ScalingPlanStatusCodeDeletionFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitScalingPlanUpdated(ctx context.Context, conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeUpdateInProgress},
		Target:  []string{autoscalingplans.ScalingPlanStatusCodeActive, autoscalingplans.ScalingPlanStatusCodeActiveWithProblems},
		Refresh: statusScalingPlanCode(ctx, conn, scalingPlanName, scalingPlanVersion),
		Timeout: scalingPlanUpdatedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*autoscalingplans.ScalingPlan); ok {
		if statusCode := aws.StringValue(output.StatusCode); statusCode == autoscalingplans.ScalingPlanStatusCodeUpdateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}
