//go:build sweep
// +build sweep

package dataexchange

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dataexchange"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_dataexchange_data_set", &resource.Sweeper{
		Name: "aws_dataexchange_data_set",
		F:    sweepDataSets,
	})
}

func sweepDataSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DataExchangeConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &dataexchange.ListDataSetsInput{}

	err = conn.ListDataSetsPagesWithContext(ctx, input, func(page *dataexchange.ListDataSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, dataSet := range page.DataSets {
			r := ResourceDataSet()
			d := r.Data(nil)

			d.SetId(aws.StringValue(dataSet.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing DataExchange DataSet for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping DataExchange DataSet for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping DataExchange DataSet sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
