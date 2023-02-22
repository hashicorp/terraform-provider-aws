package autoscalingplans

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindScalingPlanByNameAndVersion(ctx context.Context, conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	input := &autoscalingplans.DescribeScalingPlansInput{
		ScalingPlanNames:   aws.StringSlice([]string{scalingPlanName}),
		ScalingPlanVersion: aws.Int64(int64(scalingPlanVersion)),
	}

	output, err := conn.DescribeScalingPlansWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, autoscalingplans.ErrCodeObjectNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ScalingPlans) == 0 || output.ScalingPlans[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ScalingPlans[0], nil
}
