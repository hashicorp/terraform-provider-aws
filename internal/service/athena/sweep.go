//go:build sweep
// +build sweep

package athena

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_athena_database", &resource.Sweeper{
		Name: "aws_athena_database",
		F:    sweepDatabases,
	})
}

func sweepDatabases(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AthenaConn()
	input := &athena.ListDatabasesInput{
		CatalogName: aws.String("AwsDataCatalog"),
	}
	var errs *multierror.Error

	sweepResources := make([]sweep.Sweepable, 0)
	for {
		output, err := conn.ListDatabasesWithContext(ctx, input)

		for _, v := range output.DatabaseList {
			name := aws.StringValue(v.Name)
			if name == "default" {
				continue
			}
			r := ResourceDatabase()
			d := r.Data(nil)
			d.SetId(name)

			if err != nil {
				err := fmt.Errorf("error listing Athena Databases (%s): %w", name, err)
				log.Printf("[ERROR] %s", err)
				errs = multierror.Append(errs, err)
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Athena Database sweep for %s: %s", region, err)
		return nil
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Athena Databases (%s): %w", region, err))
	}

	return errs.ErrorOrNil()
}
