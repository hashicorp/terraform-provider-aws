// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resourceexplorer2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_resourceexplorer2_index", sweepIndexes,
		"aws_resourceexplorer2_view",
		// To delete AWSManagedViewForCloudWatchTelemetryEnrichment.
		"aws_observabilityadmin_telemetry_enrichment",
	)
	awsv2.Register("aws_resourceexplorer2_view", sweepViews)
}

func sweepIndexes(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResourceExplorer2Client(ctx)
	input := resourceexplorer2.ListIndexesInput{
		Regions: []string{client.Region(ctx)},
	}
	var sweepResources []sweep.Sweepable

	pages := resourceexplorer2.NewListIndexesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Indexes {
			sweepResources = append(sweepResources, framework.NewSweepResource(newIndexResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.Arn)),
			))
		}
	}

	return sweepResources, nil
}

func sweepViews(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResourceExplorer2Client(ctx)
	var input resourceexplorer2.ListViewsInput
	var sweepResources []sweep.Sweepable

	pages := resourceexplorer2.NewListViewsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Views {
			sweepResources = append(sweepResources, framework.NewSweepResource(newViewResource, client,
				framework.NewAttribute(names.AttrARN, v),
			))
		}
	}

	return sweepResources, nil
}
