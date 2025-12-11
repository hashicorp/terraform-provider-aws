// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"

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
// @FrameworkDataSource("aws_odb_networks", name="Networks")
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
				Description: "List of odb networks returns basic information about odb networks.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[odbNetworkSummary](ctx),
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
	out, err := ListOracleDBNetworks(ctx, conn)
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

func ListOracleDBNetworks(ctx context.Context, conn *odb.Client) (*odb.ListOdbNetworksOutput, error) {
	var out odb.ListOdbNetworksOutput
	paginator := odb.NewListOdbNetworksPaginator(conn, &odb.ListOdbNetworksInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		out.OdbNetworks = append(out.OdbNetworks, output.OdbNetworks...)
	}
	return &out, nil
}

type odbNetworksListModel struct {
	framework.WithRegionModel
	OdbNetworks fwtypes.ListNestedObjectValueOf[odbNetworkSummary] `tfsdk:"odb_networks"`
}

type odbNetworkSummary struct {
	OdbNetworkId       types.String `tfsdk:"id"`
	OdbNetworkArn      types.String `tfsdk:"arn"`
	OciNetworkAnchorId types.String `tfsdk:"oci_network_anchor_id"`
	OciVcnUrl          types.String `tfsdk:"oci_vcn_url"`
	OciVcnId           types.String `tfsdk:"oci_vcn_id"`
	DisplayName        types.String `tfsdk:"display_name"`
}
