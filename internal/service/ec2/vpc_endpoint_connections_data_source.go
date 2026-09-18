// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_vpc_endpoint_connections", name="VPC Endpoint Connections")
func newVPCEndpointConnectionsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &vpcEndpointConnectionsDataSource{}, nil
}

type vpcEndpointConnectionsDataSource struct {
	framework.DataSourceWithModel[vpcEndpointConnectionsDataSourceModel]
}

func (d *vpcEndpointConnectionsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"connections": framework.DataSourceComputedListOfObjectAttribute[vpcEndpointConnectionModel](ctx),
			"vpc_endpoint_service_id": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: customFiltersBlock(ctx),
		},
	}
}

func (d *vpcEndpointConnectionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data vpcEndpointConnectionsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	filters := newAttributeFilterList(map[string]string{
		"service-id": data.VPCEndpointServiceID.ValueString(),
	})
	filters = append(filters, newCustomFilterListFramework(ctx, data.Filters)...)

	input := ec2.DescribeVpcEndpointConnectionsInput{
		Filters: filters,
	}

	output, err := findVPCEndpointConnections(ctx, conn, &input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPC Endpoint Connections for service (%s)", data.VPCEndpointServiceID.ValueString()), err.Error())
		return
	}

	connections, diags := flattenVPCEndpointConnectionModels(ctx, output)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	data.Connections = connections
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// nosemgrep:ci.semgrep.framework.manual-flattener-functions
func flattenVPCEndpointConnectionModels(ctx context.Context, conns []awstypes.VpcEndpointConnection) (fwtypes.ListNestedObjectValueOf[vpcEndpointConnectionModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	models := make([]vpcEndpointConnectionModel, 0, len(conns))
	for _, conn := range conns {
		model, d := flattenVPCEndpointConnectionModel(ctx, conn)
		diags.Append(d...)
		if diags.HasError() {
			return fwtypes.NewListNestedObjectValueOfNull[vpcEndpointConnectionModel](ctx), diags
		}
		models = append(models, model)
	}

	result, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, models)
	diags.Append(d...)
	return result, diags
}

// nosemgrep:ci.semgrep.framework.manual-flattener-functions
func flattenVPCEndpointConnectionModel(ctx context.Context, conn awstypes.VpcEndpointConnection) (vpcEndpointConnectionModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var model vpcEndpointConnectionModel

	model.VPCEndpointID = types.StringPointerValue(conn.VpcEndpointId)
	model.VPCEndpointOwner = types.StringPointerValue(conn.VpcEndpointOwner)
	model.VPCEndpointState = types.StringValue(string(conn.VpcEndpointState))
	model.IPAddressType = types.StringValue(string(conn.IpAddressType))

	if conn.CreationTimestamp != nil {
		model.CreationTimestamp = timetypes.NewRFC3339TimeValue(*conn.CreationTimestamp)
	} else {
		model.CreationTimestamp = timetypes.NewRFC3339Null()
	}

	model.NetworkLoadBalancerARNs = fwflex.FlattenFrameworkStringValueListOfString(ctx, conn.NetworkLoadBalancerArns)
	model.GatewayLoadBalancerARNs = fwflex.FlattenFrameworkStringValueListOfString(ctx, conn.GatewayLoadBalancerArns)

	dnsEntries := make([]vpcEndpointConnectionDNSEntry, len(conn.DnsEntries))
	for i, entry := range conn.DnsEntries {
		dnsEntries[i] = vpcEndpointConnectionDNSEntry{
			DNSName:      types.StringPointerValue(entry.DnsName),
			HostedZoneID: types.StringPointerValue(entry.HostedZoneId),
		}
	}
	dnsEntriesList, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, dnsEntries)
	diags.Append(d...)
	if diags.HasError() {
		return model, diags
	}
	model.DNSEntries = dnsEntriesList

	kvTags := keyValueTags(ctx, conn.Tags).IgnoreAWS()
	m := kvTags.Map()
	elements := make(map[string]attr.Value, len(m))
	for k, v := range m {
		elements[k] = types.StringValue(v)
	}
	tagMap, d := tftags.NewMapValue(elements)
	diags.Append(d...)
	model.Tags = tagMap

	return model, diags
}

type vpcEndpointConnectionsDataSourceModel struct {
	framework.WithRegionModel
	Connections          fwtypes.ListNestedObjectValueOf[vpcEndpointConnectionModel] `tfsdk:"connections"`
	Filters              customFilters                                               `tfsdk:"filter"`
	VPCEndpointServiceID types.String                                                `tfsdk:"vpc_endpoint_service_id"`
}

type vpcEndpointConnectionModel struct {
	CreationTimestamp       timetypes.RFC3339                                              `tfsdk:"creation_timestamp"`
	DNSEntries              fwtypes.ListNestedObjectValueOf[vpcEndpointConnectionDNSEntry] `tfsdk:"dns_entries"`
	GatewayLoadBalancerARNs fwtypes.ListOfString                                           `tfsdk:"gateway_load_balancer_arns"`
	IPAddressType           types.String                                                   `tfsdk:"ip_address_type"`
	NetworkLoadBalancerARNs fwtypes.ListOfString                                           `tfsdk:"network_load_balancer_arns"`
	Tags                    tftags.Map                                                     `tfsdk:"tags"`
	VPCEndpointID           types.String                                                   `tfsdk:"vpc_endpoint_id"`
	VPCEndpointOwner        types.String                                                   `tfsdk:"vpc_endpoint_owner"`
	VPCEndpointState        types.String                                                   `tfsdk:"vpc_endpoint_state"`
}

type vpcEndpointConnectionDNSEntry struct {
	DNSName      types.String `tfsdk:"dns_name"`
	HostedZoneID types.String `tfsdk:"hosted_zone_id"`
}
