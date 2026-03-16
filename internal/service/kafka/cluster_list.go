// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_kafka_cluster")
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
	conn := l.Meta().KafkaClient(ctx)

	var query listClusterModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input kafka.ListClustersInput
		for item, err := range listClusters(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.ClusterArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(arn)

			if request.IncludeResource {
				outputGBB, err := findBootstrapBrokersByARN(ctx, conn, arn)

				if err != nil {
					tflog.Error(ctx, "Reading Managed Streaming for Kafka Cluster", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				if err := resourceClusterFlatten(ctx, l.Meta(), &item, outputGBB, rd); err != nil {
					tflog.Error(ctx, "Reading Managed Streaming for Kafka Cluster", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = aws.ToString(item.ClusterName)

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

func listClusters(ctx context.Context, conn *kafka.Client, input *kafka.ListClustersInput) iter.Seq2[awstypes.ClusterInfo, error] {
	return func(yield func(awstypes.ClusterInfo, error) bool) {
		pages := kafka.NewListClustersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.ClusterInfo](), fmt.Errorf("listing MSK Clusters: %w", err))
				return
			}

			for _, item := range page.ClusterInfoList {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
