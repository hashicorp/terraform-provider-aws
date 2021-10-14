package waiter

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

const (
	ScalingPlanCreatedTimeout = 5 * time.Minute
	ScalingPlanDeletedTimeout = 5 * time.Minute
	ScalingPlanUpdatedTimeout = 5 * time.Minute
)

func ScalingPlanCreated(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeCreationInProgress},
		Target:  []string{autoscalingplans.ScalingPlanStatusCodeActive, autoscalingplans.ScalingPlanStatusCodeActiveWithProblems},
		Refresh: ScalingPlanStatusCode(conn, scalingPlanName, scalingPlanVersion),
		Timeout: ScalingPlanCreatedTimeout,
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

func ScalingPlanDeleted(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeDeletionInProgress},
		Target:  []string{},
		Refresh: ScalingPlanStatusCode(conn, scalingPlanName, scalingPlanVersion),
		Timeout: ScalingPlanDeletedTimeout,
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

func ScalingPlanUpdated(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeUpdateInProgress},
		Target:  []string{autoscalingplans.ScalingPlanStatusCodeActive, autoscalingplans.ScalingPlanStatusCodeActiveWithProblems},
		Refresh: ScalingPlanStatusCode(conn, scalingPlanName, scalingPlanVersion),
		Timeout: ScalingPlanUpdatedTimeout,
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
