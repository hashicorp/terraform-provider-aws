package autoscalingplans

import (
	"time"

	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a ScalingPlan to return Created
	scalingPlanCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a ScalingPlan to return Deleted
	scalingPlanDeletedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a ScalingPlan to return Updated
	scalingPlanUpdatedTimeout = 5 * time.Minute
)

// waitScalingPlanCreated waits for a ScalingPlan to return Created
func waitScalingPlanCreated(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeCreationInProgress},
		Target:  []string{autoscalingplans.ScalingPlanStatusCodeActive, autoscalingplans.ScalingPlanStatusCodeActiveWithProblems},
		Refresh: statusScalingPlan(conn, scalingPlanName, scalingPlanVersion),
		Timeout: scalingPlanCreatedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*autoscalingplans.ScalingPlan); ok {
		return v, err
	}

	return nil, err
}

// waitScalingPlanDeleted waits for a ScalingPlan to return Deleted
func waitScalingPlanDeleted(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeDeletionInProgress},
		Target:  []string{},
		Refresh: statusScalingPlan(conn, scalingPlanName, scalingPlanVersion),
		Timeout: scalingPlanDeletedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*autoscalingplans.ScalingPlan); ok {
		return v, err
	}

	return nil, err
}

// waitScalingPlanUpdated waits for a ScalingPlan to return Updated
func waitScalingPlanUpdated(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeUpdateInProgress},
		Target:  []string{autoscalingplans.ScalingPlanStatusCodeActive, autoscalingplans.ScalingPlanStatusCodeActiveWithProblems},
		Refresh: statusScalingPlan(conn, scalingPlanName, scalingPlanVersion),
		Timeout: scalingPlanUpdatedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*autoscalingplans.ScalingPlan); ok {
		return v, err
	}

	return nil, err
}
