// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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

// @FrameworkDataSource("aws_odb_cloud_vm_cluster", name="Cloud Vm Cluster")
// @Tags(identifierAttribute="arn")
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
				Required:    true,
				Description: "The unique identifier of the VM cluster.",
			},
			"cloud_exadata_infrastructure_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Cloud Exadata Infrastructure.",
			},
			"cloud_exadata_infrastructure_arn": schema.StringAttribute{
				Computed:    true,
				Description: "The ARN of the Cloud Exadata Infrastructure.",
			},
			names.AttrClusterName: schema.StringAttribute{
				Computed:    true,
				Description: "The name of the Grid Infrastructure (GI) cluster.",
			},
			"cpu_core_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of CPU cores enabled on the VM cluster.",
			},
			"data_storage_size_in_tbs": schema.Float64Attribute{
				Computed:    true,
				Description: "The size of the data disk group, in terabytes (TB), that's allocated for the VM cluster.",
			},
			"db_node_storage_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The amount of local node storage, in gigabytes (GB), that's allocated for the VM cluster.",
			},
			"db_servers": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Description: "The list of database servers for the VM cluster.",
			},
			"disk_redundancy": schema.StringAttribute{
				CustomType:  diskRedundancyType,
				Computed:    true,
				Description: "The type of redundancy configured for the VM cluster. NORMAL is 2-way redundancy. HIGH is 3-way redundancy.",
			},
			names.AttrDisplayName: schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the VM cluster.",
			},
			names.AttrDomain: schema.StringAttribute{
				Computed:    true,
				Description: "The domain name of the VM cluster.",
			},
			"gi_version": schema.StringAttribute{
				Computed:    true,
				Description: "The software version of the Oracle Grid Infrastructure (GI) for the VM cluster.",
			},
			"hostname_prefix_computed": schema.StringAttribute{
				Computed:    true,
				Description: "The computed hostname prefix for the VM cluster.",
			},
			"is_local_backup_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates whether database backups to local Exadata storage is enabled for the VM cluster.",
			},
			"is_sparse_disk_group_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates whether the VM cluster is configured with a sparse disk group.",
			},
			"last_update_history_entry_id": schema.StringAttribute{
				Computed:    true,
				Description: "The Oracle Cloud ID (OCID) of the last maintenance update history entry.",
			},
			"license_model": schema.StringAttribute{
				CustomType:  licenseModelType,
				Computed:    true,
				Description: "The Oracle license model applied to the VM cluster.",
			},
			"listener_port": schema.Int32Attribute{
				Computed:    true,
				Description: "The port number configured for the listener on the VM cluster.",
			},
			"memory_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The amount of memory, in gigabytes (GB), that's allocated for the VM cluster.",
			},
			"node_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of nodes in the VM cluster.",
			},
			"ocid": schema.StringAttribute{
				Computed:    true,
				Description: "The OCID of the VM cluster.",
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the OCI Resource Anchor.",
			},
			"oci_url": schema.StringAttribute{
				Computed:    true,
				Description: "The HTTPS link to the VM cluster in OCI.",
			},
			"odb_network_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the ODB network.",
			},
			"odb_network_arn": schema.StringAttribute{
				Computed:    true,
				Description: "The ARN of the ODB network.",
			},
			"percent_progress": schema.Float64Attribute{
				Computed:    true,
				Description: "The amount of progress made on the current operation on the VM cluster,expressed as a percentage.",
			},
			"scan_dns_name": schema.StringAttribute{
				Computed: true,
				Description: "The FQDN of the DNS record for the Single Client Access Name (SCAN) IP\n" +
					" addresses that are associated with the VM cluster.",
			},
			"scan_dns_record_id": schema.StringAttribute{
				Computed:    true,
				Description: "The OCID of the DNS record for the SCAN IP addresses that are associated with the VM cluster.",
			},
			"scan_ip_ids": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Description: "The OCID of the SCAN IP addresses that are associated with the VM cluster.",
			},
			"shape": schema.StringAttribute{
				Computed:    true,
				Description: "The hardware model name of the Exadata infrastructure that's running the VM cluster.",
			},
			"ssh_public_keys": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Description: "The public key portion of one or more key pairs used for SSH access to the VM cluster.",
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType:  statusType,
				Computed:    true,
				Description: "The status of the VM cluster.",
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed:    true,
				Description: "Additional information about the status of the VM cluster.",
			},
			"storage_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The amount of local node storage, in gigabytes (GB), that's allocated to the VM cluster.",
			},
			"system_version": schema.StringAttribute{
				Computed:    true,
				Description: "The operating system version of the image chosen for the VM cluster.",
			},
			"timezone": schema.StringAttribute{
				Computed:    true,
				Description: "The time zone of the VM cluster.",
			},
			"vip_ids": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Description: "The virtual IP (VIP) addresses that are associated with the VM cluster.\n" +
					"Oracle's Cluster Ready Services (CRS) creates and maintains one VIP address for\n" +
					"each node in the VM cluster to enable failover. If one node fails, the VIP is\n" +
					"reassigned to another active node in the cluster.",
			},
			names.AttrCreatedAt: schema.StringAttribute{
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Description: "The time when the VM cluster was created.",
			},
			"compute_model": schema.StringAttribute{
				CustomType: computeModelType,
				Computed:   true,
				Description: "The OCI model compute model used when you create or clone an instance: ECPU or\n" +
					"OCPU. An ECPU is an abstracted measure of compute resources. ECPUs are based on\n" +
					"the number of cores elastically allocated from a pool of compute and storage\n" +
					"servers. An OCPU is a legacy physical measure of compute resources. OCPUs are\n" +
					"based on the physical core of a processor with hyper-threading enabled.",
			},
			"data_collection_options": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.NewListNestedObjectTypeOf[dataCollectionOptionsVMCDataSourceModel](ctx),
				Description: "The set of diagnostic collection options enabled for the VM cluster.",
			},
			"iorm_config_cache": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.NewListNestedObjectTypeOf[exadataIormConfigVMCDataSourceModel](ctx),
				Description: "The ExadataIormConfig cache details for the VM cluster.",
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

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
	resp.Diagnostics.Append(flex.Flatten(ctx, out.CloudVmCluster, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceCloudVmClusterModel struct {
	framework.WithRegionModel
	CloudExadataInfrastructureId  types.String                                                             `tfsdk:"cloud_exadata_infrastructure_id"`
	CloudExadataInfrastructureArn types.String                                                             `tfsdk:"cloud_exadata_infrastructure_arn" autoflex:"-"`
	CloudVmClusterArn             types.String                                                             `tfsdk:"arn"`
	CloudVmClusterId              types.String                                                             `tfsdk:"id"`
	ClusterName                   types.String                                                             `tfsdk:"cluster_name"`
	CpuCoreCount                  types.Int32                                                              `tfsdk:"cpu_core_count"`
	DataCollectionOptions         fwtypes.ListNestedObjectValueOf[dataCollectionOptionsVMCDataSourceModel] `tfsdk:"data_collection_options"`
	DataStorageSizeInTBs          types.Float64                                                            `tfsdk:"data_storage_size_in_tbs"`
	DbNodeStorageSizeInGBs        types.Int32                                                              `tfsdk:"db_node_storage_size_in_gbs"`
	DbServers                     fwtypes.ListValueOf[types.String]                                        `tfsdk:"db_servers"`
	DiskRedundancy                fwtypes.StringEnum[odbtypes.DiskRedundancy]                              `tfsdk:"disk_redundancy"`
	DisplayName                   types.String                                                             `tfsdk:"display_name"`
	Domain                        types.String                                                             `tfsdk:"domain"`
	GiVersion                     types.String                                                             `tfsdk:"gi_version"`
	HostnamePrefixComputed        types.String                                                             `tfsdk:"hostname_prefix_computed" autoflex:",noflatten"`
	IormConfigCache               fwtypes.ListNestedObjectValueOf[exadataIormConfigVMCDataSourceModel]     `tfsdk:"iorm_config_cache"`
	IsLocalBackupEnabled          types.Bool                                                               `tfsdk:"is_local_backup_enabled"`
	IsSparseDiskGroupEnabled      types.Bool                                                               `tfsdk:"is_sparse_disk_group_enabled"`
	LastUpdateHistoryEntryId      types.String                                                             `tfsdk:"last_update_history_entry_id"`
	LicenseModel                  fwtypes.StringEnum[odbtypes.LicenseModel]                                `tfsdk:"license_model"`
	ListenerPort                  types.Int32                                                              `tfsdk:"listener_port"`
	MemorySizeInGbs               types.Int32                                                              `tfsdk:"memory_size_in_gbs"`
	NodeCount                     types.Int32                                                              `tfsdk:"node_count"`
	Ocid                          types.String                                                             `tfsdk:"ocid"`
	OciResourceAnchorName         types.String                                                             `tfsdk:"oci_resource_anchor_name"`
	OciUrl                        types.String                                                             `tfsdk:"oci_url"`
	OdbNetworkId                  types.String                                                             `tfsdk:"odb_network_id"`
	OdbNetworkArn                 types.String                                                             `tfsdk:"odb_network_arn" autoflex:"-"`
	PercentProgress               types.Float64                                                            `tfsdk:"percent_progress"`
	ScanDnsName                   types.String                                                             `tfsdk:"scan_dns_name"`
	ScanDnsRecordId               types.String                                                             `tfsdk:"scan_dns_record_id"`
	ScanIpIds                     fwtypes.ListValueOf[types.String]                                        `tfsdk:"scan_ip_ids"`
	Shape                         types.String                                                             `tfsdk:"shape"`
	SshPublicKeys                 fwtypes.ListValueOf[types.String]                                        `tfsdk:"ssh_public_keys"`
	Status                        fwtypes.StringEnum[odbtypes.ResourceStatus]                              `tfsdk:"status"`
	StatusReason                  types.String                                                             `tfsdk:"status_reason"`
	StorageSizeInGBs              types.Int32                                                              `tfsdk:"storage_size_in_gbs"`
	SystemVersion                 types.String                                                             `tfsdk:"system_version"`
	Timezone                      types.String                                                             `tfsdk:"timezone"`
	VipIds                        fwtypes.ListValueOf[types.String]                                        `tfsdk:"vip_ids"`
	CreatedAt                     timetypes.RFC3339                                                        `tfsdk:"created_at"`
	ComputeModel                  fwtypes.StringEnum[odbtypes.ComputeModel]                                `tfsdk:"compute_model"`
	Tags                          tftags.Map                                                               `tfsdk:"tags"`
}

type dataCollectionOptionsVMCDataSourceModel struct {
	IsDiagnosticsEventsEnabled types.Bool `tfsdk:"is_diagnostics_events_enabled"`
	IsHealthMonitoringEnabled  types.Bool `tfsdk:"is_health_monitoring_enabled"`
	IsIncidentLogsEnabled      types.Bool `tfsdk:"is_incident_logs_enabled"`
}

type exadataIormConfigVMCDataSourceModel struct {
	DbPlans          fwtypes.ListNestedObjectValueOf[dbIormConfigVMCDatasourceModel] `tfsdk:"db_plans"`
	LifecycleDetails types.String                                                    `tfsdk:"lifecycle_details"`
	LifecycleState   fwtypes.StringEnum[odbtypes.IormLifecycleState]                 `tfsdk:"lifecycle_state"`
	Objective        fwtypes.StringEnum[odbtypes.Objective]                          `tfsdk:"objective"`
}

type dbIormConfigVMCDatasourceModel struct {
	DbName          types.String `tfsdk:"db_name"`
	FlashCacheLimit types.String `tfsdk:"flash_cache_limit"`
	Share           types.Int32  `tfsdk:"share"`
}
