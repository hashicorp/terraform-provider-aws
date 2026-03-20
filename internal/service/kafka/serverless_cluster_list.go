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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_msk_serverless_cluster")
func newServerlessClusterResourceAsListResource() inttypes.ListResourceForSDK {
	l := serverlessClusterListResource{}
	l.SetResourceSchema(resourceServerlessCluster())
	return &l
}

var _ list.ListResource = &serverlessClusterListResource{}

type serverlessClusterListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *serverlessClusterListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().KafkaClient(ctx)

	var query listServerlessClusterModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		input := kafka.ListClustersV2Input{
			ClusterTypeFilter: aws.String("SERVERLESS"),
		}
		for item, err := range listServerlessClusters(ctx, conn, &input) {
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
			rd.Set(names.AttrARN, arn)

			if request.IncludeResource {
				if err := resourceServerlessClusterFlatten(ctx, &item, rd); err != nil {
					tflog.Error(ctx, "Flattening MSK Serverless Cluster", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				output, err := findBootstrapBrokersByARN(ctx, conn, arn)

				switch {
				case errs.IsA[*awstypes.ForbiddenException](err):
					rd.Set("bootstrap_brokers_sasl_iam", nil)
				case err != nil:
					tflog.Error(ctx, "Reading MSK Cluster bootstrap brokers", map[string]any{
						"error": err.Error(),
					})
					continue
				default:
					rd.Set("bootstrap_brokers_sasl_iam", sortEndpointsString(aws.ToString(output.BootstrapBrokerStringSaslIam)))
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

type listServerlessClusterModel struct {
	framework.WithRegionModel
}

func listServerlessClusters(ctx context.Context, conn *kafka.Client, input *kafka.ListClustersV2Input) iter.Seq2[awstypes.Cluster, error] {
	return func(yield func(awstypes.Cluster, error) bool) {
		pages := kafka.NewListClustersV2Paginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.Cluster](), fmt.Errorf("listing MSK Serverless Clusters: %w", err))
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
