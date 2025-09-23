// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"github.com/hashicorp/terraform-provider-aws/internal/create"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_odb_network_peering_connections_list", name="Network Peering Connections List")
func newDataSourceNetworkPeeringConnectionsList(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceNetworkPeeringConnectionsList{}, nil
}

const (
	DSNameNetworkPeeringConnectionsList = "Network Peering Connections List Data Source"
)

type dataSourceNetworkPeeringConnectionsList struct {
	framework.DataSourceWithModel[odbNetworkPeeringConnectionsListDataSourceModel]
}

func (d *dataSourceNetworkPeeringConnectionsList) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{},
		Blocks: map[string]schema.Block{
			"odb_peering_connections": schema.ListNestedBlock{
				Description: "The list of ODB peering connections. A summary of an ODB peering connection.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[odbNetworkPeeringConnectionSummaryDataSourceModel](ctx),
			},
		},
	}
}

func (d *dataSourceNetworkPeeringConnectionsList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)
	var data odbNetworkPeeringConnectionsListDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := ListOdbPeeringConnections(ctx, conn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameNetworkPeeringConnectionsList, "", err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func ListOdbPeeringConnections(ctx context.Context, conn *odb.Client) (*odb.ListOdbPeeringConnectionsOutput, error) {
	var out odb.ListOdbPeeringConnectionsOutput
	listOfPeering, err := conn.ListOdbPeeringConnections(ctx, &odb.ListOdbPeeringConnectionsInput{})
	if err != nil {
		return nil, err
	}
	if listOfPeering != nil {
		out.OdbPeeringConnections = append(out.OdbPeeringConnections, listOfPeering.OdbPeeringConnections...)
	}
	for listOfPeering != nil && listOfPeering.NextToken != nil {
		listOfPeering, err = conn.ListOdbPeeringConnections(ctx, &odb.ListOdbPeeringConnectionsInput{
			NextToken: listOfPeering.NextToken,
		})
		if err != nil {
			return nil, err
		}
		if listOfPeering != nil {
			out.OdbPeeringConnections = append(out.OdbPeeringConnections, listOfPeering.OdbPeeringConnections...)
		}
	}
	return &out, nil
}

type odbNetworkPeeringConnectionsListDataSourceModel struct {
	framework.WithRegionModel
	OdbPeeringConnections fwtypes.ListNestedObjectValueOf[odbNetworkPeeringConnectionSummaryDataSourceModel] `tfsdk:"odb_peering_connections"`
}
type odbNetworkPeeringConnectionSummaryDataSourceModel struct {
	OdbPeeringConnectionId   types.String `tfsdk:"id"`
	OdbPeeringConnectionArn  types.String `tfsdk:"arn"`
	DisplayName              types.String `tfsdk:"display_name"`
	OdbNetworkArn            types.String `tfsdk:"odb_network_arn"`
	OdbPeeringConnectionType types.String `tfsdk:"odb_peering_connection_type"`
}
