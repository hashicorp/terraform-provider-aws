// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package scheduler

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_scheduler_schedule", sweepSchedules)
	awsv2.Register("aws_scheduler_schedule_group", sweepScheduleGroups, "aws_scheduler_schedule")
}

func sweepScheduleGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SchedulerClient(ctx)
	var input scheduler.ListScheduleGroupsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := scheduler.NewListScheduleGroupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ScheduleGroups {
			name := aws.ToString(v.Name)

			if name == "default" {
				log.Printf("[INFO] Skipping EventBridge Scheduler Schedule Group %s", name)
				continue
			}

			r := resourceScheduleGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepSchedules(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SchedulerClient(ctx)
	var input scheduler.ListSchedulesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := scheduler.NewListSchedulesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Schedules {
			r := resourceSchedule()
			d := r.Data(nil)
			d.SetId(scheduleCreateResourceID(aws.ToString(v.GroupName), aws.ToString(v.Name)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
