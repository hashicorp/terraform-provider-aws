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
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_rds_cluster")
func newClusterResourceAsListResource() inttypes.ListResourceForSDK {
	l := clusterListResource{}
	l.SetResourceSchema(resourceCluster())
	return &l
}

var _ list.ListResource = &clusterListResource{}

type clusterListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *clusterListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.RDSClient(ctx)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &rds.DescribeDBClustersInput{}
		for item, err := range listDBClusters(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.DBClusterIdentifier)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrClusterIdentifier), id)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(id)
			rd.Set(names.AttrClusterIdentifier, id)

			if request.IncludeResource {
				if err := resourceClusterFlatten(ctx, conn, &item, rd); err != nil {
					tflog.Error(ctx, "Reading RDS Cluster", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = id

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
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
