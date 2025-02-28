// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_dataexchange_data_set", &resource.Sweeper{
		Name: "aws_dataexchange_data_set",
		F:    sweepDataSets,
	})

	awsv2.Register("aws_dataexchange_event_action", sweepEventActions)
}

func sweepDataSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.DataExchangeClient(ctx)
	input := &dataexchange.ListDataSetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := dataexchange.NewListDataSetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataExchange DataSet sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing DataExchange DataSets (%s): %w", region, err)
		}

		for _, v := range page.DataSets {
			r := ResourceDataSet()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DataExchange DataSets (%s): %w", region, err)
	}

	return nil
}

func sweepEventActions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DataExchangeClient(ctx)

	var sweepResources []sweep.Sweepable

	input := dataexchange.ListEventActionsInput{}
	pages := dataexchange.NewListEventActionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, eventActions := range page.EventActions {
			sweepResources = append(sweepResources, framework.NewSweepResource(newEventActionResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(eventActions.Id)),
			))
		}
	}

	return sweepResources, nil
}
