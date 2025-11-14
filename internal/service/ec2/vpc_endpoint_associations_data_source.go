// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_vpc_endpoint_associations", name="VPC Endpoint Associations")
func newVPCEndpointAssociationsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &vpcEndpointAssociationsDataSource{}, nil
}

type vpcEndpointAssociationsDataSource struct {
	framework.DataSourceWithModel[vpcEndpointAssociationsDataSourceModel]
}

func (d *vpcEndpointAssociationsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"associations": framework.DataSourceComputedListOfObjectAttribute[vpcEndpointAssociationModel](ctx),
			names.AttrVPCEndpointID: schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *vpcEndpointAssociationsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data vpcEndpointAssociationsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	input := ec2.DescribeVpcEndpointAssociationsInput{
		VpcEndpointIds: fwflex.StringSliceValueFromFramework(ctx, data.VPCEndpointID),
	}
	output, err := findVPCEndpointAssociations(ctx, conn, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPC Endpoint Associations (%s)", data.VPCEndpointID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data.Associations)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type vpcEndpointAssociationsDataSourceModel struct {
	framework.WithRegionModel
	Associations  fwtypes.ListNestedObjectValueOf[vpcEndpointAssociationModel] `tfsdk:"associations"`
	VPCEndpointID types.String                                                 `tfsdk:"vpc_endpoint_id"`
}

type vpcEndpointAssociationModel struct {
	AssociatedResourceAccessibility types.String                                   `tfsdk:"associated_resource_accessibility"`
	AssociatedResourceARN           types.String                                   `tfsdk:"associated_resource_arn"`
	DNSEntry                        fwtypes.ListNestedObjectValueOf[dnsEntryModel] `tfsdk:"dns_entry"`
	ID                              types.String                                   `tfsdk:"id"`
	PrivateDNSEntry                 fwtypes.ListNestedObjectValueOf[dnsEntryModel] `tfsdk:"private_dns_entry"`
	ResourceConfigurationGroupARN   fwtypes.ARN                                    `tfsdk:"resource_configuration_group_arn"`
	ServiceNetworkARN               fwtypes.ARN                                    `tfsdk:"service_network_arn"`
	ServiceNetworkName              types.String                                   `tfsdk:"service_network_name"`
	Tags                            tftags.Map                                     `tfsdk:"tags"`
}

type dnsEntryModel struct {
	DnsName      types.String `tfsdk:"dns_name"`
	HostedZoneID types.String `tfsdk:"hosted_zone_id"`
}
