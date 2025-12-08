// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
// @FrameworkDataSource("aws_odb_network_peering_connection", name="Network Peering Connection")
func newDataSourceNetworkPeeringConnection(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceNetworkPeeringConnection{}, nil
}

const (
	DSNameNetworkPeeringConnection = "Network Peering Connection Data Source"
)

type dataSourceNetworkPeeringConnection struct {
	framework.DataSourceWithModel[odbNetworkPeeringConnectionDataSourceModel]
}

func (d *dataSourceNetworkPeeringConnection) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Description: "Network Peering Connection identifier.",
				Required:    true,
			},
			names.AttrDisplayName: schema.StringAttribute{
				Description: "Display name of the odb network peering connection.",
				Computed:    true,
			},
			names.AttrStatus: schema.StringAttribute{
				Description: "Status of the odb network peering connection.",
				CustomType:  fwtypes.StringEnumType[odbtypes.ResourceStatus](),
				Computed:    true,
			},
			names.AttrStatusReason: schema.StringAttribute{
				Description: "Status of the odb network peering connection.",
				Computed:    true,
			},

			"odb_network_arn": schema.StringAttribute{
				Description: "ARN of the odb network peering connection.",
				Computed:    true,
			},

			names.AttrARN: framework.ARNAttributeComputedOnly(),

			"peer_network_arn": schema.StringAttribute{
				Description: "ARN of the peer network peering connection.",
				Computed:    true,
			},
			"odb_peering_connection_type": schema.StringAttribute{
				Description: "Type of the odb peering connection.",
				Computed:    true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				Description: "Created time of the odb network peering connection.",
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
			},
			"percent_progress": schema.Float32Attribute{
				Description: "Progress of the odb network peering connection.",
				Computed:    true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *dataSourceNetworkPeeringConnection) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)
	var data odbNetworkPeeringConnectionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.GetOdbPeeringConnectionInput{
		OdbPeeringConnectionId: data.OdbPeeringConnectionId.ValueStringPointer(),
	}
	out, err := conn.GetOdbPeeringConnection(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameNetworkPeeringConnection, data.OdbPeeringConnectionId.ValueString(), err),
			err.Error(),
		)
		return
	}
	tagsRead, err := listTags(ctx, conn, *out.OdbPeeringConnection.OdbPeeringConnectionArn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameNetworkPeeringConnection, data.OdbPeeringConnectionId.ValueString(), err),
			err.Error(),
		)
		return
	}
	if tagsRead != nil {
		data.Tags = tftags.FlattenStringValueMap(ctx, tagsRead.Map())
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out.OdbPeeringConnection, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type odbNetworkPeeringConnectionDataSourceModel struct {
	framework.WithRegionModel
	OdbPeeringConnectionId   types.String                                `tfsdk:"id"`
	DisplayName              types.String                                `tfsdk:"display_name"`
	Status                   fwtypes.StringEnum[odbtypes.ResourceStatus] `tfsdk:"status"`
	StatusReason             types.String                                `tfsdk:"status_reason"`
	OdbPeeringConnectionArn  types.String                                `tfsdk:"arn"`
	OdbNetworkArn            types.String                                `tfsdk:"odb_network_arn"`
	PeerNetworkArn           types.String                                `tfsdk:"peer_network_arn"`
	OdbPeeringConnectionType types.String                                `tfsdk:"odb_peering_connection_type"`
	CreatedAt                timetypes.RFC3339                           `tfsdk:"created_at"`
	PercentProgress          types.Float32                               `tfsdk:"percent_progress"`
	Tags                     tftags.Map                                  `tfsdk:"tags"`
}
