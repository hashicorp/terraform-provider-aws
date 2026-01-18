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

// @FrameworkDataSource("aws_odb_db_node", name="Db Node")
func newDataSourceDBNode(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDbNode{}, nil
}

const (
	DSNameDBNode = "DB Node Data Source"
)

type dataSourceDbNode struct {
	framework.DataSourceWithModel[dbNodeDataSourceModel]
}

func (d *dataSourceDbNode) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			"cloud_vm_cluster_id": schema.StringAttribute{
				Required: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed:    true,
				CustomType:  fwtypes.StringEnumType[odbtypes.ResourceStatus](),
				Description: "The current status of the DB node.",
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed:    true,
				Description: "Additional information about the status of the DB node.",
			},
			"additional_details": schema.StringAttribute{
				Computed:    true,
				Description: "Additional information about the planned maintenance.",
			},
			"backup_ip_id": schema.StringAttribute{
				Computed:    true,
				Description: "The Oracle Cloud ID (OCID) of the backup IP address that's associated with the DB node.",
			},
			"backup_vnic2_id": schema.StringAttribute{
				Computed:    true,
				Description: "The OCID of the second backup VNIC.",
			},
			"backup_vnic_id": schema.StringAttribute{
				Computed:    true,
				Description: "The OCID of the backup VNIC.",
			},
			"cpu_core_count": schema.Int32Attribute{
				Computed:    true,
				Description: "Number of CPU cores enabled on the DB node.",
			},
			"db_storage_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The amount of local node storage, in gigabytes (GBs), allocated on the DB node.",
			},
			"db_server_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the DB server that is associated with the DB node.",
			},
			"db_system_id": schema.StringAttribute{
				Computed:    true,
				Description: "The OCID of the DB system.",
			},
			"fault_domain": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the fault domain the instance is contained in.",
			},
			"host_ip_id": schema.StringAttribute{
				Computed:    true,
				Description: "The OCID of the host IP address that's associated with the DB node.",
			},
			"hostname": schema.StringAttribute{
				Computed:    true,
				Description: "The host name for the DB node.",
			},
			"ocid": schema.StringAttribute{
				Computed:    true,
				Description: "The OCID of the DB node.",
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the OCI resource anchor for the DB node.",
			},
			"maintenance_type": schema.StringAttribute{
				Computed:    true,
				CustomType:  fwtypes.StringEnumType[odbtypes.DbNodeMaintenanceType](),
				Description: "The type of database node maintenance. Either VMDB_REBOOT_MIGRATION or EXADBXS_REBOOT_MIGRATION.",
			},
			"memory_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The allocated memory in GBs on the DB node.",
			},
			"software_storage_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The size (in GB) of the block storage volume allocation for the DB system.",
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "The date and time when the DB node was created.",
			},
			"time_maintenance_window_end": schema.StringAttribute{
				Computed:    true,
				Description: "End date and time of maintenance window.",
			},
			"time_maintenance_window_start": schema.StringAttribute{
				Computed:    true,
				Description: "Start date and time of maintenance window.",
			},
			"total_cpu_core_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The total number of CPU cores reserved on the DB node.",
			},
			"vnic2_id": schema.StringAttribute{
				Computed:    true,
				Description: "The OCID of the second VNIC.",
			},
			"vnic_id": schema.StringAttribute{
				Computed:    true,
				Description: "The OCID of the VNIC.",
			},
			"private_ip_address": schema.StringAttribute{
				Computed:    true,
				Description: "The private IP address assigned to the DB node.",
			},
			"floating_ip_address": schema.StringAttribute{
				Computed:    true,
				Description: "The floating IP address assigned to the DB node.",
			},
		},
	}
}

func (d *dataSourceDbNode) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)
	var data dbNodeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := odb.GetDbNodeInput{
		DbNodeId:         data.DbNodeId.ValueStringPointer(),
		CloudVmClusterId: data.CloudVmClusterId.ValueStringPointer(),
	}
	out, err := conn.GetDbNode(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameDBNode, data.DbNodeId.ValueString(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out.DbNode, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dbNodeDataSourceModel struct {
	framework.WithRegionModel
	CloudVmClusterId           types.String                                       `tfsdk:"cloud_vm_cluster_id"`
	DbNodeId                   types.String                                       `tfsdk:"id"`
	DbNodeArn                  types.String                                       `tfsdk:"arn"`
	Status                     fwtypes.StringEnum[odbtypes.ResourceStatus]        `tfsdk:"status"`
	StatusReason               types.String                                       `tfsdk:"status_reason"`
	AdditionalDetails          types.String                                       `tfsdk:"additional_details"`
	BackupIpId                 types.String                                       `tfsdk:"backup_ip_id"`
	BackupVnic2Id              types.String                                       `tfsdk:"backup_vnic2_id"`
	BackupVnicId               types.String                                       `tfsdk:"backup_vnic_id"`
	CpuCoreCount               types.Int32                                        `tfsdk:"cpu_core_count"`
	DbNodeStorageSizeInGBs     types.Int32                                        `tfsdk:"db_storage_size_in_gbs"`
	DbServerId                 types.String                                       `tfsdk:"db_server_id"`
	DbSystemId                 types.String                                       `tfsdk:"db_system_id"`
	FaultDomain                types.String                                       `tfsdk:"fault_domain"`
	HostIpId                   types.String                                       `tfsdk:"host_ip_id"`
	Hostname                   types.String                                       `tfsdk:"hostname"`
	Ocid                       types.String                                       `tfsdk:"ocid"`
	OciResourceAnchorName      types.String                                       `tfsdk:"oci_resource_anchor_name"`
	MaintenanceType            fwtypes.StringEnum[odbtypes.DbNodeMaintenanceType] `tfsdk:"maintenance_type"`
	MemorySizeInGBs            types.Int32                                        `tfsdk:"memory_size_in_gbs"`
	SoftwareStorageSizeInGB    types.Int32                                        `tfsdk:"software_storage_size_in_gbs"`
	CreatedAt                  timetypes.RFC3339                                  `tfsdk:"created_at"`
	TimeMaintenanceWindowEnd   types.String                                       `tfsdk:"time_maintenance_window_end"`
	TimeMaintenanceWindowStart types.String                                       `tfsdk:"time_maintenance_window_start"`
	TotalCpuCoreCount          types.Int32                                        `tfsdk:"total_cpu_core_count"`
	Vnic2Id                    types.String                                       `tfsdk:"vnic2_id"`
	VnicId                     types.String                                       `tfsdk:"vnic_id"`
	PrivateIpAddress           types.String                                       `tfsdk:"private_ip_address"`
	FloatingIpAddress          types.String                                       `tfsdk:"floating_ip_address"`
}
