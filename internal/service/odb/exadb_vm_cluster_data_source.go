// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"

	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_odb_exadb_vm_cluster", name="ExaDB VM Cluster")
// @Tags(identifierAttribute="arn")
func newExaDBVMClusterDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &exaDBVMClusterDataSource{}, nil
}

type exaDBVMClusterDataSource struct {
	framework.DataSourceWithModel[exaDBVMClusterDataSourceModel]
}

func (d *exaDBVMClusterDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	dataCollectionOptionsAttribute := framework.DataSourceComputedListOfObjectAttribute[exaDBVMClusterDataCollectionOptionsModel](ctx)
	dataCollectionOptionsAttribute.Description = "Diagnostic collection preferences for the ExaDB VM Cluster."
	iamRolesAttribute := framework.DataSourceComputedListOfObjectAttribute[exaDBVMClusterIAMRoleModel](ctx)
	iamRolesAttribute.Description = "IAM service roles associated with the ExaDB VM Cluster."
	iormConfigCacheAttribute := framework.DataSourceComputedListOfObjectAttribute[exaDBVMClusterIORMConfigModel](ctx)
	iormConfigCacheAttribute.Description = "IORM configuration cache details for the ExaDB VM Cluster."
	snapshotFileSystemStorageAttribute := framework.DataSourceComputedListOfObjectAttribute[exaDBVMClusterStorageDetailsModel](ctx)
	snapshotFileSystemStorageAttribute.Description = "Snapshot file system storage details for the ExaDB VM Cluster."
	totalFileSystemStorageAttribute := framework.DataSourceComputedListOfObjectAttribute[exaDBVMClusterStorageDetailsModel](ctx)
	totalFileSystemStorageAttribute.Description = "Total file system storage details for the ExaDB VM Cluster."
	vmFileSystemStorageAttribute := framework.DataSourceComputedListOfObjectAttribute[exaDBVMClusterStorageDetailsModel](ctx)
	vmFileSystemStorageAttribute.Description = "VM file system storage details for the ExaDB VM Cluster."

	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrClusterName: schema.StringAttribute{
				Computed:    true,
				Description: "Name of the Grid Infrastructure cluster.",
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "Date and time when the ExaDB VM Cluster was created.",
			},
			"data_collection_options": dataCollectionOptionsAttribute,
			names.AttrDisplayName: schema.StringAttribute{
				Computed:    true,
				Description: "User-friendly name for the ExaDB VM Cluster.",
			},
			names.AttrDomain: schema.StringAttribute{
				Computed:    true,
				Description: "Domain of the ExaDB VM Cluster.",
			},
			"enabled_ecpu_count": schema.Int32Attribute{
				Computed:    true,
				Description: "Number of ECPUs enabled for the ExaDB VM Cluster.",
			},
			"exascale_db_storage_vault_arn": schema.StringAttribute{
				Computed:    true,
				Description: "ARN of the Exascale DB Storage Vault associated with the ExaDB VM Cluster.",
			},
			"exascale_db_storage_vault_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the Exascale DB Storage Vault associated with the ExaDB VM Cluster.",
			},
			"gi_version": schema.StringAttribute{
				Computed:    true,
				Description: "Oracle Grid Infrastructure software version for the ExaDB VM Cluster.",
			},
			"grid_image_id": schema.StringAttribute{
				Computed:    true,
				Description: "Grid Infrastructure software image ID for the ExaDB VM Cluster.",
			},
			"grid_image_type": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[odbtypes.GridImageType](),
				Computed:    true,
				Description: "Type of Grid Infrastructure image used by the ExaDB VM Cluster.",
			},
			"hostname": schema.StringAttribute{
				Computed:    true,
				Description: "Host name for the ExaDB VM Cluster.",
			},
			"iam_roles": iamRolesAttribute,
			names.AttrID: schema.StringAttribute{
				Required:    true,
				Description: "Unique identifier of the ExaDB VM Cluster.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 2048),
				},
			},
			"iorm_config_cache": iormConfigCacheAttribute,
			"last_update_history_entry_id": schema.StringAttribute{
				Computed:    true,
				Description: "OCID of the last maintenance update history entry.",
			},
			"license_model": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[odbtypes.LicenseModel](),
				Computed:    true,
				Description: "Oracle license model applied to the ExaDB VM Cluster.",
			},
			"listener_port": schema.Int32Attribute{
				Computed:    true,
				Description: "Listener port configured for the ExaDB VM Cluster.",
			},
			"memory_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "Amount of memory allocated to the ExaDB VM Cluster, in GB.",
			},
			"node_count": schema.Int32Attribute{
				Computed:    true,
				Description: "Number of nodes in the ExaDB VM Cluster.",
			},
			"ocid": schema.StringAttribute{
				Computed:    true,
				Description: "OCID of the ExaDB VM Cluster.",
			},
			"oci_resource_anchor_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the OCI resource anchor for the ExaDB VM Cluster.",
			},
			"oci_url": schema.StringAttribute{
				Computed:    true,
				Description: "HTTPS URL of the ExaDB VM Cluster in OCI.",
			},
			"odb_network_arn": schema.StringAttribute{
				Computed:    true,
				Description: "ARN of the ODB network associated with the ExaDB VM Cluster.",
			},
			"odb_network_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the ODB network associated with the ExaDB VM Cluster.",
			},
			"percent_progress": schema.Float32Attribute{
				Computed:    true,
				Description: "Progress of the current operation, expressed as a percentage.",
			},
			"scan_dns_name": schema.StringAttribute{
				Computed:    true,
				Description: "FQDN of the SCAN IP addresses associated with the ExaDB VM Cluster.",
			},
			"scan_dns_record_id": schema.StringAttribute{
				Computed:    true,
				Description: "OCID of the DNS record for the SCAN IP addresses.",
			},
			"scan_ip_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				Description: "OCIDs of the SCAN IP addresses associated with the ExaDB VM Cluster.",
			},
			"scan_listener_port_tcp": schema.Int32Attribute{
				Computed:    true,
				Description: "Port for TCP connections to the SCAN listener.",
			},
			"scan_listener_port_tcp_ssl": schema.Int32Attribute{
				Computed:    true,
				Description: "Port for SSL/TCP connections to the SCAN listener.",
			},
			"shape": schema.StringAttribute{
				Computed:    true,
				Description: "Shape of the ExaDB VM Cluster.",
			},
			"shape_attribute": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[odbtypes.ShapeAttribute](),
				Computed:    true,
				Description: "Shape attribute for the ExaDB VM Cluster.",
			},
			"snapshot_file_system_storage": snapshotFileSystemStorageAttribute,
			"ssh_public_keys": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				Description: "Public keys used for SSH access to the ExaDB VM Cluster.",
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[odbtypes.ResourceStatus](),
				Computed:    true,
				Description: "Current status of the ExaDB VM Cluster.",
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed:    true,
				Description: "Additional information about the current ExaDB VM Cluster status.",
			},
			"system_version": schema.StringAttribute{
				Computed:    true,
				Description: "Operating system version of the image for the ExaDB VM Cluster.",
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"time_zone": schema.StringAttribute{
				Computed:    true,
				Description: "Time zone for the ExaDB VM Cluster.",
			},
			"total_ecpu_count": schema.Int32Attribute{
				Computed:    true,
				Description: "Total number of ECPUs for the ExaDB VM Cluster.",
			},
			"total_file_system_storage": totalFileSystemStorageAttribute,
			"vip_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				Description: "OCIDs of the virtual IP addresses associated with the ExaDB VM Cluster.",
			},
			"vm_file_system_storage": vmFileSystemStorageAttribute,
			"vm_file_system_storage_total_size_in_gbs": schema.Int32Attribute{
				Computed:    true,
				Description: "Total amount of VM file system storage for the ExaDB VM Cluster, in GB.",
			},
		},
	}
}

func (d *exaDBVMClusterDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)

	var data exaDBVMClusterDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findExaDBVMClusterByID(ctx, conn, data.ExaDBVMClusterID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ExaDBVMClusterID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, flex.Flatten(ctx, output, &data, flex.WithFieldNamePrefix("ExadbVmCluster")))
	if response.Diagnostics.HasError() {
		return
	}
	if output.VmFileSystemStorage != nil {
		data.VMFileSystemStorageTotalSizeInGBs = types.Int32PointerValue(output.VmFileSystemStorage.TotalSizeInGBs)
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type exaDBVMClusterDataSourceModel struct {
	framework.WithRegionModel
	ExaDBVMClusterARN                 types.String                                                              `tfsdk:"arn"`
	ClusterName                       types.String                                                              `tfsdk:"cluster_name"`
	CreatedAt                         timetypes.RFC3339                                                         `tfsdk:"created_at"`
	DataCollectionOptions             fwtypes.ListNestedObjectValueOf[exaDBVMClusterDataCollectionOptionsModel] `tfsdk:"data_collection_options"`
	DisplayName                       types.String                                                              `tfsdk:"display_name"`
	Domain                            types.String                                                              `tfsdk:"domain"`
	EnabledECPUCount                  types.Int32                                                               `tfsdk:"enabled_ecpu_count"`
	ExascaleDBStorageVaultARN         types.String                                                              `tfsdk:"exascale_db_storage_vault_arn"`
	ExascaleDBStorageVaultID          types.String                                                              `tfsdk:"exascale_db_storage_vault_id"`
	ExaDBVMClusterID                  types.String                                                              `tfsdk:"id"`
	GIVersion                         types.String                                                              `tfsdk:"gi_version"`
	GridImageID                       types.String                                                              `tfsdk:"grid_image_id"`
	GridImageType                     fwtypes.StringEnum[odbtypes.GridImageType]                                `tfsdk:"grid_image_type"`
	Hostname                          types.String                                                              `tfsdk:"hostname"`
	IAMRoles                          fwtypes.ListNestedObjectValueOf[exaDBVMClusterIAMRoleModel]               `tfsdk:"iam_roles"`
	IORMConfigCache                   fwtypes.ListNestedObjectValueOf[exaDBVMClusterIORMConfigModel]            `tfsdk:"iorm_config_cache"`
	LastUpdateHistoryEntryID          types.String                                                              `tfsdk:"last_update_history_entry_id"`
	LicenseModel                      fwtypes.StringEnum[odbtypes.LicenseModel]                                 `tfsdk:"license_model"`
	ListenerPort                      types.Int32                                                               `tfsdk:"listener_port"`
	MemorySizeInGBs                   types.Int32                                                               `tfsdk:"memory_size_in_gbs"`
	NodeCount                         types.Int32                                                               `tfsdk:"node_count"`
	OCID                              types.String                                                              `tfsdk:"ocid"`
	OCIResourceAnchorName             types.String                                                              `tfsdk:"oci_resource_anchor_name"`
	OCIURL                            types.String                                                              `tfsdk:"oci_url"`
	ODBNetworkARN                     types.String                                                              `tfsdk:"odb_network_arn"`
	ODBNetworkID                      types.String                                                              `tfsdk:"odb_network_id"`
	PercentProgress                   types.Float32                                                             `tfsdk:"percent_progress"`
	ScanDNSName                       types.String                                                              `tfsdk:"scan_dns_name"`
	ScanDNSRecordID                   types.String                                                              `tfsdk:"scan_dns_record_id"`
	ScanIPIDs                         fwtypes.ListValueOf[types.String]                                         `tfsdk:"scan_ip_ids"`
	ScanListenerPortTCP               types.Int32                                                               `tfsdk:"scan_listener_port_tcp"`
	ScanListenerPortTCPSSL            types.Int32                                                               `tfsdk:"scan_listener_port_tcp_ssl"`
	Shape                             types.String                                                              `tfsdk:"shape"`
	ShapeAttribute                    fwtypes.StringEnum[odbtypes.ShapeAttribute]                               `tfsdk:"shape_attribute"`
	SnapshotFileSystemStorage         fwtypes.ListNestedObjectValueOf[exaDBVMClusterStorageDetailsModel]        `tfsdk:"snapshot_file_system_storage"`
	SSHPublicKeys                     fwtypes.SetValueOf[types.String]                                          `tfsdk:"ssh_public_keys"`
	Status                            fwtypes.StringEnum[odbtypes.ResourceStatus]                               `tfsdk:"status"`
	StatusReason                      types.String                                                              `tfsdk:"status_reason"`
	SystemVersion                     types.String                                                              `tfsdk:"system_version"`
	Tags                              tftags.Map                                                                `tfsdk:"tags"`
	TimeZone                          types.String                                                              `tfsdk:"time_zone"`
	TotalECPUCount                    types.Int32                                                               `tfsdk:"total_ecpu_count"`
	TotalFileSystemStorage            fwtypes.ListNestedObjectValueOf[exaDBVMClusterStorageDetailsModel]        `tfsdk:"total_file_system_storage"`
	VIPIDs                            fwtypes.ListValueOf[types.String]                                         `tfsdk:"vip_ids"`
	VMFileSystemStorage               fwtypes.ListNestedObjectValueOf[exaDBVMClusterStorageDetailsModel]        `tfsdk:"vm_file_system_storage"`
	VMFileSystemStorageTotalSizeInGBs types.Int32                                                               `tfsdk:"vm_file_system_storage_total_size_in_gbs"`
}
