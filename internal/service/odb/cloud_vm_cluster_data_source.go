//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb

import (
	"context"
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

// @FrameworkDataSource("aws_odb_cloud_vm_cluster", name="Cloud Vm Cluster")
func newDataSourceCloudVmCluster(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceCloudVmCluster{}, nil
}

const (
	DSNameCloudVmCluster = "Cloud Vm Cluster Data Source"
)

type dataSourceCloudVmCluster struct {
	framework.DataSourceWithModel[dataSourceCloudVmClusterModel]
}

func (d *dataSourceCloudVmCluster) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	statusType := fwtypes.StringEnumType[odbtypes.ResourceStatus]()
	diskRedundancyType := fwtypes.StringEnumType[odbtypes.DiskRedundancy]()
	licenseModelType := fwtypes.StringEnumType[odbtypes.LicenseModel]()
	computeModelType := fwtypes.StringEnumType[odbtypes.ComputeModel]()
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			"cloud_exadata_infrastructure_id": schema.StringAttribute{
				Computed: true,
			},
			"cluster_name": schema.StringAttribute{
				Computed: true,
			},
			"cpu_core_count": schema.Int32Attribute{
				Computed: true,
			},
			"data_storage_size_in_tbs": schema.Float64Attribute{
				Computed: true,
			},
			"db_node_storage_size_in_gbs": schema.Int32Attribute{
				Computed: true,
			},
			"db_servers": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
			},
			"disk_redundancy": schema.StringAttribute{
				CustomType: diskRedundancyType,
				Computed:   true,
			},
			"display_name": schema.StringAttribute{
				Computed: true,
			},
			"domain": schema.StringAttribute{
				Computed: true,
			},
			"gi_version": schema.StringAttribute{
				Computed: true,
			},
			"hostname_prefix_computed": schema.StringAttribute{
				Computed: true,
			},
			"is_local_backup_enabled": schema.BoolAttribute{
				Computed: true,
			},
			"is_sparse_disk_group_enabled": schema.BoolAttribute{
				Computed: true,
			},
			"last_update_history_entry_id": schema.StringAttribute{
				Computed: true,
			},
			"license_model": schema.StringAttribute{
				CustomType: licenseModelType,
				Computed:   true,
			},
			"listener_port": schema.Int32Attribute{
				Computed: true,
			},
			"memory_size_in_gbs": schema.Int32Attribute{
				Computed: true,
			},
			"node_count": schema.Int32Attribute{
				Computed: true,
			},
			"ocid": schema.StringAttribute{
				Computed: true,
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed: true,
			},
			"oci_url": schema.StringAttribute{
				Computed: true,
			},
			"odb_network_id": schema.StringAttribute{
				Computed: true,
			},
			"percent_progress": schema.Float64Attribute{
				Computed: true,
			},
			"scan_dns_name": schema.StringAttribute{
				Computed: true,
			},
			"scan_dns_record_id": schema.StringAttribute{
				Computed: true,
			},
			"scan_ip_ids": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
			},
			"shape": schema.StringAttribute{
				Computed: true,
			},
			"ssh_public_keys": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
			},
			"status": schema.StringAttribute{
				CustomType: statusType,
				Computed:   true,
			},
			"status_reason": schema.StringAttribute{
				Computed: true,
			},
			"storage_size_in_gbs": schema.Int32Attribute{
				Computed: true,
			},
			"system_version": schema.StringAttribute{
				Computed: true,
			},
			"timezone": schema.StringAttribute{
				Computed: true,
			},
			"vip_ids": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"compute_model": schema.StringAttribute{
				CustomType: computeModelType,
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"data_collection_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataCollectionOptions](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"is_diagnostics_events_enabled": schema.BoolAttribute{
							Computed: true,
						},
						"is_health_monitoring_enabled": schema.BoolAttribute{
							Computed: true,
						},
						"is_incident_logs_enabled": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
			"iorm_config_cache": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[exadataIormConfig](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"lifecycle_details": schema.StringAttribute{
							Computed: true,
						},
						"lifecycle_state": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[odbtypes.IormLifecycleState](),
							Computed:   true,
						},
						"objective": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[odbtypes.Objective](),
							Computed:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"db_plans": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dbIormConfig](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"db_name": schema.StringAttribute{
										Computed: true,
									},
									"flash_cache_limit": schema.StringAttribute{
										Computed: true,
									},
									"share": schema.Int32Attribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Data sources only have a read method.
func (d *dataSourceCloudVmCluster) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	conn := d.Meta().ODBClient(ctx)

	var data dataSourceCloudVmClusterModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.GetCloudVmClusterInput{
		CloudVmClusterId: data.CloudVmClusterId.ValueStringPointer(),
	}

	out, err := conn.GetCloudVmCluster(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameCloudVmCluster, data.CloudVmClusterId.ValueString(), err),
			err.Error(),
		)
		return
	}

	data.HostnamePrefixComputed = types.StringValue(*out.CloudVmCluster.Hostname)
	data.CreatedAt = types.StringValue(time.Time{}.Format(time.RFC3339))

	resp.Diagnostics.Append(flex.Flatten(ctx, out.CloudVmCluster, &data, flex.WithIgnoredFieldNamesAppend("HostnamePrefixComputed"),
		flex.WithIgnoredFieldNamesAppend("CreatedAt"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceCloudVmClusterModel struct {
	framework.WithRegionModel
	CloudExadataInfrastructureId types.String                                           `tfsdk:"cloud_exadata_infrastructure_id"`
	CloudVmClusterArn            types.String                                           `tfsdk:"arn"`
	CloudVmClusterId             types.String                                           `tfsdk:"id"`
	ClusterName                  types.String                                           `tfsdk:"cluster_name"`
	CpuCoreCount                 types.Int32                                            `tfsdk:"cpu_core_count"`
	DataCollectionOptions        fwtypes.ListNestedObjectValueOf[dataCollectionOptions] `tfsdk:"data_collection_options"`
	DataStorageSizeInTBs         types.Float64                                          `tfsdk:"data_storage_size_in_tbs"`
	DbNodeStorageSizeInGBs       types.Int32                                            `tfsdk:"db_node_storage_size_in_gbs"`
	DbServers                    fwtypes.ListValueOf[types.String]                      `tfsdk:"db_servers"`
	DiskRedundancy               fwtypes.StringEnum[odbtypes.DiskRedundancy]            `tfsdk:"disk_redundancy"`
	DisplayName                  types.String                                           `tfsdk:"display_name"`
	Domain                       types.String                                           `tfsdk:"domain"`
	GiVersion                    types.String                                           `tfsdk:"gi_version"`
	HostnamePrefixComputed       types.String                                           `tfsdk:"hostname_prefix_computed"`
	IormConfigCache              fwtypes.ListNestedObjectValueOf[exadataIormConfig]     `tfsdk:"iorm_config_cache"`
	IsLocalBackupEnabled         types.Bool                                             `tfsdk:"is_local_backup_enabled"`
	IsSparseDiskGroupEnabled     types.Bool                                             `tfsdk:"is_sparse_disk_group_enabled"`
	LastUpdateHistoryEntryId     types.String                                           `tfsdk:"last_update_history_entry_id"`
	LicenseModel                 fwtypes.StringEnum[odbtypes.LicenseModel]              `tfsdk:"license_model"`
	ListenerPort                 types.Int32                                            `tfsdk:"listener_port"`
	MemorySizeInGbs              types.Int32                                            `tfsdk:"memory_size_in_gbs"`
	NodeCount                    types.Int32                                            `tfsdk:"node_count"`
	Ocid                         types.String                                           `tfsdk:"ocid"`
	OciResourceAnchorName        types.String                                           `tfsdk:"oci_resource_anchor_name"`
	OciUrl                       types.String                                           `tfsdk:"oci_url"`
	OdbNetworkId                 types.String                                           `tfsdk:"odb_network_id"`
	PercentProgress              types.Float64                                          `tfsdk:"percent_progress"`
	ScanDnsName                  types.String                                           `tfsdk:"scan_dns_name"`
	ScanDnsRecordId              types.String                                           `tfsdk:"scan_dns_record_id"`
	ScanIpIds                    fwtypes.ListValueOf[types.String]                      `tfsdk:"scan_ip_ids"`
	Shape                        types.String                                           `tfsdk:"shape"`
	SshPublicKeys                fwtypes.ListValueOf[types.String]                      `tfsdk:"ssh_public_keys"`
	Status                       fwtypes.StringEnum[odbtypes.ResourceStatus]            `tfsdk:"status"`
	StatusReason                 types.String                                           `tfsdk:"status_reason"`
	StorageSizeInGBs             types.Int32                                            `tfsdk:"storage_size_in_gbs"`
	SystemVersion                types.String                                           `tfsdk:"system_version"`
	Timezone                     types.String                                           `tfsdk:"timezone"`
	VipIds                       fwtypes.ListValueOf[types.String]                      `tfsdk:"vip_ids"`
	CreatedAt                    types.String                                           `tfsdk:"created_at"`
	ComputeModel                 fwtypes.StringEnum[odbtypes.ComputeModel]              `tfsdk:"compute_model"`
}

type dataCollectionOptions struct {
	IsDiagnosticsEventsEnabled types.Bool `tfsdk:"is_diagnostics_events_enabled"`
	IsHealthMonitoringEnabled  types.Bool `tfsdk:"is_health_monitoring_enabled"`
	IsIncidentLogsEnabled      types.Bool `tfsdk:"is_incident_logs_enabled"`
}

type exadataIormConfig struct {
	DbPlans          fwtypes.ListNestedObjectValueOf[dbIormConfig]   `tfsdk:"db_plans"`
	LifecycleDetails types.String                                    `tfsdk:"lifecycle_details"`
	LifecycleState   fwtypes.StringEnum[odbtypes.IormLifecycleState] `tfsdk:"lifecycle_state"`
	Objective        fwtypes.StringEnum[odbtypes.Objective]          `tfsdk:"objective"`
}

type dbIormConfig struct {
	DbName          types.String `tfsdk:"db_name"`
	FlashCacheLimit types.String `tfsdk:"flash_cache_limit"`
	Share           types.Int32  `tfsdk:"share"`
}
