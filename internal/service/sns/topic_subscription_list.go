// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_sns_topic_subscription")
func newTopicSubscriptionResourceAsListResource() inttypes.ListResourceForSDK {
	l := topicSubscriptionListResource{}
	l.SetResourceSchema(resourceTopicSubscription())
	return &l
}

type topicSubscriptionListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type topicSubscriptionListResourceModel struct {
	framework.WithRegionModel
}

func (l *topicSubscriptionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.SNSClient(ctx)

	var query topicSubscriptionListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input sns.ListSubscriptionsInput

	tflog.Info(ctx, "Listing SNS Topic Subscriptions")

	stream.Results = func(yield func(list.ListResult) bool) {
		for subscription, err := range listSubscriptions(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(subscription.SubscriptionArn)
			if arn == "PendingConfirmation" {
				continue
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), arn)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(arn)

			tflog.Info(ctx, "Reading SNS Topic Subscription")
			diags := resourceTopicSubscriptionRead(ctx, rd, awsClient)
			if diags.HasError() {
				tflog.Error(ctx, "Error reading SNS Topic Subscription", map[string]any{
					"arn":   arn,
					"error": diags,
				})
				continue
			}

			if rd.Id() == "" {
				continue
			}

			result.DisplayName = arn

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
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

func listSubscriptions(ctx context.Context, conn *sns.Client, input *sns.ListSubscriptionsInput) iter.Seq2[awstypes.Subscription, error] {
	return func(yield func(awstypes.Subscription, error) bool) {
		pages := sns.NewListSubscriptionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Subscription{}, fmt.Errorf("listing SNS Topic Subscriptions: %w", err))
				return
			}

			for _, subscription := range page.Subscriptions {
				if !yield(subscription, nil) {
					return
				}
			}
		}
	}
}
