// Copyright (c) HashiCorp, Inc.
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
	awsv2.Register("aws_dsql_cluster", sweepClusters)
}

func sweepClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DSQLClient(ctx)
	var input dsql.ListClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := dsql.NewListClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Clusters {
			sweepResources = append(sweepResources, framework.NewSweepResource(newClusterResource, client,
				framework.NewAttribute(names.AttrIdentifier, aws.ToString(v.Identifier))),
			)
		}
	}

	return sweepResources, nil
}
