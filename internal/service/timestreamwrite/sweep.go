// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package timestreamwrite

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func init() {
	resource.AddTestSweepers("aws_timestreamwrite_database", &resource.Sweeper{
		Name:         "aws_timestreamwrite_database",
		F:            sweepDatabases,
		Dependencies: []string{"aws_timestreamwrite_table"},
	})

	resource.AddTestSweepers("aws_timestreamwrite_table", &resource.Sweeper{
		Name: "aws_timestreamwrite_table",
		F:    sweepTables,
	})
}

func sweepDatabases(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &timestreamwrite.ListDatabasesInput{}
	conn := client.TimestreamWriteClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := timestreamwrite.NewListDatabasesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Timestream Database sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Timestream Databases (%s): %w", region, err)
		}

		for _, v := range page.Databases {
			r := resourceDatabase()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DatabaseName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Timestream Databases (%s): %w", region, err)
	}

	return nil
}

func sweepTables(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &timestreamwrite.ListTablesInput{}
	conn := client.TimestreamWriteClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := timestreamwrite.NewListTablesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Timestream Table sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Timestream Tables (%s): %w", region, err)
		}

		for _, v := range page.Tables {
			r := resourceTable()
			d := r.Data(nil)
			d.SetId(tableCreateResourceID(aws.ToString(v.TableName), aws.ToString(v.DatabaseName)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Timestream Tables (%s): %w", region, err)
	}

	return nil
}
