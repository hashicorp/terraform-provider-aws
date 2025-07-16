//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_odb_networks_list", name="Networks List")
func newDataSourceNetworksList(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceNetworksList{}, nil
}

const (
	DSNameNetworksList = "Networks List Data Source"
)

type dataSourceNetworksList struct {
	framework.DataSourceWithModel[odbNetworksListModel]
}

func (d *dataSourceNetworksList) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"odb_networks": schema.ListAttribute{
				Computed:    true,
				Description: "List of odb networks (OCID, ID, ARN, OCI URL, Display Name)",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[cloudAutonomousVmClusterSummary](ctx),
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"arn":          types.StringType,
						"id":           types.StringType,
						"oci_url":      types.StringType,
						"ocid":         types.StringType,
						"display_name": types.StringType,
					},
				},
			},
		},
	}
}

// Data sources only have a read method.
func (d *dataSourceNetworksList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	conn := d.Meta().ODBClient(ctx)

	var data odbNetworksListModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.ListOdbNetworks(ctx, &odb.ListOdbNetworksInput{})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameNetworksList, "", err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type odbNetworksListModel struct {
	framework.WithRegionModel
	OdbNetworks fwtypes.ListNestedObjectValueOf[odbNetworkSummary] `tfsdk:"odb_networks"`
}

type odbNetworkSummary struct {
	CloudExadataInfrastructureArn types.String `tfsdk:"arn"`
	CloudAutonomousVmClusterId    types.String `tfsdk:"id"`
	OciUrl                        types.String `tfsdk:"oci_url"`
	Ocid                          types.String `tfsdk:"ocid"`
	DisplayName                   types.String `tfsdk:"display_name"`
}
