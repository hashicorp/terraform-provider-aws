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

// @SDKListResource("aws_cloudwatch_log_metric_filter")
func newMetricFilterResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceMetricFilter(), framework.SDKListSpec[awstypes.MetricFilter, listMetricFilterModel]{
		ConfigSchema: listMetricFilterConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listMetricFilterModel) (iter.Seq2[awstypes.MetricFilter, error], diag.Diagnostics) {
			input := cloudwatchlogs.DescribeMetricFiltersInput{
				LogGroupName: aws.String(fwflex.StringValueFromFramework(ctx, query.LogGroupName)),
			}

			return listMetricFilters(ctx, c.LogsClient(ctx), &input), nil
		},
		ID: func(mf awstypes.MetricFilter, _ listMetricFilterModel) string {
			return aws.ToString(mf.FilterName)
		},
		Flatten: resourceMetricFilterFlatten,
	})
}

type listMetricFilterModel struct {
	framework.WithRegionModel
	LogGroupName types.String `tfsdk:"log_group_name"`
}

func listMetricFilters(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeMetricFiltersInput, optFns ...func(*cloudwatchlogs.Options)) iter.Seq2[awstypes.MetricFilter, error] {
	return tfiter.ConcatValuesWithError(listMetricFilterPages(ctx, conn, input, optFns...))
}

func listMetricFilterConfigSchema(_ context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrLogGroupName: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the log group.",
			},
		},
	}
}
