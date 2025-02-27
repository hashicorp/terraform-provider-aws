// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_cloudwatch_event_buses", name="Event Buses")
func newEventBusesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &eventBusesDataSource{}, nil
}

type eventBusesDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *eventBusesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"event_buses": framework.DataSourceComputedListOfObjectAttribute[eventBusModel](ctx),
			names.AttrNamePrefix: schema.StringAttribute{
				Optional: true,
			},
		},
	}
}
func (d *eventBusesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data eventBusesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EventsClient(ctx)

	input := eventbridge.ListEventBusesInput{
		NamePrefix: fwflex.StringFromFramework(ctx, data.NamePrefix),
	}

	output, err := findEventBuses(ctx, conn, &input)

	if err != nil {
		response.Diagnostics.AddError("reading EventBridge Event Buses", err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data.EventBuses)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type eventBusesDataSourceModel struct {
	EventBuses fwtypes.ListNestedObjectValueOf[eventBusModel] `tfsdk:"event_buses"`
	NamePrefix types.String                                   `tfsdk:"name_prefix"`
}

type eventBusModel struct {
	ARN              types.String      `tfsdk:"arn"`
	CreationTime     timetypes.RFC3339 `tfsdk:"creation_time"`
	Description      types.String      `tfsdk:"description"`
	LastModifiedTime timetypes.RFC3339 `tfsdk:"last_modified_time"`
	Name             types.String      `tfsdk:"name"`
	Policy           types.String      `tfsdk:"policy"`
}

func findEventBuses(ctx context.Context, conn *eventbridge.Client, input *eventbridge.ListEventBusesInput) ([]awstypes.EventBus, error) {
	var output []awstypes.EventBus

	err := listEventBusesPages(ctx, conn, input, func(page *eventbridge.ListEventBusesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.EventBuses...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
