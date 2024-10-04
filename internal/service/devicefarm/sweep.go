// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devicefarm

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devicefarm"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_devicefarm_project", &resource.Sweeper{
		Name: "aws_devicefarm_project",
		F:    sweepProjects,
	})

	resource.AddTestSweepers("aws_devicefarm_test_grid_project", &resource.Sweeper{
		Name: "aws_devicefarm_test_grid_project",
		F:    sweepTestGridProjects,
	})
}

func sweepProjects(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.DeviceFarmClient(ctx)
	input := &devicefarm.ListProjectsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := devicefarm.NewListProjectsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DeviceFarm Project sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing DeviceFarm Projects (%s): %w", region, err)
		}

		for _, v := range page.Projects {
			r := resourceProject()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DeviceFarm Projects (%s): %w", region, err)
	}

	return nil
}

func sweepTestGridProjects(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.DeviceFarmClient(ctx)
	input := &devicefarm.ListTestGridProjectsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := devicefarm.NewListTestGridProjectsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DeviceFarm Test Grid Project sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing DeviceFarm Test Grid Projects (%s): %w", region, err)
		}

		for _, v := range page.TestGridProjects {
			r := resourceTestGridProject()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DeviceFarm Test Grid Projects (%s): %w", region, err)
	}

	return nil
}
