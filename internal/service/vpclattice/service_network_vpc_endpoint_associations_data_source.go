// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_vpclattice_service_network_vpc_endpoint_associations", name="Service Network VPC Endpoint Associations")
func newDataSourceServiceNetworkVPCEndpointAssociations(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceServiceNetworkVPCEndpointAssociations{}, nil
}

const (
	DSNameServiceNetworkVPCEndpointAssociations = "Service Network VPC Endpoint Associations Data Source"
)

type dataSourceServiceNetworkVPCEndpointAssociations struct {
	framework.DataSourceWithModel[dataSourceServiceNetworkVPCEndpointAssociationsModel]
}

func (d *dataSourceServiceNetworkVPCEndpointAssociations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_network_identifier": schema.StringAttribute{
				Description: "ID or ARN of the Service Network.",
				Required:    true,
			},
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"associations": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrID:            schema.StringAttribute{Computed: true},
						"service_network_arn":   schema.StringAttribute{Computed: true},
						names.AttrState:         schema.StringAttribute{Computed: true},
						names.AttrVPCEndpointID: schema.StringAttribute{Computed: true},
						"vpc_endpoint_owner_id": schema.StringAttribute{Computed: true},
						names.AttrVPCID:         schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *dataSourceServiceNetworkVPCEndpointAssociations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().VPCLatticeClient(ctx)

	var associations []*awstypes.ServiceNetworkEndpointAssociation

	var data dataSourceServiceNetworkVPCEndpointAssociationsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	snID := data.ServiceNetworkIdentifier.ValueString()
	data.ID = types.StringValue(snID)

	if _, err := findServiceNetworkByID(ctx, conn, snID); err != nil {
		resp.Diagnostics.AddError("Error reading Service Network", err.Error())
		return
	}

	associations, err := listServiceNetworkVPCEndpointAssociationsByServiceNetworkIdentifier(ctx, conn, snID)
	if err != nil {
		resp.Diagnostics.AddError("Error listing associations", err.Error())
		return
	}

	data.Associations = flattenServiceNetworkVPCEndpointAssociations(associations)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func flattenServiceNetworkVPCEndpointAssociations(objs []*awstypes.ServiceNetworkEndpointAssociation) []associationModel {
	result := make([]associationModel, 0, len(objs))
	for _, obj := range objs {
		result = append(result, flattenServiceNetworkServiceAssociation(obj))
	}
	return result
}

func flattenServiceNetworkServiceAssociation(obj *awstypes.ServiceNetworkEndpointAssociation) associationModel {
	return associationModel{
		ID:                 types.StringPointerValue(obj.Id),
		ServiceNetworkArn:  types.StringPointerValue(obj.ServiceNetworkArn),
		State:              types.StringPointerValue(obj.State),
		VpcEndpointId:      types.StringPointerValue(obj.VpcEndpointId),
		VpcEndpointOwnerId: types.StringPointerValue(obj.VpcEndpointOwnerId),
		VpcId:              types.StringPointerValue(obj.VpcId),
	}
}

func listServiceNetworkVPCEndpointAssociationsByServiceNetworkIdentifier(ctx context.Context, conn *vpclattice.Client, serviceNetworkIdentifier string) ([]*awstypes.ServiceNetworkEndpointAssociation, error) {
	input := vpclattice.ListServiceNetworkVpcEndpointAssociationsInput{
		ServiceNetworkIdentifier: aws.String(serviceNetworkIdentifier),
	}
	return listServiceNetworkVPCEndpointAssociations(ctx, conn, &input)
}

func listServiceNetworkVPCEndpointAssociations(ctx context.Context, conn *vpclattice.Client, input *vpclattice.ListServiceNetworkVpcEndpointAssociationsInput) ([]*awstypes.ServiceNetworkEndpointAssociation, error) {
	var output []*awstypes.ServiceNetworkEndpointAssociation
	paginator := vpclattice.NewListServiceNetworkVpcEndpointAssociationsPaginator(conn, input, func(opts *vpclattice.ListServiceNetworkVpcEndpointAssociationsPaginatorOptions) {
		opts.Limit = 100
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for idx := range page.Items {
			v := page.Items[idx]
			output = append(output, &v)
		}
	}
	return output, nil
}

type dataSourceServiceNetworkVPCEndpointAssociationsModel struct {
	framework.WithRegionModel
	ServiceNetworkIdentifier types.String       `tfsdk:"service_network_identifier"`
	Associations             []associationModel `tfsdk:"associations"`
	ID                       types.String       `tfsdk:"id"`
}

type associationModel struct {
	ID                 types.String `tfsdk:"id"`
	ServiceNetworkArn  types.String `tfsdk:"service_network_arn"`
	State              types.String `tfsdk:"state"`
	VpcEndpointId      types.String `tfsdk:"vpc_endpoint_id"`
	VpcEndpointOwnerId types.String `tfsdk:"vpc_endpoint_owner_id"`
	VpcId              types.String `tfsdk:"vpc_id"`
}
