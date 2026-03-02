// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package docdb

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/docdb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_docdb_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
func resourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		UpdateWithoutTimeout: resourceClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				// Neither skip_final_snapshot nor final_snapshot_identifier can be fetched
				// from any API call, so we need to default skip_final_snapshot to true so
				// that final_snapshot_identifier is not required
				d.Set("skip_final_snapshot", true)
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAllowMajorVersionUpgrade: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrApplyImmediately: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"backup_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntAtMost(35),
			},
			names.AttrClusterIdentifier: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"cluster_identifier_prefix"},
				ValidateFunc:  validIdentifier,
			},
			"cluster_identifier_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrClusterIdentifier},
				ValidateFunc:  validIdentifierPrefix,
			},
			"cluster_members": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
			},
			"cluster_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_cluster_parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"db_subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrDeletionProtection: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"enabled_cloudwatch_logs_exports": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"audit",
						"profiler",
					}, false),
				},
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngine: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      engineDocDB,
				ValidateFunc: validation.StringInSlice(engine_Values(), false),
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Optional: true,
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validGlobalCusterIdentifier,
			},
			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Computed: true,
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
			"master_password": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"manage_master_user_password", "master_password_wo"},
			},
			"master_password_wo": {
				Type:          schema.TypeString,
				Optional:      true,
				WriteOnly:     true,
				ConflictsWith: []string{"manage_master_user_password", "master_password"},
				RequiredWith:  []string{"master_password_wo_version"},
			},
			"master_password_wo_version": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"master_password_wo"},
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
			"master_username": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"network_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(networkType_Values(), false),
			},
			names.AttrPort: {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      27017,
				ValidateFunc: validation.IntBetween(1150, 65535),
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
							ValidateFunc: validation.StringInSlice(restoreType_Values(), false),
						},
						"source_cluster_identifier": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
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
					"snapshot_identifier",
				},
			},
			"serverless_v2_scaling_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrMaxCapacity: {
							Type:     schema.TypeFloat,
							Required: true,
							ValidateFunc: validation.All(
								validation.FloatBetween(1.0, 256.0),
								validateServerlessCapacity,
							),
						},
						"min_capacity": {
							Type:     schema.TypeFloat,
							Required: true,
							ValidateFunc: validation.All(
								validation.FloatBetween(0.5, 256.0),
								validateServerlessCapacity,
							),
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
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// allow snapshot_idenfitier to be removed without forcing re-creation
					return new == ""
				},
				ConflictsWith: []string{
					"restore_to_point_in_time",
				},
			},
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			names.AttrStorageType: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(storageType_Values(), false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// When you create a DocumentDB DB cluster with the storage type set to "iopt1",
					// the storage type is returned in the response.
					// The storage type isn't returned when you set it to "standard".
					if old == "" && new == storageTypeStandard {
						return true
					}
					return old == new
				},
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
		CustomizeDiff: customdiff.All(
			// ForceNew only when removing serverless_v2_scaling_configuration; adding or modifying is supported in-place.
			// AWS does not support removing this configuration via ModifyDBCluster, so recreation is required.
			customdiff.ForceNewIfChange("serverless_v2_scaling_configuration", func(_ context.Context, old, new, meta any) bool {
				o := old != nil && len(old.([]any)) > 0
				n := new != nil && len(new.([]any)) > 0
				return o && !n
			}),
		),
	}
}

func validateServerlessCapacity(i any, k string) (ws []string, es []error) {
	const (
		epsilon = 1.0e-10
	)

	v, ok := i.(float64)
	if !ok {
		es = append(es, fmt.Errorf("expected type of %s to be float64", k))
		return
	}
	// add a small epsilon to avoid floating point precision issues
	if int(v*10+epsilon)%5 != 0 {
		es = append(es, fmt.Errorf("%s must be a multiple of 0.5", k))
		return
	}
	return
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	identifier := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrClusterIdentifier).(string)),
		create.WithConfiguredPrefix(d.Get("cluster_identifier_prefix").(string)),
		create.WithDefaultPrefix("tf-"),
	).Generate(ctx)

	// Some API calls (e.g. RestoreDBClusterFromSnapshot do not support all
	// parameters to correctly apply all settings in one pass. For missing
	// parameters or unsupported configurations, we may need to call
	// ModifyDBInstance afterwadocdb to prevent Terraform operators from API
	// errors or needing to double apply.
	var requiresModifyDbCluster bool
	inputMDBC := docdb.ModifyDBClusterInput{
		ApplyImmediately: aws.Bool(true),
	}

	// get write-only value from configuration
	masterPasswordWO, di := flex.GetWriteOnlyStringValue(d, cty.GetAttrPath("master_password_wo"))
	diags = append(diags, di...)
	if diags.HasError() {
		return diags
	}

	if _, ok := d.GetOk("snapshot_identifier"); ok {
		input := docdb.RestoreDBClusterFromSnapshotInput{
			DBClusterIdentifier: aws.String(identifier),
			DeletionProtection:  aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			Engine:              aws.String(d.Get(names.AttrEngine).(string)),
			SnapshotIdentifier:  aws.String(d.Get("snapshot_identifier").(string)),
			Tags:                getTagsIn(ctx),
		}

		if v := d.Get(names.AttrAvailabilityZones).(*schema.Set); v.Len() > 0 {
			input.AvailabilityZones = flex.ExpandStringValueSet(v)
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			inputMDBC.BackupRetentionPeriod = aws.Int32(int32(v.(int)))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			inputMDBC.DBClusterParameterGroupName = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && len(v.([]any)) > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringValueList(v.([]any))
		}

		if v, ok := d.GetOk(names.AttrEngineVersion); ok {
			input.EngineVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input.KmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("manage_master_user_password"); ok {
			inputMDBC.ManageMasterUserPassword = aws.Bool(v.(bool))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("master_password"); ok {
			inputMDBC.MasterUserPassword = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if masterPasswordWO != "" {
			inputMDBC.MasterUserPassword = aws.String(masterPasswordWO)
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("network_type"); ok {
			input.NetworkType = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrPort); ok {
			input.Port = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("preferred_backup_window"); ok {
			inputMDBC.PreferredBackupWindow = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk(names.AttrPreferredMaintenanceWindow); ok {
			inputMDBC.PreferredMaintenanceWindow = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("serverless_v2_scaling_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]any)[0].(map[string]any))
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v)
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
			return conn.RestoreDBClusterFromSnapshot(ctx, &input)
		}, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating DocumentDB Cluster (restore from snapshot) (%s): %s", identifier, err)
		}
	} else if v, ok := d.GetOk("restore_to_point_in_time"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		tfMap := v.([]any)[0].(map[string]any)
		input := docdb.RestoreDBClusterToPointInTimeInput{
			DBClusterIdentifier:       aws.String(identifier),
			SourceDBClusterIdentifier: aws.String(tfMap["source_cluster_identifier"].(string)),
			DeletionProtection:        aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			Tags:                      getTagsIn(ctx),
		}

		if v, ok := tfMap["restore_to_time"].(string); ok && v != "" {
			t, _ := time.Parse(time.RFC3339, v)
			input.RestoreToTime = aws.Time(t)
		}

		if v, ok := tfMap["use_latest_restorable_time"].(bool); ok && v {
			input.UseLatestRestorableTime = aws.Bool(v)
		}

		if input.RestoreToTime == nil && input.UseLatestRestorableTime == nil {
			return sdkdiag.AppendErrorf(diags, `Either "restore_to_time" or "use_latest_restorable_time" must be set`)
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && len(v.([]any)) > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringValueList(v.([]any))
		}

		if v, ok := tfMap["restore_type"].(string); ok {
			input.RestoreType = aws.String(v)
		}

		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input.KmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("network_type"); ok {
			input.NetworkType = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrPort); ok {
			input.Port = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("serverless_v2_scaling_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]any)[0].(map[string]any))
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v)
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
			return conn.RestoreDBClusterToPointInTime(ctx, &input)
		}, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating DocumentDB Cluster (restore to point-in-time) (%s): %s", identifier, err)
		}
	} else {
		// Secondary DocDB clusters part of a global cluster will not supply the master_password
		if _, ok := d.GetOk("global_cluster_identifier"); !ok {
			// If manage_master_user_password is true, we don't need master_password
			if v, ok := d.GetOk("manage_master_user_password"); !ok || !v.(bool) {
				if _, ok := d.GetOk("master_password"); !ok && masterPasswordWO == "" {
					return sdkdiag.AppendErrorf(diags, `provider.aws: aws_docdb_cluster: %s: "master_password", "master_password_wo": required field is not set`, identifier)
				}
			}
		}

		// Secondary DocDB clusters part of a global cluster will not supply the master_username
		if _, ok := d.GetOk("global_cluster_identifier"); !ok {
			if _, ok := d.GetOk("master_username"); !ok {
				return sdkdiag.AppendErrorf(diags, `provider.aws: aws_docdb_cluster: %s: "master_username": required field is not set`, identifier)
			}
		}

		input := docdb.CreateDBClusterInput{
			DBClusterIdentifier: aws.String(identifier),
			DeletionProtection:  aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			Engine:              aws.String(d.Get(names.AttrEngine).(string)),
			MasterUsername:      aws.String(d.Get("master_username").(string)),
			Tags:                getTagsIn(ctx),
		}

		if v := d.Get(names.AttrAvailabilityZones).(*schema.Set); v.Len() > 0 {
			input.AvailabilityZones = flex.ExpandStringValueSet(v)
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			input.BackupRetentionPeriod = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			input.DBClusterParameterGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && len(v.([]any)) > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringValueList(v.([]any))
		}

		if v, ok := d.GetOk(names.AttrEngineVersion); ok {
			input.EngineVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("global_cluster_identifier"); ok {
			input.GlobalClusterIdentifier = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input.KmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("manage_master_user_password"); ok {
			input.ManageMasterUserPassword = aws.Bool(v.(bool))
		} else {
			if v, ok := d.GetOk("master_password"); ok {
				input.MasterUserPassword = aws.String(v.(string))
			}

			if masterPasswordWO != "" {
				input.MasterUserPassword = aws.String(masterPasswordWO)
			}
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

		if v, ok := d.GetOk("serverless_v2_scaling_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]any)[0].(map[string]any))
		}

		if v, ok := d.GetOk(names.AttrStorageEncrypted); ok {
			input.StorageEncrypted = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v)
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
			return conn.CreateDBCluster(ctx, &input)
		}, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating DocumentDB Cluster (%s): %s", identifier, err)
		}
	}

	d.SetId(identifier)

	if _, err := waitDBClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster (%s) create: %s", d.Id(), err)
	}

	if requiresModifyDbCluster {
		inputMDBC.DBClusterIdentifier = aws.String(d.Id())

		_, err := conn.ModifyDBCluster(ctx, &inputMDBC)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying DocumentDB Cluster (%s): %s", d.Id(), err)
		}

		if _, err := waitDBClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	dbc, err := findDBClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] DocumentDB Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB Cluster (%s): %s", d.Id(), err)
	}

	// Ignore the following API error for regions/partitions that do not support DocDB Global Clusters:
	// InvalidParameterValue: Access Denied to API Version: APIGlobalDatabases
	if globalCluster, err := findGlobalClusterByClusterARN(ctx, conn, aws.ToString(dbc.DBClusterArn)); retry.NotFound(err) || tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "Access Denied to API Version: APIGlobalDatabases") {
		d.Set("global_cluster_identifier", "")
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB Global Cluster information for DocumentDB Cluster (%s): %s", d.Id(), err)
	} else {
		d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
	}

	d.Set(names.AttrARN, dbc.DBClusterArn)
	d.Set(names.AttrAvailabilityZones, dbc.AvailabilityZones)
	d.Set("backup_retention_period", dbc.BackupRetentionPeriod)
	d.Set(names.AttrClusterIdentifier, dbc.DBClusterIdentifier)
	d.Set("cluster_identifier_prefix", create.NamePrefixFromName(aws.ToString(dbc.DBClusterIdentifier)))
	var clusterMembers []string
	for _, v := range dbc.DBClusterMembers {
		clusterMembers = append(clusterMembers, aws.ToString(v.DBInstanceIdentifier))
	}
	d.Set("cluster_members", clusterMembers)
	d.Set("cluster_resource_id", dbc.DbClusterResourceId)
	d.Set("db_cluster_parameter_group_name", dbc.DBClusterParameterGroup)
	d.Set("db_subnet_group_name", dbc.DBSubnetGroup)
	d.Set(names.AttrDeletionProtection, dbc.DeletionProtection)
	d.Set("enabled_cloudwatch_logs_exports", dbc.EnabledCloudwatchLogsExports)
	d.Set(names.AttrEndpoint, dbc.Endpoint)
	d.Set(names.AttrEngineVersion, dbc.EngineVersion)
	d.Set(names.AttrEngine, dbc.Engine)
	d.Set(names.AttrHostedZoneID, dbc.HostedZoneId)
	d.Set(names.AttrKMSKeyID, dbc.KmsKeyId)
	if v := dbc.MasterUserSecret; v != nil {
		tfList := []any{map[string]any{
			names.AttrKMSKeyID: aws.ToString(v.KmsKeyId),
			"secret_arn":       aws.ToString(v.SecretArn),
			"secret_status":    aws.ToString(v.SecretStatus),
		}}
		if err := d.Set("master_user_secret", tfList); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting master_user_secret: %s", err)
		}
	} else {
		d.Set("master_user_secret", nil)
	}
	d.Set("master_username", dbc.MasterUsername)
	d.Set("network_type", dbc.NetworkType)
	d.Set(names.AttrPort, dbc.Port)
	d.Set("preferred_backup_window", dbc.PreferredBackupWindow)
	d.Set(names.AttrPreferredMaintenanceWindow, dbc.PreferredMaintenanceWindow)
	d.Set("reader_endpoint", dbc.ReaderEndpoint)
	if err := d.Set("serverless_v2_scaling_configuration", flattenServerlessV2ScalingConfiguration(dbc.ServerlessV2ScalingConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting serverless_v2_scaling_configuration: %s", err)
	}
	d.Set(names.AttrStorageEncrypted, dbc.StorageEncrypted)
	d.Set(names.AttrStorageType, dbc.StorageType)
	d.Set(names.AttrVPCSecurityGroupIDs, tfslices.ApplyToAll(dbc.VpcSecurityGroups, func(v awstypes.VpcSecurityGroupMembership) string {
		return aws.ToString(v.VpcSecurityGroupId)
	}))

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, "global_cluster_identifier", "skip_final_snapshot") {
		input := docdb.ModifyDBClusterInput{
			ApplyImmediately:    aws.Bool(d.Get(names.AttrApplyImmediately).(bool)),
			DBClusterIdentifier: aws.String(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrAllowMajorVersionUpgrade); ok {
			input.AllowMajorVersionUpgrade = aws.Bool(v.(bool))
		}

		if d.HasChange("backup_retention_period") {
			input.BackupRetentionPeriod = aws.Int32(int32(d.Get("backup_retention_period").(int)))
		}

		if d.HasChange("db_cluster_parameter_group_name") {
			input.DBClusterParameterGroupName = aws.String(d.Get("db_cluster_parameter_group_name").(string))
		}

		if d.HasChange(names.AttrDeletionProtection) {
			input.DeletionProtection = aws.Bool(d.Get(names.AttrDeletionProtection).(bool))
		}

		if d.HasChange("enabled_cloudwatch_logs_exports") {
			input.CloudwatchLogsExportConfiguration = expandCloudwatchLogsExportConfiguration(d)
		}

		if d.HasChange(names.AttrEngineVersion) {
			input.EngineVersion = aws.String(d.Get(names.AttrEngineVersion).(string))
		}

		if d.HasChange("manage_master_user_password") {
			input.ManageMasterUserPassword = aws.Bool(d.Get("manage_master_user_password").(bool))
		}

		if d.HasChange("master_password") {
			input.MasterUserPassword = aws.String(d.Get("master_password").(string))
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

		if d.HasChange("network_type") {
			input.NetworkType = aws.String(d.Get("network_type").(string))
		}

		if d.HasChange("preferred_backup_window") {
			input.PreferredBackupWindow = aws.String(d.Get("preferred_backup_window").(string))
		}

		if d.HasChange("serverless_v2_scaling_configuration") {
			if v, ok := d.GetOk("serverless_v2_scaling_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChange(names.AttrStorageType) {
			input.StorageType = aws.String(d.Get(names.AttrStorageType).(string))
		}

		if d.HasChange(names.AttrPreferredMaintenanceWindow) {
			input.PreferredMaintenanceWindow = aws.String(d.Get(names.AttrPreferredMaintenanceWindow).(string))
		}

		if d.HasChange(names.AttrVPCSecurityGroupIDs) {
			if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
				input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v)
			} else {
				input.VpcSecurityGroupIds = []string{}
			}
		}

		const (
			timeout = 5 * time.Minute
		)
		_, err := tfresource.RetryWhen(ctx, timeout,
			func(ctx context.Context) (any, error) {
				return conn.ModifyDBCluster(ctx, &input)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
					return true, err
				}

				if errs.IsAErrorMessageContains[*awstypes.InvalidDBClusterStateFault](err, "is not currently in the available state") {
					return true, err
				}

				if errs.IsAErrorMessageContains[*awstypes.InvalidDBClusterStateFault](err, "cluster is a part of a global cluster") {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying DocumentDB Cluster (%s): %s", d.Id(), err)
		}

		if _, err := waitDBClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("global_cluster_identifier") {
		oRaw, nRaw := d.GetChange("global_cluster_identifier")
		o, n := oRaw.(string), nRaw.(string)

		if o == "" {
			return sdkdiag.AppendErrorf(diags, "existing DocumentDB Clusters cannot be added to an existing DocumentDB Global Cluster")
		}

		if n != "" {
			return sdkdiag.AppendErrorf(diags, "existing DocumentDB Clusters cannot be migrated between existing DocumentDB Global Clusters")
		}

		if err := removeClusterFromGlobalCluster(ctx, conn, d.Get(names.AttrARN).(string), o, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	skipFinalSnapshot := d.Get("skip_final_snapshot").(bool)
	input := docdb.DeleteDBClusterInput{
		DBClusterIdentifier: aws.String(d.Id()),
		SkipFinalSnapshot:   aws.Bool(skipFinalSnapshot),
	}

	if !skipFinalSnapshot {
		if v, ok := d.GetOk(names.AttrFinalSnapshotIdentifier); ok {
			input.FinalDBSnapshotIdentifier = aws.String(v.(string))
		} else {
			return sdkdiag.AppendErrorf(diags, "DocumentDB Cluster FinalSnapshotIdentifier is required when a final snapshot is required")
		}
	}

	if v, ok := d.GetOk("global_cluster_identifier"); ok {
		if err := removeClusterFromGlobalCluster(ctx, conn, d.Get(names.AttrARN).(string), v.(string), d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[DEBUG] Deleting DocumentDB Cluster: %s", d.Id())
	_, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutDelete),
		func(ctx context.Context) (any, error) {
			return conn.DeleteDBCluster(ctx, &input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.InvalidDBClusterStateFault](err, "is not currently in the available state") {
				return true, err
			}

			if errs.IsAErrorMessageContains[*awstypes.InvalidDBClusterStateFault](err, "cluster is a part of a global cluster") {
				return true, err
			}

			return false, err
		},
	)

	if errs.IsA[*awstypes.DBClusterNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DocumentDB Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitDBClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandCloudwatchLogsExportConfiguration(d *schema.ResourceData) *awstypes.CloudwatchLogsExportConfiguration { // nosemgrep:ci.caps0-in-func-name
	oraw, nraw := d.GetChange("enabled_cloudwatch_logs_exports")
	o := oraw.([]any)
	n := nraw.([]any)

	create, disable := diffCloudWatchLogsExportConfiguration(o, n)

	return &awstypes.CloudwatchLogsExportConfiguration{
		EnableLogTypes:  flex.ExpandStringValueList(create),
		DisableLogTypes: flex.ExpandStringValueList(disable),
	}
}

func diffCloudWatchLogsExportConfiguration(old, new []any) ([]any, []any) {
	add := make([]any, 0)
	disable := make([]any, 0)

	for _, n := range new {
		if idx := tfslices.IndexOf(old, n.(string)); idx == -1 {
			add = append(add, n)
		}
	}

	for _, o := range old {
		if idx := tfslices.IndexOf(new, o.(string)); idx == -1 {
			disable = append(disable, o)
		}
	}

	return add, disable
}

func expandServerlessV2ScalingConfiguration(v map[string]any) *awstypes.ServerlessV2ScalingConfiguration {
	if v == nil {
		return nil
	}

	apiObject := &awstypes.ServerlessV2ScalingConfiguration{}
	if v, ok := v[names.AttrMaxCapacity].(float64); ok {
		apiObject.MaxCapacity = aws.Float64(v)
	}
	if v, ok := v["min_capacity"].(float64); ok {
		apiObject.MinCapacity = aws.Float64(v)
	}

	return apiObject
}

func flattenServerlessV2ScalingConfiguration(v *awstypes.ServerlessV2ScalingConfigurationInfo) []any {
	if v == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v.MaxCapacity != nil {
		tfMap[names.AttrMaxCapacity] = aws.ToFloat64(v.MaxCapacity)
	}
	if v.MinCapacity != nil {
		tfMap["min_capacity"] = aws.ToFloat64(v.MinCapacity)
	}
	return []any{tfMap}
}

func removeClusterFromGlobalCluster(ctx context.Context, conn *docdb.Client, clusterARN, globalClusterID string, timeout time.Duration) error {
	input := docdb.RemoveFromGlobalClusterInput{
		DbClusterIdentifier:     aws.String(clusterARN),
		GlobalClusterIdentifier: aws.String(globalClusterID),
	}

	_, err := conn.RemoveFromGlobalCluster(ctx, &input)

	if errs.IsA[*awstypes.DBClusterNotFoundFault](err) || errs.IsA[*awstypes.GlobalClusterNotFoundFault](err) ||
		tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "is not found in global cluster") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("removing DocumentDB Cluster (%s) from DocumentDB Global Cluster (%s): %w", clusterARN, globalClusterID, err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, timeout, func(ctx context.Context) (any, error) {
		return findGlobalClusterByClusterARN(ctx, conn, clusterARN)
	})

	if err != nil {
		return fmt.Errorf("waiting for DocumentDB Cluster (%s) removal from DocumentDB Global Cluster (%s): %w", clusterARN, globalClusterID, err)
	}

	return nil
}

func findDBClusterByID(ctx context.Context, conn *docdb.Client, id string) (*awstypes.DBCluster, error) {
	input := docdb.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(id),
	}
	output, err := findDBCluster(ctx, conn, &input, tfslices.PredicateTrue[*awstypes.DBCluster]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.DBClusterIdentifier) != id {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func findClusterByARN(ctx context.Context, conn *docdb.Client, arn string) (*awstypes.DBCluster, error) {
	input := docdb.DescribeDBClustersInput{}

	return findDBCluster(ctx, conn, &input, func(v *awstypes.DBCluster) bool {
		return aws.ToString(v.DBClusterArn) == arn
	})
}

func findDBCluster(ctx context.Context, conn *docdb.Client, input *docdb.DescribeDBClustersInput, filter tfslices.Predicate[*awstypes.DBCluster]) (*awstypes.DBCluster, error) {
	output, err := findDBClusters(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBClusters(ctx context.Context, conn *docdb.Client, input *docdb.DescribeDBClustersInput, filter tfslices.Predicate[*awstypes.DBCluster]) ([]awstypes.DBCluster, error) {
	var output []awstypes.DBCluster

	pages := docdb.NewDescribeDBClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.DBClusterNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusters {
			if !inttypes.IsZero(&v) && filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusDBCluster(ctx context.Context, conn *docdb.Client, id string) retry.StateRefreshFunc {
	return func(_ context.Context) (any, string, error) {
		output, err := findDBClusterByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitDBClusterAvailable(ctx context.Context, conn *docdb.Client, id string, timeout time.Duration) (*awstypes.DBCluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			clusterStatusCreating,
			clusterStatusBackingUp,
			clusterStatusModifying,
			clusterStatusPreparingDataMigration,
			clusterStatusMigrating,
			clusterStatusResettingMasterCredentials,
			clusterStatusUpgrading,
		},
		Target:     []string{clusterStatusAvailable},
		Refresh:    statusDBCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterDeleted(ctx context.Context, conn *docdb.Client, id string, timeout time.Duration) (*awstypes.DBCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			clusterStatusAvailable,
			clusterStatusDeleting,
			clusterStatusBackingUp,
			clusterStatusModifying,
		},
		Target:     []string{},
		Refresh:    statusDBCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DBCluster); ok {
		return output, err
	}

	return nil, err
}
