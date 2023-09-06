// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package autoscaling

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_autoscaling_group", &resource.Sweeper{
		Name: "aws_autoscaling_group",
		F:    sweepGroups,
	})

	resource.AddTestSweepers("aws_launch_configuration", &resource.Sweeper{
		Name:         "aws_launch_configuration",
		F:            sweepLaunchConfigurations,
		Dependencies: []string{"aws_autoscaling_group"},
	})
}

func sweepGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.AutoScalingConn(ctx)
	input := &autoscaling.DescribeAutoScalingGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeAutoScalingGroupsPagesWithContext(ctx, input, func(page *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.AutoScalingGroups {
			r := ResourceGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.AutoScalingGroupName))
			d.Set("force_delete", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Auto Scaling Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Auto Scaling Groups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Auto Scaling Groups (%s): %w", region, err)
	}

	return nil
}

func sweepLaunchConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.AutoScalingConn(ctx)
	input := &autoscaling.DescribeLaunchConfigurationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeLaunchConfigurationsPagesWithContext(ctx, input, func(page *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LaunchConfigurations {
			r := ResourceLaunchConfiguration()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.LaunchConfigurationName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Auto Scaling Launch Configuration sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Auto Scaling Launch Configurations (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Auto Scaling Launch Configurations (%s): %w", region, err)
	}

	return nil
}
