// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventbridgev2

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridgev2"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_eventbridgev2_event_bus", sweepEventBuses, "aws_eventbridgev2_event_source")
	awsv2.Register("aws_eventbridgev2_event_source", sweepEventSources)
}

func sweepEventBuses(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.EventBridgeV2Client(ctx)
	input := eventbridgev2.ListEventBusesInput{
		EventBusAccountId: aws.String(client.AccountID(ctx)),
	}
	var sweepResources []sweep.Sweepable

	pages := eventbridgev2.NewListEventBusesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.EventBuses {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newEventBusResource, client,
				sweepfw.NewAttribute(names.AttrARN, aws.ToString(v.EventBusArn))),
			)
		}
	}

	return sweepResources, nil
}

func sweepEventSources(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.EventBridgeV2Client(ctx)
	var input eventbridgev2.ListEventSourcesInput
	var sweepResources []sweep.Sweepable

	pages := eventbridgev2.NewListEventSourcesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.EventSources {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newEventSourceResource, client,
				sweepfw.NewAttribute(names.AttrARN, aws.ToString(v.EventSourceArn))),
			)
		}
	}

	return sweepResources, nil
}
