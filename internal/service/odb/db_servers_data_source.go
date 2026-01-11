// Copyright IBM Corp. 2014, 2026
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

// @FrameworkDataSource("aws_odb_db_servers", name="Db Servers")
func newDataSourceDBServers(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDbServersList{}, nil
}

const (
	DSNameDBServersList = "DB Servers List Data Source"
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
				CustomType:  fwtypes.NewListNestedObjectTypeOf[dbServerForDbServersListDataSourceModel](ctx),
				Computed:    true,
				Description: "List of database servers associated with cloud_exadata_infrastructure_id.",
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
	paginator := odb.NewListDbServersPaginator(conn, &input)
	var out odb.ListDbServersOutput
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameDBServersList, "", err),
				err.Error(),
			)
		}
		out.DbServers = append(out.DbServers, page.DbServers...)
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
	AutonomousVirtualMachineIds fwtypes.ListOfString                                                                  `tfsdk:"autonomous_virtual_machine_ids"`
	AutonomousVmClusterIds      fwtypes.ListOfString                                                                  `tfsdk:"autonomous_vm_cluster_ids"`
	ComputeModel                fwtypes.StringEnum[odbtypes.ComputeModel]                                             `tfsdk:"compute_model"`
	CreatedAt                   timetypes.RFC3339                                                                     `tfsdk:"created_at"`
	CpuCoreCount                types.Int32                                                                           `tfsdk:"cpu_core_count"`
	DbNodeStorageSizeInGBs      types.Int32                                                                           `tfsdk:"db_node_storage_size_in_gbs"`
	DbServerId                  types.String                                                                          `tfsdk:"id"`
	DbServerPatchingDetails     fwtypes.ListNestedObjectValueOf[dbNodePatchingDetailsForDbServersListDataSourceModel] `tfsdk:"db_server_patching_details"`
	DisplayName                 types.String                                                                          `tfsdk:"display_name"`
	ExadataInfrastructureId     types.String                                                                          `tfsdk:"exadata_infrastructure_id"`
	MaxCpuCount                 types.Int32                                                                           `tfsdk:"max_cpu_count"`
	MaxDbNodeStorageInGBs       types.Int32                                                                           `tfsdk:"max_db_node_storage_in_gbs"`
	MaxMemoryInGBs              types.Int32                                                                           `tfsdk:"max_memory_in_gbs"`
	MemorySizeInGBs             types.Int32                                                                           `tfsdk:"memory_size_in_gbs"`
	OCID                        types.String                                                                          `tfsdk:"ocid"`
	OciResourceAnchorName       types.String                                                                          `tfsdk:"oci_resource_anchor_name"`
	Shape                       types.String                                                                          `tfsdk:"shape"`
	Status                      fwtypes.StringEnum[odbtypes.ResourceStatus]                                           `tfsdk:"status"`
	StatusReason                types.String                                                                          `tfsdk:"status_reason"`
	VmClusterIds                fwtypes.ListOfString                                                                  `tfsdk:"vm_cluster_ids"`
}

type dbNodePatchingDetailsForDbServersListDataSourceModel struct {
	EstimatedPatchDuration types.Int32                                         `tfsdk:"estimated_patch_duration"`
	PatchingStatus         fwtypes.StringEnum[odbtypes.DbServerPatchingStatus] `tfsdk:"patching_status"`
	TimePatchingEnded      types.String                                        `tfsdk:"time_patching_ended"`
	TimePatchingStarted    types.String                                        `tfsdk:"time_patching_started"`
}
