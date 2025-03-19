// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	types2 "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_vpc_endpoint_associations", name="VPC Endpoint Associations")
func newDataSourceVPCEndpointAssociations(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceVPCEndpointAssociations{}, nil
}

const (
	DSNameVPCEndpointAssociations = "VPC Endpoint Associations Data Source"
)

type dataSourceVPCEndpointAssociations struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceVPCEndpointAssociations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"associations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataSourceVPCEndpointAssociationModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrID: framework.ARNAttributeComputedOnly(),
						"resource_accessibility": schema.StringAttribute{
							Computed: true,
						},
						"resource_arn": schema.StringAttribute{
							Computed: true,
						},
						"resource_configuration_group_arn": schema.StringAttribute{
							Computed: true,
						},
						"service_network_arn": schema.StringAttribute{
							Computed: true,
						},
						"service_network_name": schema.StringAttribute{
							Computed: true,
						},
						names.AttrTags: tftags.TagsAttributeComputedOnly(),
					},
					Blocks: map[string]schema.Block{
						"dns_entry":         schemaDNSEntry(ctx),
						"private_dns_entry": schemaDNSEntry(ctx),
					},
				},
			},
		},
	}
}

func schemaDNSEntry(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[dataSourceVPCEndpointAssociationDNSEntryModel](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrDNSName: schema.StringAttribute{
					Computed: true,
				},
				names.AttrHostedZoneID: schema.StringAttribute{
					Computed: true,
				},
			},
		},
	}
}

func (d *dataSourceVPCEndpointAssociations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().EC2Client(ctx)

	var data dataSourceVPCEndpointAssociationsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findVPCEndpointAssociationsByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, DSNameVPCEndpointAssociations, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data.Associations)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findVPCEndpointAssociationsByID(ctx context.Context, conn *ec2.Client, id string) ([]types2.VpcEndpointAssociation, error) {

	input := ec2.DescribeVpcEndpointAssociationsInput{
		VpcEndpointIds: []string{id},
	}

	output, err := conn.DescribeVpcEndpointAssociations(ctx, &input)

	if err != nil {
		return nil, err
	}

	return output.VpcEndpointAssociations, nil
}

type dataSourceVPCEndpointAssociationsModel struct {
	Associations fwtypes.ListNestedObjectValueOf[dataSourceVPCEndpointAssociationModel] `tfsdk:"associations"`
	ID           types.String                                                           `tfsdk:"id"`
}

type dataSourceVPCEndpointAssociationModel struct {
	AssociatedResourceAccessibility types.String                                                                   `tfsdk:"resource_accessibility"`
	AssociatedResourceArn           types.String                                                                   `tfsdk:"resource_arn"`
	DnsEntry                        fwtypes.ListNestedObjectValueOf[dataSourceVPCEndpointAssociationDNSEntryModel] `tfsdk:"dns_entry"`
	ID                              types.String                                                                   `tfsdk:"id"`
	PrivateDnsEntry                 fwtypes.ListNestedObjectValueOf[dataSourceVPCEndpointAssociationDNSEntryModel] `tfsdk:"private_dns_entry"`
	ResourceConfigurationGroupArn   fwtypes.ARN                                                                    `tfsdk:"resource_configuration_group_arn"`
	ServiceNetworkArn               fwtypes.ARN                                                                    `tfsdk:"service_network_arn"`
	ServiceNetworkName              types.String                                                                   `tfsdk:"service_network_name"`
	Tags                            tftags.Map                                                                     `tfsdk:"tags"`
}

type dataSourceVPCEndpointAssociationDNSEntryModel struct {
	DnsName      types.String `tfsdk:"dns_name"`
	HostedZoneId types.String `tfsdk:"hosted_zone_id"`
}
