package ec2

import (
	"context"
	"encoding/xml"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// ...DataSource
// ...DataSourceModel
// ...Model

// @FrameworkDataSource("aws_vpn_connection", name="VPN Connection")
func newVpnSiteConnectionDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &vpnSiteConnectionDataSource{}

	return d, nil
}

type vpnSiteConnectionDataSource struct {
	framework.DataSourceWithModel[vpnSiteConnectionDataSourceModel]
}

func (d *vpnSiteConnectionDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Required: true,
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
			"enable_acceleration": schema.BoolAttribute{
				Computed: true,
			},
			"local_ipv4_network_cidr": schema.StringAttribute{
				Computed: true,
			},
			"local_ipv6_network_cidr": schema.StringAttribute{
				Computed: true,
			},
			"outside_ip_address_type": schema.StringAttribute{
				Computed: true,
			},
			"preshared_key_arn": schema.StringAttribute{
				Computed: true,
			},
			/*"preshared_key_storage": schema.StringAttribute{
				Computed: true,
			},*/
			"remote_ipv4_network_cidr": schema.StringAttribute{
				Computed: true,
			},
			"remote_ipv6_network_cidr": schema.StringAttribute{
				Computed: true,
			},
			/*"routes": schema.SetAttribute{
				Computed: true,
			},*/
			"static_routes_only": schema.BoolAttribute{
				Computed: true,
			}, /*
				names.AttrTags: tftags.TagsAttributeComputedOnly(),
				/*					"tags_all": schema.SetAttribute{
									Computed: true,
								},*/
			"transit_gateway_attachment_id": schema.StringAttribute{
				Computed: true,
			},
			"transit_gateway_id": schema.StringAttribute{
				Computed: true,
			},
			"transport_transit_gateway_attachment_id": schema.StringAttribute{
				Computed: true,
			},
			"tunnel1_address": schema.StringAttribute{
				Computed: true,
			},
			"tunnel1_bgp_asn": schema.StringAttribute{
				Computed: true,
			},
			"tunnel1_bgp_holdtime": schema.Int64Attribute{
				Computed: true,
			},
			"tunnel1_cgw_inside_address": schema.StringAttribute{
				Computed: true,
			},
			"tunnel1_vgw_inside_address": schema.StringAttribute{
				Computed: true,
			},
			"tunnel2_address": schema.StringAttribute{
				Computed: true,
			},
			"tunnel2_bgp_asn": schema.StringAttribute{
				Computed: true,
			},
			"tunnel2_bgp_holdtime": schema.Int64Attribute{
				Computed: true,
			},
			"tunnel2_cgw_inside_address": schema.StringAttribute{
				Computed: true,
			},

			/*
				"tunnel1_dpd_timeout_action": schema.StringAttribute{
					Computed: true,
				},
				"tunnel1_dpd_timeout_seconds": schema.Int64Attribute{
					Computed: true,
				},
				"tunnel1_enable_tunnel_lifecycle_control": schema.BoolAttribute{
					Computed: true,
				},
				"tunnel1_ike_versions": schema.SetAttribute{
					Computed: true,
				},
				"tunnel1_inside_cidr": schema.StringAttribute{
					Computed: true,
				},
				"tunnel1_inside_ipv6_cidr": schema.StringAttribute{
					Computed: true,
				},
				"tunnel1_log_options": schema.ListAttribute{
					Computed: true,
				},
				"tunnel1_phase1_dh_group_numbers": schema.SetAttribute{
					Computed: true,
				},
				"tunnel1_phase1_encryption_algorithms": schema.SetAttribute{
					Computed: true,
				},
				"tunnel1_phase1_integrity_algorithms": schema.SetAttribute{
					Computed: true,
				},
				"tunnel1_phase1_lifetime_seconds": schema.Int64Attribute{
					Computed: true,
				},
				"tunnel1_phase2_dh_group_numbers": schema.SetAttribute{
					Computed: true,
				},
				"tunnel1_phase2_encryption_algorithms": schema.SetAttribute{
					Computed: true,
				},
				"tunnel1_phase2_integrity_algorithms": schema.SetAttribute{
					Computed: true,
				},
				"tunnel1_phase2_lifetime_seconds": schema.Int64Attribute{
					Computed: true,
				},
				"tunnel1_preshared_key": schema.StringAttribute{
					Computed: true,
				},
				"tunnel1_rekey_fuzz_percentage": schema.Int64Attribute{
					Computed: true,
				},
				"tunnel1_rekey_margin_time_seconds": schema.Int64Attribute{
					Computed: true,
				},
				"tunnel1_replay_window_size": schema.Int64Attribute{
					Computed: true,
				},
				"tunnel1_startup_action": schema.StringAttribute{
					Computed: true,
				},*/
			/*

				"tunnel2_dpd_timeout_action": schema.StringAttribute{
					Computed: true,
				},
				"tunnel2_dpd_timeout_seconds": schema.Int64Attribute{
					Computed: true,
				},
				"tunnel2_enable_tunnel_lifecycle_control": schema.BoolAttribute{
					Computed: true,
				},
				"tunnel2_ike_versions": schema.SetAttribute{
					Computed: true,
				},
				"tunnel2_inside_cidr": schema.StringAttribute{
					Computed: true,
				},
				"tunnel2_inside_ipv6_cidr": schema.StringAttribute{
					Computed: true,
				},
				"tunnel2_log_options": schema.ListAttribute{
					Computed: true,
				},
				"tunnel2_phase1_dh_group_numbers": schema.SetAttribute{
					Computed: true,
				},
				"tunnel2_phase1_encryption_algorithms": schema.SetAttribute{
					Computed: true,
				},
				"tunnel2_phase1_integrity_algorithms": schema.SetAttribute{
					Computed: true,
				},
				"tunnel2_phase1_lifetime_seconds": schema.Int64Attribute{
					Computed: true,
				},
				"tunnel2_phase2_dh_group_numbers": schema.SetAttribute{
					Computed: true,
				},
				"tunnel2_phase2_encryption_algorithms": schema.SetAttribute{
					Computed: true,
				},
				"tunnel2_phase2_integrity_algorithms": schema.SetAttribute{
					Computed: true,
				},
				"tunnel2_phase2_lifetime_seconds": schema.Int64Attribute{
					Computed: true,
				},
				"tunnel2_preshared_key": schema.StringAttribute{
					Computed: true,
				},
				"tunnel2_rekey_fuzz_percentage": schema.Int64Attribute{
					Computed: true,
				},
				"tunnel2_rekey_margin_time_seconds": schema.Int64Attribute{
					Computed: true,
				},
				"tunnel2_replay_window_size": schema.Int64Attribute{
					Computed: true,
				},
				"tunnel2_startup_action": schema.StringAttribute{
					Computed: true,
				},
				"tunnel2_vgw_inside_address": schema.StringAttribute{
					Computed: true,
				},
				"tunnel_inside_ip_version": schema.StringAttribute{
					Computed: true,
				},
				"type": schema.StringAttribute{
					Computed: true,
				},
				"vgw_telemetry": schema.SetAttribute{
					Computed: true,
				},
				"vpn_gateway_id": schema.StringAttribute{
					Computed: true,
				},*/
		},
	}
}

func (d *vpnSiteConnectionDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	tflog.Info(ctx, "Using API token for authentication")
	//response.Diagnostics.AddAttributeWarning("DDDDDDDd")

	var data vpnSiteConnectionDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	conn := d.Meta().EC2Client(ctx)
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig(ctx)
	output, err := findVPNConnectionByID(ctx, conn, data.ID.ValueString())

	/*tunnelInfo, err := customerGatewayConfigurationToTunnelInfo(
		aws.ToString(output.CustomerGatewayConfiguration),
		d.Get("tunnel1_preshared_key").(string), // Not currently available during import
		d.Get("tunnel1_inside_cidr").(string),
		d.Get("tunnel1_inside_ipv6_cidr").(string),
	)*/

	//data.ID = types.StringValue("vpn-0698886b93fc97483")
	data.ARN = types.StringValue("test-id")

	//c := meta.(*conns.AWSClient)
	//conn := c.EC2Client(ctx)
	log.Printf("READ: %s", data.ID.ValueString()) // test-id

	if err != nil {
		response.Diagnostics.AddError("reading VPN Connection", tfresource.SingularDataSourceFindError("Security Group Rule", err).Error())

		return
	}

	log.Printf("===output.VpnGatewayId: %s", *output.VpnGatewayId)
	log.Printf("===================")
	log.Printf("===output: %#v", output)
	/*
		&types.VpnConnection{Category:(*string)(0xc004433a40),
		                     CoreNetworkArn:(*string)(nil),
												 CoreNetworkAttachmentArn:(*string)(nil),
												 CustomerGatewayConfiguration:(*string)(0xc004433900), # ⭐️このなかに色々入ってる
												 CustomerGatewayId:(*string)(0xc004433910),
												 GatewayAssociationState:"associated",
												 Options:(*types.VpnConnectionOptions)(0xc001c91ce0), # この中を展開する⭐️⭐️
												 PreSharedKeyArn:(*string)(nil),
												 Routes:[]types.VpnStaticRoute{},
												 State:"available",
												 Tags:[]types.Tag{types.Tag{Key:(*string)(0xc004433930), Value:(*string)(0xc004433940),
												 noSmithyDocumentSerde:document.NoSerde{}}},
												 TransitGatewayId:(*string)(nil),
												 Type:"ipsec.1",
												 VgwTelemetry:[]types.VgwTelemetry{types.VgwTelemetry{AcceptedRouteCount:(*int32)(0xc001e2ba28),
												 CertificateArn:(*string)(nil),
												 LastStatusChange:time.Date(2025, time.October, 4, 11, 57, 6, 0, time.UTC),
												 OutsideIpAddress:(*string)(0xc004433950),
												 Status:"DOWN",
												 StatusMessage:(*string)(0xc004433960),
												 noSmithyDocumentSerde:document.NoSerde{}},
												 types.VgwTelemetry{AcceptedRouteCount:(*int32)(0xc001e2bc80),
												 CertificateArn:(*string)(nil),
												 LastStatusChange:time.Date(2025, time.October, 4, 11, 46, 33, 0, time.UTC),
												 OutsideIpAddress:(*string)(0xc004433970),
												 Status:"DOWN",
												 StatusMessage:(*string)(0xc004433980),
												 noSmithyDocumentSerde:document.NoSerde{}}},
												 VpnConnectionId:(*string)(0xc0044338f0),
												 VpnGatewayId:(*string)(0xc004433920),
												 noSmithyDocumentSerde:document.NoSerde{}}
	*/
	log.Printf("===================")
	log.Printf("===CustomerGatewayId: %#v", *output.CustomerGatewayId)
	// CustomerGatewayId
	// log.Printf("output.CoreNetworkArn: %s", *output.CoreNetworkArn)
	data.CoreNetworkARN = fwflex.StringToFramework(ctx, output.CoreNetworkArn)                             // nil も
	data.CoreNetworkAttachmentArn = fwflex.StringToFramework(ctx, output.CoreNetworkAttachmentArn)         // nil も
	data.CustomerGatewayConfiguration = fwflex.StringToFramework(ctx, output.CustomerGatewayConfiguration) // nil も
	data.CustomerGatewayId = fwflex.StringToFramework(ctx, output.CustomerGatewayId)
	data.PreSharedKeyArn = fwflex.StringToFramework(ctx, output.PreSharedKeyArn)
	data.Tags = tftags.FlattenStringValueMap(ctx, keyValueTags(ctx, output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())
	data.TransitGatewayId = fwflex.StringToFramework(ctx, output.TransitGatewayId)

	if v := output.Options; v != nil {
		data.EnableAcceleration = fwflex.BoolToFramework(ctx, output.Options.EnableAcceleration) // nil も
		data.LocalIpv4NetworkCidr = fwflex.StringToFramework(ctx, output.Options.LocalIpv4NetworkCidr)
		data.LocalIpv6NetworkCidr = fwflex.StringToFramework(ctx, output.Options.LocalIpv6NetworkCidr)
		data.OutsideIpAddressType = fwflex.StringToFramework(ctx, output.Options.OutsideIpAddressType)
		data.RemoteIpv4NetworkCidr = fwflex.StringToFramework(ctx, output.Options.RemoteIpv4NetworkCidr)
		data.RemoteIpv6NetworkCidr = fwflex.StringToFramework(ctx, output.Options.RemoteIpv6NetworkCidr)
		data.StaticRoutesOnly = fwflex.BoolToFramework(ctx, output.Options.StaticRoutesOnly)
		data.TransportTransitGatewayAttachmentId = fwflex.StringToFramework(ctx, output.Options.TransportTransitGatewayAttachmentId)
		data.TunnelInsideIpVersion = types.StringValue(string(output.Options.TunnelInsideIpVersion))

		var vpnConfig xmlVpnConnectionConfig
		xmlConfig := aws.ToString(output.CustomerGatewayConfiguration)
		xml.Unmarshal([]byte(xmlConfig), &vpnConfig)

		data.Tunnel1Address = types.StringValue(vpnConfig.Tunnels[0].OutsideAddress)
		data.Tunnel1BGPASN = types.StringValue(vpnConfig.Tunnels[0].BGPASN)
		data.Tunnel1BGPHoldTime = types.Int64Value(int64(vpnConfig.Tunnels[0].BGPHoldTime))
		data.Tunnel1CgwInsideAddress = types.StringValue(vpnConfig.Tunnels[0].CgwInsideAddress)
		data.Tunnel1VgwInsideAddress = types.StringValue(vpnConfig.Tunnels[0].VgwInsideAddress)

		data.Tunnel2Address = types.StringValue(vpnConfig.Tunnels[1].OutsideAddress)
		data.Tunnel2BGPASN = types.StringValue(vpnConfig.Tunnels[1].BGPASN)
		data.Tunnel2BGPHoldTime = types.Int64Value(int64(vpnConfig.Tunnels[1].BGPHoldTime))
		data.Tunnel2CgwInsideAddress = types.StringValue(vpnConfig.Tunnels[1].CgwInsideAddress)
		data.Tunnel2VgwInsideAddress = types.StringValue(vpnConfig.Tunnels[1].VgwInsideAddress)

		data.Tunnel1DpdTimeoutAction = types.StringValue(*output.Options.TunnelOptions[0].DpdTimeoutAction)
		data.Tunnel1DpdTimeoutSeconds = types.Int32Value(int32(*output.Options.TunnelOptions[0].DpdTimeoutSeconds))
		data.Tunnel1EnableLifecycleControl = fwflex.BoolToFramework(ctx, output.Options.TunnelOptions[0].EnableTunnelLifecycleControl)
		data.Tunnel1IkeVersions = fwflex.FlattenFrameworkStringValueListOfString(ctx, []string{*output.Options.TunnelOptions[0].IkeVersions[0].Value}) // ここあってる？
		data.Tunnel1InsideCidr = fwflex.StringToFramework(ctx, output.Options.TunnelOptions[0].TunnelInsideCidr)
		data.Tunnel1InsideIpv6Cidr = fwflex.StringToFramework(ctx, output.Options.TunnelOptions[0].TunnelInsideIpv6Cidr)
		data.Kure = types.StringValue("kure")
		data.Kure2 = fwflex.FlattenFrameworkStringValueListOfString(ctx, []string{"a", "b", "c"})

		// tunnel1_log_options ⭐️次ここから。
		// tunnel1_phase1_dh_group_numbers
		// tunnel1_phase1_encryption_algorithms
		// tunnel1_phase1_integrity_algorithms
		data.Tunnel1Phase1LifetimeSeconds = types.Int32Value(int32(*output.Options.TunnelOptions[0].Phase1LifetimeSeconds))
	} else {

	}

	if v := output.TransitGatewayId; v != nil {
		input := ec2.DescribeTransitGatewayAttachmentsInput{
			Filters: newAttributeFilterList(map[string]string{
				"resource-id":        data.ID.ValueString(),
				"resource-type":      string(awstypes.TransitGatewayAttachmentResourceTypeVpn),
				"transit-gateway-id": aws.ToString(v),
			}),
		}

		transitGatewayAttachmentOutput, _ := findTransitGatewayAttachment(ctx, conn, &input)
		data.TransitGatewayAttachmentId = fwflex.StringToFramework(ctx, transitGatewayAttachmentOutput.TransitGatewayAttachmentId)
	} else {
		data.TransitGatewayAttachmentId = basetypes.NewStringNull()
	}

	// data.StaticRoutesOnly = fwflex.StringToFramework(ctx, output.StaticRoutesOnly)
	//data.Routes = fwflex.SetTo(ctx, output.Options.Routes)
	//data.PreSharedKeyStorage = fwflex.StringToFramework(ctx, output.Options.PreSharedKeyStorage)

	// data.CoreNetworkARN = types.StringValue("test-id")
	//data.ID = fwflex.StringToFramework(ctx, output.VpnGatewayId)
	// data.ARN =
	// data.ARN = d.( ctx, data.ID.ValueString())

	// log.Printf("READ: %s", output.VpnGatewayId)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type vpnSiteConnectionDataSourceModel struct {
	framework.WithRegionModel
	ARN                                 types.String         `tfsdk:"arn"`
	ID                                  types.String         `tfsdk:"id"`
	CoreNetworkARN                      types.String         `tfsdk:"core_network_arn"`
	CoreNetworkAttachmentArn            types.String         `tfsdk:"core_network_attachment_arn"`
	CustomerGatewayConfiguration        types.String         `tfsdk:"customer_gateway_configuration"`
	CustomerGatewayId                   types.String         `tfsdk:"customer_gateway_id"`
	EnableAcceleration                  types.Bool           `tfsdk:"enable_acceleration"`
	LocalIpv4NetworkCidr                types.String         `tfsdk:"local_ipv4_network_cidr"`
	LocalIpv6NetworkCidr                types.String         `tfsdk:"local_ipv6_network_cidr"`
	OutsideIpAddressType                types.String         `tfsdk:"outside_ip_address_type"`
	PreSharedKeyArn                     types.String         `tfsdk:"preshared_key_arn"`
	RemoteIpv4NetworkCidr               types.String         `tfsdk:"remote_ipv4_network_cidr"`
	RemoteIpv6NetworkCidr               types.String         `tfsdk:"remote_ipv6_network_cidr"`
	StaticRoutesOnly                    types.Bool           `tfsdk:"static_routes_only"`
	TransitGatewayId                    types.String         `tfsdk:"transit_gateway_id"`
	Tags                                tftags.Map           `tfsdk:"tags"`
	TransitGatewayAttachmentId          types.String         `tfsdk:"transit_gateway_attachment_id"`
	TransportTransitGatewayAttachmentId types.String         `tfsdk:"transport_transit_gateway_attachment_id"`
	Tunnel1Address                      types.String         `tfsdk:"tunnel1_address"`
	Tunnel1BGPASN                       types.String         `tfsdk:"tunnel1_bgp_asn"`
	Tunnel1BGPHoldTime                  types.Int64          `tfsdk:"tunnel1_bgp_holdtime"`
	TunnelInsideIpVersion               types.String         `tfsdk:"tunnel1_inside_ip_version"`
	Tunnel1CgwInsideAddress             types.String         `tfsdk:"tunnel1_cgw_inside_address"`
	Tunnel1VgwInsideAddress             types.String         `tfsdk:"tunnel1_vgw_inside_address"`
	Tunnel2Address                      types.String         `tfsdk:"tunnel2_address"`
	Tunnel2BGPASN                       types.String         `tfsdk:"tunnel2_bgp_asn"`
	Tunnel2BGPHoldTime                  types.Int64          `tfsdk:"tunnel2_bgp_holdtime"`
	Tunnel2nsideIpVersion               types.String         `tfsdk:"tunnel2_inside_ip_version"`
	Tunnel2CgwInsideAddress             types.String         `tfsdk:"tunnel2_cgw_inside_address"`
	Tunnel2VgwInsideAddress             types.String         `tfsdk:"tunnel2_vgw_inside_address"`
	Tunnel1DpdTimeoutAction             types.String         `tfsdk:"tunnel1_dpd_timeout_action"`
	Tunnel1DpdTimeoutSeconds            types.Int32          `tfsdk:"tunnel1_dpd_timeout_seconds"`
	Tunnel1EnableLifecycleControl       types.Bool           `tfsdk:"tunnel1_enable_tunnel_lifecycle_control"`
	Tunnel1IkeVersions                  fwtypes.ListOfString `tfsdk:"tunnel1_ike_versions"`
	Tunnel1InsideCidr                   types.String         `tfsdk:"tunnel1_inside_cidr"`
	Tunnel1InsideIpv6Cidr               types.String         `tfsdk:"tunnel1_inside_ipv6_cidr"`
	Tunnel1Phase1LifetimeSeconds        types.Int32          `tfsdk:"tunnel1_phase1_lifetime_seconds"`

	Kure  types.String         `tfsdk:"kure"`
	Kure2 fwtypes.ListOfString `tfsdk:"kure2"`
	//Routes                       types.Set    `tfsdk:"routes"`
}
