// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscalingplans

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscalingplans"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscalingplans/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindScalingPlanByNameAndVersion(ctx context.Context, conn *autoscalingplans.Client, scalingPlanName string, scalingPlanVersion int) (*awstypes.ScalingPlan, error) {
	input := &autoscalingplans.DescribeScalingPlansInput{
		ScalingPlanNames:   []string{scalingPlanName},
		ScalingPlanVersion: aws.Int64(int64(scalingPlanVersion)),
	}

	output, err := conn.DescribeScalingPlans(ctx, input)

	if errs.IsA[*awstypes.ObjectNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output.ScalingPlans)
}
