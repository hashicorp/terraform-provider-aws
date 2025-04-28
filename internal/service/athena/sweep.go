// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_athena_data_catalog", &resource.Sweeper{
		Name: "aws_athena_data_catalog",
		F:    sweepDataCatalogs,
		Dependencies: []string{
			"aws_athena_database",
		},
	})

	resource.AddTestSweepers("aws_athena_database", &resource.Sweeper{
		Name: "aws_athena_database",
		F:    sweepDatabases,
	})

	resource.AddTestSweepers("aws_athena_workgroup", &resource.Sweeper{
		Name: "aws_athena_workgroup",
		F:    sweepWorkGroups,
		Dependencies: []string{
			"aws_athena_data_catalog",
		},
	})

	awsv2.Register("aws_athena_capacity_reservation", sweepCapacityReservations)
}

func sweepCapacityReservations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := athena.ListCapacityReservationsInput{}
	conn := client.AthenaClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := athena.NewListCapacityReservationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		// The Athena API returns a ThrottlingException in unsupported regions
		if tfawserr.ErrCodeEquals(err, "ThrottlingException") {
			tflog.Warn(ctx, "Skipping sweeper", map[string]any{
				"skip_reason": "Unsupported region",
				"error":       err.Error(),
			})
			return sweepResources, nil
		}
		if err != nil {
			return nil, err
		}

		for _, v := range page.CapacityReservations {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceCapacityReservation, client,
				framework.NewAttribute(names.AttrName, aws.ToString(v.Name))),
			)
		}
	}

	return sweepResources, nil
}

func sweepDataCatalogs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.AthenaClient(ctx)
	input := &athena.ListDataCatalogsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := athena.NewListDataCatalogsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Athena Data Catalog sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Athena Data Catalogs (%s): %w", region, err)
		}

		for _, v := range page.DataCatalogsSummary {
			name := aws.ToString(v.CatalogName)

			if name == "AwsDataCatalog" {
				log.Printf("[INFO] Skipping Athena Data Catalog %s", name)
				continue
			}

			r := resourceDataCatalog()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Athena Data Catalogs (%s): %w", region, err)
	}

	return nil
}

func sweepDatabases(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.AthenaClient(ctx)
	input := &athena.ListDataCatalogsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := athena.NewListDataCatalogsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Athena Database sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Athena Data Catalogs (%s): %w", region, err)
		}

		for _, v := range page.DataCatalogsSummary {
			catalogName := aws.ToString(v.CatalogName)
			input := &athena.ListDatabasesInput{
				CatalogName: aws.String(catalogName),
			}

			pages := athena.NewListDatabasesPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					continue
				}

				for _, v := range page.DatabaseList {
					name := aws.ToString(v.Name)

					if name == "default" {
						log.Printf("[INFO] Skipping Athena Database %s", name)
						continue
					}

					r := resourceDatabase()
					d := r.Data(nil)
					d.SetId(name)
					d.Set(names.AttrForceDestroy, true)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Athena Databases (%s): %w", region, err)
	}

	return nil
}

func sweepWorkGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.AthenaClient(ctx)
	input := &athena.ListWorkGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := athena.NewListWorkGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Athena WorkGroup sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Athena WorkGroups (%s): %w", region, err)
		}

		for _, v := range page.WorkGroups {
			name := aws.ToString(v.Name)

			if name == "primary" {
				log.Printf("[INFO] Skipping Athena WorkGroup %s", name)
				continue
			}

			r := resourceWorkGroup()
			d := r.Data(nil)
			d.SetId(name)
			d.Set(names.AttrForceDestroy, true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Athena WorkGroups (%s): %w", region, err)
	}

	return nil
}
