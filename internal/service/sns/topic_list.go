// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_sns_topic")
func newTopicResourceAsListResource() inttypes.ListResourceForSDK {
	l := topicListResource{}
	l.SetResourceSchema(resourceTopic())
	return &l
}

var _ list.ListResource = &topicListResource{}

type topicListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listTopicModel struct {
	framework.WithRegionModel
}

func (l *topicListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SNSClient(ctx)

	var query listTopicModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing SNS Topics")
	stream.Results = func(yield func(list.ListResult) bool) {
		input := &sns.ListTopicsInput{}
		for item, err := range listTopics(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			topicARN := aws.ToString(item.TopicArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), topicARN)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(topicARN)
			rd.Set(names.AttrARN, topicARN) //nolint:errcheck

			if request.IncludeResource {
				attributes, err := findTopicAttributesWithValidAWSPrincipalsByARN(ctx, conn, topicARN)
				if err != nil {
					tflog.Error(ctx, "Reading SNS Topic", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				if diags := resourceTopicFlatten(ctx, rd, attributes); diags.HasError() {
					tflog.Error(ctx, "Reading SNS Topic", map[string]any{
						"diags": sdkdiag.DiagnosticsString(diags),
					})
					continue
				}
			}

			// Set display name from topic name (last segment of ARN)
			result.DisplayName = topicARN[strings.LastIndex(topicARN, ":")+1:]

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &result, rd)
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

func listTopics(ctx context.Context, conn *sns.Client, input *sns.ListTopicsInput) iter.Seq2[awstypes.Topic, error] {
	return func(yield func(awstypes.Topic, error) bool) {
		pages := sns.NewListTopicsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Topic{}, fmt.Errorf("listing SNS Topics: %w", err))
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
