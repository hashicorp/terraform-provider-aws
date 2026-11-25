// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dsql

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_dsql_cluster_policy")
func newClusterPolicyResourceAsListResource() list.ListResourceWithConfigure {
	return &clusterPolicyListResource{}
}

var _ list.ListResource = &clusterPolicyListResource{}

type clusterPolicyListResource struct {
	clusterPolicyResource
	framework.WithList
}

func (l *clusterPolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().DSQLClient(ctx)

	tflog.Info(ctx, "Listing Aurora DSQL Cluster Policies")

	stream.Results = func(yield func(list.ListResult) bool) {
		var clusterInput dsql.ListClustersInput
		for cluster, err := range listClusters(ctx, conn, &clusterInput) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			id := aws.ToString(cluster.Identifier)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrIdentifier), id)

			output, err := findClusterPolicyByID(ctx, conn, id)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			result := request.NewListResult(ctx)

			var data clusterPolicyResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.Identifier = fwflex.StringValueToFramework(ctx, id)

				if request.IncludeResource {
					result.Diagnostics.Append(l.flatten(ctx, output, &data)...)
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
