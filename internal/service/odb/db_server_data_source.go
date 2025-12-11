// Copyright IBM Corp. 2014, 2025
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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_odb_db_server", name="Db Server")
func newDataSourceDBServer(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDbServer{}, nil
}

const (
	DSNameDBServer = "DB Server Data Source"
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
			"autonomous_virtual_machine_ids": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Description: "The list of unique identifiers for the Autonomous VMs associated with this database server.",
			},
			"autonomous_vm_cluster_ids": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Description: "The OCID of the autonomous VM clusters that are associated with the database server.",
			},
			"compute_model": schema.StringAttribute{
				Computed:    true,
				CustomType:  fwtypes.StringEnumType[odbtypes.ComputeModel](),
				Description: " The compute model of the database server.",
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[odbtypes.ResourceStatus](),
				Computed:    true,
				Description: "The status of the database server.",
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed:    true,
				Description: "Additional information about the current status of the database server.",
			},
			"cpu_core_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of CPU cores enabled on the database server.",
			},
			"db_node_storage_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The allocated local node storage in GBs on the database server.",
			},
			"db_server_patching_details": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dbNodePatchingDetailsDbServerDataSourceModel](ctx),
				Computed:   true,
				Description: "The scheduling details for the quarterly maintenance window. Patching and\n" +
					"system updates take place during the maintenance window.",
			},
			names.AttrDisplayName: schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the database server.",
			},
			"exadata_infrastructure_id": schema.StringAttribute{
				Computed:    true,
				Description: "The exadata infrastructure ID of the database server.",
			},
			"ocid": schema.StringAttribute{
				Computed:    true,
				Description: "The OCID of the database server to retrieve information about.",
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the OCI resource anchor.",
			},
			"max_cpu_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The total number of CPU cores available.",
			},
			"max_db_node_storage_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The total local node storage available in GBs.",
			},
			"max_memory_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The total memory available in GBs.",
			},
			"memory_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The allocated memory in GBs on the database server.",
			},
			"shape": schema.StringAttribute{
				Computed: true,
				Description: "The shape of the database server. The shape determines the amount of CPU, " +
					"storage, and memory resources available.",
			},
			names.AttrCreatedAt: schema.StringAttribute{
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Description: "The date and time when the database server was created.",
			},
			"vm_cluster_ids": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Description: "The OCID of the VM clusters that are associated with the database server.",
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
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameDBServer, data.DbServerID.ValueString(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out.DbServer, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dbServerDataSourceModel struct {
	framework.WithRegionModel
	DbServerID                   types.String                                                                  `tfsdk:"id"`
	CloudExadataInfrastructureID types.String                                                                  `tfsdk:"cloud_exadata_infrastructure_id"`
	Status                       fwtypes.StringEnum[odbtypes.ResourceStatus]                                   `tfsdk:"status"`
	StatusReason                 types.String                                                                  `tfsdk:"status_reason"`
	CpuCoreCount                 types.Int32                                                                   `tfsdk:"cpu_core_count"`
	DbNodeStorageSizeInGBs       types.Int32                                                                   `tfsdk:"db_node_storage_size_in_gbs"`
	DbServerPatchingDetails      fwtypes.ListNestedObjectValueOf[dbNodePatchingDetailsDbServerDataSourceModel] `tfsdk:"db_server_patching_details"`
	DisplayName                  types.String                                                                  `tfsdk:"display_name"`
	ExadataInfrastructureId      types.String                                                                  `tfsdk:"exadata_infrastructure_id"`
	OCID                         types.String                                                                  `tfsdk:"ocid"`
	OciResourceAnchorName        types.String                                                                  `tfsdk:"oci_resource_anchor_name"`
	MaxCpuCount                  types.Int32                                                                   `tfsdk:"max_cpu_count"`
	MaxDbNodeStorageInGBs        types.Int32                                                                   `tfsdk:"max_db_node_storage_in_gbs"`
	MaxMemoryInGBs               types.Int32                                                                   `tfsdk:"max_memory_in_gbs"`
	MemorySizeInGBs              types.Int32                                                                   `tfsdk:"memory_size_in_gbs"`
	Shape                        types.String                                                                  `tfsdk:"shape"`
	CreatedAt                    timetypes.RFC3339                                                             `tfsdk:"created_at" `
	VmClusterIds                 fwtypes.ListOfString                                                          `tfsdk:"vm_cluster_ids"`
	ComputeModel                 fwtypes.StringEnum[odbtypes.ComputeModel]                                     `tfsdk:"compute_model"`
	AutonomousVmClusterIds       fwtypes.ListOfString                                                          `tfsdk:"autonomous_vm_cluster_ids"`
	AutonomousVirtualMachineIds  fwtypes.ListOfString                                                          `tfsdk:"autonomous_virtual_machine_ids"`
}

type dbNodePatchingDetailsDbServerDataSourceModel struct {
	EstimatedPatchDuration types.Int32  `tfsdk:"estimated_patch_duration"`
	PatchingStatus         types.String `tfsdk:"patching_status"`
	TimePatchingEnded      types.String `tfsdk:"time_patching_ended"`
	TimePatchingStarted    types.String `tfsdk:"time_patching_started"`
}
