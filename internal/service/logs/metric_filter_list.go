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

// @SDKListResource("aws_cloudwatch_log_metric_filter")
func newMetricFilterResourceAsListResource() inttypes.ListResourceForSDK {
	l := metricFilterListResource{}
	l.SetResourceSchema(resourceMetricFilter())
	return &l
}

var _ list.ListResource = &metricFilterListResource{}

type metricFilterListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *metricFilterListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrLogGroupName: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the log group.",
			},
		},
	}
}

func (l *metricFilterListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().LogsClient(ctx)

	var query listMetricFilterModel
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

	// TIP: -- 4. Get information about a resource from AWS
	stream.Results = func(yield func(list.ListResult) bool) {
		input := cloudwatchlogs.DescribeMetricFiltersInput{
			LogGroupName: aws.String(logGroupName),
		}
		for item, err := range listMetricFilters(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			logGroupName, name := aws.ToString(item.LogGroupName), aws.ToString(item.FilterName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set(names.AttrLogGroupName, logGroupName)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				if err := resourceMetricFilterFlatten(ctx, &item, rd); err != nil {
					tflog.Error(ctx, "Reading CloudWatch Logs Metric Filter", map[string]any{
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

type listMetricFilterModel struct {
	framework.WithRegionModel
	LogGroupName types.String `tfsdk:"log_group_name"`
}

func listMetricFilters(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeMetricFiltersInput) iter.Seq2[awstypes.MetricFilter, error] {
	return func(yield func(awstypes.MetricFilter, error) bool) {
		pages := cloudwatchlogs.NewDescribeMetricFiltersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.MetricFilter](), fmt.Errorf("listing CloudWatch Logs Metric Filters: %w", err))
				return
			}

			for _, item := range page.MetricFilters {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
