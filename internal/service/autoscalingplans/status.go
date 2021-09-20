package autoscalingplans

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	scalingPlanStatusNotFound = "NotFound"
	scalingPlanStatusUnknown  = "Unknown"
)

// statusScalingPlan fetches the ScalingPlan and its Status
func statusScalingPlan(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		scalingPlan, err := FindScalingPlan(conn, scalingPlanName, scalingPlanVersion)

		if tfawserr.ErrCodeEquals(err, autoscalingplans.ErrCodeObjectNotFoundException) {
			return nil, scalingPlanStatusNotFound, nil
		}

		if err != nil {
			return nil, scalingPlanStatusUnknown, err
		}

		if scalingPlan == nil {
			return nil, scalingPlanStatusNotFound, nil
		}

		if statusMessage := aws.StringValue(scalingPlan.StatusMessage); statusMessage != "" {
			log.Printf("[INFO] Auto Scaling Scaling Plan (%s/%d) status message: %s", scalingPlanName, scalingPlanVersion, statusMessage)
		}

		return scalingPlan, aws.StringValue(scalingPlan.StatusCode), nil
	}
}
