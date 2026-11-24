// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dsql

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_dsql_cluster", sweepClusters, "aws_dsql_stream")
	awsv2.Register("aws_dsql_stream", sweepStreams)
}

func sweepClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DSQLClient(ctx)
	var sweepResources []sweep.Sweepable

	var input dsql.ListClustersInput
	pages := dsql.NewListClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Clusters {
			sweepResources = append(sweepResources, framework.NewSweepResource(newClusterResource, client,
				framework.NewAttribute(names.AttrIdentifier, aws.ToString(v.Identifier)),
				framework.NewAttribute(names.AttrForceDestroy, true),
			))
		}
	}

	return sweepResources, nil
}

// sweepStreams deletes every aws_dsql_stream in the Region, under every parent ListClusters returns, not only the
// acceptance tests' own.
func sweepStreams(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DSQLClient(ctx)
	var sweepResources []sweep.Sweepable

	var parentInput dsql.ListClustersInput
	parentPages := dsql.NewListClustersPaginator(conn, &parentInput)
	for parentPages.HasMorePages() {
		parentPage, err := parentPages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, parent := range parentPage.Clusters {
			input := dsql.ListStreamsInput{
				ClusterIdentifier: parent.Identifier,
			}
			for v, err := range listStreams(ctx, conn, &input) {
				if err != nil {
					return nil, err
				}

				sweepResources = append(sweepResources, framework.NewSweepResource(newStreamResource, client,
					framework.NewAttribute(names.AttrClusterIdentifier, aws.ToString(v.ClusterIdentifier)),
					framework.NewAttribute("stream_identifier", aws.ToString(v.StreamIdentifier))))
			}
		}
	}

	return sweepResources, nil
}
