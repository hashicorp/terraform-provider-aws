// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package athena

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	awstypes "github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	s3tablestypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

	awsv2.Register("aws_athena_database", sweepDatabases)

	resource.AddTestSweepers("aws_athena_workgroup", &resource.Sweeper{
		Name: "aws_athena_workgroup",
		F:    sweepWorkGroups,
		Dependencies: []string{
			"aws_athena_data_catalog",
		},
	})

	awsv2.Register("aws_athena_capacity_reservation", sweepCapacityReservations)
	awsv2.Register("aws_athena_system_table_configuration", sweepSystemTableConfigurations)
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
			sweepResources = append(sweepResources, framework.NewSweepResource(newCapacityReservationResource, client,
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
		return fmt.Errorf("getting client: %w", err)
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
				log.Printf("[INFO] Skipping Athena Data Catalog %q", name)
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

func sweepDatabases(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AthenaClient(ctx)
	var sweepResources []sweep.Sweepable

	input := athena.ListDataCatalogsInput{}
	pages := athena.NewListDataCatalogsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
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
					tflog.Warn(ctx, "Skipping resource", map[string]any{
						"error":        err.Error(),
						"catalog_name": catalogName,
					})
					break
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

	return sweepResources, nil
}

func sweepWorkGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
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

func sweepSystemTableConfigurations(_ context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	return []sweep.Sweepable{systemTablesSweeper{client: client}}, nil
}

type systemTablesSweeper struct {
	client *conns.AWSClient
}

func (s systemTablesSweeper) Delete(ctx context.Context, _ ...tfresource.OptionsFunc) error {
	return resetSystemTables(ctx, s.client)
}

func systemTablesBucketARN(ctx context.Context, client *conns.AWSClient) string {
	return client.RegionalARN(ctx, names.S3Tables, "bucket/"+systemTablesBucketName)
}

func findSystemTables(ctx context.Context, client *conns.AWSClient) ([]s3tablestypes.TableSummary, error) {
	input := s3tables.ListTablesInput{
		Namespace:      aws.String(systemTablesNamespace),
		TableBucketARN: aws.String(systemTablesBucketARN(ctx, client)),
	}
	var output []s3tablestypes.TableSummary

	pages := s3tables.NewListTablesPaginator(client.S3TablesClient(ctx), &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if errs.IsA[*s3tablestypes.NotFoundException](err) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}

		output = append(output, page.Tables...)
	}

	return output, nil
}

func deleteSystemTables(ctx context.Context, client *conns.AWSClient) error {
	tables, err := findSystemTables(ctx, client)
	if err != nil {
		return err
	}

	conn := client.S3TablesClient(ctx)
	bucketARN := systemTablesBucketARN(ctx, client)
	for _, v := range tables {
		input := s3tables.DeleteTableInput{
			Name:           v.Name,
			Namespace:      aws.String(systemTablesNamespace),
			TableBucketARN: aws.String(bucketARN),
		}
		_, err := conn.DeleteTable(ctx, &input)
		if errs.IsA[*s3tablestypes.NotFoundException](err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("deleting S3 Tables table (%s): %w", aws.ToString(v.Name), err)
		}
	}

	return nil
}

func deleteSystemTableConfigurations(ctx context.Context, client *conns.AWSClient) error {
	conn := client.AthenaClient(ctx)
	for _, name := range enum.EnumValues[awstypes.SystemTableName]() {
		input := athena.DeleteSystemTableConfigurationInput{
			TableName: name,
		}
		if _, err := conn.DeleteSystemTableConfiguration(ctx, &input); err != nil {
			return fmt.Errorf("deleting Athena System Table Configuration (%s): %w", name, err)
		}
	}

	return nil
}

func resetSystemTables(ctx context.Context, client *conns.AWSClient) error {
	if err := deleteSystemTableConfigurations(ctx, client); err != nil {
		return err
	}

	return deleteSystemTables(ctx, client)
}
