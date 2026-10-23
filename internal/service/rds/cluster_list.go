// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_rds_cluster")
func newClusterResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceCluster(), framework.SDKListSpec[awstypes.DBCluster, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.DBCluster, error], diag.Diagnostics) {
			var input rds.DescribeDBClustersInput

			return listDBClusters(ctx, c.RDSClient(ctx), &input), nil
		},
		ID: func(dbc awstypes.DBCluster, _ framework.RegionQuery) string {
			return aws.ToString(dbc.DBClusterIdentifier)
		},
		Prepare: func(_ context.Context, dbc awstypes.DBCluster, _ framework.RegionQuery, rd *schema.ResourceData) {
			// Identity is the cluster identifier, and Flatten only runs when
			// resource state is asked for.
			rd.Set(names.AttrClusterIdentifier, aws.ToString(dbc.DBClusterIdentifier))
		},
		Flatten: resourceClusterFlatten,
		// Flatten reads the cluster's parameters and members, so a listing that
		// only wants identity should not pay for those calls.
		FlattenWhenRequested: true,
	})
}

func listDBClusters(ctx context.Context, conn *rds.Client, input *rds.DescribeDBClustersInput) iter.Seq2[awstypes.DBCluster, error] {
	return func(yield func(awstypes.DBCluster, error) bool) {
		pages := rds.NewDescribeDBClustersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.DBCluster{}, fmt.Errorf("listing RDS Cluster resources: %w", err))
				return
			}

			for _, item := range page.DBClusters {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
