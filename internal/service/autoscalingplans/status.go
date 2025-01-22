// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscalingplans

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/autoscalingplans"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusScalingPlanCode(ctx context.Context, conn *autoscalingplans.Client, scalingPlanName string, scalingPlanVersion int) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		scalingPlan, err := FindScalingPlanByNameAndVersion(ctx, conn, scalingPlanName, scalingPlanVersion)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return scalingPlan, string(scalingPlan.StatusCode), nil
	}
}
