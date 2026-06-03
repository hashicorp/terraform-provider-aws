// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_cloudwatch_log_subscription_filter")
func newSubscriptionFilterResourceAsListResource() inttypes.ListResourceForSDK {
	l := subscriptionFilterListResource{}
	l.SetResourceSchema(resourceSubscriptionFilter())
	return &l
}

var _ list.ListResource = &subscriptionFilterListResource{}

type subscriptionFilterListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *subscriptionFilterListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrLogGroupName: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the log group.",
			},
		},
	}
}

func (l *subscriptionFilterListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().LogsClient(ctx)

	var query listSubscriptionFilterModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	logGroupName := fwflex.StringValueFromFramework(ctx, query.LogGroupName)

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey(names.AttrLogGroupName): logGroupName,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := cloudwatchlogs.DescribeSubscriptionFiltersInput{
			LogGroupName: aws.String(logGroupName),
		}
		for item, err := range listSubscriptionFilters(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			logGroupName, name := aws.ToString(item.LogGroupName), aws.ToString(item.FilterName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(subscriptionFilterCreateResourceID(logGroupName))
			rd.Set(names.AttrLogGroupName, logGroupName)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				if err := resourceSubscriptionFilterFlatten(ctx, &item, rd); err != nil {
					tflog.Error(ctx, "Reading CloudWatch Logs Subscription Filter", map[string]any{
						"error": err.Error(),
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

type listSubscriptionFilterModel struct {
	framework.WithRegionModel
	LogGroupName types.String `tfsdk:"log_group_name"`
}

func listSubscriptionFilters(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeSubscriptionFiltersInput) iter.Seq2[awstypes.SubscriptionFilter, error] {
	return func(yield func(awstypes.SubscriptionFilter, error) bool) {
		pages := cloudwatchlogs.NewDescribeSubscriptionFiltersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.SubscriptionFilter](), fmt.Errorf("listing CloudWatch Logs Subscription Filters: %w", err))
				return
			}

			for _, item := range page.SubscriptionFilters {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
