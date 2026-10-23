// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_cloudwatch_metric_alarm")
func newMetricAlarmResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceMetricAlarm(), framework.SDKListSpec[awstypes.MetricAlarm, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.MetricAlarm, error], diag.Diagnostics) {
			input := cloudwatch.DescribeAlarmsInput{
				AlarmTypes: []awstypes.AlarmType{awstypes.AlarmTypeMetricAlarm},
			}

			return listMetricAlarms(ctx, c.CloudWatchClient(ctx), &input), nil
		},
		ID: func(a awstypes.MetricAlarm, _ framework.RegionQuery) string {
			return aws.ToString(a.AlarmName)
		},
		Flatten: resourceMetricAlarmFlatten,
	})
}

func listMetricAlarms(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.DescribeAlarmsInput) iter.Seq2[awstypes.MetricAlarm, error] {
	return tfiter.PageItems[cloudwatch.Options](ctx, "listing CloudWatch Metric Alarms", cloudwatch.NewDescribeAlarmsPaginator(conn, input),
		func(o *cloudwatch.DescribeAlarmsOutput) []awstypes.MetricAlarm { return o.MetricAlarms })
}
