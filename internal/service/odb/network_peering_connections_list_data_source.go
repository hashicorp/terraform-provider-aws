//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
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
				Description: " The list of ODB peering connections. A summary of an ODB peering connection.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[odbNetworkPeeringConnectionSummaryDataSourceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrID: schema.StringAttribute{
							Computed:    true,
							Description: "ID of the network peering connection.",
						},

						"display_name": schema.StringAttribute{
							Computed:    true,
							Description: "Display name for the network peering connection.",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							CustomType:  fwtypes.StringEnumType[odbtypes.ResourceStatus](),
							Description: "Status of this network peering connection.",
						},
						"status_reason": schema.StringAttribute{
							Computed:    true,
							Description: "Status reason of this network peering connection.",
						},
						names.AttrARN: framework.ARNAttributeComputedOnly(),
						"odb_network_arn": schema.StringAttribute{
							Computed:    true,
							Description: "The ARN of the ODB network peering connection.",
						},
						"odb_peering_connection_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the ODB peering connection.",
						},
						"percent_progress": schema.Float32Attribute{
							Computed:    true,
							Description: "The percentage of progress made in network peering .",
						},
					},
				},
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

	out, err := conn.ListOdbPeeringConnections(ctx, &odb.ListOdbPeeringConnectionsInput{})
	listOfPeerConnections := make([]peeringConnectionSummaryRead, 0)
	if err == nil && out.OdbPeeringConnections != nil {
		for _, peerConn := range out.OdbPeeringConnections {
			peerConnSummary := peeringConnectionSummaryRead{
				OdbPeeringConnectionId:   peerConn.OdbPeeringConnectionId,
				DisplayName:              peerConn.DisplayName,
				OdbNetworkArn:            peerConn.OdbNetworkArn,
				OdbPeeringConnectionArn:  peerConn.OdbPeeringConnectionArn,
				OdbPeeringConnectionType: peerConn.OdbPeeringConnectionType,
				PeerNetworkArn:           peerConn.PeerNetworkArn,
				PercentProgress:          peerConn.PercentProgress,
				Status:                   peerConn.Status,
				StatusReason:             peerConn.StatusReason,
			}
			listOfPeerConnections = append(listOfPeerConnections, peerConnSummary)
		}
		resp.Diagnostics.Append(flex.Flatten(ctx, listOdbPeeringConnectionsOutputRead{
			OdbPeeringConnections: listOfPeerConnections,
		}, &data)...)
		if resp.Diagnostics.HasError() {
			return

		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type odbNetworkPeeringConnectionsListDataSourceModel struct {
	framework.WithRegionModel
	OdbPeeringConnections fwtypes.ListNestedObjectValueOf[odbNetworkPeeringConnectionSummaryDataSourceModel] `tfsdk:"odb_peering_connections"`
}
type odbNetworkPeeringConnectionSummaryDataSourceModel struct {
	OdbPeeringConnectionId   types.String                                `tfsdk:"id"`
	DisplayName              types.String                                `tfsdk:"display_name"`
	Status                   fwtypes.StringEnum[odbtypes.ResourceStatus] `tfsdk:"status"`
	StatusReason             types.String                                `tfsdk:"status_reason"`
	OdbPeeringConnectionArn  types.String                                `tfsdk:"arn"`
	OdbNetworkArn            types.String                                `tfsdk:"odb_network_arn"`
	OdbPeeringConnectionType types.String                                `tfsdk:"odb_peering_connection_type"`
	PercentProgress          types.Float32                               `tfsdk:"percent_progress"`
}
type peeringConnectionSummaryRead struct {
	OdbPeeringConnectionId   *string
	DisplayName              *string
	OdbNetworkArn            *string
	OdbPeeringConnectionArn  *string
	OdbPeeringConnectionType *string
	PeerNetworkArn           *string
	PercentProgress          *float32
	Status                   odbtypes.ResourceStatus
	StatusReason             *string
}
type listOdbPeeringConnectionsOutputRead struct {
	OdbPeeringConnections []peeringConnectionSummaryRead
}
