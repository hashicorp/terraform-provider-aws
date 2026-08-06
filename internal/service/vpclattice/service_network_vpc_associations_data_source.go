// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_vpclattice_service_network_vpc_associations", name="Service Network VPC Associations")
func newDataSourceServiceNetworkVPCAssociations(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceServiceNetworkVPCAssociations{}, nil
}

const (
	DSNameServiceNetworkVPCAssociations = "Service Network VPC Associations Data Source"
)

type dataSourceServiceNetworkVPCAssociations struct {
	framework.DataSourceWithModel[dataSourceServiceNetworkVPCAssociationsModel]
}

func (d *dataSourceServiceNetworkVPCAssociations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_network_identifier": schema.StringAttribute{
				Optional:    true,
				Description: "ID or ARN of the Service Network.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("vpc_identifier"),
					),
				},
			},
			"vpc_identifier": schema.StringAttribute{
				Optional:    true,
				Description: "ID the VPC.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("service_network_identifier"),
					),
				},
			},
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"associations": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrARN:          framework.ARNAttributeComputedOnly(),
						"created_by":           schema.StringAttribute{Computed: true},
						names.AttrID:           schema.StringAttribute{Computed: true},
						"service_network_arn":  schema.StringAttribute{Computed: true},
						"service_network_id":   schema.StringAttribute{Computed: true},
						"service_network_name": schema.StringAttribute{Computed: true},
						names.AttrStatus:       schema.StringAttribute{Computed: true},
						names.AttrVPCID:        schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *dataSourceServiceNetworkVPCAssociations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().VPCLatticeClient(ctx)
	ec2conn := d.Meta().EC2Client(ctx)

	var associations []*awstypes.ServiceNetworkVpcAssociationSummary

	var data dataSourceServiceNetworkVPCAssociationsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !data.ServiceNetworkIdentifier.IsNull() {
		snID := data.ServiceNetworkIdentifier.ValueString()
		data.ID = types.StringValue(snID)

		if _, err := findServiceNetworkByID(ctx, conn, snID); err != nil {
			resp.Diagnostics.AddError("Error reading Service Network", err.Error())
			return
		}

		items, err := listServiceNetworkVPCAssociationsByServiceNetworkIdentifier(ctx, conn, snID)
		if err != nil {
			resp.Diagnostics.AddError("Error listing associations", err.Error())
			return
		}
		associations = items
	} else if !data.VpcIdentifier.IsNull() {
		vID := data.VpcIdentifier.ValueString()
		data.ID = types.StringValue(vID)

		if _, err := ec2.FindVPCByID(ctx, ec2conn, vID); err != nil {
			resp.Diagnostics.AddError("Error finding VPC", err.Error())
			return
		}

		items, err := listServiceNetworkVPCAssociationsByVPCIdentifier(ctx, conn, vID)
		if err != nil {
			resp.Diagnostics.AddError("Error listing associations", err.Error())
			return
		}
		associations = items
	}

	data.Associations = flattenServiceNetworkVPCAssociations(associations)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func flattenServiceNetworkVPCAssociations(objs []*awstypes.ServiceNetworkVpcAssociationSummary) []associationModel {
	result := make([]associationModel, 0, len(objs))
	for _, obj := range objs {
		result = append(result, flattenServiceNetworkVPCAssociation(obj))
	}
	return result
}

func flattenServiceNetworkVPCAssociation(obj *awstypes.ServiceNetworkVpcAssociationSummary) associationModel {
	return associationModel{
		Arn:                types.StringPointerValue(obj.Arn),
		CreatedBy:          types.StringPointerValue(obj.CreatedBy),
		ID:                 types.StringPointerValue(obj.Id),
		ServiceNetworkArn:  types.StringPointerValue(obj.ServiceNetworkArn),
		ServiceNetworkId:   types.StringPointerValue(obj.ServiceNetworkId),
		ServiceNetworkName: types.StringPointerValue(obj.ServiceNetworkName),
		Status:             types.StringValue(string(obj.Status)),
		VpcId:              types.StringPointerValue(obj.VpcId),
	}
}

func listServiceNetworkVPCAssociationsByVPCIdentifier(ctx context.Context, conn *vpclattice.Client, vpcIdentifier string) ([]*awstypes.ServiceNetworkVpcAssociationSummary, error) {
	input := vpclattice.ListServiceNetworkVpcAssociationsInput{
		VpcIdentifier: aws.String(vpcIdentifier),
	}
	return listServiceNetworkVPCAssociations(ctx, conn, &input)
}

func listServiceNetworkVPCAssociationsByServiceNetworkIdentifier(ctx context.Context, conn *vpclattice.Client, serviceNetworkIdentifier string) ([]*awstypes.ServiceNetworkVpcAssociationSummary, error) {
	input := vpclattice.ListServiceNetworkVpcAssociationsInput{
		ServiceNetworkIdentifier: aws.String(serviceNetworkIdentifier),
	}
	return listServiceNetworkVPCAssociations(ctx, conn, &input)
}

func listServiceNetworkVPCAssociations(ctx context.Context, conn *vpclattice.Client, input *vpclattice.ListServiceNetworkVpcAssociationsInput) ([]*awstypes.ServiceNetworkVpcAssociationSummary, error) {
	var output []*awstypes.ServiceNetworkVpcAssociationSummary
	paginator := vpclattice.NewListServiceNetworkVpcAssociationsPaginator(conn, input, func(opts *vpclattice.ListServiceNetworkVpcAssociationsPaginatorOptions) {
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

type dataSourceServiceNetworkVPCAssociationsModel struct {
	framework.WithRegionModel
	ServiceNetworkIdentifier types.String       `tfsdk:"service_network_identifier"`
	VpcIdentifier            types.String       `tfsdk:"vpc_identifier"`
	Associations             []associationModel `tfsdk:"associations"`
	ID                       types.String       `tfsdk:"id"`
}

type associationModel struct {
	Arn                types.String `tfsdk:"arn"`
	CreatedBy          types.String `tfsdk:"created_by"`
	ID                 types.String `tfsdk:"id"`
	ServiceNetworkArn  types.String `tfsdk:"service_network_arn"`
	ServiceNetworkId   types.String `tfsdk:"service_network_id"`
	ServiceNetworkName types.String `tfsdk:"service_network_name"`
	Status             types.String `tfsdk:"status"`
	VpcId              types.String `tfsdk:"vpc_id"`
}
