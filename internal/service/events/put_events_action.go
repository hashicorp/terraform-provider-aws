// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_events_put_events, name="Put Events")
// nosemgrep: ci.events-in-func-name -- "PutEvents" matches AWS API operation name (PutEvents). Required for consistent generated/action naming; safe to ignore.
func newPutEventsAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &putEventsAction{}, nil
}

var (
	_ action.Action = (*putEventsAction)(nil)
)

type putEventsAction struct {
	framework.ActionWithModel[putEventsActionModel]
}

type putEventsActionModel struct {
	framework.WithRegionModel
	Entry fwtypes.ListNestedObjectValueOf[putEventEntryModel] `tfsdk:"entry"`
}

type putEventEntryModel struct {
	Detail       types.String                      `tfsdk:"detail"`
	DetailType   types.String                      `tfsdk:"detail_type"`
	EventBusName types.String                      `tfsdk:"event_bus_name"`
	Resources    fwtypes.ListValueOf[types.String] `tfsdk:"resources"`
	Source       types.String                      `tfsdk:"source"`
	Time         timetypes.RFC3339                 `tfsdk:"time"`
}

func (a *putEventsAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sends custom events to Amazon EventBridge so that they can be matched to rules.",
		Blocks: map[string]schema.Block{
			"entry": schema.ListNestedBlock{
				Description: "The entry that defines an event in your system.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[putEventEntryModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"detail": schema.StringAttribute{
							Description: "A valid JSON string. There is no other schema imposed.",
							Optional:    true,
						},
						"detail_type": schema.StringAttribute{
							Description: "Free-form string used to decide what fields to expect in the event detail.",
							Optional:    true,
						},
						"event_bus_name": schema.StringAttribute{
							Description: "The name or ARN of the event bus to receive the event.",
							Optional:    true,
						},
						names.AttrResources: schema.ListAttribute{
							Description: "AWS resources, identified by Amazon Resource Name (ARN), which the event primarily concerns.",
							CustomType:  fwtypes.ListOfStringType,
							Optional:    true,
						},
						names.AttrSource: schema.StringAttribute{
							Description: "The source of the event.",
							Required:    true,
						},
						"time": schema.StringAttribute{
							Description: "The time stamp of the event, per RFC3339.",
							Optional:    true,
							CustomType:  timetypes.RFC3339Type{},
						},
					},
				},
			},
		},
	}
}

func (a *putEventsAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var model putEventsActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().EventsClient(ctx)

	tflog.Info(ctx, "Putting events", map[string]any{
		"entry_count": len(model.Entry.Elements()),
	})

	resp.SendProgress(action.InvokeProgressEvent{
		Message: "Putting events to EventBridge...",
	})

	var input eventbridge.PutEventsInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, model, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.PutEvents(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Putting Events",
			"Could not put events: "+err.Error(),
		)
		return
	}

	if output.FailedEntryCount > 0 {
		resp.Diagnostics.AddError(
			"Putting Events",
			strconv.Itoa(int(output.FailedEntryCount))+" entries failed to be processed",
		)
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: "Events put successfully",
	})

	tflog.Info(ctx, "Put events completed", map[string]any{
		"successful_entries": len(output.Entries),
	})
}
