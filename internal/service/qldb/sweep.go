// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package qldb

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qldb"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func init() {
	resource.AddTestSweepers("aws_qldb_ledger", &resource.Sweeper{
		Name: "aws_qldb_ledger",
		F:    sweepLedgers,
		Dependencies: []string{
			"aws_qldb_stream",
		},
	})

	resource.AddTestSweepers("aws_qldb_stream", &resource.Sweeper{
		Name: "aws_qldb_stream",
		F:    sweepStreams,
	})

}

func sweepLedgers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.QLDBClient(ctx)
	input := &qldb.ListLedgersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := qldb.NewListLedgersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping QLDB Ledger sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing QLDB Ledgers (%s): %w", region, err)
		}

		for _, v := range page.Ledgers {
			r := resourceLedger()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping QLDB Ledgers (%s): %w", region, err)
	}

	return nil
}

func sweepStreams(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.QLDBClient(ctx)
	input := &qldb.ListLedgersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := qldb.NewListLedgersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping QLDB Stream sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing QLDB Ledgers (%s): %w", region, err))
		}

		for _, v := range page.Ledgers {
			input := &qldb.ListJournalKinesisStreamsForLedgerInput{
				LedgerName: v.Name,
			}

			pages := qldb.NewListJournalKinesisStreamsForLedgerPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing QLDB Streams (%s): %w", region, err))
				}

				for _, v := range page.Streams {
					r := resourceStream()
					d := r.Data(nil)
					d.SetId(aws.ToString(v.StreamId))
					d.Set("ledger_name", v.LedgerName)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping QLDB Streams (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
