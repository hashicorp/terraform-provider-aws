// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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
	TopicARN types.String `tfsdk:"topic_arn"`
}

func (l *topicSubscriptionListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrTopicARN: listschema.StringAttribute{
				Required:    true,
				Description: `The ARN of the SNS topic whose subscriptions to list.`,
			},
		},
	}
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

	input := sns.ListSubscriptionsByTopicInput{
		TopicArn: query.TopicARN.ValueStringPointer(),
	}

	tflog.Info(ctx, "Listing SNS Topic Subscriptions")

	stream.Results = func(yield func(list.ListResult) bool) {
		for subscription, err := range tfiter.ConcatValuesWithError(listSubscriptionsByTopic(ctx, conn, &input)) {
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
			rd.Set(names.AttrARN, arn)

			if request.IncludeResource {
				attributes, err := findSubscriptionAttributesByARN(ctx, conn, arn)
				if err != nil {
					tflog.Error(ctx, "Reading SNS Topic Subscription", map[string]any{
						"err": err.Error(),
					})
					continue
				}

				if err := subscriptionAttributeMap.APIAttributesToResourceData(attributes, rd); err != nil {
					tflog.Error(ctx, "Reading SNS Topic Subscription", map[string]any{
						"err": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = arn

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
