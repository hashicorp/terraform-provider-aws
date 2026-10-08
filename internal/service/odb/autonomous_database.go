// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package odb

import (
	"context"
	"math"
	"reflect"
	"slices"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_odb_autonomous_database", name="Autonomous Database")
// @IdentityAttribute("id")
// @Tags(identifierAttribute="arn")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/odb/types;odbtypes;odbtypes.AutonomousDatabase")
// @Testing(preCheck="testAccAutonomousDatabasePreCheck")
// The fixture supplies one external ODB network ID, which cannot be reused in the alternate Region.
// @Testing(identityRegionOverrideTest=false)
// @Testing(importIgnore="admin_password;admin_password_wo;admin_password_wo_version;source;source_configuration;transportable_tablespace")
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
	framework.WithImportByIdentity
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
			"admin_password_source":            adminPasswordSourceResourceBlock(ctx),
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
		"admin_password": schema.StringAttribute{
			Optional:  true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(12, 30),
				stringvalidator.ConflictsWith(path.MatchRoot("admin_password_source")),
				stringvalidator.ConflictsWith(path.MatchRoot("admin_password_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("admin_password_wo")),
			},
			Description: "Password for the ADMIN user. This value is stored in Terraform state. Use admin_password_wo with Terraform 1.11 or later to avoid storing the password in state.",
		},
		"admin_password_wo": schema.StringAttribute{
			Optional:  true,
			Sensitive: true,
			WriteOnly: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(12, 30),
				stringvalidator.ConflictsWith(path.MatchRoot("admin_password")),
				stringvalidator.ConflictsWith(path.MatchRoot("admin_password_source")),
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
			CustomType:  fwtypes.ListOfStringType,
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
		names.AttrAvailabilityZone: schema.StringAttribute{
			Computed:    true,
			Description: "Availability Zone where the Autonomous Database is located.",
		},
		"availability_zone_id": schema.StringAttribute{
			Computed:    true,
			Description: "Availability Zone ID where the Autonomous Database is located.",
		},
		"available_upgrade_versions": schema.ListAttribute{
			CustomType:  fwtypes.ListOfStringType,
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
		"data_storage_size_in_tbs": schema.Float64Attribute{
			Optional: true,
			Computed: true,
			Validators: []validator.Float64{
				float64validator.Between(1, 384),
			},
			Description: "Data volume size in TB. Configured values must be whole numbers; computed values may be fractional when storage is configured in GB.",
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
		names.AttrSource: schema.StringAttribute{
			CustomType: sourceType,
			Optional:   true,
			PlanModifiers: []planmodifier.String{
				// AWS does not return the creation source, so it is absent from imported state.
				stringplanmodifier.RequiresReplaceIf(func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
					resp.RequiresReplace = !req.StateValue.IsNull()
				}, "Replace when a known creation source changes", "Replace when a known creation source changes"),
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
			CustomType:  fwtypes.ListOfStringType,
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

func adminPasswordSourceResourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseAdminPasswordSourceModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.ConflictsWith(
				path.MatchRoot("admin_password"),
				path.MatchRoot("admin_password_wo"),
			),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"customer_managed_aws_secret": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseCustomerManagedAWSSecretModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"external_id_type": schema.StringAttribute{
								CustomType:  fwtypes.StringEnumType[odbtypes.ExternalIdType](),
								Required:    true,
								Description: "Type of OCI identifier supplied as the external ID when OCI assumes the IAM role.",
							},
							names.AttrIAMRoleARN: schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									fwvalidators.ARN(),
								},
								Description: "ARN of the customer-managed IAM role OCI assumes to retrieve the secret.",
							},
							"secret_arn": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									fwvalidators.ARN(),
								},
								Description: "ARN of the AWS Secrets Manager secret that contains the ADMIN password.",
							},
						},
					},
				},
			},
		},
		Description: "Source of the ADMIN password. Conflicts with admin_password and admin_password_wo.",
	}
}

func customerContactsResourceBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[autonomousDatabaseCustomerContactModel](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrEmail: schema.StringAttribute{
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
			// Imported state has no creation-only block because AWS does not return it.
			listplanmodifier.RequiresReplaceIf(func(_ context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = req.StateValue.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0
			}, "Replace when known creation-only settings change", "Replace when known creation-only settings change"),
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
			// Imported state has no creation-only block because AWS does not return it.
			listplanmodifier.RequiresReplaceIf(func(_ context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = req.StateValue.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0
			}, "Replace when known creation-only settings change", "Replace when known creation-only settings change"),
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
								CustomType:  autonomousDatabaseListOfInt32Type(ctx),
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
								CustomType:  autonomousDatabaseListOfInt32Type(ctx),
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

func autonomousDatabaseListOfInt32Type(ctx context.Context) basetypes.ListTypable {
	return fwtypes.NewListValueOfNull[types.Int32](ctx).Type(ctx).(basetypes.ListTypable)
}

func (r *resourceAutonomousDatabase) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var storage types.Float64
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.GetAttribute(ctx, path.Root("data_storage_size_in_tbs"), &storage))
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := expandAutonomousDatabaseDataStorageSizeInTBs(storage); err != nil {
		smerr.AddOne(ctx, &resp.Diagnostics, diag.NewAttributeErrorDiagnostic(path.Root("data_storage_size_in_tbs"), "Invalid TB storage size", err.Error()))
	}
}

func (r *resourceAutonomousDatabase) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan, config autonomousDatabaseResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &config))
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.CreateAutonomousDatabaseInput{
		Tags: getTagsIn(ctx),
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.AdminPasswordWO.IsNull() {
		input.AdminPassword = config.AdminPasswordWO.ValueStringPointer()
	} else if !plan.AdminPassword.IsNull() {
		input.AdminPassword = plan.AdminPassword.ValueStringPointer()
	}
	input.AdminPasswordSource, input.AdminPasswordSourceConfiguration = expandAutonomousDatabaseAdminPasswordSource(ctx, plan.AdminPasswordSource, &resp.Diagnostics)
	input.EncryptionKeyProvider, input.EncryptionKeyConfiguration = expandAutonomousDatabaseEncryption(plan.EncryptionKeyProvider, plan.KMSKeyID)
	input.ScheduledOperations = expandAutonomousDatabaseScheduledOperations(ctx, plan.ScheduledOperations, &resp.Diagnostics)
	input.SourceConfiguration = expandAutonomousDatabaseSourceConfiguration(ctx, plan.SourceConfiguration, &resp.Diagnostics)
	if !config.DataStorageSizeInTBs.IsNull() {
		var err error
		input.DataStorageSizeInTBs, err = expandAutonomousDatabaseDataStorageSizeInTBs(plan.DataStorageSizeInTBs)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.DisplayName.ValueString())
			return
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateAutonomousDatabase(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.DisplayName.ValueString())
		return
	}
	if out == nil || out.AutonomousDatabaseId == nil {
		err := tfresource.NewEmptyResultError()
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.DisplayName.ValueString())
		return
	}

	id := aws.ToString(out.AutonomousDatabaseId)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrID), id))

	deadline := inttypes.NewDeadline(r.CreateTimeout(ctx, plan.Timeouts))
	created, err := waitAutonomousDatabaseCreated(ctx, conn, id, deadline.Remaining())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	if autonomousDatabasePostCreateUpdateRequired(plan) {
		updateInput := expandAutonomousDatabasePostCreateUpdateInput(ctx, id, plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateAutonomousDatabase(ctx, &updateInput)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
			return
		}
		if out == nil || out.AutonomousDatabaseId == nil {
			err := tfresource.NewEmptyResultError()
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
			return
		}

		created, err = waitAutonomousDatabaseUpdated(ctx, conn, id, deadline.Remaining())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
			return
		}
	}

	// These configured blocks are not always returned by GetAutonomousDatabase.
	// Start flattening with their configuration values so omitted API fields do
	// not produce an inconsistent result after apply.
	plan.CustomerContactsToSendToOCI = config.CustomerContactsToSendToOCI
	plan.LongTermBackupSchedule = config.LongTermBackupSchedule
	flattenAutonomousDatabase(ctx, created, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceAutonomousDatabase) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state autonomousDatabaseResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAutonomousDatabaseByID(ctx, conn, state.AutonomousDatabaseID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.AutonomousDatabaseID.ValueString())
		return
	}

	flattenAutonomousDatabase(ctx, out, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
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

type autonomousDatabasePostCreateUpdateModel struct {
	AutoRefreshFrequencyInSeconds        types.Int32
	AutoRefreshPointLagInSeconds         types.Int32
	IsRefreshableClone                   types.Bool
	LocalAdgAutoFailoverMaxDataLossLimit types.Int32
	LongTermBackupSchedule               fwtypes.ListNestedObjectValueOf[autonomousDatabaseLongTermBackupScheduleModel]
	OpenMode                             fwtypes.StringEnum[odbtypes.OpenMode]
	PermissionLevel                      fwtypes.StringEnum[odbtypes.PermissionLevel]
	RefreshableMode                      fwtypes.StringEnum[odbtypes.RefreshableMode]
	TimeOfAutoRefreshStart               timetypes.RFC3339
}

// AutoFlex converts the fields while this wrapper limits the payload to properties accepted by the post-create update API.
// nosemgrep:ci.semgrep.framework.manual-expander-functions
func expandAutonomousDatabasePostCreateUpdateInput(ctx context.Context, id string, plan autonomousDatabaseResourceModel, diags *diag.Diagnostics) odb.UpdateAutonomousDatabaseInput {
	input := odb.UpdateAutonomousDatabaseInput{
		AutonomousDatabaseId: aws.String(id),
	}
	postCreateUpdate := autonomousDatabasePostCreateUpdateModel{
		AutoRefreshFrequencyInSeconds:        plan.AutoRefreshFrequencyInSeconds,
		AutoRefreshPointLagInSeconds:         plan.AutoRefreshPointLagInSeconds,
		IsRefreshableClone:                   plan.IsRefreshableClone,
		LocalAdgAutoFailoverMaxDataLossLimit: plan.LocalAdgAutoFailoverMaxDataLossLimit,
		LongTermBackupSchedule:               plan.LongTermBackupSchedule,
		OpenMode:                             plan.OpenMode,
		PermissionLevel:                      plan.PermissionLevel,
		RefreshableMode:                      plan.RefreshableMode,
		TimeOfAutoRefreshStart:               plan.TimeOfAutoRefreshStart,
	}
	smerr.AddEnrich(ctx, diags, flex.Expand(ctx, postCreateUpdate, &input))

	return input
}

type autonomousDatabaseValue interface {
	IsNull() bool
	IsUnknown() bool
}

type autonomousDatabaseCollectionValue interface {
	autonomousDatabaseValue
	Length(basetypes.CollectionLengthOptions) int
}

func isKnownAutonomousDatabaseValue(value autonomousDatabaseValue) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func (r *resourceAutonomousDatabase) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan, state, config autonomousDatabaseResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &config))
	if resp.Diagnostics.HasError() {
		return
	}

	input := expandAutonomousDatabaseUpdateInput(ctx, plan, state, config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var updated *odbtypes.AutonomousDatabase
	if autonomousDatabaseUpdateInputHasChanges(input) {
		out, err := conn.UpdateAutonomousDatabase(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.AutonomousDatabaseID.ValueString())
			return
		}
		if out == nil || out.AutonomousDatabaseId == nil {
			err := tfresource.NewEmptyResultError()
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.AutonomousDatabaseID.ValueString())
			return
		}

		updated, err = waitAutonomousDatabaseUpdated(ctx, conn, state.AutonomousDatabaseID.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.AutonomousDatabaseID.ValueString())
			return
		}
	} else {
		var err error
		updated, err = findAutonomousDatabaseByID(ctx, conn, state.AutonomousDatabaseID.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.AutonomousDatabaseID.ValueString())
			return
		}
	}

	if updated.LongTermBackupSchedule == nil && isConfiguredAutonomousDatabaseBlock(plan.LongTermBackupSchedule) {
		schedule, diags := plan.LongTermBackupSchedule.ToPtr(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
		previous, diags := state.LongTermBackupSchedule.ToPtr(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
		if previous == nil {
			previous, diags = config.LongTermBackupSchedule.ToPtr(ctx)
			smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
		}
		if resp.Diagnostics.HasError() {
			return
		}
		if previous != nil {
			if schedule.IsDisabled.IsUnknown() {
				schedule.IsDisabled = previous.IsDisabled
			}
			if schedule.RepeatCadence.IsUnknown() {
				schedule.RepeatCadence = previous.RepeatCadence
			}
			if schedule.RetentionPeriodInDays.IsUnknown() {
				schedule.RetentionPeriodInDays = previous.RetentionPeriodInDays
			}
			if schedule.TimeOfBackup.IsUnknown() {
				schedule.TimeOfBackup = previous.TimeOfBackup
			}
			plan.LongTermBackupSchedule, diags = fwtypes.NewListNestedObjectValueOfPtr(ctx, schedule)
			smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	planTags, planTagsAll := plan.Tags, plan.TagsAll
	flattenAutonomousDatabase(ctx, updated, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Tags, plan.TagsAll = planTags, planTagsAll
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

// AWS applies every supplied update field. Omit unchanged values so unrelated updates
// do not replay computed settings or trigger additional database operations.
// nosemgrep:ci.semgrep.framework.manual-expander-functions
func expandAutonomousDatabaseUpdateInput(ctx context.Context, plan, state, config autonomousDatabaseResourceModel, diags *diag.Diagnostics) odb.UpdateAutonomousDatabaseInput {
	input := odb.UpdateAutonomousDatabaseInput{
		AutonomousDatabaseId: state.AutonomousDatabaseID.ValueStringPointer(),
	}
	smerr.AddEnrich(ctx, diags, flex.Expand(ctx, plan, &input))
	if diags.HasError() {
		return input
	}

	if !config.AdminPasswordWO.IsNull() && !plan.AdminPasswordWOVersion.Equal(state.AdminPasswordWOVersion) {
		input.AdminPassword = config.AdminPasswordWO.ValueStringPointer()
	} else if !plan.AdminPassword.Equal(state.AdminPassword) {
		input.AdminPassword = plan.AdminPassword.ValueStringPointer()
	}
	if !plan.AdminPasswordSource.Equal(state.AdminPasswordSource) {
		if plan.AdminPasswordSource.IsNull() {
			input.AdminPasswordSource = odbtypes.AdminPasswordSourceApiRequestParameter
		} else {
			input.AdminPasswordSource, input.AdminPasswordSourceConfiguration = expandAutonomousDatabaseAdminPasswordSource(ctx, plan.AdminPasswordSource, diags)
		}
	}
	if plan.AllowlistedIps.Equal(state.AllowlistedIps) {
		input.AllowlistedIps = nil
	}
	if plan.AutoRefreshFrequencyInSeconds.Equal(state.AutoRefreshFrequencyInSeconds) {
		input.AutoRefreshFrequencyInSeconds = nil
	}
	if plan.AutoRefreshPointLagInSeconds.Equal(state.AutoRefreshPointLagInSeconds) {
		input.AutoRefreshPointLagInSeconds = nil
	}
	if plan.AutonomousMaintenanceScheduleType.Equal(state.AutonomousMaintenanceScheduleType) {
		input.AutonomousMaintenanceScheduleType = ""
	}
	if plan.BackupRetentionPeriodInDays.Equal(state.BackupRetentionPeriodInDays) {
		input.BackupRetentionPeriodInDays = nil
	}
	if plan.ByolComputeCountLimit.Equal(state.ByolComputeCountLimit) {
		input.ByolComputeCountLimit = nil
	}
	if plan.ComputeCount.Equal(state.ComputeCount) {
		input.ComputeCount = nil
	}
	if plan.CpuCoreCount.Equal(state.CpuCoreCount) {
		input.CpuCoreCount = nil
	}
	if plan.CustomerContactsToSendToOCI.Equal(state.CustomerContactsToSendToOCI) {
		input.CustomerContactsToSendToOCI = nil
	} else if !plan.CustomerContactsToSendToOCI.IsUnknown() && !isConfiguredAutonomousDatabaseBlock(plan.CustomerContactsToSendToOCI) && isConfiguredAutonomousDatabaseBlock(state.CustomerContactsToSendToOCI) {
		input.CustomerContactsToSendToOCI = []odbtypes.CustomerContact{}
		tflog.Debug(ctx, "Clearing ODB Autonomous Database customer contacts")
	}
	if plan.DataStorageSizeInGBs.Equal(state.DataStorageSizeInGBs) {
		input.DataStorageSizeInGBs = nil
	}
	// AWS reads fractional TB values (for GB-sized storage), but accepts only integer TB writes.
	// A computed value must not be sent back when another attribute changes.
	if !config.DataStorageSizeInTBs.IsNull() && !plan.DataStorageSizeInTBs.Equal(state.DataStorageSizeInTBs) {
		var err error
		input.DataStorageSizeInTBs, err = expandAutonomousDatabaseDataStorageSizeInTBs(plan.DataStorageSizeInTBs)
		if err != nil {
			smerr.AddError(ctx, diags, err, smerr.ID, state.AutonomousDatabaseID.ValueString())
			return input
		}
	}
	if plan.DatabaseEdition.Equal(state.DatabaseEdition) {
		input.DatabaseEdition = ""
	}
	if plan.DbName.Equal(state.DbName) {
		input.DbName = nil
	}
	if plan.DbToolsDetails.Equal(state.DbToolsDetails) {
		input.DbToolsDetails = nil
	}
	if plan.DbVersion.Equal(state.DbVersion) {
		input.DbVersion = nil
	}
	if plan.DbWorkload.Equal(state.DbWorkload) {
		input.DbWorkload = ""
	}
	if plan.DisplayName.Equal(state.DisplayName) {
		input.DisplayName = nil
	}
	if plan.EncryptionKeyProvider.Equal(state.EncryptionKeyProvider) && plan.KMSKeyID.Equal(state.KMSKeyID) {
		input.EncryptionKeyProvider = ""
		input.EncryptionKeyConfiguration = nil
	} else {
		input.EncryptionKeyProvider, input.EncryptionKeyConfiguration = expandAutonomousDatabaseEncryption(plan.EncryptionKeyProvider, plan.KMSKeyID)
	}
	if plan.IsAutoScalingEnabled.Equal(state.IsAutoScalingEnabled) {
		input.IsAutoScalingEnabled = nil
	}
	if plan.IsAutoScalingForStorageEnabled.Equal(state.IsAutoScalingForStorageEnabled) {
		input.IsAutoScalingForStorageEnabled = nil
	}
	if plan.IsBackupRetentionLocked.Equal(state.IsBackupRetentionLocked) {
		input.IsBackupRetentionLocked = nil
	}
	if plan.IsLocalDataGuardEnabled.Equal(state.IsLocalDataGuardEnabled) {
		input.IsLocalDataGuardEnabled = nil
	}
	if plan.IsMtlsConnectionRequired.Equal(state.IsMtlsConnectionRequired) {
		input.IsMtlsConnectionRequired = nil
	}
	if plan.IsRefreshableClone.Equal(state.IsRefreshableClone) {
		input.IsRefreshableClone = nil
	}
	if plan.LicenseModel.Equal(state.LicenseModel) {
		input.LicenseModel = ""
	}
	if plan.LocalAdgAutoFailoverMaxDataLossLimit.Equal(state.LocalAdgAutoFailoverMaxDataLossLimit) {
		input.LocalAdgAutoFailoverMaxDataLossLimit = nil
	}
	if plan.LongTermBackupSchedule.Equal(state.LongTermBackupSchedule) {
		input.LongTermBackupSchedule = nil
	} else if !plan.LongTermBackupSchedule.IsUnknown() && !isConfiguredAutonomousDatabaseBlock(plan.LongTermBackupSchedule) && isConfiguredAutonomousDatabaseBlock(state.LongTermBackupSchedule) {
		input.LongTermBackupSchedule = &odbtypes.LongTermBackupSchedule{IsDisabled: aws.Bool(true)}
		tflog.Debug(ctx, "Disabling ODB Autonomous Database long-term backup schedule")
	}
	if plan.OpenMode.Equal(state.OpenMode) {
		input.OpenMode = ""
	}
	if plan.PermissionLevel.Equal(state.PermissionLevel) {
		input.PermissionLevel = ""
	}
	if plan.PrivateEndpointIp.Equal(state.PrivateEndpointIp) {
		input.PrivateEndpointIp = nil
	}
	if plan.PrivateEndpointLabel.Equal(state.PrivateEndpointLabel) {
		input.PrivateEndpointLabel = nil
	}
	if plan.RefreshableMode.Equal(state.RefreshableMode) {
		input.RefreshableMode = ""
	}
	if plan.ResourcePoolLeaderId.Equal(state.ResourcePoolLeaderId) {
		input.ResourcePoolLeaderId = nil
	}
	if plan.ResourcePoolSummary.Equal(state.ResourcePoolSummary) {
		input.ResourcePoolSummary = nil
	} else if !plan.ResourcePoolSummary.IsUnknown() && !isConfiguredAutonomousDatabaseBlock(plan.ResourcePoolSummary) && isConfiguredAutonomousDatabaseBlock(state.ResourcePoolSummary) {
		input.ResourcePoolSummary = &odbtypes.ResourcePoolSummary{IsDisabled: aws.Bool(true)}
		tflog.Debug(ctx, "Disabling ODB Autonomous Database resource pool")
	}
	if plan.ScheduledOperations.Equal(state.ScheduledOperations) {
		input.ScheduledOperations = nil
	} else if !plan.ScheduledOperations.IsUnknown() && !isConfiguredAutonomousDatabaseBlock(plan.ScheduledOperations) && isConfiguredAutonomousDatabaseBlock(state.ScheduledOperations) {
		input.ScheduledOperations = []odbtypes.ScheduledOperationDetails{}
		tflog.Debug(ctx, "Clearing ODB Autonomous Database scheduled operations")
	} else {
		input.ScheduledOperations = expandAutonomousDatabaseScheduledOperations(ctx, plan.ScheduledOperations, diags)
	}
	if plan.StandbyAllowlistedIps.Equal(state.StandbyAllowlistedIps) {
		input.StandbyAllowlistedIps = nil
	}
	if plan.StandbyAllowlistedIpsSource.Equal(state.StandbyAllowlistedIpsSource) {
		input.StandbyAllowlistedIpsSource = ""
	}
	if plan.TimeOfAutoRefreshStart.Equal(state.TimeOfAutoRefreshStart) {
		input.TimeOfAutoRefreshStart = nil
	}

	return input
}

func autonomousDatabaseUpdateInputHasChanges(input odb.UpdateAutonomousDatabaseInput) bool {
	input.AutonomousDatabaseId = nil
	return !reflect.DeepEqual(input, odb.UpdateAutonomousDatabaseInput{})
}

func (r *resourceAutonomousDatabase) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state autonomousDatabaseResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.AutonomousDatabaseID.ValueString()
	input := odb.DeleteAutonomousDatabaseInput{
		AutonomousDatabaseId: aws.String(id),
	}
	_, err := conn.DeleteAutonomousDatabase(ctx, &input)
	if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	err = waitAutonomousDatabaseDeleted(ctx, conn, id, r.DeleteTimeout(ctx, state.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
	}
}

func findAutonomousDatabaseByID(ctx context.Context, conn *odb.Client, id string) (*odbtypes.AutonomousDatabase, error) {
	out, err := findAutonomousDatabase(ctx, conn, id)
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out.Status == odbtypes.AutonomousDatabaseResourceStatusTerminated {
		tflog.Debug(ctx, "ODB Autonomous Database terminated; treating as absent", map[string]any{names.AttrID: id, names.AttrStatus: out.Status})
		return nil, smarterr.NewError(&retry.NotFoundError{Message: string(out.Status)})
	}
	return out, nil
}

// Create and update waiters need the raw TERMINATED status to report the failure reason.
func findAutonomousDatabase(ctx context.Context, conn *odb.Client, id string) (*odbtypes.AutonomousDatabase, error) {
	input := odb.GetAutonomousDatabaseInput{
		AutonomousDatabaseId: aws.String(id),
	}
	out, err := conn.GetAutonomousDatabase(ctx, &input)
	if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil || out.AutonomousDatabase == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.AutonomousDatabase, nil
}

func statusAutonomousDatabase(conn *odb.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findAutonomousDatabase(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
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
	return smarterr.Assert(waitAutonomousDatabaseReady(ctx, conn, id, timeout))
}

func waitAutonomousDatabaseUpdated(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) (*odbtypes.AutonomousDatabase, error) {
	return smarterr.Assert(waitAutonomousDatabaseReady(ctx, conn, id, timeout))
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
		return nil, smarterr.NewError(err)
	}
	out, ok := outputRaw.(*odbtypes.AutonomousDatabase)
	if !ok || out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	if slices.Contains(autonomousDatabaseFailureStatuses, string(out.Status)) {
		return out, smarterr.Errorf("autonomous database (%s) entered status %s: %s", id, out.Status, aws.ToString(out.StatusReason))
	}

	return out, nil
}

func waitAutonomousDatabaseDeleted(ctx context.Context, conn *odb.Client, id string, timeout time.Duration) error {
	refresh := statusAutonomousDatabase(conn, id)
	stateConf := &retry.StateChangeConf{
		Pending: append(append([]string{}, autonomousDatabasePendingStatuses...), autonomousDatabaseSuccessStatuses...),
		Target:  []string{},
		Refresh: func(ctx context.Context) (any, string, error) {
			out, status, err := refresh(ctx)
			if status == string(odbtypes.AutonomousDatabaseResourceStatusTerminated) {
				tflog.Debug(ctx, "ODB Autonomous Database deletion completed at terminal status", map[string]any{names.AttrID: id, names.AttrStatus: status})
				return nil, "", nil
			}
			return out, status, err
		},
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return smarterr.NewError(err)
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

// AdminPasswordSourceConfigurationInput is an SDK tagged union and requires explicit member selection.
// nosemgrep:ci.semgrep.framework.manual-expander-functions
func expandAutonomousDatabaseAdminPasswordSource(ctx context.Context, value fwtypes.ListNestedObjectValueOf[autonomousDatabaseAdminPasswordSourceModel], diags *diag.Diagnostics) (odbtypes.AdminPasswordSource, odbtypes.AdminPasswordSourceConfigurationInput) {
	if value.IsNull() || value.IsUnknown() {
		return "", nil
	}

	source, d := value.ToPtr(ctx)
	smerr.AddEnrich(ctx, diags, d)
	if diags.HasError() || source == nil || source.CustomerManagedAWSSecret.IsNull() {
		return "", nil
	}

	secret, d := source.CustomerManagedAWSSecret.ToPtr(ctx)
	smerr.AddEnrich(ctx, diags, d)
	if diags.HasError() || secret == nil {
		return "", nil
	}

	return odbtypes.AdminPasswordSourceCustomerManagedAwsSecret, &odbtypes.AdminPasswordSourceConfigurationInputMemberCustomerManagedAwsSecret{
		Value: odbtypes.CustomerManagedAwsSecretConfigurationInput{
			ExternalIdType: secret.ExternalIDType.ValueEnum(),
			IamRoleArn:     secret.IAMRoleARN.ValueStringPointer(),
			SecretId:       secret.SecretARN.ValueStringPointer(),
		},
	}
}

// Scheduled operations require explicit conversion because the SDK represents day_of_week as a nested structure.
// nosemgrep:ci.semgrep.framework.manual-expander-functions
func expandAutonomousDatabaseScheduledOperations(ctx context.Context, value fwtypes.ListNestedObjectValueOf[autonomousDatabaseScheduledOperationModel], diags *diag.Diagnostics) []odbtypes.ScheduledOperationDetails {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	models, d := value.ToSlice(ctx)
	smerr.AddEnrich(ctx, diags, d)
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

// Scheduled operations require explicit conversion because the SDK represents day_of_week as a nested structure.
// nosemgrep:ci.semgrep.framework.manual-flattener-functions
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

// SourceConfiguration is an SDK tagged union and requires explicit member selection.
// nosemgrep:ci.semgrep.framework.manual-expander-functions
func expandAutonomousDatabaseSourceConfiguration(ctx context.Context, value fwtypes.ListNestedObjectValueOf[autonomousDatabaseSourceConfigurationModel], diags *diag.Diagnostics) odbtypes.SourceConfiguration {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	configuration, d := value.ToPtr(ctx)
	smerr.AddEnrich(ctx, diags, d)
	if diags.HasError() || configuration == nil {
		return nil
	}

	if !configuration.CloneToRefreshable.IsNull() {
		model, d := configuration.CloneToRefreshable.ToPtr(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if diags.HasError() || model == nil {
			return nil
		}
		var apiObject odbtypes.CloneToRefreshableConfiguration
		smerr.AddEnrich(ctx, diags, flex.Expand(ctx, model, &apiObject))
		return &odbtypes.SourceConfigurationMemberCloneToRefreshable{Value: apiObject}
	}
	if !configuration.CrossRegionDataGuard.IsNull() {
		model, d := configuration.CrossRegionDataGuard.ToPtr(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if diags.HasError() || model == nil {
			return nil
		}
		var apiObject odbtypes.CrossRegionDataGuardConfiguration
		smerr.AddEnrich(ctx, diags, flex.Expand(ctx, model, &apiObject))
		return &odbtypes.SourceConfigurationMemberCrossRegionDataGuard{Value: apiObject}
	}
	if !configuration.CrossRegionDisasterRecovery.IsNull() {
		model, d := configuration.CrossRegionDisasterRecovery.ToPtr(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if diags.HasError() || model == nil {
			return nil
		}
		var apiObject odbtypes.CrossRegionDisasterRecoveryConfiguration
		smerr.AddEnrich(ctx, diags, flex.Expand(ctx, model, &apiObject))
		return &odbtypes.SourceConfigurationMemberCrossRegionDisasterRecovery{Value: apiObject}
	}
	if !configuration.DatabaseClone.IsNull() {
		model, d := configuration.DatabaseClone.ToPtr(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if diags.HasError() || model == nil {
			return nil
		}
		var apiObject odbtypes.DatabaseCloneConfiguration
		smerr.AddEnrich(ctx, diags, flex.Expand(ctx, model, &apiObject))
		return &odbtypes.SourceConfigurationMemberDatabaseClone{Value: apiObject}
	}
	if !configuration.PointInTimeRestore.IsNull() {
		model, d := configuration.PointInTimeRestore.ToPtr(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if diags.HasError() || model == nil {
			return nil
		}
		var apiObject odbtypes.PointInTimeRestoreConfiguration
		smerr.AddEnrich(ctx, diags, flex.Expand(ctx, model, &apiObject))
		return &odbtypes.SourceConfigurationMemberPointInTimeRestore{Value: apiObject}
	}
	if !configuration.RestoreFromBackup.IsNull() {
		model, d := configuration.RestoreFromBackup.ToPtr(ctx)
		smerr.AddEnrich(ctx, diags, d)
		if diags.HasError() || model == nil {
			return nil
		}
		var apiObject odbtypes.RestoreFromBackupConfiguration
		smerr.AddEnrich(ctx, diags, flex.Expand(ctx, model, &apiObject))
		return &odbtypes.SourceConfigurationMemberRestoreFromBackup{Value: apiObject}
	}

	return nil
}

// AutoFlex handles common fields; manual post-processing preserves configured blocks omitted by the API and normalizes SDK-specific values.
// nosemgrep:ci.semgrep.framework.manual-flattener-functions
func flattenAutonomousDatabase(ctx context.Context, apiObject *odbtypes.AutonomousDatabase, model *autonomousDatabaseResourceModel, diags *diag.Diagnostics) {
	customerContacts := model.CustomerContactsToSendToOCI
	dbToolsDetails := model.DbToolsDetails
	longTermBackupSchedule := model.LongTermBackupSchedule
	resourcePoolSummary := model.ResourcePoolSummary

	smerr.AddEnrich(ctx, diags, flex.Flatten(ctx, apiObject, model))
	if diags.HasError() {
		return
	}
	if !dbToolsDetails.IsUnknown() && dbToolsDetails.Length(fwtypes.CollectionLengthUnhandledAsZero) == 0 {
		model.DbToolsDetails = dbToolsDetails
	}
	if apiObject.LongTermBackupSchedule == nil && isConfiguredAutonomousDatabaseBlock(longTermBackupSchedule) {
		model.LongTermBackupSchedule = longTermBackupSchedule
	} else if !longTermBackupSchedule.IsUnknown() && !isConfiguredAutonomousDatabaseBlock(longTermBackupSchedule) && apiObject.LongTermBackupSchedule != nil && aws.ToBool(apiObject.LongTermBackupSchedule.IsDisabled) {
		model.LongTermBackupSchedule = longTermBackupSchedule
	}
	if !resourcePoolSummary.IsUnknown() && resourcePoolSummary.Length(fwtypes.CollectionLengthUnhandledAsZero) == 0 {
		model.ResourcePoolSummary = resourcePoolSummary
	}

	model.ByolComputeCountLimit = flattenAutonomousDatabaseByolComputeCountLimit(apiObject.ByolComputeCountLimit)
	model.KMSKeyID = types.StringNull()
	smerr.AddEnrich(ctx, diags, flattenAutonomousDatabaseAdminPasswordSource(ctx, apiObject.AdminPasswordSourceSummary, &model.AdminPasswordSource))
	if diags.HasError() {
		return
	}
	if len(apiObject.CustomerContacts) == 0 && isConfiguredAutonomousDatabaseBlock(customerContacts) {
		model.CustomerContactsToSendToOCI = customerContacts
	} else {
		smerr.AddEnrich(ctx, diags, flex.Flatten(ctx, apiObject.CustomerContacts, &model.CustomerContactsToSendToOCI))
	}
	smerr.AddEnrich(ctx, diags, flattenAutonomousDatabaseScheduledOperations(ctx, apiObject.ScheduledOperations, &model.ScheduledOperations))
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

// AdminPasswordSourceConfiguration is an SDK tagged union and requires explicit member selection.
// nosemgrep:ci.semgrep.framework.manual-flattener-functions
func flattenAutonomousDatabaseAdminPasswordSource(ctx context.Context, summary *odbtypes.AdminPasswordSourceSummary, target *fwtypes.ListNestedObjectValueOf[autonomousDatabaseAdminPasswordSourceModel]) diag.Diagnostics {
	var diags diag.Diagnostics

	*target = fwtypes.NewListNestedObjectValueOfNull[autonomousDatabaseAdminPasswordSourceModel](ctx)
	if summary == nil || summary.AdminPasswordSource != odbtypes.AdminPasswordSourceCustomerManagedAwsSecret {
		return diags
	}

	configuration, ok := summary.AdminPasswordSourceConfiguration.(*odbtypes.AdminPasswordSourceConfigurationMemberCustomerManagedAwsSecret)
	if !ok {
		return diags
	}

	value, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []autonomousDatabaseAdminPasswordSourceModel{
		{
			CustomerManagedAWSSecret: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []autonomousDatabaseCustomerManagedAWSSecretModel{
				{
					ExternalIDType: fwtypes.StringEnumValue(configuration.Value.ExternalIdType),
					IAMRoleARN:     types.StringPointerValue(configuration.Value.IamRoleArn),
					SecretARN:      types.StringPointerValue(configuration.Value.SecretId),
				},
			}),
		},
	})
	smerr.AddEnrich(ctx, &diags, d)
	*target = value
	return diags
}

func isConfiguredAutonomousDatabaseBlock(value autonomousDatabaseCollectionValue) bool {
	if value.IsNull() || value.IsUnknown() {
		return false
	}

	return value.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0
}

func flattenAutonomousDatabaseByolComputeCountLimit(value *int32) types.Float64 {
	if value == nil {
		return types.Float64Null()
	}

	return types.Float64Value(float64(*value))
}

// The SDK reads float64 TB values but only accepts int32 on Create and Update.
// nosemgrep:ci.semgrep.framework.manual-expander-functions
func expandAutonomousDatabaseDataStorageSizeInTBs(value types.Float64) (*int32, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}
	v := value.ValueFloat64()
	if v < 1 || v > 384 || math.Trunc(v) != v {
		return nil, smarterr.Errorf("data_storage_size_in_tbs must be a whole number between 1 and 384; use data_storage_size_in_gbs for fractional TB sizes")
	}
	return aws.Int32(int32(v)), nil
}

type autonomousDatabaseResourceModel struct {
	framework.WithRegionModel
	ActualUsedDataStorageSizeInTBs       types.Float64                                                                   `tfsdk:"actual_used_data_storage_size_in_tbs"`
	AdminPassword                        types.String                                                                    `tfsdk:"admin_password" autoflex:"-"`
	AdminPasswordWO                      types.String                                                                    `tfsdk:"admin_password_wo" autoflex:"-"`
	AdminPasswordWOVersion               types.Int64                                                                     `tfsdk:"admin_password_wo_version" autoflex:"-"`
	AdminPasswordSource                  fwtypes.ListNestedObjectValueOf[autonomousDatabaseAdminPasswordSourceModel]     `tfsdk:"admin_password_source" autoflex:"-"`
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
	ByolComputeCountLimit                types.Float64                                                                   `tfsdk:"byol_compute_count_limit" autoflex:",noflatten"`
	CharacterSet                         types.String                                                                    `tfsdk:"character_set"`
	ComputeCount                         types.Float64                                                                   `tfsdk:"compute_count"`
	ComputeModel                         fwtypes.StringEnum[odbtypes.ComputeModel]                                       `tfsdk:"compute_model"`
	CpuCoreCount                         types.Int32                                                                     `tfsdk:"cpu_core_count"`
	CreatedAt                            timetypes.RFC3339                                                               `tfsdk:"created_at"`
	CustomerContactsToSendToOCI          fwtypes.ListNestedObjectValueOf[autonomousDatabaseCustomerContactModel]         `tfsdk:"customer_contacts_to_send_to_oci" autoflex:",noflatten"`
	DataStorageSizeInGBs                 types.Int32                                                                     `tfsdk:"data_storage_size_in_gbs"`
	DataStorageSizeInTBs                 types.Float64                                                                   `tfsdk:"data_storage_size_in_tbs" autoflex:",noexpand"`
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

type autonomousDatabaseAdminPasswordSourceModel struct {
	CustomerManagedAWSSecret fwtypes.ListNestedObjectValueOf[autonomousDatabaseCustomerManagedAWSSecretModel] `tfsdk:"customer_managed_aws_secret"`
}

type autonomousDatabaseCustomerManagedAWSSecretModel struct {
	ExternalIDType fwtypes.StringEnum[odbtypes.ExternalIdType] `tfsdk:"external_id_type"`
	IAMRoleARN     types.String                                `tfsdk:"iam_role_arn"`
	SecretARN      types.String                                `tfsdk:"secret_arn"`
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
