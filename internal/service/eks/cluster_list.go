// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_eks_cluster")
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
	conn := l.Meta().EKSClient(ctx)

	var query listClusterModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input eks.ListClustersInput
		for item, err := range listClusters(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), item)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(item)
			rd.Set(names.AttrName, item)

			if request.IncludeResource {
				cluster, err := findClusterByName(ctx, conn, item)

				if err != nil {
					tflog.Error(ctx, "Reading EKS Cluster", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				if err := resourceClusterFlatten(ctx, l.Meta(), cluster, rd); err != nil {
					tflog.Error(ctx, "Flattening EKS Cluster", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = item

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
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

type listClusterModel struct {
	framework.WithRegionModel
}

func listClusters(ctx context.Context, conn *eks.Client, input *eks.ListClustersInput) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		pages := eks.NewListClustersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[string](), fmt.Errorf("listing EKS Clusters: %w", err))
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
