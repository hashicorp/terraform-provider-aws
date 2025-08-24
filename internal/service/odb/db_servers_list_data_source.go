//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"

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
// @FrameworkDataSource("aws_odb_db_servers_list", name="Db Servers List")
func newDataSourceDbServersList(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDbServersList{}, nil
}

const (
	DSNameDbServersList = "Db Servers List Data Source"
)

type dataSourceDbServersList struct {
	framework.DataSourceWithModel[dbServersListDataSourceModel]
}

func (d *dataSourceDbServersList) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cloud_exadata_infrastructure_id": schema.StringAttribute{
				Required:    true,
				Description: "The cloud exadata infrastructure ID. Mandatory field.",
			},
			"db_servers": schema.ListAttribute{
				Description: "List of database servers associated with cloud_exadata_infrastructure_id.",
				Computed:    true,
				CustomType:  fwtypes.NewListNestedObjectTypeOf[dbServerForDbServersListDataSourceModel](ctx),
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":                      types.StringType,
						"status":                  fwtypes.StringEnumType[odbtypes.ResourceStatus](),
						"status_reason":           types.StringType,
						"cpu_core_count":          types.Int32Type,
						"cpu_core_count_per_node": types.Int32Type,
						"db_server_patching_details": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"estimated_patch_duration": types.Int32Type,
								"patching_status":          fwtypes.StringEnumType[odbtypes.DbServerPatchingStatus](),
								"time_patching_ended":      types.StringType,
								"time_patching_started":    types.StringType,
							},
						},
						"display_name":               types.StringType,
						"exadata_infrastructure_id":  types.StringType,
						"ocid":                       types.StringType,
						"oci_resource_anchor_name":   types.StringType,
						"max_cpu_count":              types.Int32Type,
						"max_db_node_storage_in_gbs": types.Int32Type,
						"max_memory_in_gbs":          types.Int32Type,
						"memory_size_in_gbs":         types.Int32Type,
						"shape":                      types.StringType,
						"vm_cluster_ids": types.ListType{
							ElemType: types.StringType,
						},
						"compute_model": fwtypes.StringEnumType[odbtypes.ComputeModel](),
						"autonomous_vm_cluster_ids": types.ListType{
							ElemType: types.StringType,
						},
						"autonomous_virtual_machine_ids": types.ListType{
							ElemType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// Data sources only have a read method.
func (d *dataSourceDbServersList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	conn := d.Meta().ODBClient(ctx)
	var data dbServersListDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := odb.ListDbServersInput{}
	if !data.CloudExadataInfrastructureId.IsNull() && !data.CloudExadataInfrastructureId.IsUnknown() {
		input.CloudExadataInfrastructureId = data.CloudExadataInfrastructureId.ValueStringPointer()
	}
	out, err := conn.ListDbServers(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameDbServersList, "", err),
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

type dbServersListDataSourceModel struct {
	framework.WithRegionModel
	CloudExadataInfrastructureId types.String                                                             `tfsdk:"cloud_exadata_infrastructure_id"`
	DbServers                    fwtypes.ListNestedObjectValueOf[dbServerForDbServersListDataSourceModel] `tfsdk:"db_servers"`
}

type dbServerForDbServersListDataSourceModel struct {
	DbServerId              types.String                                                                `tfsdk:"id"`
	Status                  fwtypes.StringEnum[odbtypes.ResourceStatus]                                 `tfsdk:"status"`
	StatusReason            types.String                                                                `tfsdk:"status_reason"`
	CpuCoreCount            types.Int32                                                                 `tfsdk:"cpu_core_count"`
	DbNodeStorageSizeInGBs  types.Int32                                                                 `tfsdk:"db_node_storage_size_in_gbs"`
	DbServerPatchingDetails fwtypes.ObjectValueOf[dbNodePatchingDetailsForDbServersListDataSourceModel] `tfsdk:"db_server_patching_details"`
	DisplayName             types.String                                                                `tfsdk:"display_name"`
	ExadataInfrastructureId types.String                                                                `tfsdk:"exadata_infrastructure_id"`
	OCID                    types.String                                                                `tfsdk:"ocid"`
	OciResourceAnchorName   types.String                                                                `tfsdk:"oci_resource_anchor_name"`
	MaxCpuCount             types.Int32                                                                 `tfsdk:"max_cpu_count"`
	MaxDbNodeStorageInGBs   types.Int32                                                                 `tfsdk:"max_db_node_storage_in_gbs"`
	MaxMemoryInGBs          types.Int32                                                                 `tfsdk:"max_memory_in_gbs"`
	MemorySizeInGBs         types.Int32                                                                 `tfsdk:"memory_size_in_gbs"`
	Shape                   types.String                                                                `tfsdk:"shape"`
	//CreatedAt                   types.String                                                                `tfsdk:"created_at"`
	VmClusterIds                fwtypes.ListOfString                      `tfsdk:"vm_cluster_ids"`
	ComputeModel                fwtypes.StringEnum[odbtypes.ComputeModel] `tfsdk:"compute_model"`
	AutonomousVmClusterIds      fwtypes.ListOfString                      `tfsdk:"autonomous_vm_cluster_ids"`
	AutonomousVirtualMachineIds fwtypes.ListOfString                      `tfsdk:"autonomous_virtual_machine_ids"`
}

type dbNodePatchingDetailsForDbServersListDataSourceModel struct {
	EstimatedPatchDuration types.Int32                                         `tfsdk:"estimated_patch_duration"`
	PatchingStatus         fwtypes.StringEnum[odbtypes.DbServerPatchingStatus] `tfsdk:"patching_status"`
	TimePatchingEnded      types.String                                        `tfsdk:"time_patching_ended"`
	TimePatchingStarted    types.String                                        `tfsdk:"time_patching_started"`
}
