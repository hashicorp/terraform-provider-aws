// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_cloudwatch_log_subscription_filter")
func newSubscriptionFilterResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceSubscriptionFilter(), framework.SDKListSpec[awstypes.SubscriptionFilter, listSubscriptionFilterModel]{
		ConfigSchema: listSubscriptionFilterConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listSubscriptionFilterModel) (iter.Seq2[awstypes.SubscriptionFilter, error], diag.Diagnostics) {
			input := cloudwatchlogs.DescribeSubscriptionFiltersInput{
				LogGroupName: fwflex.StringFromFramework(ctx, query.LogGroupName),
			}

			return listSubscriptionFilters(ctx, c.LogsClient(ctx), &input), nil
		},
		// The resource ID is the log group name, which comes from the query
		// rather than the item.
		ID: func(_ awstypes.SubscriptionFilter, query listSubscriptionFilterModel) string {
			return subscriptionFilterCreateResourceID(query.LogGroupName.ValueString())
		},
		Flatten:     resourceSubscriptionFilterFlatten,
		DisplayName: func(sf awstypes.SubscriptionFilter) string { return aws.ToString(sf.FilterName) },
	})
}

type listSubscriptionFilterModel struct {
	framework.WithRegionModel
	LogGroupName types.String `tfsdk:"log_group_name"`
}

func listSubscriptionFilters(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeSubscriptionFiltersInput, optFns ...func(*cloudwatchlogs.Options)) iter.Seq2[awstypes.SubscriptionFilter, error] {
	return tfiter.ConcatValuesWithError(listSubscriptionFilterPages(ctx, conn, input, optFns...))
}

func listSubscriptionFilterConfigSchema(_ context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrLogGroupName: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the log group.",
			},
		},
	}
}
