// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_cloudwatch_alarm_mute_rule")
func newAlarmMuteRuleResourceAsListResource() list.ListResourceWithConfigure {
	return &alarmMuteRuleListResource{}
}

var _ list.ListResource = &alarmMuteRuleListResource{}

type alarmMuteRuleListResource struct {
	alarmMuteRuleResource
	framework.WithList
}

func (l *alarmMuteRuleListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().CloudWatchClient(ctx)

	var query listAlarmMuteRuleModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input cloudwatch.ListAlarmMuteRulesInput
		for item, err := range listAlarmMuteRules(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			alarmMuteRuleARN := aws.ToString(item.AlarmMuteRuleArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), alarmMuteRuleARN)

			// arn:${Partition}:cloudwatch:${Region}:${Account}:alarm-mute-rule:${AlarmMuteRuleName}
			v, err := arn.Parse(alarmMuteRuleARN)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := strings.TrimPrefix(v.Resource, "alarm-mute-rule:")
			out, err := findAlarmMuteRuleByName(ctx, conn, name)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			var data alarmMuteRuleResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, out, &data))
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

type listAlarmMuteRuleModel struct {
	framework.WithRegionModel
}

func listAlarmMuteRules(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.ListAlarmMuteRulesInput) iter.Seq2[awstypes.AlarmMuteRuleSummary, error] {
	return func(yield func(awstypes.AlarmMuteRuleSummary, error) bool) {
		pages := cloudwatch.NewListAlarmMuteRulesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(inttypes.Zero[awstypes.AlarmMuteRuleSummary](), fmt.Errorf("listing CloudWatch Alarm Mute Rules: %w", err))
				return
			}

			for _, item := range page.AlarmMuteRuleSummaries {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
