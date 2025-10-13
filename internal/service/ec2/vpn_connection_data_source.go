// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ec2_vpn_connection", name="VPN Connection")
func newDataSourceVPNConnection(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceVPNConnection{}, nil
}

const (
	DSNameVPNConnection = "VPN Connection Data Source"
)

type dataSourceVPNConnection struct {
	framework.DataSourceWithModel[dataSourceVPNConnectionModel]
}

func (d *dataSourceVPNConnection) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"vpn_connection_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"category": schema.StringAttribute{
				Computed: true,
			},
			"core_network_arn": schema.StringAttribute{
				Computed: true,
			},
			"core_network_attachment_arn": schema.StringAttribute{
				Computed: true,
			},
			"customer_gateway_configuration": schema.StringAttribute{
				Computed: true,
			},
			"customer_gateway_id": schema.StringAttribute{
				Computed: true,
			},
			"gateway_association_state": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.GatewayAssociationState](),
				Computed:   true,
			},
			"pre_shared_key_arn": schema.StringAttribute{
				Computed: true,
			},
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.VpnState](),
				Computed:   true,
			},
			names.AttrTransitGatewayID: schema.StringAttribute{
				Computed: true,
			},
			names.AttrType: schema.StringAttribute{
				Computed: true,
			},
			"vpn_gateway_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID:      framework.IDAttribute(),
			names.AttrTags:    tftags.TagsAttributeComputedOnly(),
			"routes":          framework.DataSourceComputedListOfObjectAttribute[routeModel](ctx),
			"vgw_telemetries": framework.DataSourceComputedListOfObjectAttribute[vgwTelemetryModel](ctx),
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: customFiltersBlock(ctx),
		},
	}
}

func (d *dataSourceVPNConnection) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().EC2Client(ctx)
	
	var data dataSourceVPNConnectionModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	input := ec2.DescribeVpnConnectionsInput{
		Filters:          newCustomFilterListFramework(ctx, data.Filters),
		VpnConnectionIds: []string{data.VpnConnectionId.ValueString()},
	}

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	if input.VpnConnectionIds[0] == "" {
		// Don't send an empty ID; the EC2 API won't accept it.
		input.VpnConnectionIds = nil
	}

	if input.Filters == nil && input.VpnConnectionIds == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("missing input"), smerr.ID)
		return
	}

	out, err := findVPNConnection(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data), smerr.ID, data.VpnConnectionId.String())
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(aws.ToString(out.VpnConnectionId))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.VpnConnectionId.String())
}

type dataSourceVPNConnectionModel struct {
	framework.WithRegionModel
	Filters         customFilters `tfsdk:"filter"`
	VpnConnectionId types.String  `tfsdk:"vpn_connection_id"`
	ID              types.String  `tfsdk:"id"`

	Category                     types.String                                         `tfsdk:"category"`
	CoreNetworkArn               types.String                                         `tfsdk:"core_network_arn"`
	CoreNetworkAttachmentArn     types.String                                         `tfsdk:"core_network_attachment_arn"`
	CustomerGatewayConfiguration types.String                                         `tfsdk:"customer_gateway_configuration"`
	CustomerGatewayID            types.String                                         `tfsdk:"customer_gateway_id"`
	GatewayAssociationState      fwtypes.StringEnum[awstypes.GatewayAssociationState] `tfsdk:"gateway_association_state"`
	PreSharedKeyArn              types.String                                         `tfsdk:"pre_shared_key_arn"`
	State                        fwtypes.StringEnum[awstypes.VpnState]                `tfsdk:"state"`
	TransitGatewayId             types.String                                         `tfsdk:"transit_gateway_id"`
	Type                         types.String                                         `tfsdk:"type"`
	VpnGatewayId                 types.String                                         `tfsdk:"vpn_gateway_id"`
	Routes                       fwtypes.ListNestedObjectValueOf[routeModel]          `tfsdk:"routes"`
	VgwTelemetries               fwtypes.ListNestedObjectValueOf[vgwTelemetryModel]   `tfsdk:"vgw_telemetries"`
	Tags                         tftags.Map                                           `tfsdk:"tags"`
}

type routeModel struct {
	DestinationCidrBlock types.String                          `tfsdk:"destination_cidr_block"`
	Source               types.String                          `tfsdk:"source"`
	State                fwtypes.StringEnum[awstypes.VpnState] `tfsdk:"state"`
}

type vgwTelemetryModel struct {
	AcceptedRouteCount types.Int64                                  `tfsdk:"accepted_route_count"`
	LastStatusChange   timetypes.RFC3339                            `tfsdk:"last_status_change"`
	Status             fwtypes.StringEnum[awstypes.TelemetryStatus] `tfsdk:"status"`
	StatusMessage      types.String                                 `tfsdk:"status_message"`
	OutsideIpAddress   types.String                                 `tfsdk:"outside_ip_address"`
}
