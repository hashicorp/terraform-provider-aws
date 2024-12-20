// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

// @FrameworkDataSource(name="Event Buses")
func newDataSourceEventBuses(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceEventBuses{}, nil
}

const (
	DSNameEventBuses = "Event Buses Data Source"
)

type dataSourceEventBuses struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceEventBuses) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_cloudwatch_event_buses"
}

func (d *dataSourceEventBuses) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"name_prefix": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"event_buses": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[eventBustModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"arn": schema.StringAttribute{
							Computed: true,
						},
						"creation_time": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"last_modified_time": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"policy": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}
func (d *dataSourceEventBuses) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().EventsClient(ctx)

	var data dataSourceEventBusesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(d.Meta().AccountID)

	input := &eventbridge.ListEventBusesInput{
		NamePrefix: data.NamePrefix.ValueStringPointer(),
	}

	out, err := findEventBuses(ctx, conn, input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Events, create.ErrActionReading, DSNameEventBuses, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data.EventBuses)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceEventBusesModel struct {
	ID         types.String                                    `tfsdk:"id"`
	NamePrefix types.String                                    `tfsdk:"name_prefix"`
	EventBuses fwtypes.ListNestedObjectValueOf[eventBustModel] `tfsdk:"event_buses"`
}

type eventBustModel struct {
	Arn              types.String      `tfsdk:"arn"`
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
