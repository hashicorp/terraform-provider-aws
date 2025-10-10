// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscalingplans

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/autoscalingplans"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_autoscalingplans_scaling_plan", sweepScalingPlans)
}

func sweepScalingPlans(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AutoScalingPlansClient(ctx)

	var sweepResources []sweep.Sweepable
	r := ResourceScalingPlan()

	input := autoscalingplans.DescribeScalingPlansInput{}
	err := describeScalingPlansPages(ctx, conn, &input, func(page *autoscalingplans.DescribeScalingPlansOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, scalingPlan := range page.ScalingPlans {
			d := r.Data(nil)
			d.SetId("unused")
			d.Set(names.AttrName, scalingPlan.ScalingPlanName)
			d.Set("scaling_plan_version", scalingPlan.ScalingPlanVersion)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	return sweepResources, err
}
