package rds

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwboolplanmodifier "github.com/hashicorp/terraform-provider-aws/internal/framework/boolplanmodifier"
	fwint64planmodifier "github.com/hashicorp/terraform-provider-aws/internal/framework/int64planmodifier"
	fwstringplanmodifier "github.com/hashicorp/terraform-provider-aws/internal/framework/stringplanmodifier"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func clusterSchemaFramework(ctx context.Context, version int64) schema.Schema {
	s := schema.Schema{
		Version: version,
		Attributes: map[string]schema.Attribute{
			"allocated_storage": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"allow_major_version_upgrade": schema.BoolAttribute{
				Optional: true,
			},
			"apply_immediately": schema.BoolAttribute{
				Optional: true,
			},
			"arn": framework.ARNAttributeComputedOnly(),
			"availability_zones": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"backtrack_window": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(0, 259200),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"backup_retention_period": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					fwint64planmodifier.DefaultValue(1),
				},
				Validators: []validator.Int64{
					int64validator.AtMost(35),
				},
			},
			"cluster_identifier": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					fwvalidators.ClusterIdentifier(),
				},
			},
			"cluster_identifier_prefix": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					fwvalidators.ClusterIdentifierPrefix(),
				},
			},
			"cluster_members": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_resource_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"copy_tags_to_snapshot": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					fwboolplanmodifier.DefaultValue(false),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"database_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"db_cluster_instance_class": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"db_cluster_parameter_group_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"db_cluster_parameter_group_name_actual": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"db_instance_parameter_group_name": schema.StringAttribute{
				Optional: true,
			},
			"db_subnet_group_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deletion_protection": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enable_global_write_forwarding": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					fwboolplanmodifier.DefaultValue(false),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enable_http_endpoint": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					fwboolplanmodifier.DefaultValue(false),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled_cloudwatch_logs_exports": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.OneOf(ClusterExportableLogType_Values()...)),
				},
			},
			"endpoint": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"engine": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					fwstringplanmodifier.DefaultValue(ClusterEngineAurora),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(ClusterEngine_Values()...),
				},
			},
			"engine_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					fwstringplanmodifier.DefaultValue(EngineModeProvisioned),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(EngineMode_Values()...),
				},
			},
			"engine_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"engine_version_actual": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"final_snapshot_identifier": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					fwvalidators.ClusterFinalSnapshotIdentifier(),
				},
			},
			"global_cluster_identifier": schema.StringAttribute{
				Optional: true,
			},
			"hosted_zone_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"iam_database_authentication_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"iam_roles": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"id": framework.IDAttribute(),
			"iops": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"kms_key_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
			},
			"master_password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"master_username": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"master_username_actual": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf(NetworkType_Values()...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"option_group_name": schema.StringAttribute{
				Optional: true,
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"preferred_backup_window": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					fwvalidators.OnceADayWindowFormat(),
				},
			},
			"preferred_maintenance_window": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					fwvalidators.OnceAWeekWindowFormat(),
				},
			},
			"reader_endpoint": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"replication_source_identifier": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"scaling_configuration_actual": schema.ListAttribute{
				ElementType: types.ObjectType{AttrTypes: scalingConfigurationAttrTypes},
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"skip_final_snapshot": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					fwboolplanmodifier.DefaultValue(false),
				},
			},
			"snapshot_identifier": schema.StringAttribute{
				Optional: true,
			},
			"source_region": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"storage_encrypted": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"storage_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags":     tftags.TagsAttribute(),
			"tags_all": tftags.TagsAttributeComputedOnly(),
			"vpc_security_group_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"restore_to_point_in_time": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"restore_to_time": schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								fwvalidators.UTCTimestamp(),
							},
						},
						"restore_type": schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(RestoreType_Values()...),
							},
						},
						"source_cluster_identifier": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.Any(
									fwvalidators.ClusterIdentifier(),
									fwvalidators.ARN(),
								),
							},
						},
						"use_latest_restorable_time": schema.BoolAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"s3_import": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"bucket_name": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"bucket_prefix": schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"ingestion_role": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"source_engine": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"source_engine_version": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"scaling_configuration": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"auto_pause": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Bool{
								fwboolplanmodifier.DefaultValue(true),
							},
						},
						"max_capacity": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								fwint64planmodifier.DefaultValue(clusterScalingConfiguration_DefaultMaxCapacity),
							},
						},
						"min_capacity": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								fwint64planmodifier.DefaultValue(clusterScalingConfiguration_DefaultMinCapacity),
							},
						},
						"seconds_until_auto_pause": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								fwint64planmodifier.DefaultValue(300), //nolint:gomnd
							},
							Validators: []validator.Int64{
								int64validator.Between(300, 86400),
							},
						},
						"timeout_action": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								fwstringplanmodifier.DefaultValue(TimeoutActionRollbackCapacityChange),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(TimeoutAction_Values()...),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"serverlessv2_scaling_configuration": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_capacity": schema.Float64Attribute{
							Required: true,
							Validators: []validator.Float64{
								float64validator.Between(0.5, 128), //nolint:gomnd
							},
						},
						"min_capacity": schema.Float64Attribute{
							Required: true,
							Validators: []validator.Float64{
								float64validator.Between(0.5, 128), //nolint:gomnd
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
		},
	}

	if s.Blocks == nil {
		s.Blocks = make(map[string]schema.Block)
	}
	s.Blocks["timeouts"] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Update: true,
		Delete: true,
	})

	return s
}
