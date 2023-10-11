// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package kendra

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func init() {
	resource.AddTestSweepers("aws_kendra_index", &resource.Sweeper{
		Name: "aws_kendra_index",
		F:    sweepIndex,
	})
}

func sweepIndex(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.KendraClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &kendra.ListIndicesInput{}
	var errs *multierror.Error

	pages := kendra.NewListIndicesPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Kendra Indices sweep for %s: %s", region, err)
			return errs.ErrorOrNil()
		}

		if err != nil {
			return multierror.Append(errs, fmt.Errorf("retrieving Kendra Indices: %w", err))
		}

		for _, index := range page.IndexConfigurationSummaryItems {
			r := ResourceIndex()
			d := r.Data(nil)
			d.SetId(aws.ToString(index.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Kendra Indices for %s: %w", region, err))
	}

	return errs.ErrorOrNil()
}
