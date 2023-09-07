// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package autoscalingplans

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_autoscalingplans_scaling_plan", &resource.Sweeper{
		Name: "aws_autoscalingplans_scaling_plan",
		F:    sweepScalingPlans,
	})
}

func sweepScalingPlans(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.AutoScalingPlansConn(ctx)
	input := &autoscalingplans.DescribeScalingPlansInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeScalingPlansPages(ctx, conn, input, func(page *autoscalingplans.DescribeScalingPlansOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, scalingPlan := range page.ScalingPlans {
			scalingPlanName := aws.StringValue(scalingPlan.ScalingPlanName)
			scalingPlanVersion := int(aws.Int64Value(scalingPlan.ScalingPlanVersion))

			r := ResourceScalingPlan()
			d := r.Data(nil)
			d.SetId("unused")
			d.Set("name", scalingPlanName)
			d.Set("scaling_plan_version", scalingPlanVersion)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Auto Scaling Scaling Plan sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Auto Scaling Scaling Plans (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Auto Scaling Scaling Plans (%s): %w", region, err)
	}

	return nil
}
