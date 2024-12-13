// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package evidently

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/evidently"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
)

func RegisterSweepers() {
	awsv2.Register("aws_evidently_project", sweepProjects)

	awsv2.Register("aws_evidently_segment", sweepSegments)
}

func sweepProjects(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.EvidentlyClient(ctx)

	var sweepResources []sweep.Sweepable

	pages := evidently.NewListProjectsPaginator(conn, &evidently.ListProjectsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, project := range page.Projects {
			r := ResourceProject()
			d := r.Data(nil)
			d.SetId(aws.ToString(project.Name))

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepSegments(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.EvidentlyClient(ctx)

	var sweepResources []sweep.Sweepable

	pages := evidently.NewListSegmentsPaginator(conn, &evidently.ListSegmentsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, segment := range page.Segments {
			r := ResourceSegment()
			d := r.Data(nil)
			d.SetId(aws.ToString(segment.Arn))

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
