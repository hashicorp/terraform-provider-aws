//go:build sweep
// +build sweep

package timestreamwrite

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).TimestreamWriteConn
	ctx := context.Background()

	var sweeperErrs *multierror.Error

	input := &timestreamwrite.ListDatabasesInput{}

	err = conn.ListDatabasesPagesWithContext(ctx, input, func(page *timestreamwrite.ListDatabasesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, database := range page.Databases {
			if database == nil {
				continue
			}

			dbName := aws.StringValue(database.DatabaseName)

			log.Printf("[INFO] Deleting Timestream Database (%s)", dbName)
			r := ResourceDatabase()
			d := r.Data(nil)
			d.SetId(dbName)

			diags := r.DeleteWithoutTimeout(ctx, d, client)

			if diags != nil && diags.HasError() {
				for _, d := range diags {
					if d.Severity == diag.Error {
						sweeperErr := fmt.Errorf("error deleting Timestream Database (%s): %s", dbName, d.Summary)
						log.Printf("[ERROR] %s", sweeperErr)
						sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					}
				}
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Timestream Database sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Timestream Databases: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepTables(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).TimestreamWriteConn
	ctx := context.Background()

	var sweeperErrs *multierror.Error

	input := &timestreamwrite.ListTablesInput{}

	err = conn.ListTablesPagesWithContext(ctx, input, func(page *timestreamwrite.ListTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, table := range page.Tables {
			if table == nil {
				continue
			}

			tableName := aws.StringValue(table.TableName)
			dbName := aws.StringValue(table.TableName)

			log.Printf("[INFO] Deleting Timestream Table (%s) from Database (%s)", tableName, dbName)
			r := ResourceTable()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s:%s", tableName, dbName))

			diags := r.DeleteWithoutTimeout(ctx, d, client)

			if diags != nil && diags.HasError() {
				for _, d := range diags {
					if d.Severity == diag.Error {
						sweeperErr := fmt.Errorf("error deleting Timestream Table (%s): %s", dbName, d.Summary)
						log.Printf("[ERROR] %s", sweeperErr)
						sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					}
				}
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Timestream Table sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Timestream Tables: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
