//go:build sweep
// +build sweep

package timestreamwrite

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
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
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &timestreamwrite.ListDatabasesInput{}
	conn := client.(*conns.AWSClient).TimestreamWriteConn()
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListDatabasesPagesWithContext(ctx, input, func(page *timestreamwrite.ListDatabasesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Databases {
			r := ResourceDatabase()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.DatabaseName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Timestream Database sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Timestream Databases (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Timestream Databases (%s): %w", region, err)
	}

	return nil
}

func sweepTables(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &timestreamwrite.ListTablesInput{}
	conn := client.(*conns.AWSClient).TimestreamWriteConn()
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListTablesPagesWithContext(ctx, input, func(page *timestreamwrite.ListTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Tables {
			r := ResourceTable()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(v.TableName), aws.StringValue(v.DatabaseName)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Timestream Table sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Timestream Tables (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Timestream Tables (%s): %w", region, err)
	}

	return nil
}
