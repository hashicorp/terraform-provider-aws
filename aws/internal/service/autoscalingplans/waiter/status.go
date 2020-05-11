package waiter

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	// ScalingPlan NotFound
	ScalingPlanStatusNotFound = "NotFound"

	// ScalingPlan Unknown
	ScalingPlanStatusUnknown = "Unknown"
)

// ScalingPlanStatus fetches the ScalingPlan and its Status
func ScalingPlanStatus(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &autoscalingplans.DescribeScalingPlansInput{
			ScalingPlanNames:   aws.StringSlice([]string{scalingPlanName}),
			ScalingPlanVersion: aws.Int64(int64(scalingPlanVersion)),
		}

		output, err := conn.DescribeScalingPlans(input)

		if err != nil {
			return nil, ScalingPlanStatusUnknown, err
		}

		if len(output.ScalingPlans) == 0 {
			return "", ScalingPlanStatusNotFound, nil
		}

		scalingPlan := output.ScalingPlans[0]
		if statusMessage := aws.StringValue(scalingPlan.StatusMessage); statusMessage != "" {
			log.Printf("[INFO] Auto Scaling Scaling Plan (%s/%d) status message: %s", scalingPlanName, scalingPlanVersion, statusMessage)
		}

		return scalingPlan, aws.StringValue(scalingPlan.StatusCode), nil
	}
}
