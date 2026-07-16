// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package odb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_odb_autonomous_database", name="Autonomous Database")
// @Tags(identifierAttribute="arn")
// @Testing(importIgnore="admin_password_wo;admin_password_wo_version;source;source_configuration;transportable_tablespace")
func newResourceAutonomousDatabase(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAutonomousDatabase{}
	r.SetDefaultCreateTimeout(24 * time.Hour)
	r.SetDefaultUpdateTimeout(24 * time.Hour)
	r.SetDefaultDeleteTimeout(24 * time.Hour)

	return r, nil
}

const ResNameAutonomousDatabase = "Autonomous Database"

type resourceAutonomousDatabase struct {
	framework.ResourceWithModel[autonomousDatabaseResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceAutonomousDatabase) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: autonomousDatabaseResourceAttributes(),
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"customer_contacts_to_send_to_oci": customerContactsResourceBlock(ctx),
			"db_tools_details":                 databaseToolsResourceBlock(ctx),
			"long_term_backup_schedule":        longTermBackupScheduleResourceBlock(ctx),
			"resource_pool_summary":            resourcePoolSummaryResourceBlock(ctx),
			"scheduled_operations":             scheduledOperationsResourceBlock(ctx),
			"source_configuration":             sourceConfigurationResourceBlock(ctx),
			"transportable_tablespace":         transportableTablespaceResourceBlock(ctx),
		},
	}
}

func autonomousDatabaseResourceAttributes() map[string]schema.Attribute {
	statusType := fwtypes.StringEnumType[odbtypes.AutonomousDatabaseResourceStatus]()
	maintenanceScheduleType := fwtypes.StringEnumType[odbtypes.AutonomousMaintenanceScheduleType]()
	computeModelType := fwtypes.StringEnumType[odbtypes.ComputeModel]()
	databaseEditionType := fwtypes.StringEnumType[odbtypes.DatabaseEdition]()
	databaseType := fwtypes.StringEnumType[odbtypes.DatabaseType]()
	dbWorkloadType := fwtypes.StringEnumType[odbtypes.DbWorkload]()
	licenseModelType := fwtypes.StringEnumType[odbtypes.LicenseModel]()
	openModeType := fwtypes.StringEnumType[odbtypes.OpenMode]()
	permissionLevelType := fwtypes.StringEnumType[odbtypes.PermissionLevel]()
	refreshableModeType := fwtypes.StringEnumType[odbtypes.RefreshableMode]()
	sourceType := fwtypes.StringEnumType[odbtypes.SourceType]()
	standbyAllowlistedIPsSourceType := fwtypes.StringEnumType[odbtypes.StandbyAllowlistedIpsSource]()

	return map[string]schema.Attribute{
		names.AttrARN: framework.ARNAttributeComputedOnly(),
		names.AttrID:  framework.IDAttribute(),
		"admin_password_wo": schema.StringAttribute{
			Optional:  true,
			Sensitive: true,
			WriteOnly: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(12, 30),
				stringvalidator.AlsoRequires(path.MatchRoot("admin_password_wo_version")),
			},
			Description: "Password for the ADMIN user. This write-only value is never stored in Terraform state.",
		},
		"admin_password_wo_version": schema.Int64Attribute{
			Optional: true,
			Validators: []validator.Int64{
				int64validator.AlsoRequires(path.MatchRoot("admin_password_wo")),
			},
			Description: "Arbitrary version used to trigger an ADMIN password update.",
		},
		"actual_used_data_storage_size_in_tbs": schema.Float64Attribute{
			Computed:    true,
			Description: "Actual amount of data storage currently in use, in TB.",
		},
		"allocated_storage_size_in_tbs": schema.Float64Attribute{
			Computed:    true,
			Description: "Amount of storage currently allocated, in TB.",
		},
		"allowlisted_ips": schema.ListAttribute{
			Optional:    true,
			Computed:    true,
			ElementType: types.StringType,
			Validators: []validator.List{
				listvalidator.SizeBetween(1, 1024),
			},
			Description: "IP addresses allowed to access the Autonomous Database.",
		},
		"auto_refresh_frequency_in_seconds": schema.Int32Attribute{
			Optional:    true,
			Computed:    true,
			Description: "Frequency at which a refreshable clone is automatically refreshed, in seconds.",
		},
		"auto_refresh_point_lag_in_seconds": schema.Int32Attribute{
			Optional:    true,
			Computed:    true,
			Description: "Time lag between a refreshable clone and its source, in seconds.",
		},
		"autonomous_maintenance_schedule_type": schema.StringAttribute{
			CustomType:  maintenanceScheduleType,
			Optional:    true,
			Computed:    true,
			Description: "Maintenance schedule type for the Autonomous Database.",
		},
		"availability_zone": schema.StringAttribute{
			Computed:    true,
			Description: "Availability Zone where the Autonomous Database is located.",
		},
		"availability_zone_id": schema.StringAttribute{
			Computed:    true,
			Description: "Availability Zone ID where the Autonomous Database is located.",
		},
		"available_upgrade_versions": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
			Description: "Oracle Database versions to which the Autonomous Database can be upgraded.",
		},
		"backup_retention_period_in_days": schema.Int32Attribute{
			Optional:    true,
			Computed:    true,
			Description: "Retention period for automatic backups, in days.",
		},
		"byol_compute_count_limit": schema.Float64Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Float64{
				float64validator.AtLeast(2),
			},
			Description: "Maximum compute capacity under the bring-your-own-license model.",
		},
		"character_set": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 255),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "Character set of the Autonomous Database.",
		},
		"compute_count": schema.Float64Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Float64{
				float64validator.Between(0.1, 512),
			},
			Description: "Compute capacity in ECPUs or OCPUs.",
		},
		"compute_model": schema.StringAttribute{
			CustomType:  computeModelType,
			Computed:    true,
			Description: "Compute model of the Autonomous Database.",
		},
		"cpu_core_count": schema.Int32Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Int32{
				int32validator.Between(1, 128),
			},
			Description: "Number of CPU cores allocated to the Autonomous Database.",
		},
		names.AttrCreatedAt: schema.StringAttribute{
			CustomType:  timetypes.RFC3339Type{},
			Computed:    true,
			Description: "Date and time when the Autonomous Database was created.",
		},
		"data_storage_size_in_gbs": schema.Int32Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Int32{
				int32validator.Between(20, 393216),
			},
			Description: "Data volume size in GB.",
		},
		"data_storage_size_in_tbs": schema.Int32Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Int32{
				int32validator.Between(1, 384),
			},
			Description: "Data volume size in TB.",
		},
		"database_edition": schema.StringAttribute{
			CustomType:  databaseEditionType,
			Optional:    true,
			Computed:    true,
			Description: "Oracle Database edition.",
		},
		"database_type": schema.StringAttribute{
			CustomType:  databaseType,
			Computed:    true,
			Description: "Type of Autonomous Database.",
		},
		"db_name": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 30),
				stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`), "must start with a letter and contain only alphanumeric characters"),
			},
			Description: "Name of the Autonomous Database.",
		},
		"db_version": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 255),
			},
			Description: "Oracle Database software version.",
		},
		"db_workload": schema.StringAttribute{
			CustomType:  dbWorkloadType,
			Optional:    true,
			Computed:    true,
			Description: "Intended database workload.",
		},
		names.AttrDisplayName: schema.StringAttribute{
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 255),
			},
			Description: "User-friendly name for the Autonomous Database.",
		},
		"encryption_key_provider": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.OneOf(enum.Slice(odbtypes.EncryptionKeyProviderInputOracleManaged, odbtypes.EncryptionKeyProviderInputAwsKms)...),
			},
			Description: "Encryption key provider. Configurable values are ORACLE_MANAGED and AWS_KMS.",
		},
		"is_auto_scaling_enabled": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether automatic compute scaling is enabled.",
		},
		"is_auto_scaling_for_storage_enabled": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether automatic storage scaling is enabled.",
		},
		"is_backup_retention_locked": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether the backup retention period is locked.",
		},
		"is_local_data_guard_enabled": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether local Oracle Data Guard is enabled.",
		},
		"is_mtls_connection_required": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether mutual TLS authentication is required.",
		},
		"is_refreshable_clone": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether the Autonomous Database is a refreshable clone.",
		},
		names.AttrKMSKeyID: schema.StringAttribute{
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
			Description: "ARN of the AWS KMS key used to encrypt the Autonomous Database.",
		},
		"license_model": schema.StringAttribute{
			CustomType:  licenseModelType,
			Optional:    true,
			Computed:    true,
			Description: "Oracle license model.",
		},
		"local_adg_auto_failover_max_data_loss_limit": schema.Int32Attribute{
			Optional:    true,
			Computed:    true,
			Description: "Maximum data-loss limit for automatic local Data Guard failover, in seconds.",
		},
		"ncharacter_set": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 255),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "National character set of the Autonomous Database.",
		},
		"oci_resource_anchor_name": schema.StringAttribute{
			Computed:    true,
			Description: "Name of the OCI resource anchor.",
		},
		"oci_url": schema.StringAttribute{
			Computed:    true,
			Description: "URL for the Autonomous Database in the OCI console.",
		},
		"ocid": schema.StringAttribute{
			Computed:    true,
			Description: "Oracle Cloud Identifier of the Autonomous Database.",
		},
		"odb_network_arn": schema.StringAttribute{
			Computed:    true,
			Description: "ARN of the associated ODB network.",
		},
		"odb_network_id": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(6, 2048),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "ID of the associated ODB network.",
		},
		"open_mode": schema.StringAttribute{
			CustomType:  openModeType,
			Optional:    true,
			Computed:    true,
			Description: "Open mode of the Autonomous Database.",
		},
		"percent_progress": schema.Float32Attribute{
			Computed:    true,
			Description: "Progress of the current operation, as a percentage.",
		},
		"permission_level": schema.StringAttribute{
			CustomType:  permissionLevelType,
			Optional:    true,
			Computed:    true,
			Description: "Permission level of the Autonomous Database.",
		},
		"private_endpoint": schema.StringAttribute{
			Computed:    true,
			Description: "Private endpoint of the Autonomous Database.",
		},
		"private_endpoint_ip": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Private endpoint IP address.",
		},
		"private_endpoint_label": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Private endpoint label.",
		},
		"refreshable_mode": schema.StringAttribute{
			CustomType:  refreshableModeType,
			Optional:    true,
			Computed:    true,
			Description: "Refresh mode of a refreshable clone.",
		},
		"resource_pool_leader_id": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(6, 2048),
			},
			Description: "ID of the resource-pool leader Autonomous Database.",
		},
		"service_console_url": schema.StringAttribute{
			Computed:    true,
			Description: "URL for the Oracle service console.",
		},
		"source": schema.StringAttribute{
			CustomType: sourceType,
			Optional:   true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Description: "Source from which to create the Autonomous Database.",
		},
		"source_id": schema.StringAttribute{
			Computed:    true,
			Description: "ID of the source used to create the Autonomous Database.",
		},
		"sql_web_developer_url": schema.StringAttribute{
			Computed:    true,
			Description: "URL for Oracle SQL Developer Web.",
		},
		"standby_allowlisted_ips": schema.ListAttribute{
			Optional:    true,
			Computed:    true,
			ElementType: types.StringType,
			Validators: []validator.List{
				listvalidator.SizeBetween(1, 1024),
			},
			Description: "IP addresses allowed to access the standby Autonomous Database.",
		},
		"standby_allowlisted_ips_source": schema.StringAttribute{
			CustomType:  standbyAllowlistedIPsSourceType,
			Optional:    true,
			Computed:    true,
			Description: "Source of the standby allowlisted IP addresses.",
		},
		names.AttrStatus: schema.StringAttribute{
			CustomType:  statusType,
			Computed:    true,
			Description: "Current status of the Autonomous Database.",
		},
		names.AttrStatusReason: schema.StringAttribute{
			Computed:    true,
			Description: "Additional information about the current status.",
		},
		"time_of_auto_refresh_start": schema.StringAttribute{
			CustomType:  timetypes.RFC3339Type{},
			Optional:    true,
			Computed:    true,
			Description: "Date and time when automatic refresh begins.",
		},
		names.AttrTags:    tftags.TagsAttribute(),
		names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
	}
}

func customerContactsResourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseCustomerContactModel](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"email": schema.StringAttribute{
					Required:    true,
					Description: "Email address of the customer contact.",
				},
			},
		},
		Description: "Customer contacts that receive operational notifications from OCI.",
	}
}

func databaseToolsResourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseToolModel](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"compute_count": schema.Float64Attribute{
					Optional:    true,
					Computed:    true,
					Description: "Compute capacity allocated to the database tool.",
				},
				"is_enabled": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Whether the database tool is enabled.",
				},
				"max_idle_time_in_minutes": schema.Int32Attribute{
					Optional:    true,
					Computed:    true,
					Description: "Maximum idle time before the database tool is shut down, in minutes.",
				},
				names.AttrName: schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Name of the database tool.",
				},
			},
		},
		Description: "Database management tools enabled for the Autonomous Database.",
	}
}

func longTermBackupScheduleResourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseLongTermBackupScheduleModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"is_disabled": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Whether the long-term backup schedule is disabled.",
				},
				"repeat_cadence": schema.StringAttribute{
					CustomType:  fwtypes.StringEnumType[odbtypes.RepeatCadence](),
					Optional:    true,
					Computed:    true,
					Description: "Cadence at which long-term backups are taken.",
				},
				"retention_period_in_days": schema.Int32Attribute{
					Optional: true,
					Computed: true,
					Validators: []validator.Int32{
						int32validator.Between(90, 3650),
					},
					Description: "Retention period for long-term backups, in days.",
				},
				"time_of_backup": schema.StringAttribute{
					CustomType:  timetypes.RFC3339Type{},
					Optional:    true,
					Computed:    true,
					Description: "Date and time at which the long-term backup is taken.",
				},
			},
		},
		Description: "Long-term backup schedule for the Autonomous Database.",
	}
}

func resourcePoolSummaryResourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseResourcePoolSummaryModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"available_compute_capacity": schema.Int32Attribute{
					Computed:    true,
					Description: "Available compute capacity in the resource pool.",
				},
				"available_storage_capacity_in_tbs": schema.Float64Attribute{
					Computed:    true,
					Description: "Available storage capacity in the resource pool, in TB.",
				},
				"is_disabled": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Whether the resource pool is disabled.",
				},
				"pool_size": schema.Int32Attribute{
					Optional:    true,
					Computed:    true,
					Description: "Number of Autonomous Databases the resource pool can contain.",
				},
				"pool_storage_size_in_tbs": schema.Int32Attribute{
					Optional:    true,
					Computed:    true,
					Description: "Total storage size of the resource pool, in TB.",
				},
				"total_compute_capacity": schema.Int32Attribute{
					Computed:    true,
					Description: "Total compute capacity of the resource pool.",
				},
			},
		},
		Description: "Resource pool configuration for the Autonomous Database.",
	}
}

func scheduledOperationsResourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseScheduledOperationModel](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"day_of_week": schema.StringAttribute{
					CustomType:  fwtypes.StringEnumType[odbtypes.DayOfWeekName](),
					Required:    true,
					Description: "Day of the week.",
				},
				"scheduled_start_time": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Scheduled start time in UTC.",
				},
				"scheduled_stop_time": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Scheduled stop time in UTC.",
				},
			},
		},
		Description: "Scheduled start and stop times for the Autonomous Database.",
	}
}

func transportableTablespaceResourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseTransportableTablespaceModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"tts_bundle_url": schema.StringAttribute{
					Optional:    true,
					Description: "URL of the transportable tablespace bundle.",
				},
			},
		},
		Description: "Transportable tablespace configuration used during creation.",
	}
}

func sourceConfigurationResourceBlock(ctx context.Context) schema.ListNestedBlock {
	unionPaths := path.Expressions{
		path.MatchRelative().AtParent().AtName("clone_to_refreshable"),
		path.MatchRelative().AtParent().AtName("cross_region_data_guard"),
		path.MatchRelative().AtParent().AtName("cross_region_disaster_recovery"),
		path.MatchRelative().AtParent().AtName("database_clone"),
		path.MatchRelative().AtParent().AtName("point_in_time_restore"),
		path.MatchRelative().AtParent().AtName("restore_from_backup"),
	}
	unionValidators := func() []validator.List {
		return []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.ExactlyOneOf(unionPaths...),
		}
	}

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseSourceConfigurationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"clone_to_refreshable": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseCloneToRefreshableModel](ctx),
					Validators: unionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"auto_refresh_frequency_in_seconds": schema.Int32Attribute{
								Optional:    true,
								Description: "Frequency at which the refreshable clone is automatically refreshed, in seconds.",
							},
							"auto_refresh_point_lag_in_seconds": schema.Int32Attribute{
								Optional:    true,
								Description: "Time lag between the refreshable clone and its source, in seconds.",
							},
							"clone_type": schema.StringAttribute{
								CustomType:  fwtypes.StringEnumType[odbtypes.CloneType](),
								Optional:    true,
								Description: "Type of clone to create.",
							},
							"open_mode": schema.StringAttribute{
								CustomType:  fwtypes.StringEnumType[odbtypes.OpenMode](),
								Optional:    true,
								Description: "Open mode of the refreshable clone.",
							},
							"refreshable_mode": schema.StringAttribute{
								CustomType:  fwtypes.StringEnumType[odbtypes.RefreshableMode](),
								Optional:    true,
								Description: "Refresh mode of the clone.",
							},
							"source_autonomous_database_id": schema.StringAttribute{
								Required:    true,
								Description: "ID of the source Autonomous Database.",
							},
							"time_of_auto_refresh_start": schema.StringAttribute{
								CustomType:  timetypes.RFC3339Type{},
								Optional:    true,
								Description: "Date and time when automatic refresh starts.",
							},
						},
					},
				},
				"cross_region_data_guard": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseCrossRegionDataGuardModel](ctx),
					Validators: unionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"source_autonomous_database_arn": schema.StringAttribute{
								Required:    true,
								Description: "ARN of the source Autonomous Database.",
							},
						},
					},
				},
				"cross_region_disaster_recovery": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseCrossRegionDisasterRecoveryModel](ctx),
					Validators: unionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"is_replicate_automatic_backups": schema.BoolAttribute{
								Optional:    true,
								Description: "Whether automatic backups are replicated to the disaster recovery database.",
							},
							"remote_disaster_recovery_type": schema.StringAttribute{
								CustomType:  fwtypes.StringEnumType[odbtypes.DisasterRecoveryType](),
								Required:    true,
								Description: "Type of remote disaster recovery.",
							},
							"source_autonomous_database_arn": schema.StringAttribute{
								Required:    true,
								Description: "ARN of the source Autonomous Database.",
							},
						},
					},
				},
				"database_clone": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseCloneModel](ctx),
					Validators: unionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"clone_type": schema.StringAttribute{
								CustomType:  fwtypes.StringEnumType[odbtypes.CloneType](),
								Required:    true,
								Description: "Type of clone to create.",
							},
							"source_autonomous_database_id": schema.StringAttribute{
								Required:    true,
								Description: "ID of the source Autonomous Database.",
							},
						},
					},
				},
				"point_in_time_restore": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabasePointInTimeRestoreModel](ctx),
					Validators: unionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"clone_table_space_list": schema.ListAttribute{
								Optional:    true,
								ElementType: types.Int32Type,
								Description: "Tablespace IDs to clone.",
							},
							"clone_type": schema.StringAttribute{
								CustomType:  fwtypes.StringEnumType[odbtypes.CloneType](),
								Required:    true,
								Description: "Type of clone to create.",
							},
							"source_autonomous_database_id": schema.StringAttribute{
								Required:    true,
								Description: "ID of the source Autonomous Database.",
							},
							"timestamp": schema.StringAttribute{
								CustomType:  timetypes.RFC3339Type{},
								Optional:    true,
								Description: "Date and time to which the database is restored.",
							},
							"use_latest_available_backup_timestamp": schema.BoolAttribute{
								Optional:    true,
								Description: "Whether to use the latest available backup timestamp.",
							},
						},
					},
				},
				"restore_from_backup": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseRestoreFromBackupModel](ctx),
					Validators: unionValidators(),
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"autonomous_database_backup_id": schema.StringAttribute{
								Required:    true,
								Description: "ID of the Autonomous Database backup.",
							},
							"clone_table_space_list": schema.ListAttribute{
								Optional:    true,
								ElementType: types.Int32Type,
								Description: "Tablespace IDs to clone.",
							},
							"clone_type": schema.StringAttribute{
								CustomType:  fwtypes.StringEnumType[odbtypes.CloneType](),
								Required:    true,
								Description: "Type of clone to create from the backup.",
							},
						},
					},
				},
			},
		},
		Description: "Source-specific configuration used during creation. Exactly one nested source block must be configured.",
	}
}

func (r *resourceAutonomousDatabase) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan, config autonomousDatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.CreateAutonomousDatabaseInput{
		Tags: getTagsIn(ctx),
	}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.AdminPasswordWO.IsNull() {
		input.AdminPassword = config.AdminPasswordWO.ValueStringPointer()
	}
	input.EncryptionKeyProvider, input.EncryptionKeyConfiguration = expandAutonomousDatabaseEncryption(plan.EncryptionKeyProvider, plan.KMSKeyID)
	input.ScheduledOperations = expandAutonomousDatabaseScheduledOperations(ctx, plan.ScheduledOperations, &resp.Diagnostics)
	input.SourceConfiguration = expandAutonomousDatabaseSourceConfiguration(ctx, plan.SourceConfiguration, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateAutonomousDatabase(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameAutonomousDatabase, plan.DisplayName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.AutonomousDatabaseId == nil {
		err := errors.New("empty output")
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameAutonomousDatabase, plan.DisplayName.String(), err),
			err.Error(),
		)
		return
	}

	id := aws.ToString(out.AutonomousDatabaseId)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), id)...)

	deadline := inttypes.NewDeadline(r.CreateTimeout(ctx, plan.Timeouts))
	created, err := waitAutonomousDatabaseCreated(ctx, conn, id, deadline.Remaining())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameAutonomousDatabase, id, err),
			err.Error(),
		)
		return
	}

	if autonomousDatabasePostCreateUpdateRequired(plan) {
		updateInput := odb.UpdateAutonomousDatabaseInput{
			AutonomousDatabaseId: aws.String(id),
		}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &updateInput)...)
		updateInput.EncryptionKeyProvider, updateInput.EncryptionKeyConfiguration = expandAutonomousDatabaseEncryption(plan.EncryptionKeyProvider, plan.KMSKeyID)
		updateInput.ScheduledOperations = expandAutonomousDatabaseScheduledOperations(ctx, plan.ScheduledOperations, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateAutonomousDatabase(ctx, &updateInput)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionUpdating, ResNameAutonomousDatabase, id, err),
				err.Error(),
			)
			return
		}
		if out == nil || out.AutonomousDatabaseId == nil {
			err := errors.New("empty output")
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionUpdating, ResNameAutonomousDatabase, id, err),
				err.Error(),
			)
			return
		}

		created, err = waitAutonomousDatabaseUpdated(ctx, conn, id, deadline.Remaining())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameAutonomousDatabase, id, err),
				err.Error(),
			)
			return
		}
	}

	flattenAutonomousDatabase(ctx, created, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAutonomousDatabase) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state autonomousDatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAutonomousDatabaseByID(ctx, conn, state.AutonomousDatabaseID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameAutonomousDatabase, state.AutonomousDatabaseID.String(), err),
			err.Error(),
		)
		return
	}

	flattenAutonomousDatabase(ctx, out, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func autonomousDatabasePostCreateUpdateRequired(plan autonomousDatabaseResourceModel) bool {
	return isKnownAutonomousDatabaseValue(plan.AutoRefreshFrequencyInSeconds) ||
		isKnownAutonomousDatabaseValue(plan.AutoRefreshPointLagInSeconds) ||
		isKnownAutonomousDatabaseValue(plan.IsRefreshableClone) ||
		isKnownAutonomousDatabaseValue(plan.LocalAdgAutoFailoverMaxDataLossLimit) ||
		isKnownAutonomousDatabaseValue(plan.LongTermBackupSchedule) ||
		isKnownAutonomousDatabaseValue(plan.OpenMode) ||
		isKnownAutonomousDatabaseValue(plan.PermissionLevel) ||
		isKnownAutonomousDatabaseValue(plan.RefreshableMode) ||
		isKnownAutonomousDatabaseValue(plan.TimeOfAutoRefreshStart)
}

type autonomousDatabaseValue interface {
	IsNull() bool
	IsUnknown() bool
}

func isKnownAutonomousDatabaseValue(value autonomousDatabaseValue) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func autonomousDatabaseUpdateRequired(plan, state autonomousDatabaseResourceModel) bool {
	return !plan.AdminPasswordWOVersion.Equal(state.AdminPasswordWOVersion) ||
		!plan.AllowlistedIps.Equal(state.AllowlistedIps) ||
		!plan.AutoRefreshFrequencyInSeconds.Equal(state.AutoRefreshFrequencyInSeconds) ||
		!plan.AutoRefreshPointLagInSeconds.Equal(state.AutoRefreshPointLagInSeconds) ||
		!plan.AutonomousMaintenanceScheduleType.Equal(state.AutonomousMaintenanceScheduleType) ||
		!plan.BackupRetentionPeriodInDays.Equal(state.BackupRetentionPeriodInDays) ||
		!plan.ByolComputeCountLimit.Equal(state.ByolComputeCountLimit) ||
		!plan.ComputeCount.Equal(state.ComputeCount) ||
		!plan.CpuCoreCount.Equal(state.CpuCoreCount) ||
		!plan.CustomerContactsToSendToOCI.Equal(state.CustomerContactsToSendToOCI) ||
		!plan.DataStorageSizeInGBs.Equal(state.DataStorageSizeInGBs) ||
		!plan.DataStorageSizeInTBs.Equal(state.DataStorageSizeInTBs) ||
		!plan.DatabaseEdition.Equal(state.DatabaseEdition) ||
		!plan.DbName.Equal(state.DbName) ||
		!plan.DbToolsDetails.Equal(state.DbToolsDetails) ||
		!plan.DbVersion.Equal(state.DbVersion) ||
		!plan.DbWorkload.Equal(state.DbWorkload) ||
		!plan.DisplayName.Equal(state.DisplayName) ||
		!plan.EncryptionKeyProvider.Equal(state.EncryptionKeyProvider) ||
		!plan.IsAutoScalingEnabled.Equal(state.IsAutoScalingEnabled) ||
		!plan.IsAutoScalingForStorageEnabled.Equal(state.IsAutoScalingForStorageEnabled) ||
		!plan.IsBackupRetentionLocked.Equal(state.IsBackupRetentionLocked) ||
		!plan.IsLocalDataGuardEnabled.Equal(state.IsLocalDataGuardEnabled) ||
		!plan.IsMtlsConnectionRequired.Equal(state.IsMtlsConnectionRequired) ||
		!plan.IsRefreshableClone.Equal(state.IsRefreshableClone) ||
		!plan.KMSKeyID.Equal(state.KMSKeyID) ||
		!plan.LicenseModel.Equal(state.LicenseModel) ||
		!plan.LocalAdgAutoFailoverMaxDataLossLimit.Equal(state.LocalAdgAutoFailoverMaxDataLossLimit) ||
		!plan.LongTermBackupSchedule.Equal(state.LongTermBackupSchedule) ||
		!plan.OpenMode.Equal(state.OpenMode) ||
		!plan.PermissionLevel.Equal(state.PermissionLevel) ||
		!plan.PrivateEndpointIp.Equal(state.PrivateEndpointIp) ||
		!plan.PrivateEndpointLabel.Equal(state.PrivateEndpointLabel) ||
		!plan.RefreshableMode.Equal(state.RefreshableMode) ||
		!plan.ResourcePoolLeaderId.Equal(state.ResourcePoolLeaderId) ||
		!plan.ResourcePoolSummary.Equal(state.ResourcePoolSummary) ||
		!plan.ScheduledOperations.Equal(state.ScheduledOperations) ||
		!plan.StandbyAllowlistedIps.Equal(state.StandbyAllowlistedIps) ||
		!plan.StandbyAllowlistedIpsSource.Equal(state.StandbyAllowlistedIpsSource) ||
		!plan.TimeOfAutoRefreshStart.Equal(state.TimeOfAutoRefreshStart)
}

func (r *resourceAutonomousDatabase) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan, state, config autonomousDatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !autonomousDatabaseUpdateRequired(plan, state) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	input := odb.UpdateAutonomousDatabaseInput{
		AutonomousDatabaseId: state.AutonomousDatabaseID.ValueStringPointer(),
	}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.AdminPasswordWO.IsNull() && !plan.AdminPasswordWOVersion.Equal(state.AdminPasswordWOVersion) {
		input.AdminPassword = config.AdminPasswordWO.ValueStringPointer()
	}
	input.EncryptionKeyProvider, input.EncryptionKeyConfiguration = expandAutonomousDatabaseEncryption(plan.EncryptionKeyProvider, plan.KMSKeyID)
	input.ScheduledOperations = expandAutonomousDatabaseScheduledOperations(ctx, plan.ScheduledOperations, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.UpdateAutonomousDatabase(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionUpdating, ResNameAutonomousDatabase, state.AutonomousDatabaseID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.AutonomousDatabaseId == nil {
		err := errors.New("empty output")
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionUpdating, ResNameAutonomousDatabase, state.AutonomousDatabaseID.String(), err),
			err.Error(),
		)
		return
	}

	updated, err := waitAutonomousDatabaseUpdated(ctx, conn, state.AutonomousDatabaseID.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForUpdate, ResNameAutonomousDatabase, state.AutonomousDatabaseID.String(), err),
			err.Error(),
		)
		return
	}

	flattenAutonomousDatabase(ctx, updated, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAutonomousDatabase) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state autonomousDatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.AutonomousDatabaseID.ValueString()
	_, err := conn.DeleteAutonomousDatabase(ctx, &odb.DeleteAutonomousDatabaseInput{
		AutonomousDatabaseId: aws.String(id),
	})
	if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionDeleting, ResNameAutonomousDatabase, id, err),
			err.Error(),
		)
		return
	}

	_, err = waitAutonomousDatabaseDeleted(ctx, conn, id, r.DeleteTimeout(ctx, state.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameAutonomousDatabase, id, err),
			err.Error(),
		)
	}
}

func findAutonomousDatabaseByID(ctx context.Context, conn *odb.Client, id string) (*odbtypes.AutonomousDatabase, error) {
	out, err := conn.GetAutonomousDatabase(ctx, &odb.GetAutonomousDatabaseInput{
		AutonomousDatabaseId: aws.String(id),
	})
	if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{LastError: err}
	}
	if err != nil {
		return nil, err
	}
	if out == nil || out.AutonomousDatabase == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.AutonomousDatabase, nil
}

func statusAutonomousDatabase(conn *odb.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findAutonomousDatabaseByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

var autonomousDatabasePendingStatuses = enum.Slice(
	odbtypes.AutonomousDatabaseResourceStatusProvisioning,
	odbtypes.AutonomousDatabaseResourceStatusTerminating,
	odbtypes.AutonomousDatabaseResourceStatusUpdating,
	odbtypes.AutonomousDatabaseResourceStatusMaintenanceInProgress,
	odbtypes.AutonomousDatabaseResourceStatusStopping,
	odbtypes.AutonomousDatabaseResourceStatusStarting,
	odbtypes.AutonomousDatabaseResourceStatusRestoreInProgress,
	odbtypes.AutonomousDatabaseResourceStatusBackupInProgress,
	odbtypes.AutonomousDatabaseResourceStatusScaleInProgress,
	odbtypes.AutonomousDatabaseResourceStatusRestarting,
	odbtypes.AutonomousDatabaseResourceStatusRecreating,
	odbtypes.AutonomousDatabaseResourceStatusRoleChangeInProgress,
	odbtypes.AutonomousDatabaseResourceStatusUpgrading,
)

var autonomousDatabaseSuccessStatuses = enum.Slice(
	odbtypes.AutonomousDatabaseResourceStatusAvailable,
	odbtypes.AutonomousDatabaseResourceStatusAvailableNeedsAttention,
	odbtypes.AutonomousDatabaseResourceStatusStopped,
	odbtypes.AutonomousDatabaseResourceStatusStandby,
)

var autonomousDatabaseFailureStatuses = enum.Slice(
	odbtypes.AutonomousDatabaseResourceStatusFailed,
	odbtypes.AutonomousDatabaseResourceStatusRestoreFailed,
	odbtypes.AutonomousDatabaseResourceStatusUnavailable,
	odbtypes.AutonomousDatabaseResourceStatusInaccessible,
	odbtypes.AutonomousDatabaseResourceStatusTerminated,
)

func waitAutonomousDatabaseCreated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.AutonomousDatabase, error) {
	return waitAutonomousDatabaseReady(ctx, conn, id, timeout)
}

func waitAutonomousDatabaseUpdated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.AutonomousDatabase, error) {
	return waitAutonomousDatabaseReady(ctx, conn, id, timeout)
}

func waitAutonomousDatabaseReady(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.AutonomousDatabase, error) {
	targets := append(append([]string{}, autonomousDatabaseSuccessStatuses...), autonomousDatabaseFailureStatuses...)
	stateConf := &retry.StateChangeConf{
		Pending: append([]string{""}, autonomousDatabasePendingStatuses...),
		Target:  targets,
		Refresh: statusAutonomousDatabase(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}
	out, ok := outputRaw.(*odbtypes.AutonomousDatabase)
	if !ok || out == nil {
		return nil, tfresource.NewEmptyResultError()
	}
	for _, status := range autonomousDatabaseFailureStatuses {
		if string(out.Status) == status {
			return out, fmt.Errorf("Autonomous Database (%s) entered status %s: %s", id, out.Status, aws.ToString(out.StatusReason))
		}
	}

	return out, nil
}

func waitAutonomousDatabaseDeleted(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.AutonomousDatabase, error) {
	stateConf := &retry.StateChangeConf{
		Pending: append(append([]string{""}, autonomousDatabasePendingStatuses...), autonomousDatabaseSuccessStatuses...),
		Target:  enum.Slice(odbtypes.AutonomousDatabaseResourceStatusTerminated),
		Refresh: statusAutonomousDatabase(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.AutonomousDatabase); ok {
		return out, err
	}
	return nil, err
}

func expandAutonomousDatabaseEncryption(provider, kmsKeyID types.String) (odbtypes.EncryptionKeyProviderInput, odbtypes.EncryptionKeyConfigurationInput) {
	var configuration odbtypes.EncryptionKeyConfigurationInput
	if !kmsKeyID.IsNull() && !kmsKeyID.IsUnknown() {
		configuration = &odbtypes.EncryptionKeyConfigurationInputMemberAwsEncryptionKey{
			Value: odbtypes.AwsEncryptionKeyConfigurationInput{
				KmsKeyId: kmsKeyID.ValueStringPointer(),
			},
		}
	}

	if provider.IsNull() || provider.IsUnknown() {
		return "", configuration
	}
	return odbtypes.EncryptionKeyProviderInput(provider.ValueString()), configuration
}

func expandAutonomousDatabaseScheduledOperations(ctx context.Context, value fwtypes.ListNestedObjectValueOf[autonomousDatabaseScheduledOperationModel], diags *diag.Diagnostics) []odbtypes.ScheduledOperationDetails {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	models, d := value.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	apiObjects := make([]odbtypes.ScheduledOperationDetails, 0, len(models))
	for _, model := range models {
		apiObjects = append(apiObjects, odbtypes.ScheduledOperationDetails{
			DayOfWeek: &odbtypes.DayOfWeek{
				Name: model.DayOfWeek.ValueEnum(),
			},
			ScheduledStartTime: model.ScheduledStartTime.ValueStringPointer(),
			ScheduledStopTime:  model.ScheduledStopTime.ValueStringPointer(),
		})
	}

	return apiObjects
}

func flattenAutonomousDatabaseScheduledOperations(ctx context.Context, apiObjects []odbtypes.ScheduledOperationDetails, target *fwtypes.ListNestedObjectValueOf[autonomousDatabaseScheduledOperationModel]) diag.Diagnostics {
	models := make([]autonomousDatabaseScheduledOperationModel, 0, len(apiObjects))
	for _, apiObject := range apiObjects {
		model := autonomousDatabaseScheduledOperationModel{
			ScheduledStartTime: types.StringPointerValue(apiObject.ScheduledStartTime),
			ScheduledStopTime:  types.StringPointerValue(apiObject.ScheduledStopTime),
		}
		if apiObject.DayOfWeek != nil {
			model.DayOfWeek = fwtypes.StringEnumValue(apiObject.DayOfWeek.Name)
		}
		models = append(models, model)
	}

	value, diags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, models)
	*target = value
	return diags
}

func expandAutonomousDatabaseSourceConfiguration(ctx context.Context, value fwtypes.ListNestedObjectValueOf[autonomousDatabaseSourceConfigurationModel], diags *diag.Diagnostics) odbtypes.SourceConfiguration {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	configuration, d := value.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || configuration == nil {
		return nil
	}

	if !configuration.CloneToRefreshable.IsNull() {
		model, d := configuration.CloneToRefreshable.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() || model == nil {
			return nil
		}
		var apiObject odbtypes.CloneToRefreshableConfiguration
		diags.Append(flex.Expand(ctx, model, &apiObject)...)
		return &odbtypes.SourceConfigurationMemberCloneToRefreshable{Value: apiObject}
	}
	if !configuration.CrossRegionDataGuard.IsNull() {
		model, d := configuration.CrossRegionDataGuard.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() || model == nil {
			return nil
		}
		var apiObject odbtypes.CrossRegionDataGuardConfiguration
		diags.Append(flex.Expand(ctx, model, &apiObject)...)
		return &odbtypes.SourceConfigurationMemberCrossRegionDataGuard{Value: apiObject}
	}
	if !configuration.CrossRegionDisasterRecovery.IsNull() {
		model, d := configuration.CrossRegionDisasterRecovery.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() || model == nil {
			return nil
		}
		var apiObject odbtypes.CrossRegionDisasterRecoveryConfiguration
		diags.Append(flex.Expand(ctx, model, &apiObject)...)
		return &odbtypes.SourceConfigurationMemberCrossRegionDisasterRecovery{Value: apiObject}
	}
	if !configuration.DatabaseClone.IsNull() {
		model, d := configuration.DatabaseClone.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() || model == nil {
			return nil
		}
		var apiObject odbtypes.DatabaseCloneConfiguration
		diags.Append(flex.Expand(ctx, model, &apiObject)...)
		return &odbtypes.SourceConfigurationMemberDatabaseClone{Value: apiObject}
	}
	if !configuration.PointInTimeRestore.IsNull() {
		model, d := configuration.PointInTimeRestore.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() || model == nil {
			return nil
		}
		var apiObject odbtypes.PointInTimeRestoreConfiguration
		diags.Append(flex.Expand(ctx, model, &apiObject)...)
		return &odbtypes.SourceConfigurationMemberPointInTimeRestore{Value: apiObject}
	}
	if !configuration.RestoreFromBackup.IsNull() {
		model, d := configuration.RestoreFromBackup.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() || model == nil {
			return nil
		}
		var apiObject odbtypes.RestoreFromBackupConfiguration
		diags.Append(flex.Expand(ctx, model, &apiObject)...)
		return &odbtypes.SourceConfigurationMemberRestoreFromBackup{Value: apiObject}
	}

	return nil
}

func flattenAutonomousDatabase(ctx context.Context, apiObject *odbtypes.AutonomousDatabase, model *autonomousDatabaseResourceModel, diags *diag.Diagnostics) {
	diags.Append(flex.Flatten(ctx, apiObject, model)...)
	if diags.HasError() {
		return
	}

	diags.Append(flex.Flatten(ctx, apiObject.CustomerContacts, &model.CustomerContactsToSendToOCI)...)
	diags.Append(flattenAutonomousDatabaseScheduledOperations(ctx, apiObject.ScheduledOperations, &model.ScheduledOperations)...)
	if diags.HasError() {
		return
	}
	if apiObject.EncryptionSummary == nil {
		return
	}

	model.EncryptionKeyProvider = types.StringValue(string(apiObject.EncryptionSummary.EncryptionKeyProvider))
	switch configuration := apiObject.EncryptionSummary.EncryptionKeyConfiguration.(type) {
	case *odbtypes.EncryptionKeyConfigurationMemberAwsEncryptionKey:
		model.KMSKeyID = types.StringPointerValue(configuration.Value.KmsKeyId)
	}
}

type autonomousDatabaseResourceModel struct {
	framework.WithRegionModel
	ActualUsedDataStorageSizeInTBs       types.Float64                                                                   `tfsdk:"actual_used_data_storage_size_in_tbs"`
	AdminPasswordWO                      types.String                                                                    `tfsdk:"admin_password_wo" autoflex:"-"`
	AdminPasswordWOVersion               types.Int64                                                                     `tfsdk:"admin_password_wo_version" autoflex:"-"`
	AllocatedStorageSizeInTBs            types.Float64                                                                   `tfsdk:"allocated_storage_size_in_tbs"`
	AllowlistedIps                       fwtypes.ListValueOf[types.String]                                               `tfsdk:"allowlisted_ips"`
	AutoRefreshFrequencyInSeconds        types.Int32                                                                     `tfsdk:"auto_refresh_frequency_in_seconds"`
	AutoRefreshPointLagInSeconds         types.Int32                                                                     `tfsdk:"auto_refresh_point_lag_in_seconds"`
	AutonomousDatabaseARN                types.String                                                                    `tfsdk:"arn"`
	AutonomousDatabaseID                 types.String                                                                    `tfsdk:"id"`
	AutonomousMaintenanceScheduleType    fwtypes.StringEnum[odbtypes.AutonomousMaintenanceScheduleType]                  `tfsdk:"autonomous_maintenance_schedule_type"`
	AvailabilityZone                     types.String                                                                    `tfsdk:"availability_zone"`
	AvailabilityZoneID                   types.String                                                                    `tfsdk:"availability_zone_id"`
	AvailableUpgradeVersions             fwtypes.ListValueOf[types.String]                                               `tfsdk:"available_upgrade_versions"`
	BackupRetentionPeriodInDays          types.Int32                                                                     `tfsdk:"backup_retention_period_in_days"`
	ByolComputeCountLimit                types.Float64                                                                   `tfsdk:"byol_compute_count_limit"`
	CharacterSet                         types.String                                                                    `tfsdk:"character_set"`
	ComputeCount                         types.Float64                                                                   `tfsdk:"compute_count"`
	ComputeModel                         fwtypes.StringEnum[odbtypes.ComputeModel]                                       `tfsdk:"compute_model"`
	CpuCoreCount                         types.Int32                                                                     `tfsdk:"cpu_core_count"`
	CreatedAt                            timetypes.RFC3339                                                               `tfsdk:"created_at"`
	CustomerContactsToSendToOCI          fwtypes.ListNestedObjectValueOf[autonomousDatabaseCustomerContactModel]         `tfsdk:"customer_contacts_to_send_to_oci" autoflex:",noflatten"`
	DataStorageSizeInGBs                 types.Int32                                                                     `tfsdk:"data_storage_size_in_gbs"`
	DataStorageSizeInTBs                 types.Int32                                                                     `tfsdk:"data_storage_size_in_tbs"`
	DatabaseEdition                      fwtypes.StringEnum[odbtypes.DatabaseEdition]                                    `tfsdk:"database_edition"`
	DatabaseType                         fwtypes.StringEnum[odbtypes.DatabaseType]                                       `tfsdk:"database_type"`
	DbName                               types.String                                                                    `tfsdk:"db_name"`
	DbToolsDetails                       fwtypes.ListNestedObjectValueOf[autonomousDatabaseToolModel]                    `tfsdk:"db_tools_details"`
	DbVersion                            types.String                                                                    `tfsdk:"db_version"`
	DbWorkload                           fwtypes.StringEnum[odbtypes.DbWorkload]                                         `tfsdk:"db_workload"`
	DisplayName                          types.String                                                                    `tfsdk:"display_name"`
	EncryptionKeyProvider                types.String                                                                    `tfsdk:"encryption_key_provider" autoflex:"-"`
	IsAutoScalingEnabled                 types.Bool                                                                      `tfsdk:"is_auto_scaling_enabled"`
	IsAutoScalingForStorageEnabled       types.Bool                                                                      `tfsdk:"is_auto_scaling_for_storage_enabled"`
	IsBackupRetentionLocked              types.Bool                                                                      `tfsdk:"is_backup_retention_locked"`
	IsLocalDataGuardEnabled              types.Bool                                                                      `tfsdk:"is_local_data_guard_enabled"`
	IsMtlsConnectionRequired             types.Bool                                                                      `tfsdk:"is_mtls_connection_required"`
	IsRefreshableClone                   types.Bool                                                                      `tfsdk:"is_refreshable_clone"`
	KMSKeyID                             types.String                                                                    `tfsdk:"kms_key_id" autoflex:"-"`
	LicenseModel                         fwtypes.StringEnum[odbtypes.LicenseModel]                                       `tfsdk:"license_model"`
	LocalAdgAutoFailoverMaxDataLossLimit types.Int32                                                                     `tfsdk:"local_adg_auto_failover_max_data_loss_limit"`
	LongTermBackupSchedule               fwtypes.ListNestedObjectValueOf[autonomousDatabaseLongTermBackupScheduleModel]  `tfsdk:"long_term_backup_schedule"`
	NcharacterSet                        types.String                                                                    `tfsdk:"ncharacter_set"`
	OciResourceAnchorName                types.String                                                                    `tfsdk:"oci_resource_anchor_name"`
	OciUrl                               types.String                                                                    `tfsdk:"oci_url"`
	Ocid                                 types.String                                                                    `tfsdk:"ocid"`
	OdbNetworkArn                        types.String                                                                    `tfsdk:"odb_network_arn"`
	OdbNetworkId                         types.String                                                                    `tfsdk:"odb_network_id"`
	OpenMode                             fwtypes.StringEnum[odbtypes.OpenMode]                                           `tfsdk:"open_mode"`
	PercentProgress                      types.Float32                                                                   `tfsdk:"percent_progress"`
	PermissionLevel                      fwtypes.StringEnum[odbtypes.PermissionLevel]                                    `tfsdk:"permission_level"`
	PrivateEndpoint                      types.String                                                                    `tfsdk:"private_endpoint"`
	PrivateEndpointIp                    types.String                                                                    `tfsdk:"private_endpoint_ip"`
	PrivateEndpointLabel                 types.String                                                                    `tfsdk:"private_endpoint_label"`
	RefreshableMode                      fwtypes.StringEnum[odbtypes.RefreshableMode]                                    `tfsdk:"refreshable_mode"`
	ResourcePoolLeaderId                 types.String                                                                    `tfsdk:"resource_pool_leader_id"`
	ResourcePoolSummary                  fwtypes.ListNestedObjectValueOf[autonomousDatabaseResourcePoolSummaryModel]     `tfsdk:"resource_pool_summary"`
	ScheduledOperations                  fwtypes.ListNestedObjectValueOf[autonomousDatabaseScheduledOperationModel]      `tfsdk:"scheduled_operations" autoflex:"-"`
	ServiceConsoleUrl                    types.String                                                                    `tfsdk:"service_console_url"`
	Source                               fwtypes.StringEnum[odbtypes.SourceType]                                         `tfsdk:"source" autoflex:",noflatten"`
	SourceConfiguration                  fwtypes.ListNestedObjectValueOf[autonomousDatabaseSourceConfigurationModel]     `tfsdk:"source_configuration" autoflex:"-"`
	SourceId                             types.String                                                                    `tfsdk:"source_id"`
	SqlWebDeveloperUrl                   types.String                                                                    `tfsdk:"sql_web_developer_url"`
	StandbyAllowlistedIps                fwtypes.ListValueOf[types.String]                                               `tfsdk:"standby_allowlisted_ips"`
	StandbyAllowlistedIpsSource          fwtypes.StringEnum[odbtypes.StandbyAllowlistedIpsSource]                        `tfsdk:"standby_allowlisted_ips_source"`
	Status                               fwtypes.StringEnum[odbtypes.AutonomousDatabaseResourceStatus]                   `tfsdk:"status"`
	StatusReason                         types.String                                                                    `tfsdk:"status_reason"`
	Tags                                 tftags.Map                                                                      `tfsdk:"tags"`
	TagsAll                              tftags.Map                                                                      `tfsdk:"tags_all"`
	TimeOfAutoRefreshStart               timetypes.RFC3339                                                               `tfsdk:"time_of_auto_refresh_start"`
	Timeouts                             timeouts.Value                                                                  `tfsdk:"timeouts"`
	TransportableTablespace              fwtypes.ListNestedObjectValueOf[autonomousDatabaseTransportableTablespaceModel] `tfsdk:"transportable_tablespace" autoflex:",noflatten"`
}

type autonomousDatabaseCustomerContactModel struct {
	Email types.String `tfsdk:"email"`
}

type autonomousDatabaseToolModel struct {
	ComputeCount         types.Float64 `tfsdk:"compute_count"`
	IsEnabled            types.Bool    `tfsdk:"is_enabled"`
	MaxIdleTimeInMinutes types.Int32   `tfsdk:"max_idle_time_in_minutes"`
	Name                 types.String  `tfsdk:"name"`
}

type autonomousDatabaseLongTermBackupScheduleModel struct {
	IsDisabled            types.Bool                                 `tfsdk:"is_disabled"`
	RepeatCadence         fwtypes.StringEnum[odbtypes.RepeatCadence] `tfsdk:"repeat_cadence"`
	RetentionPeriodInDays types.Int32                                `tfsdk:"retention_period_in_days"`
	TimeOfBackup          timetypes.RFC3339                          `tfsdk:"time_of_backup"`
}

type autonomousDatabaseResourcePoolSummaryModel struct {
	AvailableComputeCapacity      types.Int32   `tfsdk:"available_compute_capacity"`
	AvailableStorageCapacityInTBs types.Float64 `tfsdk:"available_storage_capacity_in_tbs"`
	IsDisabled                    types.Bool    `tfsdk:"is_disabled"`
	PoolSize                      types.Int32   `tfsdk:"pool_size"`
	PoolStorageSizeInTBs          types.Int32   `tfsdk:"pool_storage_size_in_tbs"`
	TotalComputeCapacity          types.Int32   `tfsdk:"total_compute_capacity"`
}

type autonomousDatabaseScheduledOperationModel struct {
	DayOfWeek          fwtypes.StringEnum[odbtypes.DayOfWeekName] `tfsdk:"day_of_week"`
	ScheduledStartTime types.String                               `tfsdk:"scheduled_start_time"`
	ScheduledStopTime  types.String                               `tfsdk:"scheduled_stop_time"`
}

type autonomousDatabaseTransportableTablespaceModel struct {
	TtsBundleUrl types.String `tfsdk:"tts_bundle_url"`
}

type autonomousDatabaseSourceConfigurationModel struct {
	CloneToRefreshable          fwtypes.ListNestedObjectValueOf[autonomousDatabaseCloneToRefreshableModel]          `tfsdk:"clone_to_refreshable"`
	CrossRegionDataGuard        fwtypes.ListNestedObjectValueOf[autonomousDatabaseCrossRegionDataGuardModel]        `tfsdk:"cross_region_data_guard"`
	CrossRegionDisasterRecovery fwtypes.ListNestedObjectValueOf[autonomousDatabaseCrossRegionDisasterRecoveryModel] `tfsdk:"cross_region_disaster_recovery"`
	DatabaseClone               fwtypes.ListNestedObjectValueOf[autonomousDatabaseCloneModel]                       `tfsdk:"database_clone"`
	PointInTimeRestore          fwtypes.ListNestedObjectValueOf[autonomousDatabasePointInTimeRestoreModel]          `tfsdk:"point_in_time_restore"`
	RestoreFromBackup           fwtypes.ListNestedObjectValueOf[autonomousDatabaseRestoreFromBackupModel]           `tfsdk:"restore_from_backup"`
}

type autonomousDatabaseCloneToRefreshableModel struct {
	AutoRefreshFrequencyInSeconds types.Int32                                  `tfsdk:"auto_refresh_frequency_in_seconds"`
	AutoRefreshPointLagInSeconds  types.Int32                                  `tfsdk:"auto_refresh_point_lag_in_seconds"`
	CloneType                     fwtypes.StringEnum[odbtypes.CloneType]       `tfsdk:"clone_type"`
	OpenMode                      fwtypes.StringEnum[odbtypes.OpenMode]        `tfsdk:"open_mode"`
	RefreshableMode               fwtypes.StringEnum[odbtypes.RefreshableMode] `tfsdk:"refreshable_mode"`
	SourceAutonomousDatabaseId    types.String                                 `tfsdk:"source_autonomous_database_id"`
	TimeOfAutoRefreshStart        timetypes.RFC3339                            `tfsdk:"time_of_auto_refresh_start"`
}

type autonomousDatabaseCrossRegionDataGuardModel struct {
	SourceAutonomousDatabaseArn types.String `tfsdk:"source_autonomous_database_arn"`
}

type autonomousDatabaseCrossRegionDisasterRecoveryModel struct {
	IsReplicateAutomaticBackups types.Bool                                        `tfsdk:"is_replicate_automatic_backups"`
	RemoteDisasterRecoveryType  fwtypes.StringEnum[odbtypes.DisasterRecoveryType] `tfsdk:"remote_disaster_recovery_type"`
	SourceAutonomousDatabaseArn types.String                                      `tfsdk:"source_autonomous_database_arn"`
}

type autonomousDatabaseCloneModel struct {
	CloneType                  fwtypes.StringEnum[odbtypes.CloneType] `tfsdk:"clone_type"`
	SourceAutonomousDatabaseId types.String                           `tfsdk:"source_autonomous_database_id"`
}

type autonomousDatabasePointInTimeRestoreModel struct {
	CloneTableSpaceList               fwtypes.ListValueOf[types.Int32]       `tfsdk:"clone_table_space_list"`
	CloneType                         fwtypes.StringEnum[odbtypes.CloneType] `tfsdk:"clone_type"`
	SourceAutonomousDatabaseId        types.String                           `tfsdk:"source_autonomous_database_id"`
	Timestamp                         timetypes.RFC3339                      `tfsdk:"timestamp"`
	UseLatestAvailableBackupTimestamp types.Bool                             `tfsdk:"use_latest_available_backup_timestamp"`
}

type autonomousDatabaseRestoreFromBackupModel struct {
	AutonomousDatabaseBackupId types.String                           `tfsdk:"autonomous_database_backup_id"`
	CloneTableSpaceList        fwtypes.ListValueOf[types.Int32]       `tfsdk:"clone_table_space_list"`
	CloneType                  fwtypes.StringEnum[odbtypes.CloneType] `tfsdk:"clone_type"`
}
