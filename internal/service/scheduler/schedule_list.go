// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package scheduler

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	awstypes "github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_scheduler_schedule", name="Schedule")
func newScheduleResourceAsListResource() inttypes.ListResourceForSDK {
	l := scheduleListResource{}
	l.SetResourceSchema(resourceSchedule())
	return &l
}

var _ list.ListResource = &scheduleListResource{}

type scheduleListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *scheduleListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SchedulerClient(ctx)

	var query listScheduleModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing Resources", map[string]any{})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := scheduler.ListSchedulesInput{}
		for item, err := range listSchedules(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			groupName := aws.ToString(item.GroupName)
			scheduleName := aws.ToString(item.Name)
			id := scheduleCreateResourceID(groupName, scheduleName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(id)
			rd.Set(names.AttrGroupName, groupName)
			rd.Set(names.AttrName, scheduleName)

			if request.IncludeResource {
				out, err := findScheduleByTwoPartKey(ctx, conn, groupName, scheduleName)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					result = fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading EventBridge Scheduler Schedule (%s): %w", id, err))
					yield(result)
					return
				}

				if err := resourceScheduleFlatten(ctx, out, rd); err != nil {
					tflog.Error(ctx, "Flatten EventBridge Scheduler Schedule", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = scheduleName
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

type listScheduleModel struct {
	framework.WithRegionModel
}

func listSchedules(ctx context.Context, conn *scheduler.Client, input *scheduler.ListSchedulesInput) iter.Seq2[awstypes.ScheduleSummary, error] {
	return func(yield func(awstypes.ScheduleSummary, error) bool) {
		pages := scheduler.NewListSchedulesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ScheduleSummary{}, fmt.Errorf("listing EventBridge Scheduler Schedule resources: %w", err))
				return
			}

			for _, item := range page.Schedules {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
