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
// @Tags
// @Testing(tagsTest=false)
func newVPNConnectionDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &vpnConnectionDataSource{}, nil
}

type vpnConnectionDataSource struct {
	framework.DataSourceWithModel[vpnConnectionDataSourceModel]
}

func (d *vpnConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"routes": framework.DataSourceComputedListOfObjectAttribute[vpnStaticRouteModel](ctx),
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
			"vpn_concentrator_id": schema.StringAttribute{
				Computed: true,
			},
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

func (d *vpnConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().EC2Client(ctx)
	var data vpnConnectionDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	input := ec2.DescribeVpnConnectionsInput{}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &input, flex.WithIgnoredFieldNamesAppend("VpnConnectionId")), smerr.ID)

	if !data.VPNConnectionID.IsNull() && !data.VPNConnectionID.IsUnknown() {
		input.VpnConnectionIds = []string{data.VPNConnectionID.ValueString()}
	}

	out, err := findVPNConnection(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data), smerr.ID, data.VPNConnectionID.String())
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.VPNConnectionID.String())
}

func (d *vpnConnectionDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("vpn_connection_id"),
			path.MatchRoot(names.AttrFilter),
		),
	}
}

type vpnConnectionDataSourceModel struct {
	framework.WithRegionModel
	Category                     types.String                                         `tfsdk:"category"`
	CoreNetworkARN               types.String                                         `tfsdk:"core_network_arn"`
	CoreNetworkAttachmentARN     types.String                                         `tfsdk:"core_network_attachment_arn"`
	CustomerGatewayConfiguration types.String                                         `tfsdk:"customer_gateway_configuration"`
	CustomerGatewayID            types.String                                         `tfsdk:"customer_gateway_id"`
	Filters                      customFilters                                        `tfsdk:"filter"`
	GatewayAssociationState      fwtypes.StringEnum[awstypes.GatewayAssociationState] `tfsdk:"gateway_association_state"`
	PreSharedKeyARN              types.String                                         `tfsdk:"pre_shared_key_arn"`
	Routes                       fwtypes.ListNestedObjectValueOf[vpnStaticRouteModel] `tfsdk:"routes"`
	State                        fwtypes.StringEnum[awstypes.VpnState]                `tfsdk:"state"`
	Tags                         tftags.Map                                           `tfsdk:"tags"`
	TransitGatewayID             types.String                                         `tfsdk:"transit_gateway_id"`
	Type                         types.String                                         `tfsdk:"type"`
	VGWTelemetry                 fwtypes.ListNestedObjectValueOf[vgwTelemetryModel]   `tfsdk:"vgw_telemetries"`
	VPNConcentratorID            types.String                                         `tfsdk:"vpn_concentrator_id"`
	VPNConnectionID              types.String                                         `tfsdk:"vpn_connection_id"`
	VPNGatewayID                 types.String                                         `tfsdk:"vpn_gateway_id"`
}

type vpnStaticRouteModel struct {
	DestinationCIDRBlock types.String                                      `tfsdk:"destination_cidr_block"`
	Source               fwtypes.StringEnum[awstypes.VpnStaticRouteSource] `tfsdk:"source"`
	State                fwtypes.StringEnum[awstypes.VpnState]             `tfsdk:"state"`
}

type vgwTelemetryModel struct {
	AcceptedRouteCount types.Int64                                  `tfsdk:"accepted_route_count"`
	LastStatusChange   timetypes.RFC3339                            `tfsdk:"last_status_change"`
	OutsideIPAddress   types.String                                 `tfsdk:"outside_ip_address"`
	Status             fwtypes.StringEnum[awstypes.TelemetryStatus] `tfsdk:"status"`
	StatusMessage      types.String                                 `tfsdk:"status_message"`
}
