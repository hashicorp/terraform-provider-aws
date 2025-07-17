// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"

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
// @FrameworkDataSource("aws_odb_db_server", name="Db Server")
func newDataSourceDbServer(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDbServer{}, nil
}

const (
	DSNameDbServer = "Db Server Data Source"
)

type dataSourceDbServer struct {
	framework.DataSourceWithModel[dbServerDataSourceModel]
}

func (d *dataSourceDbServer) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Description: "The identifier of the the database server.",
				Required:    true,
			},
			"cloud_exadata_infrastructure_id": schema.StringAttribute{
				Description: "The identifier of the database server to retrieve information about.",
				Required:    true,
			},
			"status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[odbtypes.ResourceStatus](),
				Computed:   true,
			},
			"status_reason": schema.StringAttribute{
				Computed: true,
			},
			"cpu_core_count": schema.Int32Attribute{
				Computed: true,
			},
			"db_node_ids": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
			},
			"db_node_storage_size_in_gbs": schema.Int32Attribute{
				Computed: true,
			},
			"db_server_patching_details": schema.ObjectAttribute{
				Computed:   true,
				CustomType: fwtypes.NewObjectTypeOf[dbNodePatchingDetailsDbServerDataSourceModel](ctx),
				AttributeTypes: map[string]attr.Type{
					"estimated_patch_duration": types.Int32Type,
					"patching_status":          types.StringType,
					"time_patching_ended":      types.StringType,
					"time_patching_started":    types.StringType,
				},
			},
			"display_name": schema.StringAttribute{
				Computed: true,
			},
			"exadata_infrastructure_id": schema.StringAttribute{
				Computed: true,
			},
			"ocid": schema.StringAttribute{
				Computed: true,
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed: true,
			},
			"max_cpu_count": schema.Int32Attribute{
				Computed: true,
			},
			"max_db_node_storage_in_gbs": schema.Int32Attribute{
				Computed: true,
			},
			"max_memory_in_gbs": schema.Int32Attribute{
				Computed: true,
			},
			"memory_size_in_gbs": schema.Int32Attribute{
				Computed: true,
			},
			"shape": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"vm_cluster_ids": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
			},
			"compute_model": schema.StringAttribute{
				Computed:   true,
				CustomType: fwtypes.StringEnumType[odbtypes.ComputeModel](),
			},
			"autonomous_vm_cluster_ids": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
			},
			"autonomous_virtual_machine_ids": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *dataSourceDbServer) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	conn := d.Meta().ODBClient(ctx)

	var data dbServerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.GetDbServerInput{
		DbServerId:                   data.DbServerID.ValueStringPointer(),
		CloudExadataInfrastructureId: data.CloudExadataInfrastructureID.ValueStringPointer(),
	}

	out, err := conn.GetDbServer(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameDbServer, data.DbServerID.ValueString(), err),
			err.Error(),
		)
		return
	}

	if out.DbServer.CreatedAt != nil {
		data.CreatedAt = types.StringValue(out.DbServer.CreatedAt.Format(time.RFC3339))
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.DbServer, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dbServerDataSourceModel struct {
	framework.WithRegionModel
	DbServerID                   types.String                                                        `tfsdk:"id"`
	CloudExadataInfrastructureID types.String                                                        `tfsdk:"cloud_exadata_infrastructure_id"`
	Status                       fwtypes.StringEnum[odbtypes.ResourceStatus]                         `tfsdk:"status"`
	StatusReason                 types.String                                                        `tfsdk:"status_reason"`
	CpuCoreCount                 types.Int32                                                         `tfsdk:"cpu_core_count"`
	DbNodeIds                    fwtypes.ListOfString                                                `tfsdk:"db_node_ids"`
	DbNodeStorageSizeInGBs       types.Int32                                                         `tfsdk:"db_node_storage_size_in_gbs"`
	DbServerPatchingDetails      fwtypes.ObjectValueOf[dbNodePatchingDetailsDbServerDataSourceModel] `tfsdk:"db_server_patching_details"`
	DisplayName                  types.String                                                        `tfsdk:"display_name"`
	ExadataInfrastructureId      types.String                                                        `tfsdk:"exadata_infrastructure_id"`
	OCID                         types.String                                                        `tfsdk:"ocid"`
	OciResourceAnchorName        types.String                                                        `tfsdk:"oci_resource_anchor_name"`
	MaxCpuCount                  types.Int32                                                         `tfsdk:"max_cpu_count"`
	MaxDbNodeStorageInGBs        types.Int32                                                         `tfsdk:"max_db_node_storage_in_gbs"`
	MaxMemoryInGBs               types.Int32                                                         `tfsdk:"max_memory_in_gbs"`
	MemorySizeInGBs              types.Int32                                                         `tfsdk:"memory_size_in_gbs"`
	Shape                        types.String                                                        `tfsdk:"shape"`
	CreatedAt                    types.String                                                        `tfsdk:"created_at" autoflex:",noflatten"`
	VmClusterIds                 fwtypes.ListOfString                                                `tfsdk:"vm_cluster_ids"`
	ComputeModel                 fwtypes.StringEnum[odbtypes.ComputeModel]                           `tfsdk:"compute_model"`
	AutonomousVmClusterIds       fwtypes.ListOfString                                                `tfsdk:"autonomous_vm_cluster_ids"`
	AutonomousVirtualMachineIds  fwtypes.ListOfString                                                `tfsdk:"autonomous_virtual_machine_ids"`
}

type dbNodePatchingDetailsDbServerDataSourceModel struct {
	EstimatedPatchDuration types.Int32  `tfsdk:"estimated_patch_duration"`
	PatchingStatus         types.String `tfsdk:"patching_status"`
	TimePatchingEnded      types.String `tfsdk:"time_patching_ended"`
	TimePatchingStarted    types.String `tfsdk:"time_patching_started"`
}
