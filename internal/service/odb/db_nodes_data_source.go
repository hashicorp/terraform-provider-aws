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

// @FrameworkDataSource("aws_odb_db_nodes", name="Db Nodes")
func newDataSourceDBNodes(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDbNodesList{}, nil
}

const (
	DSNameDBNodesList = "DB Nodes List Data Source"
)

type dataSourceDbNodesList struct {
	framework.DataSourceWithModel[dbNodesListDataSourceModel]
}

func (d *dataSourceDbNodesList) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cloud_vm_cluster_id": schema.StringAttribute{
				Required:    true,
				Description: "Id of the cloud VM cluster. The unique identifier of the VM cluster.",
			},
			"db_nodes": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[dbNodeForDbNodesListDataSourceModel](ctx),
				Computed:    true,
				Description: "The list of DB nodes along with their properties.",
			},
		},
	}
}

func (d *dataSourceDbNodesList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)
	var data dbNodesListDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := odb.ListDbNodesInput{
		CloudVmClusterId: data.CloudVmClusterId.ValueStringPointer(),
	}
	var out odb.ListDbNodesOutput
	paginator := odb.NewListDbNodesPaginator(conn, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameDBNodesList, data.CloudVmClusterId.ValueString(), err),
				err.Error(),
			)
			return
		}
		out.DbNodes = append(out.DbNodes, page.DbNodes...)
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dbNodesListDataSourceModel struct {
	framework.WithRegionModel
	CloudVmClusterId types.String                                                         `tfsdk:"cloud_vm_cluster_id"`
	DbNodes          fwtypes.ListNestedObjectValueOf[dbNodeForDbNodesListDataSourceModel] `tfsdk:"db_nodes"`
}

type dbNodeForDbNodesListDataSourceModel struct {
	AdditionalDetails          types.String                                       `tfsdk:"additional_details"`
	BackupIpId                 types.String                                       `tfsdk:"backup_ip_id"`
	BackupVnic2Id              types.String                                       `tfsdk:"backup_vnic2_id"`
	BackupVnicId               types.String                                       `tfsdk:"backup_vnic_id"`
	CpuCoreCount               types.Int32                                        `tfsdk:"cpu_core_count"`
	CreatedAt                  timetypes.RFC3339                                  `tfsdk:"created_at"`
	DbNodeArn                  types.String                                       `tfsdk:"arn"`
	DbNodeId                   types.String                                       `tfsdk:"id"`
	DbNodeStorageSizeInGBs     types.Int32                                        `tfsdk:"db_node_storage_size"`
	DbServerId                 types.String                                       `tfsdk:"db_server_id"`
	DbSystemId                 types.String                                       `tfsdk:"db_system_id"`
	FaultDomain                types.String                                       `tfsdk:"fault_domain"`
	HostIpId                   types.String                                       `tfsdk:"host_ip_id"`
	Hostname                   types.String                                       `tfsdk:"hostname"`
	MaintenanceType            fwtypes.StringEnum[odbtypes.DbNodeMaintenanceType] `tfsdk:"maintenance_type"`
	MemorySizeInGBs            types.Int32                                        `tfsdk:"memory_size"`
	OciResourceAnchorName      types.String                                       `tfsdk:"oci_resource_anchor_name"`
	Ocid                       types.String                                       `tfsdk:"ocid"`
	SoftwareStorageSizeInGB    types.Int32                                        `tfsdk:"software_storage_size"`
	Status                     fwtypes.StringEnum[odbtypes.DbNodeResourceStatus]  `tfsdk:"status"`
	StatusReason               types.String                                       `tfsdk:"status_reason"`
	TimeMaintenanceWindowEnd   types.String                                       `tfsdk:"time_maintenance_window_end"`
	TimeMaintenanceWindowStart types.String                                       `tfsdk:"time_maintenance_window_start"`
	TotalCpuCoreCount          types.Int32                                        `tfsdk:"total_cpu_core_count"`
	Vnic2Id                    types.String                                       `tfsdk:"vnic2_id"`
	VnicId                     types.String                                       `tfsdk:"vnic_id"`
}
