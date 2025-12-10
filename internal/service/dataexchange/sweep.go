// Copyright IBM Corp. 2014, 2025
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
	awsv2.Register("aws_dataexchange_data_set", sweepDataSets, "aws_dataexchange_event_action", "aws_dataexchange_revision")
	awsv2.Register("aws_dataexchange_event_action", sweepEventActions)
	awsv2.Register("aws_dataexchange_revision", sweepRevisions)
}

func sweepDataSets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DataExchangeClient(ctx)
	var input dataexchange.ListDataSetsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := dataexchange.NewListDataSetsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DataSets {
			r := resourceDataSet()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepEventActions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DataExchangeClient(ctx)
	var input dataexchange.ListEventActionsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := dataexchange.NewListEventActionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.EventActions {
			sweepResources = append(sweepResources, framework.NewSweepResource(newEventActionResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Id)),
			))
		}
	}

	return sweepResources, nil
}

func sweepRevisions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DataExchangeClient(ctx)
	var input dataexchange.ListDataSetsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := dataexchange.NewListDataSetsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DataSets {
			var input dataexchange.ListDataSetRevisionsInput
			input.DataSetId = v.Id

			pages := dataexchange.NewListDataSetRevisionsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					return nil, err
				}

				for _, v := range page.Revisions {
					sweepResources = append(sweepResources, framework.NewSweepResource(newRevisionAssetsResource, client,
						framework.NewAttribute(names.AttrID, aws.ToString(v.Id)),
						framework.NewAttribute("data_set_id", v.DataSetId)),
					)
				}
			}
		}
	}

	return sweepResources, nil
}
