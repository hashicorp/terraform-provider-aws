// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package odb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_odb_autonomous_database", name="Autonomous Database")
// @Tags(identifierAttribute="arn")
func newDataSourceAutonomousDatabase(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAutonomousDatabase{}, nil
}

const DSNameAutonomousDatabase = "Autonomous Database Data Source"

type dataSourceAutonomousDatabase struct {
	framework.DataSourceWithModel[autonomousDatabaseDataSourceModel]
}

func (d *dataSourceAutonomousDatabase) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Required:    true,
				Description: "Unique identifier of the Autonomous Database.",
			},
			names.AttrARN:                          framework.ARNAttributeComputedOnly(),
			"actual_used_data_storage_size_in_tbs": schema.Float64Attribute{Computed: true, Description: "Actual amount of data storage currently in use, in TB."},
			"allocated_storage_size_in_tbs":        schema.Float64Attribute{Computed: true, Description: "Amount of storage currently allocated, in TB."},
			"allowlisted_ips": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: tftypes.StringType,
				Description: "IP addresses allowed to access the Autonomous Database.",
			},
			"auto_refresh_frequency_in_seconds":    schema.Int32Attribute{Computed: true, Description: "Automatic refresh frequency, in seconds."},
			"auto_refresh_point_lag_in_seconds":    schema.Int32Attribute{Computed: true, Description: "Refresh lag from the source, in seconds."},
			"autonomous_maintenance_schedule_type": schema.StringAttribute{CustomType: fwtypes.StringEnumType[types.AutonomousMaintenanceScheduleType](), Computed: true, Description: "Maintenance schedule type."},
			names.AttrAvailabilityZone:             schema.StringAttribute{Computed: true, Description: "Availability Zone of the Autonomous Database."},
			"availability_zone_id":                 schema.StringAttribute{Computed: true, Description: "Availability Zone ID of the Autonomous Database."},
			"available_upgrade_versions": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: tftypes.StringType,
				Description: "Oracle Database versions available for upgrade.",
			},
			"backup_retention_period_in_days":             schema.Int32Attribute{Computed: true, Description: "Automatic backup retention period, in days."},
			"byol_compute_count_limit":                    schema.Float64Attribute{Computed: true, Description: "Maximum BYOL compute capacity."},
			"character_set":                               schema.StringAttribute{Computed: true, Description: "Database character set."},
			"compute_count":                               schema.Float64Attribute{Computed: true, Description: "Compute capacity in ECPUs or OCPUs."},
			"compute_model":                               schema.StringAttribute{CustomType: fwtypes.StringEnumType[types.ComputeModel](), Computed: true, Description: "Compute model."},
			"cpu_core_count":                              schema.Int32Attribute{Computed: true, Description: "Allocated CPU core count."},
			names.AttrCreatedAt:                           schema.StringAttribute{CustomType: timetypes.RFC3339Type{}, Computed: true, Description: "Creation date and time."},
			"data_storage_size_in_gbs":                    schema.Int32Attribute{Computed: true, Description: "Data volume size in GB."},
			"data_storage_size_in_tbs":                    schema.Float64Attribute{Computed: true, Description: "Data volume size in TB."},
			"database_edition":                            schema.StringAttribute{CustomType: fwtypes.StringEnumType[types.DatabaseEdition](), Computed: true, Description: "Oracle Database edition."},
			"database_type":                               schema.StringAttribute{CustomType: fwtypes.StringEnumType[types.DatabaseType](), Computed: true, Description: "Autonomous Database type."},
			"db_name":                                     schema.StringAttribute{Computed: true, Description: "Database name."},
			"db_version":                                  schema.StringAttribute{Computed: true, Description: "Oracle Database software version."},
			"db_workload":                                 schema.StringAttribute{CustomType: fwtypes.StringEnumType[types.DbWorkload](), Computed: true, Description: "Database workload."},
			names.AttrDisplayName:                         schema.StringAttribute{Computed: true, Description: "User-friendly database name."},
			"encryption_key_provider":                     schema.StringAttribute{Computed: true, Description: "Encryption key provider."},
			"is_auto_scaling_enabled":                     schema.BoolAttribute{Computed: true, Description: "Whether automatic compute scaling is enabled."},
			"is_auto_scaling_for_storage_enabled":         schema.BoolAttribute{Computed: true, Description: "Whether automatic storage scaling is enabled."},
			"is_backup_retention_locked":                  schema.BoolAttribute{Computed: true, Description: "Whether backup retention is locked."},
			"is_local_data_guard_enabled":                 schema.BoolAttribute{Computed: true, Description: "Whether local Data Guard is enabled."},
			"is_mtls_connection_required":                 schema.BoolAttribute{Computed: true, Description: "Whether mTLS is required."},
			"is_refreshable_clone":                        schema.BoolAttribute{Computed: true, Description: "Whether the database is a refreshable clone."},
			names.AttrKMSKeyID:                            schema.StringAttribute{Computed: true, Description: "ARN of the AWS KMS encryption key."},
			"license_model":                               schema.StringAttribute{CustomType: fwtypes.StringEnumType[types.LicenseModel](), Computed: true, Description: "Oracle license model."},
			"local_adg_auto_failover_max_data_loss_limit": schema.Int32Attribute{Computed: true, Description: "Maximum automatic-failover data-loss limit, in seconds."},
			"ncharacter_set":                              schema.StringAttribute{Computed: true, Description: "National character set."},
			"oci_resource_anchor_name":                    schema.StringAttribute{Computed: true, Description: "OCI resource anchor name."},
			"oci_url":                                     schema.StringAttribute{Computed: true, Description: "OCI console URL."},
			"ocid":                                        schema.StringAttribute{Computed: true, Description: "Oracle Cloud Identifier."},
			"odb_network_arn":                             schema.StringAttribute{Computed: true, Description: "ARN of the associated ODB network."},
			"odb_network_id":                              schema.StringAttribute{Computed: true, Description: "ID of the associated ODB network."},
			"open_mode":                                   schema.StringAttribute{CustomType: fwtypes.StringEnumType[types.OpenMode](), Computed: true, Description: "Database open mode."},
			"percent_progress":                            schema.Float32Attribute{Computed: true, Description: "Progress of the current operation."},
			"permission_level":                            schema.StringAttribute{CustomType: fwtypes.StringEnumType[types.PermissionLevel](), Computed: true, Description: "Database permission level."},
			"private_endpoint":                            schema.StringAttribute{Computed: true, Description: "Private endpoint."},
			"private_endpoint_ip":                         schema.StringAttribute{Computed: true, Description: "Private endpoint IP address."},
			"private_endpoint_label":                      schema.StringAttribute{Computed: true, Description: "Private endpoint label."},
			"refreshable_mode":                            schema.StringAttribute{CustomType: fwtypes.StringEnumType[types.RefreshableMode](), Computed: true, Description: "Refreshable clone mode."},
			"resource_pool_leader_id":                     schema.StringAttribute{Computed: true, Description: "Resource-pool leader database ID."},
			"service_console_url":                         schema.StringAttribute{Computed: true, Description: "Oracle service console URL."},
			"source_id":                                   schema.StringAttribute{Computed: true, Description: "Source database or backup ID."},
			"sql_web_developer_url":                       schema.StringAttribute{Computed: true, Description: "Oracle SQL Developer Web URL."},
			"standby_allowlisted_ips": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: tftypes.StringType,
				Description: "IP addresses allowed to access the standby database.",
			},
			"standby_allowlisted_ips_source":   schema.StringAttribute{CustomType: fwtypes.StringEnumType[types.StandbyAllowlistedIpsSource](), Computed: true, Description: "Source of standby allowlisted IPs."},
			names.AttrStatus:                   schema.StringAttribute{CustomType: fwtypes.StringEnumType[types.AutonomousDatabaseResourceStatus](), Computed: true, Description: "Current database status."},
			names.AttrStatusReason:             schema.StringAttribute{Computed: true, Description: "Additional status information."},
			"time_of_auto_refresh_start":       schema.StringAttribute{CustomType: timetypes.RFC3339Type{}, Computed: true, Description: "Automatic refresh start time."},
			names.AttrTags:                     tftags.TagsAttributeComputedOnly(),
			"customer_contacts_to_send_to_oci": computedCustomerContactsAttribute(ctx),
			"db_tools_details":                 computedDatabaseToolsAttribute(ctx),
			"long_term_backup_schedule":        computedLongTermBackupScheduleAttribute(ctx),
			"resource_pool_summary":            computedResourcePoolSummaryAttribute(ctx),
			"scheduled_operations":             computedScheduledOperationsAttribute(ctx),
		},
	}
}

func computedCustomerContactsAttribute(ctx context.Context) schema.ListAttribute {
	attribute := framework.DataSourceComputedListOfObjectAttribute[autonomousDatabaseCustomerContactModel](ctx)
	attribute.Description = "Customer contacts that receive operational notifications from OCI."
	return attribute
}

func computedDatabaseToolsAttribute(ctx context.Context) schema.ListAttribute {
	attribute := framework.DataSourceComputedListOfObjectAttribute[autonomousDatabaseToolModel](ctx)
	attribute.Description = "Database management tools enabled for the Autonomous Database."
	return attribute
}

func computedLongTermBackupScheduleAttribute(ctx context.Context) schema.ListAttribute {
	attribute := framework.DataSourceComputedListOfObjectAttribute[autonomousDatabaseLongTermBackupScheduleModel](ctx)
	attribute.Description = "Long-term backup schedule."
	return attribute
}

func computedResourcePoolSummaryAttribute(ctx context.Context) schema.ListAttribute {
	attribute := framework.DataSourceComputedListOfObjectAttribute[autonomousDatabaseResourcePoolSummaryModel](ctx)
	attribute.Description = "Resource pool configuration."
	return attribute
}

func computedScheduledOperationsAttribute(ctx context.Context) schema.ListAttribute {
	attribute := framework.DataSourceComputedListOfObjectAttribute[autonomousDatabaseScheduledOperationModel](ctx)
	attribute.Description = "Scheduled database start and stop times."
	return attribute
}

func (d *dataSourceAutonomousDatabase) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)

	var data autonomousDatabaseDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAutonomousDatabaseByID(ctx, conn, data.AutonomousDatabaseID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameAutonomousDatabase, data.AutonomousDatabaseID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	data.ByolComputeCountLimit = flattenAutonomousDatabaseByolComputeCountLimit(out.ByolComputeCountLimit)
	resp.Diagnostics.Append(flex.Flatten(ctx, out.CustomerContacts, &data.CustomerContactsToSendToOCI)...)
	resp.Diagnostics.Append(flattenAutonomousDatabaseScheduledOperations(ctx, out.ScheduledOperations, &data.ScheduledOperations)...)
	if out.EncryptionSummary != nil {
		data.EncryptionKeyProvider = tftypes.StringValue(string(out.EncryptionSummary.EncryptionKeyProvider))
		if configuration, ok := out.EncryptionSummary.EncryptionKeyConfiguration.(*types.EncryptionKeyConfigurationMemberAwsEncryptionKey); ok {
			data.KMSKeyID = tftypes.StringPointerValue(configuration.Value.KmsKeyId)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type autonomousDatabaseDataSourceModel struct {
	framework.WithRegionModel
	ActualUsedDataStorageSizeInTBs       tftypes.Float64                                                                `tfsdk:"actual_used_data_storage_size_in_tbs"`
	AllocatedStorageSizeInTBs            tftypes.Float64                                                                `tfsdk:"allocated_storage_size_in_tbs"`
	AllowlistedIps                       fwtypes.ListValueOf[tftypes.String]                                            `tfsdk:"allowlisted_ips"`
	AutoRefreshFrequencyInSeconds        tftypes.Int32                                                                  `tfsdk:"auto_refresh_frequency_in_seconds"`
	AutoRefreshPointLagInSeconds         tftypes.Int32                                                                  `tfsdk:"auto_refresh_point_lag_in_seconds"`
	AutonomousDatabaseARN                tftypes.String                                                                 `tfsdk:"arn"`
	AutonomousDatabaseID                 tftypes.String                                                                 `tfsdk:"id"`
	AutonomousMaintenanceScheduleType    fwtypes.StringEnum[types.AutonomousMaintenanceScheduleType]                    `tfsdk:"autonomous_maintenance_schedule_type"`
	AvailabilityZone                     tftypes.String                                                                 `tfsdk:"availability_zone"`
	AvailabilityZoneID                   tftypes.String                                                                 `tfsdk:"availability_zone_id"`
	AvailableUpgradeVersions             fwtypes.ListValueOf[tftypes.String]                                            `tfsdk:"available_upgrade_versions"`
	BackupRetentionPeriodInDays          tftypes.Int32                                                                  `tfsdk:"backup_retention_period_in_days"`
	ByolComputeCountLimit                tftypes.Float64                                                                `tfsdk:"byol_compute_count_limit" autoflex:",noflatten"`
	CharacterSet                         tftypes.String                                                                 `tfsdk:"character_set"`
	ComputeCount                         tftypes.Float64                                                                `tfsdk:"compute_count"`
	ComputeModel                         fwtypes.StringEnum[types.ComputeModel]                                         `tfsdk:"compute_model"`
	CpuCoreCount                         tftypes.Int32                                                                  `tfsdk:"cpu_core_count"`
	CreatedAt                            timetypes.RFC3339                                                              `tfsdk:"created_at"`
	CustomerContactsToSendToOCI          fwtypes.ListNestedObjectValueOf[autonomousDatabaseCustomerContactModel]        `tfsdk:"customer_contacts_to_send_to_oci" autoflex:"-"`
	DataStorageSizeInGBs                 tftypes.Int32                                                                  `tfsdk:"data_storage_size_in_gbs"`
	DataStorageSizeInTBs                 tftypes.Float64                                                                `tfsdk:"data_storage_size_in_tbs"`
	DatabaseEdition                      fwtypes.StringEnum[types.DatabaseEdition]                                      `tfsdk:"database_edition"`
	DatabaseType                         fwtypes.StringEnum[types.DatabaseType]                                         `tfsdk:"database_type"`
	DbName                               tftypes.String                                                                 `tfsdk:"db_name"`
	DbToolsDetails                       fwtypes.ListNestedObjectValueOf[autonomousDatabaseToolModel]                   `tfsdk:"db_tools_details"`
	DbVersion                            tftypes.String                                                                 `tfsdk:"db_version"`
	DbWorkload                           fwtypes.StringEnum[types.DbWorkload]                                           `tfsdk:"db_workload"`
	DisplayName                          tftypes.String                                                                 `tfsdk:"display_name"`
	EncryptionKeyProvider                tftypes.String                                                                 `tfsdk:"encryption_key_provider" autoflex:"-"`
	IsAutoScalingEnabled                 tftypes.Bool                                                                   `tfsdk:"is_auto_scaling_enabled"`
	IsAutoScalingForStorageEnabled       tftypes.Bool                                                                   `tfsdk:"is_auto_scaling_for_storage_enabled"`
	IsBackupRetentionLocked              tftypes.Bool                                                                   `tfsdk:"is_backup_retention_locked"`
	IsLocalDataGuardEnabled              tftypes.Bool                                                                   `tfsdk:"is_local_data_guard_enabled"`
	IsMtlsConnectionRequired             tftypes.Bool                                                                   `tfsdk:"is_mtls_connection_required"`
	IsRefreshableClone                   tftypes.Bool                                                                   `tfsdk:"is_refreshable_clone"`
	KMSKeyID                             tftypes.String                                                                 `tfsdk:"kms_key_id" autoflex:"-"`
	LicenseModel                         fwtypes.StringEnum[types.LicenseModel]                                         `tfsdk:"license_model"`
	LocalAdgAutoFailoverMaxDataLossLimit tftypes.Int32                                                                  `tfsdk:"local_adg_auto_failover_max_data_loss_limit"`
	LongTermBackupSchedule               fwtypes.ListNestedObjectValueOf[autonomousDatabaseLongTermBackupScheduleModel] `tfsdk:"long_term_backup_schedule"`
	NcharacterSet                        tftypes.String                                                                 `tfsdk:"ncharacter_set"`
	OciResourceAnchorName                tftypes.String                                                                 `tfsdk:"oci_resource_anchor_name"`
	OciUrl                               tftypes.String                                                                 `tfsdk:"oci_url"`
	Ocid                                 tftypes.String                                                                 `tfsdk:"ocid"`
	OdbNetworkArn                        tftypes.String                                                                 `tfsdk:"odb_network_arn"`
	OdbNetworkId                         tftypes.String                                                                 `tfsdk:"odb_network_id"`
	OpenMode                             fwtypes.StringEnum[types.OpenMode]                                             `tfsdk:"open_mode"`
	PercentProgress                      tftypes.Float32                                                                `tfsdk:"percent_progress"`
	PermissionLevel                      fwtypes.StringEnum[types.PermissionLevel]                                      `tfsdk:"permission_level"`
	PrivateEndpoint                      tftypes.String                                                                 `tfsdk:"private_endpoint"`
	PrivateEndpointIp                    tftypes.String                                                                 `tfsdk:"private_endpoint_ip"`
	PrivateEndpointLabel                 tftypes.String                                                                 `tfsdk:"private_endpoint_label"`
	RefreshableMode                      fwtypes.StringEnum[types.RefreshableMode]                                      `tfsdk:"refreshable_mode"`
	ResourcePoolLeaderId                 tftypes.String                                                                 `tfsdk:"resource_pool_leader_id"`
	ResourcePoolSummary                  fwtypes.ListNestedObjectValueOf[autonomousDatabaseResourcePoolSummaryModel]    `tfsdk:"resource_pool_summary"`
	ScheduledOperations                  fwtypes.ListNestedObjectValueOf[autonomousDatabaseScheduledOperationModel]     `tfsdk:"scheduled_operations" autoflex:"-"`
	ServiceConsoleUrl                    tftypes.String                                                                 `tfsdk:"service_console_url"`
	SourceId                             tftypes.String                                                                 `tfsdk:"source_id"`
	SqlWebDeveloperUrl                   tftypes.String                                                                 `tfsdk:"sql_web_developer_url"`
	StandbyAllowlistedIps                fwtypes.ListValueOf[tftypes.String]                                            `tfsdk:"standby_allowlisted_ips"`
	StandbyAllowlistedIpsSource          fwtypes.StringEnum[types.StandbyAllowlistedIpsSource]                          `tfsdk:"standby_allowlisted_ips_source"`
	Status                               fwtypes.StringEnum[types.AutonomousDatabaseResourceStatus]                     `tfsdk:"status"`
	StatusReason                         tftypes.String                                                                 `tfsdk:"status_reason"`
	Tags                                 tftags.Map                                                                     `tfsdk:"tags"`
	TimeOfAutoRefreshStart               timetypes.RFC3339                                                              `tfsdk:"time_of_auto_refresh_start"`
}
