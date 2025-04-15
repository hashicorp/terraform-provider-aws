// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	clusterScalingConfiguration_DefaultMinCapacity = 1
	clusterScalingConfiguration_DefaultMaxCapacity = 16
)

// @SDKResource("aws_rds_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		UpdateWithoutTimeout: resourceClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceClusterResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: clusterStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrAllocatedStorage: {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrAllowMajorVersionUpgrade: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			// apply_immediately is used to determine when the update modifications take place.
			// See http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html
			names.AttrApplyImmediately: {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"backup_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtMost(35),
			},
			"backtrack_window": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 259200),
			},
			"ca_certificate_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ca_certificate_valid_till": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrClusterIdentifier: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validIdentifier,
				ConflictsWith: []string{"cluster_identifier_prefix"},
			},
			"cluster_identifier_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validIdentifierPrefix,
				ConflictsWith: []string{names.AttrClusterIdentifier},
			},
			"cluster_members": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"cluster_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_scalability_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.ClusterScalabilityType](),
			},
			"copy_tags_to_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"database_insights_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.DatabaseInsightsMode](),
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"db_cluster_instance_class": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"db_cluster_parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"db_instance_parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"db_subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"db_system_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"delete_automated_backups": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrDeletionProtection: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrDomain: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain_iam_role_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enable_global_write_forwarding": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_http_endpoint": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_local_write_forwarding": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enabled_cloudwatch_logs_exports": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(clusterExportableLogType_Values(), false),
				},
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.StringMatch(regexache.MustCompile(fmt.Sprintf(`^%s.*$`, InstanceEngineCustomPrefix)), fmt.Sprintf("must begin with %s", InstanceEngineCustomPrefix)),
					validation.StringInSlice(clusterEngine_Values(), false),
				),
			},
			"engine_lifecycle_support": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(engineLifecycleSupport_Values(), false),
			},
			"engine_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      engineModeProvisioned,
				ValidateFunc: validation.StringInSlice(append(engineMode_Values(), ""), false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == engineModeProvisioned && new == "" {
						return true
					}
					return new == old
				},
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"engine_version_actual": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFinalSnapshotIdentifier: {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: func(v any, k string) (ws []string, es []error) {
					value := v.(string)
					if !regexache.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
						es = append(es, fmt.Errorf(
							"only alphanumeric characters and hyphens allowed in %q", k))
					}
					if regexache.MustCompile(`--`).MatchString(value) {
						es = append(es, fmt.Errorf("%q cannot contain two consecutive hyphens", k))
					}
					if regexache.MustCompile(`-$`).MatchString(value) {
						es = append(es, fmt.Errorf("%q cannot end in a hyphen", k))
					}
					return
				},
			},
			"global_cluster_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_database_authentication_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"iam_roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrIOPS: {
				Type:     schema.TypeInt,
				Optional: true,
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"manage_master_user_password": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"master_password", "master_password_wo"},
			},
			"master_user_secret": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"secret_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"secret_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"master_user_secret_kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidKMSKeyID,
			},
			"master_password": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"manage_master_user_password", "master_password_wo"},
			},
			"master_password_wo": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				WriteOnly:     true,
				ConflictsWith: []string{"manage_master_user_password", "master_password"},
				RequiredWith:  []string{"master_password_wo_version"},
			},
			"master_password_wo_version": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"master_password_wo"},
			},
			"master_username": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"monitoring_interval": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"monitoring_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"network_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(networkType_Values(), false),
			},
			"performance_insights_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"performance_insights_kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"performance_insights_retention_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"preferred_backup_window": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidOnceADayWindowFormat,
			},
			names.AttrPreferredMaintenanceWindow: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(val any) string {
					if val == nil {
						return ""
					}
					return strings.ToLower(val.(string))
				},
				ValidateFunc: verify.ValidOnceAWeekWindowFormat,
			},
			"reader_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_source_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"restore_to_point_in_time": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"restore_to_time": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidUTCTimestamp,
							ExactlyOneOf: []string{
								"restore_to_point_in_time.0.restore_to_time",
								"restore_to_point_in_time.0.use_latest_restorable_time",
							},
						},
						"restore_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(restoreType_Values(), false),
						},
						"source_cluster_identifier": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.Any(
								verify.ValidARN,
								validIdentifier,
							),
							ExactlyOneOf: []string{
								"restore_to_point_in_time.0.source_cluster_identifier",
								"restore_to_point_in_time.0.source_cluster_resource_id",
							},
						},
						"source_cluster_resource_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ExactlyOneOf: []string{
								"restore_to_point_in_time.0.source_cluster_identifier",
								"restore_to_point_in_time.0.source_cluster_resource_id",
							},
						},
						"use_latest_restorable_time": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
							ExactlyOneOf: []string{
								"restore_to_point_in_time.0.restore_to_time",
								"restore_to_point_in_time.0.use_latest_restorable_time",
							},
						},
					},
				},
				ConflictsWith: []string{
					"s3_import",
					"snapshot_identifier",
				},
			},
			"s3_import": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucketName: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						names.AttrBucketPrefix: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"ingestion_role": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"source_engine": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"source_engine_version": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
				ConflictsWith: []string{
					"snapshot_identifier",
					"restore_to_point_in_time",
				},
			},
			"scaling_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_pause": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						names.AttrMaxCapacity: {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  clusterScalingConfiguration_DefaultMaxCapacity,
						},
						"min_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  clusterScalingConfiguration_DefaultMinCapacity,
						},
						"seconds_before_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      300,
							ValidateFunc: validation.IntBetween(60, 600),
						},
						"seconds_until_auto_pause": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      300,
							ValidateFunc: validation.IntBetween(300, 86400),
						},
						"timeout_action": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      timeoutActionRollbackCapacityChange,
							ValidateFunc: validation.StringInSlice(timeoutAction_Values(), false),
						},
					},
				},
			},
			"serverlessv2_scaling_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrMaxCapacity: {
							Type:         schema.TypeFloat,
							Required:     true,
							ValidateFunc: validation.FloatBetween(1, 256),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// Handles a breaking regression. On v5.79.0 and earlier,
								// serverlessv2_scaling_configuration block could be removed from
								// configuration. Although doing so doesn't actually remove the scaling
								// configuration from AWS this dsf does allow the user to remove the block
								// from their configuration without perpetual diffs.
								// https://github.com/hashicorp/terraform-provider-aws/issues/40473
								config := d.GetRawConfig()
								raw := config.GetAttr("serverlessv2_scaling_configuration")

								if raw.LengthInt() == 0 && (old != "0" && old != "") && (new == "0" || new == "") {
									return true
								}
								return false
							},
						},
						"min_capacity": {
							Type:         schema.TypeFloat,
							Required:     true,
							ValidateFunc: validation.FloatBetween(0, 256),
						},
						"seconds_until_auto_pause": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(300, 86400),
						},
					},
				},
			},
			"skip_final_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"snapshot_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ConflictsWith: []string{
					// Clusters cannot be joined to an existing global cluster as part of
					// the "restore from snapshot" operation. Trigger an error during plan
					// to prevent an apply with unexpected results (ie. a regional
					// cluster which is not joined to the provided global cluster).
					"global_cluster_identifier",
				},
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// allow snapshot_idenfitier to be removed without forcing re-creation
					return new == ""
				},
			},
			"source_region": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrStorageType: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIf(names.AttrStorageType, func(_ context.Context, d *schema.ResourceDiff, meta any) bool {
				// Aurora supports mutation of the storage_type parameter, other engines do not
				return !strings.HasPrefix(d.Get(names.AttrEngine).(string), "aurora")
			}),
			func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
				if diff.Id() == "" {
					return nil
				}
				// The control plane will always return an empty string if a cluster is created with a storage_type of aurora
				old, new := diff.GetChange(names.AttrStorageType)

				if new.(string) == "aurora" && old.(string) == "" {
					if err := diff.SetNew(names.AttrStorageType, ""); err != nil {
						return err
					}
					return nil
				}
				return nil
			},
		),
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	identifier := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrClusterIdentifier).(string)),
		create.WithConfiguredPrefix(d.Get("cluster_identifier_prefix").(string)),
		create.WithDefaultPrefix("tf-"),
	).Generate()

	// get write-only value from configuration
	masterPasswordWO, di := flex.GetWriteOnlyStringValue(d, cty.GetAttrPath("master_password_wo"))
	diags = append(diags, di...)
	if diags.HasError() {
		return diags
	}

	// Some API calls (e.g. RestoreDBClusterFromSnapshot do not support all
	// parameters to correctly apply all settings in one pass. For missing
	// parameters or unsupported configurations, we may need to call
	// ModifyDBInstance afterwards to prevent Terraform operators from API
	// errors or needing to double apply.
	var requiresModifyDbCluster bool
	modifyDbClusterInput := &rds.ModifyDBClusterInput{
		ApplyImmediately: aws.Bool(true),
	}

	if v, ok := d.GetOk("snapshot_identifier"); ok {
		input := &rds.RestoreDBClusterFromSnapshotInput{
			CopyTagsToSnapshot:  aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBClusterIdentifier: aws.String(identifier),
			DeletionProtection:  aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			Engine:              aws.String(d.Get(names.AttrEngine).(string)),
			SnapshotIdentifier:  aws.String(v.(string)),
			Tags:                getTagsIn(ctx),
		}

		if v, ok := d.GetOk(names.AttrAvailabilityZones); ok && v.(*schema.Set).Len() > 0 {
			input.AvailabilityZones = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("backtrack_window"); ok {
			input.BacktrackWindow = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			modifyDbClusterInput.BackupRetentionPeriod = aws.Int32(int32(v.(int)))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk(names.AttrDatabaseName); ok {
			input.DatabaseName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			input.DBClusterParameterGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrDomain); ok {
			input.Domain = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_iam_role_name"); ok {
			input.DomainIAMRoleName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("engine_lifecycle_support"); ok {
			input.EngineLifecycleSupport = aws.String(v.(string))
		}

		if v, ok := d.GetOk("engine_mode"); ok {
			input.EngineMode = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrEngineVersion); ok {
			input.EngineVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input.KmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("manage_master_user_password"); ok {
			modifyDbClusterInput.ManageMasterUserPassword = aws.Bool(v.(bool))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("master_password"); ok {
			modifyDbClusterInput.MasterUserPassword = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if masterPasswordWO != "" {
			modifyDbClusterInput.MasterUserPassword = aws.String(masterPasswordWO)
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
			modifyDbClusterInput.MasterUserSecretKmsKeyId = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("monitoring_interval"); ok {
			input.MonitoringInterval = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("monitoring_role_arn"); ok {
			input.MonitoringRoleArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("network_type"); ok {
			input.NetworkType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("option_group_name"); ok {
			input.OptionGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrPort); ok {
			input.Port = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("preferred_backup_window"); ok {
			modifyDbClusterInput.PreferredBackupWindow = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk(names.AttrPreferredMaintenanceWindow); ok {
			modifyDbClusterInput.PreferredMaintenanceWindow = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("scaling_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.ScalingConfiguration = expandScalingConfiguration(v.([]any)[0].(map[string]any))
		}

		if v, ok := d.GetOk("serverlessv2_scaling_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			modifyDbClusterInput.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]any)[0].(map[string]any))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout,
			func() (any, error) {
				return conn.RestoreDBClusterFromSnapshot(ctx, input)
			},
			errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating RDS Cluster (restore from snapshot) (%s): %s", identifier, err)
		}
	} else if v, ok := d.GetOk("s3_import"); ok {
		if _, ok := d.GetOk("master_username"); !ok {
			diags = sdkdiag.AppendErrorf(diags, `"master_username": required field is not set`)
		}
		if diags.HasError() {
			return diags
		}

		tfMap := v.([]any)[0].(map[string]any)
		input := &rds.RestoreDBClusterFromS3Input{
			CopyTagsToSnapshot:  aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBClusterIdentifier: aws.String(identifier),
			DeletionProtection:  aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			Engine:              aws.String(d.Get(names.AttrEngine).(string)),
			MasterUsername:      aws.String(d.Get("master_username").(string)),
			S3BucketName:        aws.String(tfMap[names.AttrBucketName].(string)),
			S3IngestionRoleArn:  aws.String(tfMap["ingestion_role"].(string)),
			S3Prefix:            aws.String(tfMap[names.AttrBucketPrefix].(string)),
			SourceEngine:        aws.String(tfMap["source_engine"].(string)),
			SourceEngineVersion: aws.String(tfMap["source_engine_version"].(string)),
			Tags:                getTagsIn(ctx),
		}

		if v, ok := d.GetOk(names.AttrAvailabilityZones); ok && v.(*schema.Set).Len() > 0 {
			input.AvailabilityZones = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("backtrack_window"); ok {
			input.BacktrackWindow = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			input.BackupRetentionPeriod = aws.Int32(int32(v.(int)))
		}

		if v := d.Get(names.AttrDatabaseName); v.(string) != "" {
			input.DatabaseName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			input.DBClusterParameterGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrDomain); ok {
			input.Domain = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_iam_role_name"); ok {
			input.DomainIAMRoleName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("engine_lifecycle_support"); ok {
			input.EngineLifecycleSupport = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrEngineVersion); ok {
			input.EngineVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			input.EnableIAMDatabaseAuthentication = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input.KmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("manage_master_user_password"); ok {
			input.ManageMasterUserPassword = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
			input.MasterUserSecretKmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("master_password"); ok {
			input.MasterUserPassword = aws.String(v.(string))
		}

		if masterPasswordWO != "" {
			input.MasterUserPassword = aws.String(masterPasswordWO)
		}

		if v, ok := d.GetOk("network_type"); ok {
			input.NetworkType = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrPort); ok {
			input.Port = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("preferred_backup_window"); ok {
			input.PreferredBackupWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrPreferredMaintenanceWindow); ok {
			input.PreferredMaintenanceWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOkExists(names.AttrStorageEncrypted); ok {
			input.StorageEncrypted = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		_, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func() (any, error) {
				return conn.RestoreDBClusterFromS3(ctx, input)
			},
			func(err error) (bool, error) {
				// InvalidParameterValue: Files from the specified Amazon S3 bucket cannot be downloaded.
				// Make sure that you have created an AWS Identity and Access Management (IAM) role that lets Amazon RDS access Amazon S3 for you.
				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "Files from the specified Amazon S3 bucket cannot be downloaded") {
					return true, err
				}

				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "S3_SNAPSHOT_INGESTION") {
					return true, err
				}

				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "S3 bucket cannot be found") {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating RDS Cluster (restore from S3) (%s): %s", identifier, err)
		}
	} else if v, ok := d.GetOk("restore_to_point_in_time"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		tfMap := v.([]any)[0].(map[string]any)
		input := &rds.RestoreDBClusterToPointInTimeInput{
			CopyTagsToSnapshot:  aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBClusterIdentifier: aws.String(identifier),
			DeletionProtection:  aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			Tags:                getTagsIn(ctx),
		}

		if v, ok := d.GetOk("backtrack_window"); ok {
			input.BacktrackWindow = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			modifyDbClusterInput.BackupRetentionPeriod = aws.Int32(int32(v.(int)))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			input.DBClusterParameterGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrDomain); ok {
			input.Domain = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_iam_role_name"); ok {
			input.DomainIAMRoleName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			input.EnableIAMDatabaseAuthentication = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input.KmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("manage_master_user_password"); ok {
			modifyDbClusterInput.ManageMasterUserPassword = aws.Bool(v.(bool))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("master_password"); ok {
			modifyDbClusterInput.MasterUserPassword = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if masterPasswordWO != "" {
			modifyDbClusterInput.MasterUserPassword = aws.String(masterPasswordWO)
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
			modifyDbClusterInput.MasterUserSecretKmsKeyId = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("monitoring_interval"); ok {
			input.MonitoringInterval = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("monitoring_role_arn"); ok {
			input.MonitoringRoleArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("network_type"); ok {
			input.NetworkType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("option_group_name"); ok {
			input.OptionGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrPort); ok {
			input.Port = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("preferred_backup_window"); ok {
			modifyDbClusterInput.PreferredBackupWindow = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk(names.AttrPreferredMaintenanceWindow); ok {
			modifyDbClusterInput.PreferredMaintenanceWindow = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := tfMap["restore_to_time"].(string); ok && v != "" {
			v, _ := time.Parse(time.RFC3339, v)
			input.RestoreToTime = aws.Time(v)
		}

		if v, ok := tfMap["restore_type"].(string); ok {
			input.RestoreType = aws.String(v)
		}

		if v, ok := d.GetOk("scaling_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			modifyDbClusterInput.ScalingConfiguration = expandScalingConfiguration(v.([]any)[0].(map[string]any))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("serverlessv2_scaling_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			modifyDbClusterInput.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]any)[0].(map[string]any))
			requiresModifyDbCluster = true
		}

		if v, ok := tfMap["source_cluster_identifier"].(string); ok && v != "" {
			input.SourceDBClusterIdentifier = aws.String(v)
		}

		if v, ok := tfMap["source_cluster_resource_id"].(string); ok && v != "" {
			input.SourceDbClusterResourceId = aws.String(v)
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v, ok := tfMap["use_latest_restorable_time"].(bool); ok && v {
			input.UseLatestRestorableTime = aws.Bool(v)
		}

		if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		_, err := conn.RestoreDBClusterToPointInTime(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating RDS Cluster (restore to point-in-time) (%s): %s", identifier, err)
		}
	} else {
		input := &rds.CreateDBClusterInput{
			CopyTagsToSnapshot:  aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBClusterIdentifier: aws.String(identifier),
			DeletionProtection:  aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			Engine:              aws.String(d.Get(names.AttrEngine).(string)),
			Tags:                getTagsIn(ctx),
		}

		if v, ok := d.GetOkExists(names.AttrAllocatedStorage); ok {
			input.AllocatedStorage = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk(names.AttrAvailabilityZones); ok && v.(*schema.Set).Len() > 0 {
			input.AvailabilityZones = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("backtrack_window"); ok {
			input.BacktrackWindow = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			input.BackupRetentionPeriod = aws.Int32(int32(v.(int)))
		}

		if v := d.Get("ca_certificate_identifier"); v.(string) != "" {
			input.CACertificateIdentifier = aws.String(v.(string))
		}

		if v := d.Get("cluster_scalability_type"); v.(string) != "" {
			input.ClusterScalabilityType = types.ClusterScalabilityType(v.(string))
		}

		if v := d.Get("database_insights_mode"); v.(string) != "" {
			input.DatabaseInsightsMode = types.DatabaseInsightsMode(v.(string))
		}

		if v := d.Get(names.AttrDatabaseName); v.(string) != "" {
			input.DatabaseName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_cluster_instance_class"); ok {
			input.DBClusterInstanceClass = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			input.DBClusterParameterGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_system_id"); ok {
			input.DBSystemId = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrDomain); ok {
			input.Domain = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_iam_role_name"); ok {
			input.DomainIAMRoleName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enable_global_write_forwarding"); ok {
			input.EnableGlobalWriteForwarding = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("enable_http_endpoint"); ok {
			input.EnableHttpEndpoint = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("enable_local_write_forwarding"); ok {
			input.EnableLocalWriteForwarding = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("engine_lifecycle_support"); ok {
			input.EngineLifecycleSupport = aws.String(v.(string))
		}

		if v, ok := d.GetOk("engine_mode"); ok {
			input.EngineMode = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrEngineVersion); ok {
			input.EngineVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("global_cluster_identifier"); ok {
			input.GlobalClusterIdentifier = aws.String(v.(string))
		}

		if v, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			input.EnableIAMDatabaseAuthentication = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOkExists(names.AttrIOPS); ok {
			input.Iops = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input.KmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("manage_master_user_password"); ok {
			input.ManageMasterUserPassword = aws.Bool(v.(bool))
		}

		// Note: Username and password credentials are required and valid
		// unless the cluster password is managed by RDS, or it is a read-replica.
		// This also applies to clusters within a global cluster.
		// Providing a password and/or username for a replica
		// will result in an InvalidParameterValue error.
		if v, ok := d.GetOk("master_password"); ok {
			input.MasterUserPassword = aws.String(v.(string))
		}

		if masterPasswordWO != "" {
			input.MasterUserPassword = aws.String(masterPasswordWO)
		}

		if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
			input.MasterUserSecretKmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("master_username"); ok {
			input.MasterUsername = aws.String(v.(string))
		}

		if v, ok := d.GetOk("monitoring_interval"); ok {
			input.MonitoringInterval = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("monitoring_role_arn"); ok {
			input.MonitoringRoleArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("network_type"); ok {
			input.NetworkType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("performance_insights_enabled"); ok {
			// If the cluster is part of a global cluster, defer Performance Insights settings
			// to the modifyDbClusterInput to prevent them from being reset.
			if _, ok := d.GetOk("global_cluster_identifier"); ok {
				modifyDbClusterInput.EnablePerformanceInsights = aws.Bool(v.(bool))
				requiresModifyDbCluster = true
			} else {
				input.EnablePerformanceInsights = aws.Bool(v.(bool))
			}
		}

		if v, ok := d.GetOk("performance_insights_kms_key_id"); ok {
			if _, ok := d.GetOk("global_cluster_identifier"); ok {
				modifyDbClusterInput.PerformanceInsightsKMSKeyId = aws.String(v.(string))
				requiresModifyDbCluster = true
			} else {
				input.PerformanceInsightsKMSKeyId = aws.String(v.(string))
			}
		}

		if v, ok := d.GetOk("performance_insights_retention_period"); ok {
			if _, ok := d.GetOk("global_cluster_identifier"); ok {
				modifyDbClusterInput.PerformanceInsightsRetentionPeriod = aws.Int32(int32(v.(int)))
				requiresModifyDbCluster = true
			} else {
				input.PerformanceInsightsRetentionPeriod = aws.Int32(int32(v.(int)))
			}
		}

		if v, ok := d.GetOk(names.AttrPort); ok {
			input.Port = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("preferred_backup_window"); ok {
			input.PreferredBackupWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrPreferredMaintenanceWindow); ok {
			input.PreferredMaintenanceWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk("replication_source_identifier"); ok && input.GlobalClusterIdentifier == nil {
			input.ReplicationSourceIdentifier = aws.String(v.(string))
		}

		if v, ok := d.GetOk("scaling_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.ScalingConfiguration = expandScalingConfiguration(v.([]any)[0].(map[string]any))
		}

		if v, ok := d.GetOk("serverlessv2_scaling_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]any)[0].(map[string]any))
		}

		if v, ok := d.GetOk("source_region"); ok {
			input.SourceRegion = aws.String(v.(string))
		}

		if v, ok := d.GetOkExists(names.AttrStorageEncrypted); ok {
			input.StorageEncrypted = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout,
			func() (any, error) {
				return conn.CreateDBCluster(ctx, input)
			},
			errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating RDS Cluster (%s): %s", identifier, err)
		}
	}

	d.SetId(identifier)

	if _, err := waitDBClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("iam_roles"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range v.(*schema.Set).List() {
			if err := addIAMRoleToCluster(ctx, conn, d.Id(), v.(string)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if requiresModifyDbCluster {
		modifyDbClusterInput.DBClusterIdentifier = aws.String(d.Id())

		_, err := conn.ModifyDBCluster(ctx, modifyDbClusterInput)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS Cluster (%s): %s", d.Id(), err)
		}

		if _, err := waitDBClusterUpdated(ctx, conn, d.Id(), true, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbc, err := findDBClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAllocatedStorage, dbc.AllocatedStorage)
	clusterARN := aws.ToString(dbc.DBClusterArn)
	d.Set(names.AttrARN, clusterARN)
	d.Set(names.AttrAvailabilityZones, dbc.AvailabilityZones)
	d.Set("backtrack_window", dbc.BacktrackWindow)
	d.Set("backup_retention_period", dbc.BackupRetentionPeriod)
	if dbc.CertificateDetails != nil {
		d.Set("ca_certificate_identifier", dbc.CertificateDetails.CAIdentifier)
		d.Set("ca_certificate_valid_till", dbc.CertificateDetails.ValidTill.Format(time.RFC3339))
	}
	d.Set("cluster_scalability_type", dbc.ClusterScalabilityType)
	d.Set(names.AttrClusterIdentifier, dbc.DBClusterIdentifier)
	d.Set("cluster_identifier_prefix", create.NamePrefixFromName(aws.ToString(dbc.DBClusterIdentifier)))
	d.Set("cluster_members", tfslices.ApplyToAll(dbc.DBClusterMembers, func(v types.DBClusterMember) string {
		return aws.ToString(v.DBInstanceIdentifier)
	}))
	d.Set("cluster_resource_id", dbc.DbClusterResourceId)
	d.Set("copy_tags_to_snapshot", dbc.CopyTagsToSnapshot)
	d.Set("database_insights_mode", dbc.DatabaseInsightsMode)
	// Only set the DatabaseName if it is not nil. There is a known API bug where
	// RDS accepts a DatabaseName but does not return it, causing a perpetual
	// diff.
	//	See https://github.com/hashicorp/terraform/issues/4671 for backstory
	if dbc.DatabaseName != nil { // nosemgrep: ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set(names.AttrDatabaseName, dbc.DatabaseName)
	}
	d.Set("db_cluster_instance_class", dbc.DBClusterInstanceClass)
	d.Set("db_cluster_parameter_group_name", dbc.DBClusterParameterGroup)
	d.Set("db_subnet_group_name", dbc.DBSubnetGroup)
	d.Set("db_system_id", dbc.DBSystemId)
	d.Set(names.AttrDeletionProtection, dbc.DeletionProtection)
	if len(dbc.DomainMemberships) > 0 {
		domainMembership := dbc.DomainMemberships[0]
		d.Set(names.AttrDomain, domainMembership.Domain)
		d.Set("domain_iam_role_name", domainMembership.IAMRoleName)
	} else {
		d.Set(names.AttrDomain, nil)
		d.Set("domain_iam_role_name", nil)
	}
	d.Set("enabled_cloudwatch_logs_exports", dbc.EnabledCloudwatchLogsExports)
	d.Set("enable_http_endpoint", dbc.HttpEndpointEnabled)
	d.Set(names.AttrEndpoint, dbc.Endpoint)
	d.Set(names.AttrEngine, dbc.Engine)
	d.Set("engine_lifecycle_support", dbc.EngineLifecycleSupport)
	d.Set("engine_mode", dbc.EngineMode)
	clusterSetResourceDataEngineVersionFromCluster(d, dbc)
	d.Set(names.AttrHostedZoneID, dbc.HostedZoneId)
	d.Set("iam_database_authentication_enabled", dbc.IAMDatabaseAuthenticationEnabled)
	d.Set("iam_roles", tfslices.ApplyToAll(dbc.AssociatedRoles, func(v types.DBClusterRole) string {
		return aws.ToString(v.RoleArn)
	}))
	d.Set(names.AttrIOPS, dbc.Iops)
	d.Set(names.AttrKMSKeyID, dbc.KmsKeyId)
	// Note: the following attributes are not returned by the API
	// when conducting a read after a create, so we rely on Terraform's
	// implicit state passthrough, and they are treated as virtual attributes.
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#implicit-state-passthrough
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#virtual-attributes
	//
	// manage_master_user_password
	// master_password
	//
	// Expose the MasterUserSecret structure as a computed attribute
	// https://awscli.amazonaws.com/v2/documentation/api/latest/reference/rds/create-db-cluster.html#:~:text=for%20future%20use.-,MasterUserSecret,-%2D%3E%20(structure)
	if dbc.MasterUserSecret != nil {
		if err := d.Set("master_user_secret", []any{flattenManagedMasterUserSecret(dbc.MasterUserSecret)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting master_user_secret: %s", err)
		}
	} else {
		d.Set("master_user_secret", nil)
	}
	d.Set("master_username", dbc.MasterUsername)
	d.Set("monitoring_interval", dbc.MonitoringInterval)
	d.Set("monitoring_role_arn", dbc.MonitoringRoleArn)
	d.Set("network_type", dbc.NetworkType)
	d.Set("performance_insights_enabled", dbc.PerformanceInsightsEnabled)
	d.Set("performance_insights_kms_key_id", dbc.PerformanceInsightsKMSKeyId)
	d.Set("performance_insights_retention_period", dbc.PerformanceInsightsRetentionPeriod)
	d.Set(names.AttrPort, dbc.Port)
	d.Set("preferred_backup_window", dbc.PreferredBackupWindow)
	d.Set(names.AttrPreferredMaintenanceWindow, dbc.PreferredMaintenanceWindow)
	d.Set("reader_endpoint", dbc.ReaderEndpoint)
	d.Set("replication_source_identifier", dbc.ReplicationSourceIdentifier)
	if dbc.ScalingConfigurationInfo != nil {
		if err := d.Set("scaling_configuration", []any{flattenScalingConfigurationInfo(dbc.ScalingConfigurationInfo)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting scaling_configuration: %s", err)
		}
	} else {
		d.Set("scaling_configuration", nil)
	}
	if dbc.ServerlessV2ScalingConfiguration != nil {
		if err := d.Set("serverlessv2_scaling_configuration", []any{flattenServerlessV2ScalingConfigurationInfo(dbc.ServerlessV2ScalingConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting serverlessv2_scaling_configuration: %s", err)
		}
	} else {
		d.Set("serverlessv2_scaling_configuration", nil)
	}
	d.Set(names.AttrStorageEncrypted, dbc.StorageEncrypted)
	d.Set(names.AttrStorageType, dbc.StorageType)
	d.Set(names.AttrVPCSecurityGroupIDs, tfslices.ApplyToAll(dbc.VpcSecurityGroups, func(v types.VpcSecurityGroupMembership) string {
		return aws.ToString(v.VpcSecurityGroupId)
	}))

	// Fetch and save Global Cluster if engine mode global
	d.Set("global_cluster_identifier", "")

	if engineMode := aws.ToString(dbc.EngineMode); engineMode == engineModeGlobal || engineMode == engineModeProvisioned {
		globalCluster, err := findGlobalClusterByDBClusterARN(ctx, conn, clusterARN)

		if err == nil {
			d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
		} else if tfresource.NotFound(err) || tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "Access Denied to API Version: APIGlobalDatabases") { //nolint:revive // Keep comments
			// Ignore the following API error for regions/partitions that do not support RDS Global Clusters:
			// InvalidParameterValue: Access Denied to API Version: APIGlobalDatabases
		} else {
			return sdkdiag.AppendErrorf(diags, "reading RDS Global Cluster for RDS Cluster (%s): %s", d.Id(), err)
		}
	}

	setTagsOut(ctx, dbc.TagList)

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	// There are two ways to enable the HTTP endpoint: new way and old way.
	// This is the new way for provisioned engine mode (covers provisioned and serverlessv2).
	// The old way is modifying the DB cluster and setting the EnableHttpEndpoint field (below).
	// Both need a wait for update so when it's provisioned it will do old (not necessary but does the wait) & new ways, otherwise just old way.
	if d.HasChange("enable_http_endpoint") && d.Get("engine_mode").(string) == engineModeProvisioned {
		o, n := d.GetChange("enable_http_endpoint")
		if err := enableHTTPEndpointProvisioned(ctx, conn, d.Get(names.AttrARN).(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling HTTP endpoint for RDS Cluster (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("replication_source_identifier") {
		if d.Get("replication_source_identifier").(string) == "" {
			input := rds.PromoteReadReplicaDBClusterInput{
				DBClusterIdentifier: aws.String(d.Id()),
			}
			_, err := conn.PromoteReadReplicaDBCluster(ctx, &input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "promoting read replica to primary for RDS Cluster (%s): %s", d.Id(), err)
			}

			if _, err := waitDBClusterAvailable(ctx, conn, d.Id(), false, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster (%s) update: %s", d.Id(), err)
			}
		} else {
			return sdkdiag.AppendErrorf(diags, "promoting to standalone is not supported for RDS Cluster (%s)", d.Id())
		}
	}

	if d.HasChangesExcept(
		names.AttrAllowMajorVersionUpgrade,
		"delete_automated_backups",
		names.AttrFinalSnapshotIdentifier,
		"global_cluster_identifier",
		"iam_roles",
		"replication_source_identifier",
		"skip_final_snapshot",
		names.AttrTags, names.AttrTagsAll) {
		applyImmediately := d.Get(names.AttrApplyImmediately).(bool)
		input := &rds.ModifyDBClusterInput{
			ApplyImmediately:    aws.Bool(applyImmediately),
			DBClusterIdentifier: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrAllocatedStorage) {
			input.AllocatedStorage = aws.Int32(int32(d.Get(names.AttrAllocatedStorage).(int)))
		}

		if v, ok := d.GetOk(names.AttrAllowMajorVersionUpgrade); ok {
			input.AllowMajorVersionUpgrade = aws.Bool(v.(bool))
		}

		if d.HasChange("backtrack_window") {
			input.BacktrackWindow = aws.Int64(int64(d.Get("backtrack_window").(int)))
		}

		if d.HasChange("backup_retention_period") {
			input.BackupRetentionPeriod = aws.Int32(int32(d.Get("backup_retention_period").(int)))
		}

		if d.HasChange("ca_certificate_identifier") {
			input.CACertificateIdentifier = aws.String(d.Get("ca_certificate_identifier").(string))
		}

		if d.HasChange("copy_tags_to_snapshot") {
			input.CopyTagsToSnapshot = aws.Bool(d.Get("copy_tags_to_snapshot").(bool))
		}

		if d.HasChange("database_insights_mode") {
			input.DatabaseInsightsMode = types.DatabaseInsightsMode(d.Get("database_insights_mode").(string))
			input.EnablePerformanceInsights = aws.Bool(d.Get("performance_insights_enabled").(bool))
		}

		if d.HasChange("db_cluster_instance_class") {
			input.DBClusterInstanceClass = aws.String(d.Get("db_cluster_instance_class").(string))
		}

		if d.HasChange("db_cluster_parameter_group_name") {
			input.DBClusterParameterGroupName = aws.String(d.Get("db_cluster_parameter_group_name").(string))
		}

		// DB instance parameter group name is not currently returned from the
		// DescribeDBClusters API. This means there is no drift detection, so when
		// set, the configured attribute should always be sent on modify.
		// Except, this causes an error on a minor version upgrade, so it is
		// removed during update retry, if necessary.
		if v, ok := d.GetOk("db_instance_parameter_group_name"); ok || d.HasChange("db_instance_parameter_group_name") {
			input.DBInstanceParameterGroupName = aws.String(v.(string))
		}

		if d.HasChange(names.AttrDeletionProtection) {
			input.DeletionProtection = aws.Bool(d.Get(names.AttrDeletionProtection).(bool))
		}

		if d.HasChanges(names.AttrDomain, "domain_iam_role_name") {
			input.Domain = aws.String(d.Get(names.AttrDomain).(string))
			input.DomainIAMRoleName = aws.String(d.Get("domain_iam_role_name").(string))
		}

		if d.HasChange("enable_global_write_forwarding") {
			input.EnableGlobalWriteForwarding = aws.Bool(d.Get("enable_global_write_forwarding").(bool))
		}

		// for provisioned and serverlessv2 (also "provisioned"), data api must be enabled using conn.EnableHttpEndpoint() as below
		if d.HasChange("enable_http_endpoint") && d.Get("engine_mode").(string) != engineModeProvisioned {
			input.EnableHttpEndpoint = aws.Bool(d.Get("enable_http_endpoint").(bool))
		}

		if d.HasChange("enable_local_write_forwarding") {
			input.EnableLocalWriteForwarding = aws.Bool(d.Get("enable_local_write_forwarding").(bool))
		}

		if d.HasChange("enabled_cloudwatch_logs_exports") {
			o, n := d.GetChange("enabled_cloudwatch_logs_exports")
			os, ns := o.(*schema.Set), n.(*schema.Set)

			input.CloudwatchLogsExportConfiguration = &types.CloudwatchLogsExportConfiguration{
				DisableLogTypes: flex.ExpandStringValueSet(os.Difference(ns)),
				EnableLogTypes:  flex.ExpandStringValueSet(ns.Difference(os)),
			}
		}

		if d.HasChange(names.AttrEngineVersion) {
			input.EngineVersion = aws.String(d.Get(names.AttrEngineVersion).(string))
		}

		// This can happen when updates are deferred (apply_immediately = false), and
		// multiple applies occur before the maintenance window. In this case,
		// continue sending the desired engine_version as part of the modify request.
		if d.Get(names.AttrEngineVersion).(string) != d.Get("engine_version_actual").(string) {
			input.EngineVersion = aws.String(d.Get(names.AttrEngineVersion).(string))
		}

		if d.HasChange("iam_database_authentication_enabled") {
			input.EnableIAMDatabaseAuthentication = aws.Bool(d.Get("iam_database_authentication_enabled").(bool))
		}

		if d.HasChange(names.AttrIOPS) {
			input.Iops = aws.Int32(int32(d.Get(names.AttrIOPS).(int)))
		}

		if d.HasChange("manage_master_user_password") {
			input.ManageMasterUserPassword = aws.Bool(d.Get("manage_master_user_password").(bool))
		}

		if d.HasChange("master_password") {
			if v, ok := d.GetOk("master_password"); ok {
				input.MasterUserPassword = aws.String(v.(string))
			}
		}

		if d.HasChange("master_password_wo_version") {
			masterPasswordWO, di := flex.GetWriteOnlyStringValue(d, cty.GetAttrPath("master_password_wo"))
			diags = append(diags, di...)
			if diags.HasError() {
				return diags
			}

			if masterPasswordWO != "" {
				input.MasterUserPassword = aws.String(masterPasswordWO)
			}
		}

		if d.HasChange("master_user_secret_kms_key_id") {
			if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
				input.MasterUserSecretKmsKeyId = aws.String(v.(string))
			}
		}

		if d.HasChange("monitoring_interval") {
			input.MonitoringInterval = aws.Int32(int32(d.Get("monitoring_interval").(int)))
		}

		if d.HasChange("monitoring_role_arn") {
			input.MonitoringRoleArn = aws.String(d.Get("monitoring_role_arn").(string))
		}

		if d.HasChange("network_type") {
			input.NetworkType = aws.String(d.Get("network_type").(string))
		}

		if d.HasChange("performance_insights_enabled") {
			input.EnablePerformanceInsights = aws.Bool(d.Get("performance_insights_enabled").(bool))
		}

		if d.HasChange("performance_insights_kms_key_id") {
			input.PerformanceInsightsKMSKeyId = aws.String(d.Get("performance_insights_kms_key_id").(string))
		}

		if d.HasChange("performance_insights_retention_period") {
			input.PerformanceInsightsRetentionPeriod = aws.Int32(int32(d.Get("performance_insights_retention_period").(int)))
		}

		if d.HasChange(names.AttrPort) {
			input.Port = aws.Int32(int32(d.Get(names.AttrPort).(int)))
		}

		if d.HasChange("preferred_backup_window") {
			input.PreferredBackupWindow = aws.String(d.Get("preferred_backup_window").(string))
		}

		if d.HasChange(names.AttrPreferredMaintenanceWindow) {
			input.PreferredMaintenanceWindow = aws.String(d.Get(names.AttrPreferredMaintenanceWindow).(string))
		}

		if d.HasChange("scaling_configuration") {
			if v, ok := d.GetOk("scaling_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.ScalingConfiguration = expandScalingConfiguration(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChange("serverlessv2_scaling_configuration") {
			if v, ok := d.GetOk("serverlessv2_scaling_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChange(names.AttrStorageType) {
			input.StorageType = aws.String(d.Get(names.AttrStorageType).(string))
		}

		if d.HasChange(names.AttrVPCSecurityGroupIDs) {
			if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
				input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
			} else {
				input.VpcSecurityGroupIds = []string{}
			}
		}

		const (
			timeout = 5 * time.Minute
		)
		_, err := tfresource.RetryWhen(ctx, timeout,
			func() (any, error) {
				return conn.ModifyDBCluster(ctx, input)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
					return true, err
				}

				if errs.IsA[*types.InvalidDBClusterStateFault](err) {
					return true, err
				}

				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterCombination, "db-instance-parameter-group-name can only be specified for a major") {
					input.DBInstanceParameterGroupName = nil
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS Cluster (%s): %s", d.Id(), err)
		}

		if _, err := waitDBClusterUpdated(ctx, conn, d.Id(), applyImmediately, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("global_cluster_identifier") {
		o, n := d.GetChange("global_cluster_identifier")
		os, ns := o.(string), n.(string)

		if os == "" {
			return sdkdiag.AppendErrorf(diags, "existing RDS Clusters cannot be added to an existing RDS Global Cluster")
		}

		if ns != "" {
			return sdkdiag.AppendErrorf(diags, "existing RDS Clusters cannot be migrated between existing RDS Global Clusters")
		}

		clusterARN := d.Get(names.AttrARN).(string)
		input := &rds.RemoveFromGlobalClusterInput{
			DbClusterIdentifier:     aws.String(clusterARN),
			GlobalClusterIdentifier: aws.String(os),
		}

		_, err := conn.RemoveFromGlobalCluster(ctx, input)

		if err != nil && !errs.IsA[*types.GlobalClusterNotFoundFault](err) && !tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "is not found in global cluster") {
			return sdkdiag.AppendErrorf(diags, "removing RDS Cluster (%s) from RDS Global Cluster: %s", d.Id(), err)
		}

		// Removal from a global cluster puts the cluster into 'promoting' state. Wait for it to become available again.
		if _, err := waitDBClusterAvailable(ctx, conn, d.Id(), true, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster (%s) available: %s", d.Id(), err)
		}
	}

	if d.HasChange("iam_roles") {
		o, n := d.GetChange("iam_roles")
		os, ns := o.(*schema.Set), n.(*schema.Set)

		for _, v := range ns.Difference(os).List() {
			if err := addIAMRoleToCluster(ctx, conn, d.Id(), v.(string)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		for _, v := range os.Difference(ns).List() {
			if err := removeIAMRoleFromCluster(ctx, conn, d.Id(), v.(string)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	// Automatically remove from global cluster to bypass this error on deletion:
	// InvalidDBClusterStateFault: This cluster is a part of a global cluster, please remove it from globalcluster first
	if globalClusterID := d.Get("global_cluster_identifier").(string); globalClusterID != "" {
		clusterARN := d.Get(names.AttrARN).(string)
		input := &rds.RemoveFromGlobalClusterInput{
			DbClusterIdentifier:     aws.String(clusterARN),
			GlobalClusterIdentifier: aws.String(globalClusterID),
		}

		_, err := conn.RemoveFromGlobalCluster(ctx, input)

		if err != nil && !errs.IsA[*types.GlobalClusterNotFoundFault](err) && !tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "is not found in global cluster") {
			return sdkdiag.AppendErrorf(diags, "removing RDS Cluster (%s) from RDS Global Cluster (%s): %s", d.Id(), globalClusterID, err)
		}

		if _, err := waitDBClusterAvailable(ctx, conn, d.Id(), true, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster (%s) available: %s", d.Id(), err)
		}
	}

	skipFinalSnapshot := d.Get("skip_final_snapshot").(bool)
	input := &rds.DeleteDBClusterInput{
		DBClusterIdentifier:    aws.String(d.Id()),
		DeleteAutomatedBackups: aws.Bool(d.Get("delete_automated_backups").(bool)),
		SkipFinalSnapshot:      aws.Bool(skipFinalSnapshot),
	}

	if !skipFinalSnapshot {
		if v, ok := d.GetOk(names.AttrFinalSnapshotIdentifier); ok {
			input.FinalDBSnapshotIdentifier = aws.String(v.(string))
		} else {
			return sdkdiag.AppendErrorf(diags, "RDS Cluster final_snapshot_identifier is required when skip_final_snapshot is false")
		}
	}

	log.Printf("[DEBUG] Deleting RDS Cluster: %s", d.Id())
	const (
		timeout = 2 * time.Minute
	)
	_, err := tfresource.RetryWhen(ctx, timeout,
		func() (any, error) {
			return conn.DeleteDBCluster(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, errCodeInvalidParameterCombination, "disable deletion pro") {
				if v, ok := d.GetOk(names.AttrDeletionProtection); (!ok || !v.(bool)) && d.Get(names.AttrApplyImmediately).(bool) {
					_, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutDelete),
						func() (any, error) {
							return conn.ModifyDBCluster(ctx, &rds.ModifyDBClusterInput{
								ApplyImmediately:    aws.Bool(true),
								DBClusterIdentifier: aws.String(d.Id()),
								DeletionProtection:  aws.Bool(false),
							})
						},
						func(err error) (bool, error) {
							if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
								return true, err
							}

							if errs.IsA[*types.InvalidDBClusterStateFault](err) {
								return true, err
							}

							return false, err
						},
					)

					if err != nil {
						return false, fmt.Errorf("modifying RDS Cluster (%s) DeletionProtection=false: %s", d.Id(), err)
					}

					if _, err := waitDBClusterUpdated(ctx, conn, d.Id(), false, d.Timeout(schema.TimeoutDelete)); err != nil {
						return false, fmt.Errorf("waiting for RDS Cluster (%s) update: %s", d.Id(), err)
					}
				}

				return true, err
			}

			if errs.IsAErrorMessageContains[*types.InvalidDBClusterStateFault](err, "is not currently in the available state") {
				return true, err
			}

			if errs.IsAErrorMessageContains[*types.InvalidDBClusterStateFault](err, "cluster is a part of a global cluster") {
				return true, err
			}

			return false, err
		},
	)

	if errs.IsA[*types.DBClusterNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitDBClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceClusterImport(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	// Neither skip_final_snapshot nor final_snapshot_identifier can be fetched
	// from any API call, so we need to default skip_final_snapshot to true so
	// that final_snapshot_identifier is not required
	d.Set("skip_final_snapshot", true)
	d.Set("delete_automated_backups", true)
	return []*schema.ResourceData{d}, nil
}

func enableHTTPEndpointProvisioned(ctx context.Context, conn *rds.Client, arn string, o, n any) error {
	if o == nil {
		return nil
	}

	if n == nil {
		return nil
	}

	if !o.(bool) && n.(bool) {
		_, err := conn.EnableHttpEndpoint(ctx, &rds.EnableHttpEndpointInput{
			ResourceArn: aws.String(arn),
		})
		if err != nil {
			return err
		}
	}

	if o.(bool) && !n.(bool) {
		_, err := conn.DisableHttpEndpoint(ctx, &rds.DisableHttpEndpointInput{
			ResourceArn: aws.String(arn),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func addIAMRoleToCluster(ctx context.Context, conn *rds.Client, clusterID, roleARN string) error {
	input := &rds.AddRoleToDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		RoleArn:             aws.String(roleARN),
	}

	_, err := conn.AddRoleToDBCluster(ctx, input)

	if err != nil {
		return fmt.Errorf("adding IAM Role (%s) to RDS Cluster (%s): %w", roleARN, clusterID, err)
	}

	return nil
}

func removeIAMRoleFromCluster(ctx context.Context, conn *rds.Client, clusterID, roleARN string) error {
	input := &rds.RemoveRoleFromDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		RoleArn:             aws.String(roleARN),
	}

	_, err := conn.RemoveRoleFromDBCluster(ctx, input)

	if err != nil {
		return fmt.Errorf("removing IAM Role (%s) from RDS Cluster (%s): %w", roleARN, clusterID, err)
	}

	return err
}

func clusterSetResourceDataEngineVersionFromCluster(d *schema.ResourceData, c *types.DBCluster) {
	oldVersion := d.Get(names.AttrEngineVersion).(string)
	newVersion := aws.ToString(c.EngineVersion)
	var pendingVersion string
	if c.PendingModifiedValues != nil && c.PendingModifiedValues.EngineVersion != nil {
		pendingVersion = aws.ToString(c.PendingModifiedValues.EngineVersion)
	}
	compareActualEngineVersion(d, oldVersion, newVersion, pendingVersion)
}

func findDBClusterByID(ctx context.Context, conn *rds.Client, id string, optFns ...func(*rds.Options)) (*types.DBCluster, error) {
	input := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(id),
	}
	output, err := findDBCluster(ctx, conn, input, tfslices.PredicateTrue[*types.DBCluster](), optFns...)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if arn.IsARN(id) {
		if aws.ToString(output.DBClusterArn) != id {
			return nil, &retry.NotFoundError{
				LastRequest: input,
			}
		}
	} else if aws.ToString(output.DBClusterIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBCluster(ctx context.Context, conn *rds.Client, input *rds.DescribeDBClustersInput, filter tfslices.Predicate[*types.DBCluster], optFns ...func(*rds.Options)) (*types.DBCluster, error) {
	output, err := findDBClusters(ctx, conn, input, filter, optFns...)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBClusters(ctx context.Context, conn *rds.Client, input *rds.DescribeDBClustersInput, filter tfslices.Predicate[*types.DBCluster], optFns ...func(*rds.Options)) ([]types.DBCluster, error) {
	var output []types.DBCluster

	pages := rds.NewDescribeDBClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx, optFns...)

		if errs.IsA[*types.DBClusterNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusters {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusDBCluster(ctx context.Context, conn *rds.Client, id string, waitNoPendingModifiedValues bool, optFns ...func(*rds.Options)) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDBClusterByID(ctx, conn, id, optFns...)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		status := aws.ToString(output.Status)

		if status == clusterStatusAvailable && waitNoPendingModifiedValues && !itypes.IsZero(output.PendingModifiedValues) {
			status = clusterStatusAvailableWithPendingModifiedValues
		}

		return output, status, nil
	}
}

func waitDBClusterAvailable(ctx context.Context, conn *rds.Client, id string, waitNoPendingModifiedValues bool, timeout time.Duration) (*types.DBCluster, error) { //nolint:unparam
	pendingStatuses := []string{
		clusterStatusBackingUp,
		clusterStatusConfiguringIAMDatabaseAuth,
		clusterStatusConfiguringEnhancedMonitoring,
		clusterStatusCreating,
		clusterStatusMigrating,
		clusterStatusModifying,
		clusterStatusPreparingDataMigration,
		clusterStatusPromoting,
		clusterStatusRebooting,
		clusterStatusRenaming,
		clusterStatusResettingMasterCredentials,
		clusterStatusScalingCompute,
		clusterStatusUpgrading,
	}

	stateConf := &retry.StateChangeConf{
		Pending:    pendingStatuses,
		Target:     []string{clusterStatusAvailable},
		Refresh:    statusDBCluster(ctx, conn, id, waitNoPendingModifiedValues),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterCreated(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.DBCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			clusterStatusBackingUp,
			clusterStatusCreating,
			clusterStatusMigrating,
			clusterStatusModifying,
			clusterStatusPreparingDataMigration,
			clusterStatusRebooting,
			clusterStatusResettingMasterCredentials,
		},
		Target:     []string{clusterStatusAvailable},
		Refresh:    statusDBCluster(ctx, conn, id, false),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterUpdated(ctx context.Context, conn *rds.Client, id string, waitNoPendingModifiedValues bool, timeout time.Duration) (*types.DBCluster, error) { //nolint:unparam
	pendingStatuses := []string{
		clusterStatusBackingUp,
		clusterStatusConfiguringIAMDatabaseAuth,
		clusterStatusConfiguringEnhancedMonitoring,
		clusterStatusModifying,
		clusterStatusRenaming,
		clusterStatusResettingMasterCredentials,
		clusterStatusScalingCompute,
		clusterStatusScalingStorage,
		clusterStatusUpgrading,
	}
	if waitNoPendingModifiedValues {
		pendingStatuses = append(pendingStatuses, clusterStatusAvailableWithPendingModifiedValues)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    pendingStatuses,
		Target:     []string{clusterStatusAvailable},
		Refresh:    statusDBCluster(ctx, conn, id, waitNoPendingModifiedValues),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterDeleted(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.DBCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			clusterStatusAvailable,
			clusterStatusBackingUp,
			clusterStatusDeleting,
			clusterStatusModifying,
			clusterStatusPromoting,
			clusterStatusScalingCompute,
		},
		Target:     []string{},
		Refresh:    statusDBCluster(ctx, conn, id, false),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func expandScalingConfiguration(tfMap map[string]any) *types.ScalingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ScalingConfiguration{}

	if v, ok := tfMap["auto_pause"].(bool); ok {
		apiObject.AutoPause = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrMaxCapacity].(int); ok {
		apiObject.MaxCapacity = aws.Int32(int32(v))
	}

	if v, ok := tfMap["min_capacity"].(int); ok {
		apiObject.MinCapacity = aws.Int32(int32(v))
	}

	if v, ok := tfMap["seconds_before_timeout"].(int); ok {
		apiObject.SecondsBeforeTimeout = aws.Int32(int32(v))
	}

	if v, ok := tfMap["seconds_until_auto_pause"].(int); ok {
		apiObject.SecondsUntilAutoPause = aws.Int32(int32(v))
	}

	if v, ok := tfMap["timeout_action"].(string); ok && v != "" {
		apiObject.TimeoutAction = aws.String(v)
	}

	return apiObject
}

func flattenScalingConfigurationInfo(apiObject *types.ScalingConfigurationInfo) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AutoPause; v != nil {
		tfMap["auto_pause"] = aws.ToBool(v)
	}

	if v := apiObject.MaxCapacity; v != nil {
		tfMap[names.AttrMaxCapacity] = aws.ToInt32(v)
	}

	if v := apiObject.MaxCapacity; v != nil {
		tfMap[names.AttrMaxCapacity] = aws.ToInt32(v)
	}

	if v := apiObject.MinCapacity; v != nil {
		tfMap["min_capacity"] = aws.ToInt32(v)
	}

	if v := apiObject.SecondsBeforeTimeout; v != nil {
		tfMap["seconds_before_timeout"] = aws.ToInt32(v)
	}

	if v := apiObject.SecondsUntilAutoPause; v != nil {
		tfMap["seconds_until_auto_pause"] = aws.ToInt32(v)
	}

	if v := apiObject.TimeoutAction; v != nil {
		tfMap["timeout_action"] = aws.ToString(v)
	}

	return tfMap
}

func expandServerlessV2ScalingConfiguration(tfMap map[string]any) *types.ServerlessV2ScalingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ServerlessV2ScalingConfiguration{}

	// Removing the serverlessv2_scaling_configuration block from config for an update, sets max_capacity
	// to 0. Sending 0 to the API causes an error.
	if v, ok := tfMap[names.AttrMaxCapacity].(float64); ok && v != 0.0 {
		apiObject.MaxCapacity = aws.Float64(v)
	}

	if v, ok := tfMap["min_capacity"].(float64); ok {
		apiObject.MinCapacity = aws.Float64(v)
	}

	if v, ok := tfMap["seconds_until_auto_pause"].(int); ok && v != 0 {
		apiObject.SecondsUntilAutoPause = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenServerlessV2ScalingConfigurationInfo(apiObject *types.ServerlessV2ScalingConfigurationInfo) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.MaxCapacity; v != nil {
		tfMap[names.AttrMaxCapacity] = aws.ToFloat64(v)
	}

	if v := apiObject.MinCapacity; v != nil {
		tfMap["min_capacity"] = aws.ToFloat64(v)
	}

	if v := apiObject.SecondsUntilAutoPause; v != nil {
		tfMap["seconds_until_auto_pause"] = aws.ToInt32(v)
	}

	return tfMap
}
