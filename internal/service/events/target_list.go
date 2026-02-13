// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_cloudwatch_event_target")
func newTargetResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceTarget{}
	l.SetResourceSchema(resourceTarget())
	return &l
}

var _ list.ListResource = &listResourceTarget{}

type listResourceTarget struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceTarget) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"event_bus_name": listschema.StringAttribute{
				Required:    true,
				Description: "Name or ARN of the event bus associated with the rule.",
			},
			names.AttrRule: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the rule.",
			},
		},
	}
}

func (l *listResourceTarget) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EventsClient(ctx)

	var query listTargetModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing EventBridge Target")
	stream.Results = func(yield func(list.ListResult) bool) {
		input := &eventbridge.ListTargetsByRuleInput{
			EventBusName: query.EventBusName.ValueStringPointer(),
			Rule:         query.Rule.ValueStringPointer(),
			Limit:        aws.Int32(100),
		}

		for item, err := range listTargets(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			targetID := aws.ToString(item.Id)
			eventBusName := query.EventBusName.ValueString()
			ruleName := query.Rule.ValueString()

			id := targetCreateResourceID(eventBusName, ruleName, targetID)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(id)
			rd.Set("event_bus_name", eventBusName)
			rd.Set(names.AttrRule, ruleName)
			rd.Set("target_id", targetID)

			tflog.Info(ctx, "Reading EventBridge Target")
			diags := resourceTargetRead(ctx, rd, l.Meta())
			if diags.HasError() {
				tflog.Error(ctx, "Reading EventBridge Target", map[string]any{
					names.AttrID: id,
					"diags":      sdkdiag.DiagnosticsString(diags),
				})
				continue
			}
			if rd.Id() == "" {
				// Resource is logically deleted
				continue
			}

			result.DisplayName = targetID

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

type listTargetModel struct {
	framework.WithRegionModel
	EventBusName types.String `tfsdk:"event_bus_name"`
	Rule         types.String `tfsdk:"rule"`
}

func listTargets(ctx context.Context, conn *eventbridge.Client, input *eventbridge.ListTargetsByRuleInput) iter.Seq2[awstypes.Target, error] {
	return func(yield func(awstypes.Target, error) bool) {
		err := listTargetsByRulePages(ctx, conn, input, func(page *eventbridge.ListTargetsByRuleOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, item := range page.Targets {
				if !yield(item, nil) {
					return !lastPage
				}
			}

			return !lastPage
		})

		if err != nil {
			yield(awstypes.Target{}, fmt.Errorf("listing EventBridge Target resources: %w", err))
			return
		}
	}
}
