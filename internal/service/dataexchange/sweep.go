// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_dataexchange_data_set", &resource.Sweeper{
		Name: "aws_dataexchange_data_set",
		F:    sweepDataSets,
	})
}

func sweepDataSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.DataExchangeClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &dataexchange.ListDataSetsInput{}

	pages := dataexchange.NewListDataSetsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing DataExchange DataSet for %s: %w", region, err))
		}

		for _, dataSet := range page.DataSets {
			r := ResourceDataSet()
			d := r.Data(nil)

			d.SetId(aws.ToString(dataSet.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping DataExchange DataSet for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping DataExchange DataSet sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
