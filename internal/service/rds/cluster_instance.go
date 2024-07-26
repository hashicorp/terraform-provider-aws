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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rds_cluster_instance", name="Cluster Instance")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func ResourceClusterInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterInstanceCreate,
		ReadWithoutTimeout:   resourceClusterInstanceRead,
		UpdateWithoutTimeout: resourceClusterInstanceUpdate,
		DeleteWithoutTimeout: resourceClusterInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
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
			names.AttrAutoMinorVersionUpgrade: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"ca_cert_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrClusterIdentifier: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"copy_tags_to_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"custom_iam_instance_profile": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^AWSRDSCustom.*$`), "must begin with AWSRDSCustom"),
			},
			"db_parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"db_subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"dbi_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
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
					validation.StringInSlice(ClusterInstanceEngine_Values(), false),
				),
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
			names.AttrIdentifier: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"identifier_prefix"},
				ValidateFunc:  validIdentifier,
			},
			"identifier_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrIdentifier},
				ValidateFunc:  validIdentifierPrefix,
			},
			"instance_class": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"monitoring_interval": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"monitoring_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"network_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"performance_insights_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
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
				ValidateFunc: validation.Any(
					validation.IntInSlice([]int{7, 731}),
					validation.All(
						validation.IntAtLeast(7),
						validation.IntAtMost(731),
						validation.IntDivisibleBy(31),
					),
				),
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
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
				StateFunc: func(v interface{}) string {
					if v != nil {
						value := v.(string)
						return strings.ToLower(value)
					}
					return ""
				},
				ValidateFunc: verify.ValidOnceAWeekWindowFormat,
			},
			"promotion_tier": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			names.AttrPubliclyAccessible: {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"writer": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	clusterID := d.Get(names.AttrClusterIdentifier).(string)
	identifier := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrIdentifier).(string)),
		create.WithConfiguredPrefix(d.Get("identifier_prefix").(string)),
		create.WithDefaultPrefix("tf-"),
	).Generate()
	input := &rds.CreateDBInstanceInput{
		AutoMinorVersionUpgrade: aws.Bool(d.Get(names.AttrAutoMinorVersionUpgrade).(bool)),
		CopyTagsToSnapshot:      aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
		DBClusterIdentifier:     aws.String(clusterID),
		DBInstanceClass:         aws.String(d.Get("instance_class").(string)),
		DBInstanceIdentifier:    aws.String(identifier),
		Engine:                  aws.String(d.Get(names.AttrEngine).(string)),
		PromotionTier:           aws.Int64(int64(d.Get("promotion_tier").(int))),
		PubliclyAccessible:      aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool)),
		Tags:                    getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		input.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("db_parameter_group_name"); ok {
		input.DBParameterGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("db_subnet_group_name"); ok {
		input.DBSubnetGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("custom_iam_instance_profile"); ok {
		input.CustomIamInstanceProfile = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("monitoring_interval"); ok {
		input.MonitoringInterval = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("monitoring_role_arn"); ok {
		input.MonitoringRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("performance_insights_enabled"); ok {
		input.EnablePerformanceInsights = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("performance_insights_kms_key_id"); ok {
		input.PerformanceInsightsKMSKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("performance_insights_retention_period"); ok {
		input.PerformanceInsightsRetentionPeriod = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("preferred_backup_window"); ok {
		input.PreferredBackupWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPreferredMaintenanceWindow); ok {
		input.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateDBInstanceWithContext(ctx, input)
		},
		errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Cluster (%s) Instance (%s): %s", clusterID, identifier, err)
	}

	output := outputRaw.(*rds.CreateDBInstanceOutput)

	d.SetId(aws.StringValue(output.DBInstance.DBInstanceIdentifier))

	if _, err := waitDBClusterInstanceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster Instance (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("ca_cert_identifier"); ok && v.(string) != aws.StringValue(output.DBInstance.CACertificateIdentifier) {
		input := &rds.ModifyDBInstanceInput{
			ApplyImmediately:        aws.Bool(true),
			CACertificateIdentifier: aws.String(v.(string)),
			DBInstanceIdentifier:    aws.String(d.Id()),
		}

		if _, err := conn.ModifyDBInstanceWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS Cluster Instance (%s): %s", d.Id(), err)
		}

		if _, err := waitDBInstanceAvailableSDKv1(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster Instance (%s) update: %s", d.Id(), err)
		}

		_, err = conn.RebootDBInstanceWithContext(ctx, &rds.RebootDBInstanceInput{
			DBInstanceIdentifier: aws.String(d.Id()),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "rebooting RDS Cluster Instance (%s): %s", d.Id(), err)
		}

		if _, err := waitDBInstanceAvailableSDKv1(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster Instance (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterInstanceRead(ctx, d, meta)...)
}

func resourceClusterInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	db, err := findDBInstanceByIDSDKv1(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Cluster Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster Instance (%s): %s", d.Id(), err)
	}

	dbClusterID := aws.StringValue(db.DBClusterIdentifier)

	if dbClusterID == "" {
		return sdkdiag.AppendErrorf(diags, "DBClusterIdentifier is missing from RDS Cluster Instance (%s). The aws_db_instance resource should be used for non-Aurora instances", d.Id())
	}

	dbc, err := FindDBClusterByID(ctx, conn, dbClusterID)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster (%s): %s", dbClusterID, err)
	}

	for _, m := range dbc.DBClusterMembers {
		if aws.StringValue(m.DBInstanceIdentifier) == d.Id() {
			if aws.BoolValue(m.IsClusterWriter) {
				d.Set("writer", true)
			} else {
				d.Set("writer", false)
			}
		}
	}

	if db.Endpoint != nil {
		d.Set(names.AttrEndpoint, db.Endpoint.Address)
		d.Set(names.AttrPort, db.Endpoint.Port)
	}

	d.Set(names.AttrARN, db.DBInstanceArn)
	d.Set(names.AttrAutoMinorVersionUpgrade, db.AutoMinorVersionUpgrade)
	d.Set(names.AttrAvailabilityZone, db.AvailabilityZone)
	d.Set("ca_cert_identifier", db.CACertificateIdentifier)
	d.Set(names.AttrClusterIdentifier, db.DBClusterIdentifier)
	d.Set("copy_tags_to_snapshot", db.CopyTagsToSnapshot)
	d.Set("custom_iam_instance_profile", db.CustomIamInstanceProfile)
	if len(db.DBParameterGroups) > 0 && db.DBParameterGroups[0] != nil {
		d.Set("db_parameter_group_name", db.DBParameterGroups[0].DBParameterGroupName)
	}
	if db.DBSubnetGroup != nil {
		d.Set("db_subnet_group_name", db.DBSubnetGroup.DBSubnetGroupName)
	}
	d.Set("dbi_resource_id", db.DbiResourceId)
	d.Set(names.AttrEngine, db.Engine)
	d.Set(names.AttrIdentifier, db.DBInstanceIdentifier)
	d.Set("identifier_prefix", create.NamePrefixFromName(aws.StringValue(db.DBInstanceIdentifier)))
	d.Set("instance_class", db.DBInstanceClass)
	d.Set(names.AttrKMSKeyID, db.KmsKeyId)
	d.Set("monitoring_interval", db.MonitoringInterval)
	d.Set("monitoring_role_arn", db.MonitoringRoleArn)
	d.Set("network_type", db.NetworkType)
	d.Set("performance_insights_enabled", db.PerformanceInsightsEnabled)
	d.Set("performance_insights_kms_key_id", db.PerformanceInsightsKMSKeyId)
	d.Set("performance_insights_retention_period", db.PerformanceInsightsRetentionPeriod)
	d.Set("preferred_backup_window", db.PreferredBackupWindow)
	d.Set(names.AttrPreferredMaintenanceWindow, db.PreferredMaintenanceWindow)
	d.Set("promotion_tier", db.PromotionTier)
	d.Set(names.AttrPubliclyAccessible, db.PubliclyAccessible)
	d.Set(names.AttrStorageEncrypted, db.StorageEncrypted)

	clusterSetResourceDataEngineVersionFromClusterInstance(d, db)

	setTagsOut(ctx, db.TagList)

	return diags
}

func resourceClusterInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &rds.ModifyDBInstanceInput{
			ApplyImmediately:     aws.Bool(d.Get(names.AttrApplyImmediately).(bool)),
			DBInstanceIdentifier: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrAutoMinorVersionUpgrade) {
			input.AutoMinorVersionUpgrade = aws.Bool(d.Get(names.AttrAutoMinorVersionUpgrade).(bool))
		}

		if d.HasChange("ca_cert_identifier") {
			input.CACertificateIdentifier = aws.String(d.Get("ca_cert_identifier").(string))
		}

		if d.HasChange("copy_tags_to_snapshot") {
			input.CopyTagsToSnapshot = aws.Bool(d.Get("copy_tags_to_snapshot").(bool))
		}

		if d.HasChange("db_parameter_group_name") {
			input.DBParameterGroupName = aws.String(d.Get("db_parameter_group_name").(string))
		}

		if d.HasChange("instance_class") {
			input.DBInstanceClass = aws.String(d.Get("instance_class").(string))
		}

		if d.HasChange("monitoring_interval") {
			input.MonitoringInterval = aws.Int64(int64(d.Get("monitoring_interval").(int)))
		}

		if d.HasChange("monitoring_role_arn") {
			input.MonitoringRoleArn = aws.String(d.Get("monitoring_role_arn").(string))
		}

		if d.HasChanges("performance_insights_enabled", "performance_insights_kms_key_id", "performance_insights_retention_period") {
			input.EnablePerformanceInsights = aws.Bool(d.Get("performance_insights_enabled").(bool))

			if v, ok := d.GetOk("performance_insights_kms_key_id"); ok {
				input.PerformanceInsightsKMSKeyId = aws.String(v.(string))
			}

			if v, ok := d.GetOk("performance_insights_retention_period"); ok {
				input.PerformanceInsightsRetentionPeriod = aws.Int64(int64(v.(int)))
			}
		}

		if d.HasChange("preferred_backup_window") {
			input.PreferredBackupWindow = aws.String(d.Get("preferred_backup_window").(string))
		}

		if d.HasChange(names.AttrPreferredMaintenanceWindow) {
			input.PreferredMaintenanceWindow = aws.String(d.Get(names.AttrPreferredMaintenanceWindow).(string))
		}

		if d.HasChange("promotion_tier") {
			input.PromotionTier = aws.Int64(int64(d.Get("promotion_tier").(int)))
		}

		if d.HasChange(names.AttrPubliclyAccessible) {
			input.PubliclyAccessible = aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool))
		}

		log.Printf("[DEBUG] Updating RDS Cluster Instance: %s", input)
		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.ModifyDBInstanceWithContext(ctx, input)
			},
			errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS Cluster Instance (%s): %s", d.Id(), err)
		}

		if _, err := waitDBClusterInstanceUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster Instance (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterInstanceRead(ctx, d, meta)...)
}

func resourceClusterInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(d.Id()),
	}

	// Automatically set skip_final_snapshot = true for RDS Custom instances
	if strings.HasPrefix(d.Get(names.AttrEngine).(string), InstanceEngineCustomPrefix) {
		log.Printf("[DEBUG] RDS Custom engine detected (%s) applying SkipFinalSnapshot: %s", d.Get(names.AttrEngine).(string), "true")
		input.SkipFinalSnapshot = aws.Bool(true)
	}

	log.Printf("[DEBUG] Deleting RDS Cluster Instance: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, d.Timeout(schema.TimeoutDelete),
		func() (interface{}, error) {
			return conn.DeleteDBInstanceWithContext(ctx, input)
		},
		rds.ErrCodeInvalidDBClusterStateFault, "Delete the replica cluster before deleting")

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceNotFoundFault) {
		return diags
	}

	if err != nil && !tfawserr.ErrMessageContains(err, rds.ErrCodeInvalidDBInstanceStateFault, "is already being deleted") {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Cluster Instance (%s): %s", d.Id(), err)
	}

	if _, err := waitDBClusterInstanceDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster Instance (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func waitDBClusterInstanceCreated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			InstanceStatusBackingUp,
			InstanceStatusConfiguringEnhancedMonitoring,
			InstanceStatusConfiguringIAMDatabaseAuth,
			InstanceStatusConfiguringLogExports,
			InstanceStatusCreating,
			InstanceStatusMaintenance,
			InstanceStatusModifying,
			InstanceStatusRebooting,
			InstanceStatusRenaming,
			InstanceStatusResettingMasterCredentials,
			InstanceStatusStarting,
			InstanceStatusStorageOptimization,
			InstanceStatusUpgrading,
		},
		Target:     []string{InstanceStatusAvailable},
		Refresh:    statusDBInstanceSDKv1(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBInstance); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterInstanceUpdated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			InstanceStatusBackingUp,
			InstanceStatusConfiguringEnhancedMonitoring,
			InstanceStatusConfiguringIAMDatabaseAuth,
			InstanceStatusConfiguringLogExports,
			InstanceStatusCreating,
			InstanceStatusMaintenance,
			InstanceStatusModifying,
			InstanceStatusRebooting,
			InstanceStatusRenaming,
			InstanceStatusResettingMasterCredentials,
			InstanceStatusStarting,
			InstanceStatusStorageOptimization,
			InstanceStatusUpgrading,
		},
		Target:     []string{InstanceStatusAvailable},
		Refresh:    statusDBInstanceSDKv1(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBInstance); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterInstanceDeleted(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			InstanceStatusConfiguringLogExports,
			InstanceStatusDeletePreCheck,
			InstanceStatusDeleting,
			InstanceStatusModifying,
		},
		Target:     []string{},
		Refresh:    statusDBInstanceSDKv1(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBInstance); ok {
		return output, err
	}

	return nil, err
}

func clusterSetResourceDataEngineVersionFromClusterInstance(d *schema.ResourceData, c *rds.DBInstance) {
	oldVersion := d.Get(names.AttrEngineVersion).(string)
	newVersion := aws.StringValue(c.EngineVersion)
	var pendingVersion string
	if c.PendingModifiedValues != nil && c.PendingModifiedValues.EngineVersion != nil {
		pendingVersion = aws.StringValue(c.PendingModifiedValues.EngineVersion)
	}
	compareActualEngineVersion(d, oldVersion, newVersion, pendingVersion)
}
