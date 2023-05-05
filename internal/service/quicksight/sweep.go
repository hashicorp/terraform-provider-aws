//go:build sweep
// +build sweep

package quicksight

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_quicksight_data_source", &resource.Sweeper{
		Name: "aws_quicksight_data_source",
		F:    sweepsDataSource,
	})
}

func sweepsDataSource(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).QuickSightConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	awsAccountId := client.(*conns.AWSClient).AccountID

	input := &quicksight.ListDataSourcesInput{
		AwsAccountId: aws.String(awsAccountId),
	}

	err = conn.ListDataSourcesPagesWithContext(ctx, input, func(page *quicksight.ListDataSourcesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, ds := range page.DataSources {
			if ds == nil {
				continue
			}

			r := ResourceDataSource()

			d := r.Data(nil)

			d.SetId(fmt.Sprintf("%s/%s", awsAccountId, aws.StringValue(ds.DataSourceId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing QuickSigth Data Sources: %w", err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping QuickSight Data Sources for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping QuickSight Data Source sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
