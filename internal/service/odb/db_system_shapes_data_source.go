// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
// @FrameworkDataSource("aws_odb_db_system_shapes", name="Db System Shapes")
func newDataSourceDBSystemShapes(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDBSystemShapesList{}, nil
}

const (
	DSNameDBSystemShapesList = "Db System Shapes List Data Source"
)

type dataSourceDBSystemShapesList struct {
	framework.DataSourceWithModel[dbSystemShapesListDataSourceModel]
}

func (d *dataSourceDBSystemShapesList) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"availability_zone_id": schema.StringAttribute{
				Optional:    true,
				Description: "The physical ID of the AZ, for example, use1-az4. This ID persists across accounts.",
			},
			"db_system_shapes": schema.ListAttribute{
				Computed:   true,
				CustomType: fwtypes.NewListNestedObjectTypeOf[dbSystemShapeDataSourceModel](ctx),
				Description: "The list of shapes and their properties. Information about a hardware system model (shape) that's available for an Exadata infrastructure." +
					"The shape determines resources, such as CPU cores, memory, and storage, to allocate to the Exadata infrastructure.",
			},
		},
	}
}

func (d *dataSourceDBSystemShapesList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)

	var data dbSystemShapesListDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := odb.ListDbSystemShapesInput{}
	if !data.AvailabilityZoneId.IsNull() && !data.AvailabilityZoneId.IsUnknown() {
		input.AvailabilityZoneId = data.AvailabilityZoneId.ValueStringPointer()
	}
	paginator := odb.NewListDbSystemShapesPaginator(conn, &input)
	var out odb.ListDbSystemShapesOutput
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameDBSystemShapesList, "", err),
				err.Error(),
			)
			return
		}

		if page != nil && len(page.DbSystemShapes) > 0 {
			out.DbSystemShapes = append(out.DbSystemShapes, page.DbSystemShapes...)
		}
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dbSystemShapesListDataSourceModel struct {
	framework.WithRegionModel
	AvailabilityZoneId types.String                                                  `tfsdk:"availability_zone_id"`
	DbSystemShapes     fwtypes.ListNestedObjectValueOf[dbSystemShapeDataSourceModel] `tfsdk:"db_system_shapes"`
}

type dbSystemShapeDataSourceModel struct {
	AvailableCoreCount                 types.Int32  `tfsdk:"available_core_count"`
	AvailableCoreCountPerNode          types.Int32  `tfsdk:"available_core_count_per_node"`
	AvailableDataStorageInTBs          types.Int32  `tfsdk:"available_data_storage_in_tbs"`
	AvailableDataStoragePerServerInTBs types.Int32  `tfsdk:"available_data_storage_per_server_in_tbs"`
	AvailableDbNodePerNodeInGBs        types.Int32  `tfsdk:"available_db_node_per_node_in_gbs"`
	AvailableDbNodeStorageInGBs        types.Int32  `tfsdk:"available_db_node_storage_in_gbs"`
	AvailableMemoryInGBs               types.Int32  `tfsdk:"available_memory_in_gbs"`
	AvailableMemoryPerNodeInGBs        types.Int32  `tfsdk:"available_memory_per_node_in_gbs"`
	CoreCountIncrement                 types.Int32  `tfsdk:"core_count_increment"`
	MaxStorageCount                    types.Int32  `tfsdk:"max_storage_count"`
	MaximumNodeCount                   types.Int32  `tfsdk:"maximum_node_count"`
	MinCoreCountPerNode                types.Int32  `tfsdk:"min_core_count_per_node"`
	MinDataStorageInTBs                types.Int32  `tfsdk:"min_data_storage_in_tbs"`
	MinDbNodeStoragePerNodeInGBs       types.Int32  `tfsdk:"min_db_node_storage_per_node_in_gbs"`
	MinMemoryPerNodeInGBs              types.Int32  `tfsdk:"min_memory_per_node_in_gbs"`
	MinStorageCount                    types.Int32  `tfsdk:"min_storage_count"`
	MinimumCoreCount                   types.Int32  `tfsdk:"minimum_core_count"`
	MinimumNodeCount                   types.Int32  `tfsdk:"minimum_node_count"`
	Name                               types.String `tfsdk:"name"`
	RuntimeMinimumCoreCount            types.Int32  `tfsdk:"runtime_minimum_core_count"`
	ShapeFamily                        types.String `tfsdk:"shape_family"`
	ShapeType                          types.String `tfsdk:"shape_type"`
}
