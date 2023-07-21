// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package scheduler

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func init() {
	resource.AddTestSweepers("aws_scheduler_schedule_group", &resource.Sweeper{
		Name: "aws_scheduler_schedule_group",
		F:    sweepScheduleGroups,
		Dependencies: []string{
			"aws_scheduler_schedule",
		},
	})

	resource.AddTestSweepers("aws_scheduler_schedule", &resource.Sweeper{
		Name: "aws_scheduler_schedule",
		F:    sweepSchedules,
	})
}

func sweepScheduleGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.SchedulerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	paginator := scheduler.NewListScheduleGroupsPaginator(conn, &scheduler.ListScheduleGroupsInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Schedule Group sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("listing Schedule Groups for %s: %w", region, err)
		}

		for _, it := range page.ScheduleGroups {
			name := aws.ToString(it.Name)

			if name == "default" {
				// Can't delete the default schedule group.
				continue
			}

			r := ResourceScheduleGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping Schedule Group for %s: %w", region, err)
	}

	return nil
}

func sweepSchedules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.SchedulerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	paginator := scheduler.NewListSchedulesPaginator(conn, &scheduler.ListSchedulesInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Schedule sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("listing Schedules for %s: %w", region, err)
		}

		for _, it := range page.Schedules {
			groupName := aws.ToString(it.GroupName)
			scheduleName := aws.ToString(it.Name)

			r := resourceSchedule()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s/%s", groupName, scheduleName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping Schedule for %s: %w", region, err)
	}

	return nil
}
