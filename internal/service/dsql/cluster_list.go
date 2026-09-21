// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dsql

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dsql/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_dsql_cluster")
func newClusterResourceAsListResource() list.ListResourceWithConfigure {
	return &clusterListResource{}
}

var _ list.ListResource = &clusterListResource{}

type clusterListResource struct {
	clusterResource
	framework.WithList
}

func (l *clusterListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().DSQLClient(ctx)

	tflog.Info(ctx, "Listing Aurora DSQL Clusters")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input dsql.ListClustersInput
		for item, err := range listClusters(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.Identifier)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), aws.ToString(item.Arn))

			var output *dsql.GetClusterOutput
			if request.IncludeResource {
				var err error
				output, err = findClusterByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data clusterResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.Identifier = fwflex.StringValueToFramework(ctx, id)
				data.ARN = fwflex.StringToFramework(ctx, item.Arn)

				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, conn, output, &data)...)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = id
			})

			if !yield(result) {
				return
			}
		}
	}
}

func listClusters(ctx context.Context, conn *dsql.Client, input *dsql.ListClustersInput) iter.Seq2[awstypes.ClusterSummary, error] {
	return func(yield func(awstypes.ClusterSummary, error) bool) {
		pages := dsql.NewListClustersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ClusterSummary{}, fmt.Errorf("listing Aurora DSQL Clusters: %w", err))
				return
			}

			for _, item := range page.Clusters {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
