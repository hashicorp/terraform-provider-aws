// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventbridgev2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eventbridgev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eventbridgev2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_eventbridgev2_event_buses", name="Event Buses")
func newEventBusesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &eventBusesDataSource{}, nil
}

type eventBusesDataSource struct {
	framework.DataSourceWithModel[eventBusesDataSourceModel]
}

func (d *eventBusesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"event_bus_account_id": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
			},
			"event_buses": framework.DataSourceComputedListOfObjectAttribute[eventBusSummaryModel](ctx),
			names.AttrNamePrefix: schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (d *eventBusesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data eventBusesDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EventBridgeV2Client(ctx)

	input := eventbridgev2.ListEventBusesInput{
		EventBusAccountId: fwflex.StringFromFramework(ctx, data.EventBusAccountID),
		NamePrefix:        fwflex.StringFromFramework(ctx, data.NamePrefix),
	}

	eventBuses := make([]awstypes.EventBusSummary, 0)
	for v, err := range listEventBuses(ctx, conn, &input) {
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err)
			return
		}

		eventBuses = append(eventBuses, v)
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, eventBuses, &data.EventBuses, fwflex.WithFieldNamePrefix("EventBus")))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type eventBusesDataSourceModel struct {
	framework.WithRegionModel
	EventBusAccountID types.String                                          `tfsdk:"event_bus_account_id"`
	EventBuses        fwtypes.ListNestedObjectValueOf[eventBusSummaryModel] `tfsdk:"event_buses"`
	NamePrefix        types.String                                          `tfsdk:"name_prefix"`
}

type eventBusSummaryModel struct {
	AccountID        types.String                          `tfsdk:"account_id"`
	ARN              types.String                          `tfsdk:"arn"`
	CreationTime     timetypes.RFC3339                     `tfsdk:"creation_time"`
	Description      types.String                          `tfsdk:"description"`
	LastModifiedTime timetypes.RFC3339                     `tfsdk:"last_modified_time"`
	Name             types.String                          `tfsdk:"name"`
	State            fwtypes.StringEnum[awstypes.BusState] `tfsdk:"state"`
	StateReason      types.String                          `tfsdk:"state_reason"`
}
