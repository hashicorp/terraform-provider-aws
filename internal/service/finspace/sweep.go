// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package finspace

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func init() {
	resource.AddTestSweepers("aws_finspace_kx_environment", &resource.Sweeper{
		Name: "aws_finspace_kx_environment",
		F:    sweepKxEnvironments,
	})
}

func sweepKxEnvironments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.FinSpaceClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &finspace.ListKxEnvironmentsInput{}
	pages := finspace.NewListKxEnvironmentsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping FinSpace Kx Environment sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("listing FinSpace Kx Environments (%s): %w", region, err))
		}

		for _, env := range page.Environments {
			r := ResourceKxEnvironment()
			d := r.Data(nil)
			id := aws.ToString(env.EnvironmentId)
			d.SetId(id)

			log.Printf("[INFO] Deleting FinSpace Kx Environment: %s", id)
			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)
	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping FinSpace Kx Environments (%s): %w", region, err))
	}

	return errs.ErrorOrNil()
}
