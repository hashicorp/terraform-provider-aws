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

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_odb_cloud_autonomous_vm_cluster", name="Cloud Autonomous Vm Cluster")
// @Tags(identifierAttribute="arn")
func newDataSourceCloudAutonomousVmCluster(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceCloudAutonomousVmCluster{}, nil
}

const (
	DSNameCloudAutonomousVmCluster = "Cloud Autonomous Vm Cluster Data Source"
)

type dataSourceCloudAutonomousVmCluster struct {
	framework.DataSourceWithModel[cloudAutonomousVmClusterDataSourceModel]
}

func (d *dataSourceCloudAutonomousVmCluster) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	status := fwtypes.StringEnumType[odbtypes.ResourceStatus]()
	licenseModel := fwtypes.StringEnumType[odbtypes.LicenseModel]()
	computeModel := fwtypes.StringEnumType[odbtypes.ComputeModel]()
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),

			names.AttrID: schema.StringAttribute{
				Required:    true,
				Description: "Unique ID of the Autonomous VM cluster.",
			},
			"cloud_exadata_infrastructure_id": schema.StringAttribute{
				Computed:    true,
				Description: "Cloud exadata infrastructure id associated with this cloud autonomous VM cluster.",
			},
			"cloud_exadata_infrastructure_arn": schema.StringAttribute{
				Computed:    true,
				Description: "Cloud exadata infrastructure arn associated with this cloud autonomous VM cluster.",
			},
			"autonomous_data_storage_percentage": schema.Float32Attribute{
				Computed:    true,
				Description: "The percentage of data storage currently in use for Autonomous Databases in the Autonomous VM cluster.",
			},
			"autonomous_data_storage_size_in_tbs": schema.Float64Attribute{
				Computed:    true,
				Description: "The data storage size allocated for Autonomous Databases in the Autonomous VM cluster, in TB.",
			},
			"available_autonomous_data_storage_size_in_tbs": schema.Float64Attribute{
				Computed:    true,
				Description: "The available data storage space for Autonomous Databases in the Autonomous VM cluster, in TB.",
			},
			"available_container_databases": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of Autonomous CDBs that you can create with the currently available storage.",
			},
			"available_cpus": schema.Float32Attribute{
				Computed:    true,
				Description: "The number of CPU cores available for allocation to Autonomous Databases.",
			},
			"compute_model": schema.StringAttribute{
				CustomType:  computeModel,
				Computed:    true,
				Description: " The compute model of the Autonomous VM cluster: ECPU or OCPU.",
			},
			"cpu_core_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The total number of CPU cores in the Autonomous VM cluster.",
			},
			"cpu_core_count_per_node": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of CPU cores enabled per node in the Autonomous VM cluster.",
			},
			"cpu_percentage": schema.Float32Attribute{
				Computed:    true,
				Description: "The percentage of total CPU cores currently in use in the Autonomous VM cluster.",
			},
			names.AttrCreatedAt: schema.StringAttribute{
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Description: "The date and time when the Autonomous VM cluster was created.",
			},
			"data_storage_size_in_gbs": schema.Float64Attribute{
				Computed:    true,
				Description: "The total data storage allocated to the Autonomous VM cluster, in GB.",
			},
			"data_storage_size_in_tbs": schema.Float64Attribute{
				Computed:    true,
				Description: "The total data storage allocated to the Autonomous VM cluster, in TB.",
			},
			"odb_node_storage_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The local node storage allocated to the Autonomous VM cluster, in gigabytes (GB).",
			},
			"db_servers": schema.SetAttribute{
				Computed:    true,
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Description: "The list of database servers associated with the Autonomous VM cluster.",
			},
			names.AttrDescription: schema.StringAttribute{
				Computed:    true,
				Description: "The user-provided description of the Autonomous VM cluster.",
			},
			names.AttrDisplayName: schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the Autonomous VM cluster.",
			},
			names.AttrDomain: schema.StringAttribute{
				Computed:    true,
				Description: "The domain name of the Autonomous VM cluster.",
			},
			"exadata_storage_in_tbs_lowest_scaled_value": schema.Float64Attribute{
				Computed:    true,
				Description: "The minimum value to which you can scale down the Exadata storage, in TB.",
			},
			"hostname": schema.StringAttribute{
				Computed:    true,
				Description: "The hostname of the Autonomous VM cluster.",
			},
			"is_mtls_enabled_vm_cluster": schema.BoolAttribute{
				Computed:    true,
				Description: " Indicates whether mutual TLS (mTLS) authentication is enabled for the Autonomous VM cluster.",
			},
			"license_model": schema.StringAttribute{
				CustomType:  licenseModel,
				Computed:    true,
				Description: "The Oracle license model that applies to the Autonomous VM cluster. Valid values are LICENSE_INCLUDED or BRING_YOUR_OWN_LICENSE .",
			},
			"max_acds_lowest_scaled_value": schema.Int32Attribute{
				Computed:    true,
				Description: "The minimum value to which you can scale down the maximum number of Autonomous CDBs.",
			},
			"memory_per_oracle_compute_unit_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The amount of memory allocated per Oracle Compute Unit, in GB.",
			},
			"memory_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "The total amount of memory allocated to the Autonomous VM cluster, in gigabytes (GB).",
			},
			"node_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of database server nodes in the Autonomous VM cluster.",
			},
			"non_provisionable_autonomous_container_databases": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of Autonomous CDBs that can't be provisioned because of resource  constraints.",
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the OCI resource anchor associated with this Autonomous VM cluster.",
			},
			"oci_url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL for accessing the OCI console page for this Autonomous VM cluster.",
			},
			"ocid": schema.StringAttribute{
				Computed:    true,
				Description: "The Oracle Cloud Identifier (OCID) of the Autonomous VM cluster.",
			},
			"odb_network_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the ODB network associated with this Autonomous VM cluster.",
			},
			"odb_network_arn": schema.StringAttribute{
				Computed:    true,
				Description: "The arn of the ODB network associated with this Autonomous VM cluster.",
			},
			"percent_progress": schema.Float32Attribute{
				Computed:    true,
				Description: "The progress of the current operation on the Autonomous VM cluster, as a percentage.",
			},
			"provisionable_autonomous_container_databases": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of Autonomous CDBs that can be provisioned in the Autonomous VM cluster.",
			},
			"provisioned_autonomous_container_databases": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of Autonomous CDBs currently provisioned in the Autonomous VM cluster.",
			},
			"provisioned_cpus": schema.Float32Attribute{
				Computed:    true,
				Description: "The number of CPU cores currently provisioned in the Autonomous VM cluster.",
			},
			"reclaimable_cpus": schema.Float32Attribute{
				Computed:    true,
				Description: "The number of CPU cores that can be reclaimed from terminated or scaled-down Autonomous Databases.",
			},
			"reserved_cpus": schema.Float32Attribute{
				Computed:    true,
				Description: "The number of CPU cores reserved for system operations and redundancy.",
			},
			"scan_listener_port_non_tls": schema.Int32Attribute{
				Computed:    true,
				Description: "The SCAN listener port for non-TLS (TCP) protocol. The default is 1521.",
			},
			"scan_listener_port_tls": schema.Int32Attribute{
				Computed:    true,
				Description: "The SCAN listener port for TLS (TCP) protocol. The default is 2484.",
			},
			"shape": schema.StringAttribute{
				Computed:    true,
				Description: "The shape of the Exadata infrastructure for the Autonomous VM cluster.",
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType:  status,
				Computed:    true,
				Description: "The status of the Autonomous VM cluster.",
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed:    true,
				Description: "Additional information about the current status of the Autonomous VM cluster.",
			},
			"time_database_ssl_certificate_expires": schema.StringAttribute{
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Description: "The expiration date and time of the database SSL certificate.",
			},
			"time_ords_certificate_expires": schema.StringAttribute{
				Computed:    true,
				CustomType:  timetypes.RFC3339Type{},
				Description: "The expiration date and time of the Oracle REST Data Services (ORDS)certificate .",
			},
			"time_zone": schema.StringAttribute{
				Computed:    true,
				Description: "The time zone of the Autonomous VM cluster.",
			},
			"total_container_databases": schema.Int32Attribute{
				Computed:    true,
				Description: "The total number of Autonomous Container Databases that can be created with the allocated local storage.",
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"maintenance_window": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.NewListNestedObjectTypeOf[cloudAutonomousVmClusterMaintenanceWindowDataSourceModel](ctx),
				Description: "The maintenance window for the Autonomous VM cluster.",
			},
		},
	}
}

func (d *dataSourceCloudAutonomousVmCluster) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)
	var data cloudAutonomousVmClusterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := odb.GetCloudAutonomousVmClusterInput{
		CloudAutonomousVmClusterId: data.CloudAutonomousVmClusterId.ValueStringPointer(),
	}

	out, err := conn.GetCloudAutonomousVmCluster(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameCloudAutonomousVmCluster, data.CloudAutonomousVmClusterId.ValueString(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out.CloudAutonomousVmCluster, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type cloudAutonomousVmClusterDataSourceModel struct {
	framework.WithRegionModel
	CloudAutonomousVmClusterArn                  types.String                                                                              `tfsdk:"arn"`
	CloudAutonomousVmClusterId                   types.String                                                                              `tfsdk:"id"`
	CloudExadataInfrastructureId                 types.String                                                                              `tfsdk:"cloud_exadata_infrastructure_id"`
	CloudExadataInfrastructureArn                types.String                                                                              `tfsdk:"cloud_exadata_infrastructure_arn" autoflex:"-"`
	AutonomousDataStoragePercentage              types.Float32                                                                             `tfsdk:"autonomous_data_storage_percentage"`
	AutonomousDataStorageSizeInTBs               types.Float64                                                                             `tfsdk:"autonomous_data_storage_size_in_tbs"`
	AvailableAutonomousDataStorageSizeInTBs      types.Float64                                                                             `tfsdk:"available_autonomous_data_storage_size_in_tbs"`
	AvailableContainerDatabases                  types.Int32                                                                               `tfsdk:"available_container_databases"`
	AvailableCpus                                types.Float32                                                                             `tfsdk:"available_cpus"`
	ComputeModel                                 fwtypes.StringEnum[odbtypes.ComputeModel]                                                 `tfsdk:"compute_model"`
	CpuCoreCount                                 types.Int32                                                                               `tfsdk:"cpu_core_count"`
	CpuCoreCountPerNode                          types.Int32                                                                               `tfsdk:"cpu_core_count_per_node"`
	CpuPercentage                                types.Float32                                                                             `tfsdk:"cpu_percentage"`
	CreatedAt                                    timetypes.RFC3339                                                                         `tfsdk:"created_at" `
	DataStorageSizeInGBs                         types.Float64                                                                             `tfsdk:"data_storage_size_in_gbs"`
	DataStorageSizeInTBs                         types.Float64                                                                             `tfsdk:"data_storage_size_in_tbs"`
	DbNodeStorageSizeInGBs                       types.Int32                                                                               `tfsdk:"odb_node_storage_size_in_gbs"`
	DbServers                                    fwtypes.SetValueOf[types.String]                                                          `tfsdk:"db_servers"`
	Description                                  types.String                                                                              `tfsdk:"description"`
	DisplayName                                  types.String                                                                              `tfsdk:"display_name"`
	Domain                                       types.String                                                                              `tfsdk:"domain"`
	ExadataStorageInTBsLowestScaledValue         types.Float64                                                                             `tfsdk:"exadata_storage_in_tbs_lowest_scaled_value"`
	Hostname                                     types.String                                                                              `tfsdk:"hostname"`
	IsMtlsEnabledVmCluster                       types.Bool                                                                                `tfsdk:"is_mtls_enabled_vm_cluster"`
	LicenseModel                                 fwtypes.StringEnum[odbtypes.LicenseModel]                                                 `tfsdk:"license_model"`
	MaxAcdsLowestScaledValue                     types.Int32                                                                               `tfsdk:"max_acds_lowest_scaled_value"`
	MemoryPerOracleComputeUnitInGBs              types.Int32                                                                               `tfsdk:"memory_per_oracle_compute_unit_in_gbs"`
	MemorySizeInGBs                              types.Int32                                                                               `tfsdk:"memory_size_in_gbs"`
	NodeCount                                    types.Int32                                                                               `tfsdk:"node_count"`
	NonProvisionableAutonomousContainerDatabases types.Int32                                                                               `tfsdk:"non_provisionable_autonomous_container_databases"`
	OciResourceAnchorName                        types.String                                                                              `tfsdk:"oci_resource_anchor_name"`
	OciUrl                                       types.String                                                                              `tfsdk:"oci_url"`
	Ocid                                         types.String                                                                              `tfsdk:"ocid"`
	OdbNetworkId                                 types.String                                                                              `tfsdk:"odb_network_id"`
	OdbNetworkArn                                types.String                                                                              `tfsdk:"odb_network_arn" autoflex:"-"`
	PercentProgress                              types.Float32                                                                             `tfsdk:"percent_progress"`
	ProvisionableAutonomousContainerDatabases    types.Int32                                                                               `tfsdk:"provisionable_autonomous_container_databases"`
	ProvisionedAutonomousContainerDatabases      types.Int32                                                                               `tfsdk:"provisioned_autonomous_container_databases"`
	ProvisionedCpus                              types.Float32                                                                             `tfsdk:"provisioned_cpus"`
	ReclaimableCpus                              types.Float32                                                                             `tfsdk:"reclaimable_cpus"`
	ReservedCpus                                 types.Float32                                                                             `tfsdk:"reserved_cpus"`
	ScanListenerPortNonTls                       types.Int32                                                                               `tfsdk:"scan_listener_port_non_tls"`
	ScanListenerPortTls                          types.Int32                                                                               `tfsdk:"scan_listener_port_tls"`
	Shape                                        types.String                                                                              `tfsdk:"shape"`
	Status                                       fwtypes.StringEnum[odbtypes.ResourceStatus]                                               `tfsdk:"status"`
	StatusReason                                 types.String                                                                              `tfsdk:"status_reason"`
	TimeDatabaseSslCertificateExpires            timetypes.RFC3339                                                                         `tfsdk:"time_database_ssl_certificate_expires"`
	TimeOrdsCertificateExpires                   timetypes.RFC3339                                                                         `tfsdk:"time_ords_certificate_expires" `
	TimeZone                                     types.String                                                                              `tfsdk:"time_zone"`
	TotalContainerDatabases                      types.Int32                                                                               `tfsdk:"total_container_databases"`
	MaintenanceWindow                            fwtypes.ListNestedObjectValueOf[cloudAutonomousVmClusterMaintenanceWindowDataSourceModel] `tfsdk:"maintenance_window" `
	Tags                                         tftags.Map                                                                                `tfsdk:"tags"`
}
type cloudAutonomousVmClusterMaintenanceWindowDataSourceModel struct {
	DaysOfWeek      fwtypes.SetNestedObjectValueOf[dayWeekNameAutonomousVmClusterMaintenanceWindowDataSourceModel] `tfsdk:"days_of_week"`
	HoursOfDay      fwtypes.SetValueOf[types.Int64]                                                                `tfsdk:"hours_of_day"`
	LeadTimeInWeeks types.Int32                                                                                    `tfsdk:"lead_time_in_weeks"`
	Months          fwtypes.SetNestedObjectValueOf[monthNameAutonomousVmClusterMaintenanceWindowDataSourceModel]   `tfsdk:"months"`
	Preference      fwtypes.StringEnum[odbtypes.PreferenceType]                                                    `tfsdk:"preference"`
	WeeksOfMonth    fwtypes.SetValueOf[types.Int64]                                                                `tfsdk:"weeks_of_month"`
}
type dayWeekNameAutonomousVmClusterMaintenanceWindowDataSourceModel struct {
	Name fwtypes.StringEnum[odbtypes.DayOfWeekName] `tfsdk:"name"`
}

type monthNameAutonomousVmClusterMaintenanceWindowDataSourceModel struct {
	Name fwtypes.StringEnum[odbtypes.MonthName] `tfsdk:"name"`
}
