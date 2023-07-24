// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	clusterScalingConfiguration_DefaultMinCapacity = 1
	clusterScalingConfiguration_DefaultMaxCapacity = 16
	clusterTimeoutDelete                           = 2 * time.Minute
)

// @SDKResource("aws_rds_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
func ResourceCluster() *schema.Resource {
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

		Schema: map[string]*schema.Schema{
			"allocated_storage": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"allow_major_version_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			// apply_immediately is used to determine when the update modifications take place.
			// See http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html
			"apply_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zones": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"backup_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntAtMost(35),
			},
			"backtrack_window": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 259200),
			},
			"cluster_identifier": {
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
				ConflictsWith: []string{"cluster_identifier"},
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
			"copy_tags_to_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"database_name": {
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
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"enable_global_write_forwarding": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enabled_cloudwatch_logs_exports": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(ClusterExportableLogType_Values(), false),
				},
			},
			"enable_http_endpoint": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterEngine(),
			},
			"engine_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      EngineModeProvisioned,
				ValidateFunc: validation.StringInSlice(EngineMode_Values(), false),
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"engine_version_actual": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"final_snapshot_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
					value := v.(string)
					if !regexp.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
						es = append(es, fmt.Errorf(
							"only alphanumeric characters and hyphens allowed in %q", k))
					}
					if regexp.MustCompile(`--`).MatchString(value) {
						es = append(es, fmt.Errorf("%q cannot contain two consecutive hyphens", k))
					}
					if regexp.MustCompile(`-$`).MatchString(value) {
						es = append(es, fmt.Errorf("%q cannot end in a hyphen", k))
					}
					return
				},
			},
			"global_cluster_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hosted_zone_id": {
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
			"iops": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"manage_master_user_password": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"master_password"},
			},
			"master_user_secret": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_id": {
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
				ConflictsWith: []string{"manage_master_user_password"},
			},
			"master_username": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"network_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(NetworkType_Values(), false),
			},
			"port": {
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
			"preferred_maintenance_window": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(val interface{}) string {
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
							Type:          schema.TypeString,
							Optional:      true,
							ForceNew:      true,
							ValidateFunc:  verify.ValidUTCTimestamp,
							ConflictsWith: []string{"restore_to_point_in_time.0.use_latest_restorable_time"},
						},
						"restore_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(RestoreType_Values(), false),
						},
						"source_cluster_identifier": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.Any(
								verify.ValidARN,
								validIdentifier,
							),
						},
						"use_latest_restorable_time": {
							Type:          schema.TypeBool,
							Optional:      true,
							ForceNew:      true,
							ConflictsWith: []string{"restore_to_point_in_time.0.restore_to_time"},
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
						"bucket_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"bucket_prefix": {
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
						"max_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  clusterScalingConfiguration_DefaultMaxCapacity,
						},
						"min_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  clusterScalingConfiguration_DefaultMinCapacity,
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
							Default:      TimeoutActionRollbackCapacityChange,
							ValidateFunc: validation.StringInSlice(TimeoutAction_Values(), false),
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
						"max_capacity": {
							Type:         schema.TypeFloat,
							Required:     true,
							ValidateFunc: validation.FloatBetween(0.5, 128),
						},
						"min_capacity": {
							Type:         schema.TypeFloat,
							Required:     true,
							ValidateFunc: validation.FloatBetween(0.5, 128),
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
			"storage_encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"storage_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			customdiff.ForceNewIf("storage_type", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				// Aurora supports mutation of the storage_type parameter, other engines do not
				return !strings.HasPrefix(d.Get("engine").(string), "aurora")
			}),
			func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
				if diff.Id() == "" {
					return nil
				}
				// The control plane will always return an empty string if a cluster is created with a storage_type of aurora
				old, new := diff.GetChange("storage_type")

				if new.(string) == "aurora" && old.(string) == "" {
					if err := diff.SetNew("storage_type", ""); err != nil {
						return err
					}
					return nil
				}
				return nil
			},
		),
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	// Some API calls (e.g. RestoreDBClusterFromSnapshot do not support all
	// parameters to correctly apply all settings in one pass. For missing
	// parameters or unsupported configurations, we may need to call
	// ModifyDBInstance afterwards to prevent Terraform operators from API
	// errors or needing to double apply.
	var requiresModifyDbCluster bool
	modifyDbClusterInput := &rds.ModifyDBClusterInput{
		ApplyImmediately: aws.Bool(true),
	}

	var identifier string
	if v, ok := d.GetOk("cluster_identifier"); ok {
		identifier = v.(string)
	} else if v, ok := d.GetOk("cluster_identifier_prefix"); ok {
		identifier = id.PrefixedUniqueId(v.(string))
	} else {
		identifier = id.PrefixedUniqueId("tf-")
	}

	if v, ok := d.GetOk("snapshot_identifier"); ok {
		input := &rds.RestoreDBClusterFromSnapshotInput{
			CopyTagsToSnapshot:  aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBClusterIdentifier: aws.String(identifier),
			DeletionProtection:  aws.Bool(d.Get("deletion_protection").(bool)),
			Engine:              aws.String(d.Get("engine").(string)),
			EngineMode:          aws.String(d.Get("engine_mode").(string)),
			SnapshotIdentifier:  aws.String(v.(string)),
			Tags:                getTagsIn(ctx),
		}

		if v, ok := d.GetOk("availability_zones"); ok && v.(*schema.Set).Len() > 0 {
			input.AvailabilityZones = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("backtrack_window"); ok {
			input.BacktrackWindow = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			modifyDbClusterInput.BackupRetentionPeriod = aws.Int64(int64(v.(int)))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("database_name"); ok {
			input.DatabaseName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			input.DBClusterParameterGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("engine_version"); ok {
			input.EngineVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("kms_key_id"); ok {
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

		if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
			modifyDbClusterInput.MasterUserSecretKmsKeyId = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("network_type"); ok {
			input.NetworkType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("option_group_name"); ok {
			input.OptionGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("port"); ok {
			input.Port = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("preferred_backup_window"); ok {
			modifyDbClusterInput.PreferredBackupWindow = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("preferred_maintenance_window"); ok {
			modifyDbClusterInput.PreferredMaintenanceWindow = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("scaling_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ScalingConfiguration = expandScalingConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("serverlessv2_scaling_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			modifyDbClusterInput.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]interface{})[0].(map[string]interface{}))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("vpc_security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		}

		log.Printf("[DEBUG] Creating RDS Cluster: %s", input)
		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.RestoreDBClusterFromSnapshotWithContext(ctx, input)
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

		tfMap := v.([]interface{})[0].(map[string]interface{})
		input := &rds.RestoreDBClusterFromS3Input{
			CopyTagsToSnapshot:  aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBClusterIdentifier: aws.String(identifier),
			DeletionProtection:  aws.Bool(d.Get("deletion_protection").(bool)),
			Engine:              aws.String(d.Get("engine").(string)),
			MasterUsername:      aws.String(d.Get("master_username").(string)),
			S3BucketName:        aws.String(tfMap["bucket_name"].(string)),
			S3IngestionRoleArn:  aws.String(tfMap["ingestion_role"].(string)),
			S3Prefix:            aws.String(tfMap["bucket_prefix"].(string)),
			SourceEngine:        aws.String(tfMap["source_engine"].(string)),
			SourceEngineVersion: aws.String(tfMap["source_engine_version"].(string)),
			Tags:                getTagsIn(ctx),
		}

		if v, ok := d.GetOk("availability_zones"); ok && v.(*schema.Set).Len() > 0 {
			input.AvailabilityZones = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("backtrack_window"); ok {
			input.BacktrackWindow = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			input.BackupRetentionPeriod = aws.Int64(int64(v.(int)))
		}

		if v := d.Get("database_name"); v.(string) != "" {
			input.DatabaseName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			input.DBClusterParameterGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("engine_version"); ok {
			input.EngineVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			input.EnableIAMDatabaseAuthentication = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("kms_key_id"); ok {
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

		if v, ok := d.GetOk("network_type"); ok {
			input.NetworkType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("port"); ok {
			input.Port = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("preferred_backup_window"); ok {
			input.PreferredBackupWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk("preferred_maintenance_window"); ok {
			input.PreferredMaintenanceWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOkExists("storage_encrypted"); ok {
			input.StorageEncrypted = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("vpc_security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		}

		_, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.RestoreDBClusterFromS3WithContext(ctx, input)
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
	} else if v, ok := d.GetOk("restore_to_point_in_time"); ok {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		input := &rds.RestoreDBClusterToPointInTimeInput{
			DBClusterIdentifier:       aws.String(identifier),
			DeletionProtection:        aws.Bool(d.Get("deletion_protection").(bool)),
			SourceDBClusterIdentifier: aws.String(tfMap["source_cluster_identifier"].(string)),
			Tags:                      getTagsIn(ctx),
		}

		if v, ok := tfMap["restore_to_time"].(string); ok && v != "" {
			v, _ := time.Parse(time.RFC3339, v)
			input.RestoreToTime = aws.Time(v)
		}

		if v, ok := tfMap["use_latest_restorable_time"].(bool); ok && v {
			input.UseLatestRestorableTime = aws.Bool(v)
		}

		if input.RestoreToTime == nil && input.UseLatestRestorableTime == nil {
			return sdkdiag.AppendErrorf(diags, `Either "restore_to_time" or "use_latest_restorable_time" must be set`)
		}

		if v, ok := d.GetOk("backtrack_window"); ok {
			input.BacktrackWindow = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			modifyDbClusterInput.BackupRetentionPeriod = aws.Int64(int64(v.(int)))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			input.DBClusterParameterGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			input.EnableIAMDatabaseAuthentication = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("kms_key_id"); ok {
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

		if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
			modifyDbClusterInput.MasterUserSecretKmsKeyId = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("network_type"); ok {
			input.NetworkType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("option_group_name"); ok {
			input.OptionGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("port"); ok {
			input.Port = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("preferred_backup_window"); ok {
			modifyDbClusterInput.PreferredBackupWindow = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("preferred_maintenance_window"); ok {
			modifyDbClusterInput.PreferredMaintenanceWindow = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := tfMap["restore_type"].(string); ok {
			input.RestoreType = aws.String(v)
		}

		if v, ok := d.GetOk("scaling_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			modifyDbClusterInput.ScalingConfiguration = expandScalingConfiguration(v.([]interface{})[0].(map[string]interface{}))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("serverlessv2_scaling_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			modifyDbClusterInput.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]interface{})[0].(map[string]interface{}))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("vpc_security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		}

		log.Printf("[DEBUG] Creating RDS Cluster: %s", input)
		_, err := conn.RestoreDBClusterToPointInTimeWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating RDS Cluster (restore to point-in-time) (%s): %s", identifier, err)
		}
	} else {
		input := &rds.CreateDBClusterInput{
			CopyTagsToSnapshot:  aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBClusterIdentifier: aws.String(identifier),
			DeletionProtection:  aws.Bool(d.Get("deletion_protection").(bool)),
			Engine:              aws.String(d.Get("engine").(string)),
			EngineMode:          aws.String(d.Get("engine_mode").(string)),
			Tags:                getTagsIn(ctx),
		}

		if v, ok := d.GetOkExists("allocated_storage"); ok {
			input.AllocatedStorage = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("availability_zones"); ok && v.(*schema.Set).Len() > 0 {
			input.AvailabilityZones = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("backtrack_window"); ok {
			input.BacktrackWindow = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			input.BackupRetentionPeriod = aws.Int64(int64(v.(int)))
		}

		if v := d.Get("database_name"); v.(string) != "" {
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

		if v, ok := d.GetOk("enable_global_write_forwarding"); ok {
			input.EnableGlobalWriteForwarding = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("enable_http_endpoint"); ok {
			input.EnableHttpEndpoint = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("engine_version"); ok {
			input.EngineVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("global_cluster_identifier"); ok {
			input.GlobalClusterIdentifier = aws.String(v.(string))
		}

		if v, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			input.EnableIAMDatabaseAuthentication = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOkExists("iops"); ok {
			input.Iops = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("kms_key_id"); ok {
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
		if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
			input.MasterUserSecretKmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("master_username"); ok {
			input.MasterUsername = aws.String(v.(string))
		}

		if v, ok := d.GetOk("network_type"); ok {
			input.NetworkType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("port"); ok {
			input.Port = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("preferred_backup_window"); ok {
			input.PreferredBackupWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk("preferred_maintenance_window"); ok {
			input.PreferredMaintenanceWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk("replication_source_identifier"); ok && input.GlobalClusterIdentifier == nil {
			input.ReplicationSourceIdentifier = aws.String(v.(string))
		}

		if v, ok := d.GetOk("scaling_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ScalingConfiguration = expandScalingConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("serverlessv2_scaling_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("source_region"); ok {
			input.SourceRegion = aws.String(v.(string))
		}

		if v, ok := d.GetOkExists("storage_encrypted"); ok {
			input.StorageEncrypted = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOkExists("storage_type"); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("vpc_security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.CreateDBClusterWithContext(ctx, input)
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
				return sdkdiag.AppendErrorf(diags, "adding IAM Role (%s) to RDS Cluster (%s): %s", v, d.Id(), err)
			}
		}
	}

	if requiresModifyDbCluster {
		modifyDbClusterInput.DBClusterIdentifier = aws.String(d.Id())

		_, err := conn.ModifyDBClusterWithContext(ctx, modifyDbClusterInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS Cluster (%s): %s", d.Id(), err)
		}

		if _, err := waitDBClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	dbc, err := FindDBClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster (%s): %s", d.Id(), err)
	}

	d.Set("allocated_storage", dbc.AllocatedStorage)
	clusterARN := aws.StringValue(dbc.DBClusterArn)
	d.Set("arn", clusterARN)
	d.Set("availability_zones", aws.StringValueSlice(dbc.AvailabilityZones))
	d.Set("backtrack_window", dbc.BacktrackWindow)
	d.Set("backup_retention_period", dbc.BackupRetentionPeriod)
	d.Set("cluster_identifier", dbc.DBClusterIdentifier)
	var clusterMembers []string
	for _, v := range dbc.DBClusterMembers {
		clusterMembers = append(clusterMembers, aws.StringValue(v.DBInstanceIdentifier))
	}
	d.Set("cluster_members", clusterMembers)
	d.Set("cluster_resource_id", dbc.DbClusterResourceId)
	d.Set("copy_tags_to_snapshot", dbc.CopyTagsToSnapshot)
	// Only set the DatabaseName if it is not nil. There is a known API bug where
	// RDS accepts a DatabaseName but does not return it, causing a perpetual
	// diff.
	//	See https://github.com/hashicorp/terraform/issues/4671 for backstory
	if dbc.DatabaseName != nil { // nosemgrep: ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set("database_name", dbc.DatabaseName)
	}
	d.Set("db_cluster_instance_class", dbc.DBClusterInstanceClass)
	d.Set("db_cluster_parameter_group_name", dbc.DBClusterParameterGroup)
	d.Set("db_subnet_group_name", dbc.DBSubnetGroup)
	d.Set("deletion_protection", dbc.DeletionProtection)
	d.Set("enabled_cloudwatch_logs_exports", aws.StringValueSlice(dbc.EnabledCloudwatchLogsExports))
	d.Set("enable_http_endpoint", dbc.HttpEndpointEnabled)
	d.Set("endpoint", dbc.Endpoint)
	d.Set("engine", dbc.Engine)
	d.Set("engine_mode", dbc.EngineMode)
	clusterSetResourceDataEngineVersionFromCluster(d, dbc)
	d.Set("hosted_zone_id", dbc.HostedZoneId)
	d.Set("iam_database_authentication_enabled", dbc.IAMDatabaseAuthenticationEnabled)
	var iamRoleARNs []string
	for _, v := range dbc.AssociatedRoles {
		iamRoleARNs = append(iamRoleARNs, aws.StringValue(v.RoleArn))
	}
	d.Set("iam_roles", iamRoleARNs)
	d.Set("iops", dbc.Iops)
	d.Set("kms_key_id", dbc.KmsKeyId)

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
		if err := d.Set("master_user_secret", []interface{}{flattenManagedMasterUserSecret(dbc.MasterUserSecret)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting master_user_secret: %s", err)
		}
	} else {
		d.Set("master_user_secret", nil)
	}
	d.Set("master_username", dbc.MasterUsername)
	d.Set("network_type", dbc.NetworkType)
	d.Set("port", dbc.Port)
	d.Set("preferred_backup_window", dbc.PreferredBackupWindow)
	d.Set("preferred_maintenance_window", dbc.PreferredMaintenanceWindow)
	d.Set("reader_endpoint", dbc.ReaderEndpoint)
	d.Set("replication_source_identifier", dbc.ReplicationSourceIdentifier)
	if dbc.ScalingConfigurationInfo != nil {
		if err := d.Set("scaling_configuration", []interface{}{flattenScalingConfigurationInfo(dbc.ScalingConfigurationInfo)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting scaling_configuration: %s", err)
		}
	} else {
		d.Set("scaling_configuration", nil)
	}
	if dbc.ServerlessV2ScalingConfiguration != nil {
		if err := d.Set("serverlessv2_scaling_configuration", []interface{}{flattenServerlessV2ScalingConfigurationInfo(dbc.ServerlessV2ScalingConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting serverlessv2_scaling_configuration: %s", err)
		}
	} else {
		d.Set("serverlessv2_scaling_configuration", nil)
	}
	d.Set("storage_encrypted", dbc.StorageEncrypted)
	d.Set("storage_type", dbc.StorageType)
	var securityGroupIDs []string
	for _, v := range dbc.VpcSecurityGroups {
		securityGroupIDs = append(securityGroupIDs, aws.StringValue(v.VpcSecurityGroupId))
	}
	d.Set("vpc_security_group_ids", securityGroupIDs)

	// Fetch and save Global Cluster if engine mode global
	d.Set("global_cluster_identifier", "")

	if aws.StringValue(dbc.EngineMode) == EngineModeGlobal || aws.StringValue(dbc.EngineMode) == EngineModeProvisioned {
		globalCluster, err := FindGlobalClusterByDBClusterARN(ctx, conn, aws.StringValue(dbc.DBClusterArn))

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

	return nil
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	if d.HasChangesExcept(
		"allow_major_version_upgrade",
		"final_snapshot_identifier",
		"global_cluster_identifier",
		"iam_roles",
		"replication_source_identifier",
		"skip_final_snapshot",
		"tags", "tags_all") {
		input := &rds.ModifyDBClusterInput{
			ApplyImmediately:    aws.Bool(d.Get("apply_immediately").(bool)),
			DBClusterIdentifier: aws.String(d.Id()),
		}

		if d.HasChange("allocated_storage") {
			input.AllocatedStorage = aws.Int64(int64(d.Get("allocated_storage").(int)))
		}

		if v, ok := d.GetOk("allow_major_version_upgrade"); ok {
			input.AllowMajorVersionUpgrade = aws.Bool(v.(bool))
		}

		if d.HasChange("backtrack_window") {
			input.BacktrackWindow = aws.Int64(int64(d.Get("backtrack_window").(int)))
		}

		if d.HasChange("backup_retention_period") {
			input.BackupRetentionPeriod = aws.Int64(int64(d.Get("backup_retention_period").(int)))
		}

		if d.HasChange("copy_tags_to_snapshot") {
			input.CopyTagsToSnapshot = aws.Bool(d.Get("copy_tags_to_snapshot").(bool))
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

		if d.HasChange("deletion_protection") {
			input.DeletionProtection = aws.Bool(d.Get("deletion_protection").(bool))
		}

		if d.HasChange("enable_global_write_forwarding") {
			input.EnableGlobalWriteForwarding = aws.Bool(d.Get("enable_global_write_forwarding").(bool))
		}

		if d.HasChange("enable_http_endpoint") {
			input.EnableHttpEndpoint = aws.Bool(d.Get("enable_http_endpoint").(bool))
		}

		if d.HasChange("enabled_cloudwatch_logs_exports") {
			oraw, nraw := d.GetChange("enabled_cloudwatch_logs_exports")
			o := oraw.(*schema.Set)
			n := nraw.(*schema.Set)

			input.CloudwatchLogsExportConfiguration = &rds.CloudwatchLogsExportConfiguration{
				DisableLogTypes: flex.ExpandStringSet(o.Difference(n)),
				EnableLogTypes:  flex.ExpandStringSet(n.Difference(o)),
			}
		}

		if d.HasChange("engine_version") {
			input.EngineVersion = aws.String(d.Get("engine_version").(string))
		}

		// This can happen when updates are deferred (apply_immediately = false), and
		// multiple applies occur before the maintenance window. In this case,
		// continue sending the desired engine_version as part of the modify request.
		if d.Get("engine_version").(string) != d.Get("engine_version_actual").(string) {
			input.EngineVersion = aws.String(d.Get("engine_version").(string))
		}

		if d.HasChange("iam_database_authentication_enabled") {
			input.EnableIAMDatabaseAuthentication = aws.Bool(d.Get("iam_database_authentication_enabled").(bool))
		}

		if d.HasChange("iops") {
			input.Iops = aws.Int64(int64(d.Get("iops").(int)))
		}

		if d.HasChange("manage_master_user_password") {
			input.ManageMasterUserPassword = aws.Bool(d.Get("manage_master_user_password").(bool))
		}
		if d.HasChange("master_password") {
			if v, ok := d.GetOk("master_password"); ok {
				input.MasterUserPassword = aws.String(v.(string))
			}
		}
		if d.HasChange("master_user_secret_kms_key_id") {
			if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
				input.MasterUserSecretKmsKeyId = aws.String(v.(string))
			}
		}

		if d.HasChange("network_type") {
			input.NetworkType = aws.String(d.Get("network_type").(string))
		}

		if d.HasChange("port") {
			input.Port = aws.Int64(int64(d.Get("port").(int)))
		}

		if d.HasChange("preferred_backup_window") {
			input.PreferredBackupWindow = aws.String(d.Get("preferred_backup_window").(string))
		}

		if d.HasChange("preferred_maintenance_window") {
			input.PreferredMaintenanceWindow = aws.String(d.Get("preferred_maintenance_window").(string))
		}

		if d.HasChange("scaling_configuration") {
			if v, ok := d.GetOk("scaling_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.ScalingConfiguration = expandScalingConfiguration(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("serverlessv2_scaling_configuration") {
			if v, ok := d.GetOk("serverlessv2_scaling_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("storage_type") {
			input.StorageType = aws.String(d.Get("storage_type").(string))
		}

		if d.HasChange("vpc_security_group_ids") {
			if v, ok := d.GetOk("vpc_security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
				input.VpcSecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
			} else {
				input.VpcSecurityGroupIds = aws.StringSlice(nil)
			}
		}

		_, err := tfresource.RetryWhen(ctx, 5*time.Minute,
			func() (interface{}, error) {
				return conn.ModifyDBClusterWithContext(ctx, input)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
					return true, err
				}

				if tfawserr.ErrCodeEquals(err, rds.ErrCodeInvalidDBClusterStateFault) {
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

		if _, err := waitDBClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("global_cluster_identifier") {
		oRaw, nRaw := d.GetChange("global_cluster_identifier")
		o := oRaw.(string)
		n := nRaw.(string)

		if o == "" {
			return sdkdiag.AppendErrorf(diags, "existing RDS Clusters cannot be added to an existing RDS Global Cluster")
		}

		if n != "" {
			return sdkdiag.AppendErrorf(diags, "existing RDS Clusters cannot be migrated between existing RDS Global Clusters")
		}

		clusterARN := d.Get("arn").(string)
		input := &rds.RemoveFromGlobalClusterInput{
			DbClusterIdentifier:     aws.String(clusterARN),
			GlobalClusterIdentifier: aws.String(o),
		}

		log.Printf("[DEBUG] Removing RDS Cluster (%s) from RDS Global Cluster: %s", clusterARN, o)
		_, err := conn.RemoveFromGlobalClusterWithContext(ctx, input)

		if err != nil && !tfawserr.ErrCodeEquals(err, rds.ErrCodeGlobalClusterNotFoundFault) && !tfawserr.ErrMessageContains(err, "InvalidParameterValue", "is not found in global cluster") {
			return sdkdiag.AppendErrorf(diags, "removing RDS Cluster (%s) from RDS Global Cluster: %s", d.Id(), err)
		}
	}

	if d.HasChange("iam_roles") {
		oraw, nraw := d.GetChange("iam_roles")
		if oraw == nil {
			oraw = new(schema.Set)
		}
		if nraw == nil {
			nraw = new(schema.Set)
		}
		os := oraw.(*schema.Set)
		ns := nraw.(*schema.Set)

		for _, v := range ns.Difference(os).List() {
			if err := addIAMRoleToCluster(ctx, conn, d.Id(), v.(string)); err != nil {
				return sdkdiag.AppendErrorf(diags, "adding IAM Role (%s) to RDS Cluster (%s): %s", v, d.Id(), err)
			}
		}

		for _, v := range os.Difference(ns).List() {
			if err := removeIAMRoleFromCluster(ctx, conn, d.Id(), v.(string)); err != nil {
				return sdkdiag.AppendErrorf(diags, "removing IAM Role (%s) from RDS Cluster (%s): %s", v, d.Id(), err)
			}
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	// Automatically remove from global cluster to bypass this error on deletion:
	// InvalidDBClusterStateFault: This cluster is a part of a global cluster, please remove it from globalcluster first
	if d.Get("global_cluster_identifier").(string) != "" {
		clusterARN := d.Get("arn").(string)
		globalClusterID := d.Get("global_cluster_identifier").(string)
		input := &rds.RemoveFromGlobalClusterInput{
			DbClusterIdentifier:     aws.String(clusterARN),
			GlobalClusterIdentifier: aws.String(globalClusterID),
		}

		log.Printf("[DEBUG] Removing RDS Cluster (%s) from RDS Global Cluster: %s", clusterARN, globalClusterID)
		_, err := conn.RemoveFromGlobalClusterWithContext(ctx, input)

		if err != nil && !tfawserr.ErrCodeEquals(err, rds.ErrCodeGlobalClusterNotFoundFault) && !tfawserr.ErrMessageContains(err, "InvalidParameterValue", "is not found in global cluster") {
			return sdkdiag.AppendErrorf(diags, "removing RDS Cluster (%s) from RDS Global Cluster (%s): %s", d.Id(), globalClusterID, err)
		}
	}

	skipFinalSnapshot := d.Get("skip_final_snapshot").(bool)
	input := &rds.DeleteDBClusterInput{
		DBClusterIdentifier: aws.String(d.Id()),
		SkipFinalSnapshot:   aws.Bool(skipFinalSnapshot),
	}

	if !skipFinalSnapshot {
		if v, ok := d.GetOk("final_snapshot_identifier"); ok {
			input.FinalDBSnapshotIdentifier = aws.String(v.(string))
		} else {
			return sdkdiag.AppendErrorf(diags, "RDS Cluster final_snapshot_identifier is required when skip_final_snapshot is false")
		}
	}

	log.Printf("[DEBUG] Deleting RDS Cluster: %s", d.Id())
	_, err := tfresource.RetryWhen(ctx, clusterTimeoutDelete,
		func() (interface{}, error) {
			return conn.DeleteDBClusterWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, "InvalidParameterCombination", "disable deletion pro") {
				if v, ok := d.GetOk("deletion_protection"); (!ok || !v.(bool)) && d.Get("apply_immediately").(bool) {
					_, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutDelete),
						func() (interface{}, error) {
							return conn.ModifyDBClusterWithContext(ctx, &rds.ModifyDBClusterInput{
								ApplyImmediately:    aws.Bool(true),
								DBClusterIdentifier: aws.String(d.Id()),
								DeletionProtection:  aws.Bool(false),
							})
						},
						func(err error) (bool, error) {
							if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
								return true, err
							}

							if tfawserr.ErrCodeEquals(err, rds.ErrCodeInvalidDBClusterStateFault) {
								return true, err
							}

							return false, err
						},
					)
					if err != nil {
						return false, fmt.Errorf("modifying RDS Cluster (%s) DeletionProtection=false: %s", d.Id(), err)
					}

					if _, err := waitDBClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
						return false, fmt.Errorf("waiting for RDS Cluster (%s) update: %s", d.Id(), err)
					}
				}

				return true, err
			}

			if tfawserr.ErrMessageContains(err, rds.ErrCodeInvalidDBClusterStateFault, "is not currently in the available state") {
				return true, err
			}

			if tfawserr.ErrMessageContains(err, rds.ErrCodeInvalidDBClusterStateFault, "cluster is a part of a global cluster") {
				return true, err
			}

			return false, err
		},
	)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterNotFoundFault) {
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitDBClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func resourceClusterImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Neither skip_final_snapshot nor final_snapshot_identifier can be fetched
	// from any API call, so we need to default skip_final_snapshot to true so
	// that final_snapshot_identifier is not required
	d.Set("skip_final_snapshot", true)
	return []*schema.ResourceData{d}, nil
}

func addIAMRoleToCluster(ctx context.Context, conn *rds.RDS, clusterID, roleARN string) error {
	input := &rds.AddRoleToDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		RoleArn:             aws.String(roleARN),
	}

	_, err := conn.AddRoleToDBClusterWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("adding IAM Role (%s) to RDS Cluster (%s): %s", roleARN, clusterID, err)
	}

	return nil
}

func removeIAMRoleFromCluster(ctx context.Context, conn *rds.RDS, clusterID, roleARN string) error {
	input := &rds.RemoveRoleFromDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		RoleArn:             aws.String(roleARN),
	}

	_, err := conn.RemoveRoleFromDBClusterWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("removing IAM Role (%s) from RDS Cluster (%s): %s", roleARN, clusterID, err)
	}

	return err
}

func clusterSetResourceDataEngineVersionFromCluster(d *schema.ResourceData, c *rds.DBCluster) {
	oldVersion := d.Get("engine_version").(string)
	newVersion := aws.StringValue(c.EngineVersion)
	var pendingVersion string
	if c.PendingModifiedValues != nil && c.PendingModifiedValues.EngineVersion != nil {
		pendingVersion = aws.StringValue(c.PendingModifiedValues.EngineVersion)
	}
	compareActualEngineVersion(d, oldVersion, newVersion, pendingVersion)
}

func FindDBClusterByID(ctx context.Context, conn *rds.RDS, id string) (*rds.DBCluster, error) {
	input := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(id),
	}

	output, err := conn.DescribeDBClustersWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DBClusters) == 0 || output.DBClusters[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.DBClusters); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	dbCluster := output.DBClusters[0]

	// Eventual consistency check.
	if arn.IsARN(id) {
		if aws.StringValue(dbCluster.DBClusterArn) != id {
			return nil, &retry.NotFoundError{
				LastRequest: input,
			}
		}
	} else if aws.StringValue(dbCluster.DBClusterIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return dbCluster, nil
}

func statusDBCluster(ctx context.Context, conn *rds.RDS, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitDBClusterCreated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			ClusterStatusBackingUp,
			ClusterStatusCreating,
			ClusterStatusMigrating,
			ClusterStatusModifying,
			ClusterStatusPreparingDataMigration,
			ClusterStatusRebooting,
			ClusterStatusResettingMasterCredentials,
		},
		Target:     []string{ClusterStatusAvailable},
		Refresh:    statusDBCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterUpdated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBCluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			ClusterStatusBackingUp,
			ClusterStatusConfiguringIAMDatabaseAuth,
			ClusterStatusModifying,
			ClusterStatusRenaming,
			ClusterStatusResettingMasterCredentials,
			ClusterStatusScalingCompute,
			ClusterStatusUpgrading,
		},
		Target:     []string{ClusterStatusAvailable},
		Refresh:    statusDBCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterDeleted(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			ClusterStatusAvailable,
			ClusterStatusBackingUp,
			ClusterStatusDeleting,
			ClusterStatusModifying,
		},
		Target:     []string{},
		Refresh:    statusDBCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBCluster); ok {
		return output, err
	}

	return nil, err
}
