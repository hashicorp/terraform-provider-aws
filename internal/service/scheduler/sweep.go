//go:build sweep
// +build sweep

package scheduler

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_scheduler_schedule_group", &resource.Sweeper{
		Name: "aws_scheduler_schedule_group",
		F:    sweepScheduleGroups,
	})
}

func sweepScheduleGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).SchedulerClient
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	paginator := scheduler.NewListScheduleGroupsPaginator(conn, &scheduler.ListScheduleGroupsInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("listing Schedule Groups for %s: %w", region, err))
			break
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

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Schedule Group for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Schedule Group sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
