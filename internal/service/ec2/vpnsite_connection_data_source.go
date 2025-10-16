// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_vpn_connection", name="VPN Connection")
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
			"routes": framework.DataSourceComputedListOfObjectAttribute[routeModel](ctx),
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.VpnState](),
				Computed:   true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			names.AttrTransitGatewayID: schema.StringAttribute{
				Computed: true,
			},
			names.AttrType: schema.StringAttribute{
				Computed: true,
			},
			"vgw_telemetries": framework.DataSourceComputedListOfObjectAttribute[vgwTelemetryModel](ctx),
			"vpn_connection_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"vpn_gateway_id": schema.StringAttribute{
				Computed: true,
			},
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

	input := ec2.DescribeVpnConnectionsInput{}
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &input, flex.WithIgnoredFieldNamesAppend("VpnConnectionId")), smerr.ID)

	if !data.VpnConnectionId.IsNull() && !data.VpnConnectionId.IsUnknown() {
		input.VpnConnectionIds = []string{data.VpnConnectionId.ValueString()}
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
	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.VpnConnectionId.String())
}

func (d *dataSourceVPNConnection) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("vpn_connection_id"),
			path.MatchRoot(names.AttrFilter),
		),
	}
}

type dataSourceVPNConnectionModel struct {
	framework.WithRegionModel
	Category                     types.String                                         `tfsdk:"category"`
	CoreNetworkArn               types.String                                         `tfsdk:"core_network_arn"`
	CoreNetworkAttachmentArn     types.String                                         `tfsdk:"core_network_attachment_arn"`
	CustomerGatewayConfiguration types.String                                         `tfsdk:"customer_gateway_configuration"`
	CustomerGatewayID            types.String                                         `tfsdk:"customer_gateway_id"`
	Filters                      customFilters                                        `tfsdk:"filter"`
	GatewayAssociationState      fwtypes.StringEnum[awstypes.GatewayAssociationState] `tfsdk:"gateway_association_state"`
	PreSharedKeyArn              types.String                                         `tfsdk:"pre_shared_key_arn"`
	State                        fwtypes.StringEnum[awstypes.VpnState]                `tfsdk:"state"`
	TransitGatewayId             types.String                                         `tfsdk:"transit_gateway_id"`
	Type                         types.String                                         `tfsdk:"type"`
	VpnGatewayId                 types.String                                         `tfsdk:"vpn_gateway_id"`
	Routes                       fwtypes.ListNestedObjectValueOf[routeModel]          `tfsdk:"routes"`
	VgwTelemetries               fwtypes.ListNestedObjectValueOf[vgwTelemetryModel]   `tfsdk:"vgw_telemetries"`
	VpnConnectionId              types.String                                         `tfsdk:"vpn_connection_id"`
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
