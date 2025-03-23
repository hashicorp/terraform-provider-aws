// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_dataexchange_data_set", sweepDataSets,
		"aws_dataexchange_event_action",
	)

	awsv2.Register("aws_dataexchange_event_action", sweepEventActions)
}

func sweepDataSets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DataExchangeClient(ctx)

	var sweepResources []sweep.Sweepable
	r := ResourceDataSet()

	input := dataexchange.ListDataSetsInput{}

	pages := dataexchange.NewListDataSetsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.DataSets {
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
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
