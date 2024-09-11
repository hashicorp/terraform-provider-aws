// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
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
	conn := client.AutoScalingClient(ctx)
	input := &autoscaling.DescribeAutoScalingGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := autoscaling.NewDescribeAutoScalingGroupsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Auto Scaling Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Auto Scaling Groups (%s): %w", region, err)
		}

		for _, v := range page.AutoScalingGroups {
			r := resourceGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AutoScalingGroupName))
			d.Set(names.AttrForceDelete, true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
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
	conn := client.AutoScalingClient(ctx)
	input := &autoscaling.DescribeLaunchConfigurationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := autoscaling.NewDescribeLaunchConfigurationsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Auto Scaling Launch Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Auto Scaling Launch Configurations (%s): %w", region, err)
		}

		for _, v := range page.LaunchConfigurations {
			r := resourceLaunchConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.LaunchConfigurationName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Auto Scaling Launch Configurations (%s): %w", region, err)
	}

	return nil
}
