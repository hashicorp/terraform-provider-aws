// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package devicefarm

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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

	conn := client.DeviceFarmConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &devicefarm.ListProjectsInput{}

	err = conn.ListProjectsPagesWithContext(ctx, input, func(page *devicefarm.ListProjectsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, project := range page.Projects {
			r := ResourceProject()
			d := r.Data(nil)

			id := aws.StringValue(project.Arn)
			d.SetId(id)

			if err != nil {
				err := fmt.Errorf("error reading DeviceFarm Project (%s): %w", id, err)
				log.Printf("[ERROR] %s", err)
				errs = multierror.Append(errs, err)
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing DeviceFarm Project for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping DeviceFarm Project for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DeviceFarm Project sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepTestGridProjects(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.DeviceFarmConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &devicefarm.ListTestGridProjectsInput{}

	err = conn.ListTestGridProjectsPagesWithContext(ctx, input, func(page *devicefarm.ListTestGridProjectsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, project := range page.TestGridProjects {
			r := ResourceTestGridProject()
			d := r.Data(nil)

			id := aws.StringValue(project.Arn)
			d.SetId(id)

			if err != nil {
				err := fmt.Errorf("error reading DeviceFarm Project (%s): %w", id, err)
				log.Printf("[ERROR] %s", err)
				errs = multierror.Append(errs, err)
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing DeviceFarm Test Grid Project for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping DeviceFarm Test Grid Project for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DeviceFarm Test Grid Project sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
