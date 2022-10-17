package autoscalingplans

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	scalingPlanCreatedTimeout = 5 * time.Minute
	scalingPlanDeletedTimeout = 5 * time.Minute
	scalingPlanUpdatedTimeout = 5 * time.Minute
)

func waitScalingPlanCreated(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeCreationInProgress},
		Target:  []string{autoscalingplans.ScalingPlanStatusCodeActive, autoscalingplans.ScalingPlanStatusCodeActiveWithProblems},
		Refresh: statusScalingPlanCode(conn, scalingPlanName, scalingPlanVersion),
		Timeout: scalingPlanCreatedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*autoscalingplans.ScalingPlan); ok {
		if statusCode := aws.StringValue(output.StatusCode); statusCode == autoscalingplans.ScalingPlanStatusCodeCreationFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitScalingPlanDeleted(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeDeletionInProgress},
		Target:  []string{},
		Refresh: statusScalingPlanCode(conn, scalingPlanName, scalingPlanVersion),
		Timeout: scalingPlanDeletedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*autoscalingplans.ScalingPlan); ok {
		if statusCode := aws.StringValue(output.StatusCode); statusCode == autoscalingplans.ScalingPlanStatusCodeDeletionFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitScalingPlanUpdated(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeUpdateInProgress},
		Target:  []string{autoscalingplans.ScalingPlanStatusCodeActive, autoscalingplans.ScalingPlanStatusCodeActiveWithProblems},
		Refresh: statusScalingPlanCode(conn, scalingPlanName, scalingPlanVersion),
		Timeout: scalingPlanUpdatedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*autoscalingplans.ScalingPlan); ok {
		if statusCode := aws.StringValue(output.StatusCode); statusCode == autoscalingplans.ScalingPlanStatusCodeUpdateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}
