// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package evidently

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_evidently_project", &resource.Sweeper{
		Name: "aws_evidently_project",
		F:    sweepProject,
	})
}

func sweepProject(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.EvidentlyConn(ctx)
	input := &cloudwatchevidently.ListProjectsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListProjectsPagesWithContext(ctx, input, func(page *cloudwatchevidently.ListProjectsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, project := range page.Projects {
			r := ResourceProject()
			d := r.Data(nil)
			d.SetId(aws.StringValue(project.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Evidently Project sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Evidently Projects for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Evidently Projects for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
