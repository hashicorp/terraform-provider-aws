// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_ecs_cluster")
func newClusterResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceCluster{}
	l.SetResourceSchema(resourceCluster())
	return &l
}

var _ list.ListResource = &listResourceCluster{}

type listResourceCluster struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceCluster) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().ECSClient(ctx)

	var query listClusterModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing ECS (Elastic Container) Cluster")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input ecs.ListClustersInput
		for arnStr, err := range listClusterARNs(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := clusterNameFromARN(arnStr)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), arnStr)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(arnStr)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				cluster, err := findClusterByNameOrARN(ctx, conn, arnStr)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					tflog.Error(ctx, "Reading ECS (Elastic Container) Cluster", map[string]any{
						"err": err.Error(),
					})
					continue
				}

				diags := resourceClusterFlatten(ctx, rd, cluster)
				if diags.HasError() {
					tflog.Error(ctx, "Flatten ECS (Elastic Container) Cluster", map[string]any{
						"diags": sdkdiag.DiagnosticsString(diags),
					})
					continue
				}
			}

			result.DisplayName = name

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

func listClusterARNs(ctx context.Context, conn *ecs.Client, input *ecs.ListClustersInput) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		pages := ecs.NewListClustersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield("", fmt.Errorf("listing ECS (Elastic Container) Cluster resources: %w", err))
				return
			}

			for _, arnStr := range page.ClusterArns {
				if !yield(arnStr, nil) {
					return
				}
			}
		}
	}
}
