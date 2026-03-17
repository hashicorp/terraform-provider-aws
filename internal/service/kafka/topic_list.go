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
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_msk_topic")
func newTopicResourceAsListResource() list.ListResourceWithConfigure {
	return &topicListResource{}
}

var _ list.ListResource = &topicListResource{}

type topicListResource struct {
	topicResource
	framework.WithList
}

func (l *topicListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"cluster_arn": listschema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the cluster to list Topics from.",
			},
		},
	}
}

func (l *topicListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().KafkaClient(ctx)

	var query listTopicModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	clusterARN := fwflex.StringValueFromFramework(ctx, query.ClusterARN)

	stream.Results = func(yield func(list.ListResult) bool) {
		input := kafka.ListTopicsInput{
			ClusterArn: aws.String(clusterARN),
		}
		for item, err := range listTopics(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn, name := aws.ToString(item.TopicArn), aws.ToString(item.TopicName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			var data topicResourceModel
			// TIP: -- 6. Set the ID, arguments, and attributes
			// Using a field name prefix allows mapping fields such as `TopicId` to `ID`
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				out, err := findTopicByTwoPartKey(ctx, conn, arn, name)
				if err != nil {
					result.Diagnostics.AddError("Reading MSK Topic", err.Error())
					return
				}

				result.Diagnostics.Append(l.flatten(ctx, out, &data, true)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = name
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listTopicModel struct {
	framework.WithRegionModel
	ClusterARN fwtypes.ARN `tfsdk:"cluster_arn"`
}

func listTopics(ctx context.Context, conn *kafka.Client, input *kafka.ListTopicsInput) iter.Seq2[awstypes.TopicInfo, error] {
	return func(yield func(awstypes.TopicInfo, error) bool) {
		pages := kafka.NewListTopicsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.TopicInfo](), fmt.Errorf("listing MSK Topics: %w", err))
				return
			}

			for _, item := range page.Topics {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
