// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_cloudwatch_metric_alarm")
func newMetricAlarmResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceMetricAlarm{}
	l.SetResourceSchema(resourceMetricAlarm())
	return &l
}

var _ list.ListResource = &listResourceMetricAlarm{}

type listResourceMetricAlarm struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceMetricAlarm) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().CloudWatchClient(ctx)

	var query listMetricAlarmModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing CloudWatch Metric Alarm")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input cloudwatch.DescribeAlarmsInput
		input.AlarmTypes = []awstypes.AlarmType{awstypes.AlarmTypeMetricAlarm}
		for item, err := range listMetricAlarms(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(item.AlarmName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), name)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(name)

			tflog.Info(ctx, "Reading CloudWatch Metric Alarm")
			diags := resourceMetricAlarmRead(ctx, rd, l.Meta())
			if diags.HasError() {
				tflog.Error(ctx, "Reading CloudWatch Metric Alarm", map[string]any{
					names.AttrID: name,
					"diags":      sdkdiag.DiagnosticsString(diags),
				})
				continue
			}
			if rd.Id() == "" {
				// Resource is logically deleted
				continue
			}

			result.DisplayName = name

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

type listMetricAlarmModel struct {
	framework.WithRegionModel
}

func listMetricAlarms(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.DescribeAlarmsInput) iter.Seq2[awstypes.MetricAlarm, error] {
	return func(yield func(awstypes.MetricAlarm, error) bool) {
		pages := cloudwatch.NewDescribeAlarmsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.MetricAlarm{}, fmt.Errorf("listing CloudWatch Metric Alarm resources: %w", err))
				return
			}

			for _, item := range page.MetricAlarms {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
