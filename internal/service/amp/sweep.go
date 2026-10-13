// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_prometheus_anomaly_detector", sweepAnomalyDetectors)

	awsv2.Register("aws_prometheus_scraper", sweepScraper)

	awsv2.Register("aws_prometheus_workspace", sweepWorkspace,
		"aws_prometheus_scraper", "aws_prometheus_anomaly_detector",
	)
}

func sweepAnomalyDetectors(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := amp.ListAnomalyDetectorsInput{}
	conn := client.AMPClient(ctx)
	var sweepResources []sweep.Sweepable

	workspacePages := amp.NewListWorkspacesPaginator(conn, &amp.ListWorkspacesInput{})
	for workspacePages.HasMorePages() {
		workspacePage, err := workspacePages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, workspace := range workspacePage.Workspaces {
			input.WorkspaceId = workspace.WorkspaceId
			pages := amp.NewListAnomalyDetectorsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, v := range page.AnomalyDetectors {
					sweepResources = append(sweepResources, framework.NewSweepResource(newAnomalyDetectorResource, client,
						framework.NewAttribute(names.AttrID, aws.ToString(v.AnomalyDetectorId)),
						framework.NewAttribute("workspace_id", aws.ToString(workspace.WorkspaceId)),
					))
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepScraper(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AMPClient(ctx)

	var sweepResources []sweep.Sweepable

	pages := amp.NewListScrapersPaginator(conn, &amp.ListScrapersInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, scraper := range page.Scrapers {
			sweepResources = append(sweepResources, framework.NewSweepResource(newScraperResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(scraper.ScraperId)),
			))
		}
	}

	return sweepResources, nil
}

func sweepWorkspace(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AMPClient(ctx)

	var sweepResources []sweep.Sweepable
	r := resourceWorkspace()

	pages := amp.NewListWorkspacesPaginator(conn, &amp.ListWorkspacesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, workspace := range page.Workspaces {
			d := r.Data(nil)
			d.SetId(aws.ToString(workspace.WorkspaceId))

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
