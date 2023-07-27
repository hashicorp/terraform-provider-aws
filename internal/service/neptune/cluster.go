// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
	// A constant for the supported CloudwatchLogsExports types
	// is not currently available in the AWS sdk-for-go
	// https://docs.aws.amazon.com/sdk-for-go/api/service/neptune/#pkg-constants
	cloudWatchLogsExportsAudit = "audit"

	DefaultPort = 8182

	oldServerlessMinNCUs = 2.5
	ServerlessMinNCUs    = 1.0
	ServerlessMaxNCUs    = 128.0
)

// @SDKResource("aws_neptune_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		UpdateWithoutTimeout: resourceClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// allow_major_version_upgrade is used to indicate whether upgrades between different major versions
			// are allowed.
			"allow_major_version_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			// apply_immediately is used to determine when the update modifications
			// take place.
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
				MaxItems: 3,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"backup_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntAtMost(35),
			},
			"cluster_identifier": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"cluster_identifier_prefix"},
				ValidateFunc:  validIdentifier,
			},
			"cluster_identifier_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifierPrefix,
			},
			"cluster_members": {
				Type:     schema.TypeSet,
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
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"enable_cloudwatch_logs_exports": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						cloudWatchLogsExportsAudit,
					}, false),
				},
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "neptune",
				ValidateFunc: validEngine(),
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validGlobalCusterIdentifier,
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
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"neptune_cluster_parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default.neptune1",
			},
			"neptune_instance_parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"neptune_subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  DefaultPort,
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
			"serverless_v2_scaling_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_capacity": {
							Type:     schema.TypeFloat,
							Optional: true,
							Default:  ServerlessMaxNCUs,
							// Maximum capacity is 128 NCUs
							// see: https://docs.aws.amazon.com/neptune/latest/userguide/neptune-serverless-capacity-scaling.html
							ValidateFunc: validation.FloatAtMost(ServerlessMaxNCUs),
						},
						"min_capacity": {
							Type:     schema.TypeFloat,
							Optional: true,
							Default:  oldServerlessMinNCUs,
							// Minimum capacity is 1.0 NCU
							// see: https://docs.aws.amazon.com/neptune/latest/userguide/neptune-serverless-capacity-scaling.html
							ValidateFunc: validation.FloatAtLeast(ServerlessMinNCUs),
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
			},
			"storage_encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	// Check if any of the parameters that require a cluster modification after creation are set.
	// See https://docs.aws.amazon.com/neptune/latest/userguide/backup-restore-restore-snapshot.html#backup-restore-restore-snapshot-considerations.
	clusterUpdate := false
	restoreDBClusterFromSnapshot := false
	if _, ok := d.GetOk("snapshot_identifier"); ok {
		restoreDBClusterFromSnapshot = true
	}

	var clusterID string
	if v, ok := d.GetOk("cluster_identifier"); ok {
		clusterID = v.(string)
	} else if v, ok := d.GetOk("cluster_identifier_prefix"); ok {
		clusterID = id.PrefixedUniqueId(v.(string))
	} else {
		clusterID = id.PrefixedUniqueId("tf-")
	}
	serverlessConfiguration := expandServerlessConfiguration(d.Get("serverless_v2_scaling_configuration").([]interface{}))
	inputC := &neptune.CreateDBClusterInput{
		DBClusterIdentifier:              aws.String(clusterID),
		CopyTagsToSnapshot:               aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
		Engine:                           aws.String(d.Get("engine").(string)),
		Port:                             aws.Int64(int64(d.Get("port").(int))),
		StorageEncrypted:                 aws.Bool(d.Get("storage_encrypted").(bool)),
		DeletionProtection:               aws.Bool(d.Get("deletion_protection").(bool)),
		Tags:                             getTagsIn(ctx),
		ServerlessV2ScalingConfiguration: serverlessConfiguration,
	}
	inputR := &neptune.RestoreDBClusterFromSnapshotInput{
		DBClusterIdentifier:              aws.String(clusterID),
		CopyTagsToSnapshot:               aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
		Engine:                           aws.String(d.Get("engine").(string)),
		Port:                             aws.Int64(int64(d.Get("port").(int))),
		SnapshotIdentifier:               aws.String(d.Get("snapshot_identifier").(string)),
		DeletionProtection:               aws.Bool(d.Get("deletion_protection").(bool)),
		Tags:                             getTagsIn(ctx),
		ServerlessV2ScalingConfiguration: serverlessConfiguration,
	}
	inputM := &neptune.ModifyDBClusterInput{
		ApplyImmediately:    aws.Bool(true),
		DBClusterIdentifier: aws.String(clusterID),
	}

	if v, ok := d.GetOk("availability_zones"); ok && v.(*schema.Set).Len() > 0 {
		v := v.(*schema.Set)

		inputC.AvailabilityZones = flex.ExpandStringSet(v)
		inputR.AvailabilityZones = flex.ExpandStringSet(v)
	}

	if v, ok := d.GetOk("backup_retention_period"); ok {
		v := int64(v.(int))

		inputC.BackupRetentionPeriod = aws.Int64(v)
		if restoreDBClusterFromSnapshot {
			clusterUpdate = true
			inputM.BackupRetentionPeriod = aws.Int64(v)
		}
	}

	if v, ok := d.GetOk("global_cluster_identifier"); ok {
		v := v.(string)

		inputC.GlobalClusterIdentifier = aws.String(v)
	}

	if v, ok := d.GetOk("enable_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
		v := v.(*schema.Set)

		inputC.EnableCloudwatchLogsExports = flex.ExpandStringSet(v)
		inputR.EnableCloudwatchLogsExports = flex.ExpandStringSet(v)
	}

	if v, ok := d.GetOk("engine_version"); ok {
		v := v.(string)

		inputC.EngineVersion = aws.String(v)
		inputR.EngineVersion = aws.String(v)
	}

	if v, ok := d.GetOk("iam_database_authentication_enabled"); ok {
		v := v.(bool)

		inputC.EnableIAMDatabaseAuthentication = aws.Bool(v)
		inputR.EnableIAMDatabaseAuthentication = aws.Bool(v)
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		v := v.(string)

		inputC.KmsKeyId = aws.String(v)
	}

	if v, ok := d.GetOk("neptune_cluster_parameter_group_name"); ok {
		v := v.(string)

		inputC.DBClusterParameterGroupName = aws.String(v)
		if restoreDBClusterFromSnapshot {
			clusterUpdate = true
			inputM.DBClusterParameterGroupName = aws.String(v)
		}
	}

	if v, ok := d.GetOk("neptune_subnet_group_name"); ok {
		v := v.(string)

		inputC.DBSubnetGroupName = aws.String(v)
		inputR.DBSubnetGroupName = aws.String(v)
	}

	if v, ok := d.GetOk("preferred_backup_window"); ok {
		v := v.(string)

		inputC.PreferredBackupWindow = aws.String(v)
	}

	if v, ok := d.GetOk("preferred_maintenance_window"); ok {
		v := v.(string)

		inputC.PreferredMaintenanceWindow = aws.String(v)
	}

	if v, ok := d.GetOk("replication_source_identifier"); ok {
		v := v.(string)

		inputC.ReplicationSourceIdentifier = aws.String(v)
	}

	if v, ok := d.GetOk("vpc_security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		v := v.(*schema.Set)

		inputC.VpcSecurityGroupIds = flex.ExpandStringSet(v)
		inputR.VpcSecurityGroupIds = flex.ExpandStringSet(v)
		if restoreDBClusterFromSnapshot {
			clusterUpdate = true
			inputM.VpcSecurityGroupIds = flex.ExpandStringSet(v)
		}
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		if restoreDBClusterFromSnapshot {
			return conn.RestoreDBClusterFromSnapshotWithContext(ctx, inputR)
		}

		return conn.CreateDBClusterWithContext(ctx, inputC)
	}, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Cluster (%s): %s", clusterID, err)
	}

	d.SetId(clusterID)

	if _, err = waitClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune Cluster (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("iam_roles"); ok {
		for _, v := range v.(*schema.Set).List() {
			v := v.(string)

			if err := addIAMRoleToCluster(ctx, conn, d.Id(), v); err != nil {
				return sdkdiag.AppendErrorf(diags, "adding IAM Role (%s) to Neptune Cluster (%s): %s", v, d.Id(), err)
			}
		}
	}

	if clusterUpdate {
		_, err := conn.ModifyDBClusterWithContext(ctx, inputM)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Neptune Cluster (%s): %s", d.Id(), err)
		}

		if _, err = waitClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Neptune Cluster (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	dbc, err := FindClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Neptune Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Cluster (%s): %s", d.Id(), err)
	}

	// Ignore the following API error for regions/partitions that do not support Neptune Global Clusters:
	// InvalidParameterValue: Access Denied to API Version: APIGlobalDatabases
	if globalCluster, err := findGlobalClusterByClusterARN(ctx, conn, aws.StringValue(dbc.DBClusterArn)); tfresource.NotFound(err) || tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Access Denied to API Version: APIGlobalDatabases") {
		d.Set("global_cluster_identifier", "")
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Global Cluster information for Neptune Cluster (%s): %s", d.Id(), err)
	} else {
		d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
	}

	arn := aws.StringValue(dbc.DBClusterArn)
	d.Set("arn", arn)
	d.Set("availability_zones", aws.StringValueSlice(dbc.AvailabilityZones))
	d.Set("backup_retention_period", dbc.BackupRetentionPeriod)
	d.Set("cluster_identifier", dbc.DBClusterIdentifier)
	var clusterMembers []string
	for _, v := range dbc.DBClusterMembers {
		clusterMembers = append(clusterMembers, aws.StringValue(v.DBInstanceIdentifier))
	}
	d.Set("cluster_members", clusterMembers)
	d.Set("cluster_resource_id", dbc.DbClusterResourceId)
	d.Set("copy_tags_to_snapshot", dbc.CopyTagsToSnapshot)
	d.Set("deletion_protection", dbc.DeletionProtection)
	d.Set("enable_cloudwatch_logs_exports", aws.StringValueSlice(dbc.EnabledCloudwatchLogsExports))
	d.Set("endpoint", dbc.Endpoint)
	d.Set("engine_version", dbc.EngineVersion)
	d.Set("engine", dbc.Engine)
	d.Set("hosted_zone_id", dbc.HostedZoneId)
	d.Set("iam_database_authentication_enabled", dbc.IAMDatabaseAuthenticationEnabled)
	var iamRoles []string
	for _, v := range dbc.AssociatedRoles {
		iamRoles = append(iamRoles, aws.StringValue(v.RoleArn))
	}
	d.Set("iam_roles", iamRoles)
	d.Set("kms_key_arn", dbc.KmsKeyId)
	d.Set("neptune_cluster_parameter_group_name", dbc.DBClusterParameterGroup)
	d.Set("neptune_subnet_group_name", dbc.DBSubnetGroup)
	d.Set("port", dbc.Port)
	d.Set("preferred_backup_window", dbc.PreferredBackupWindow)
	d.Set("preferred_maintenance_window", dbc.PreferredMaintenanceWindow)
	d.Set("reader_endpoint", dbc.ReaderEndpoint)
	d.Set("replication_source_identifier", dbc.ReplicationSourceIdentifier)
	if err := d.Set("serverless_v2_scaling_configuration", flattenServerlessV2ScalingConfigurationInfo(dbc.ServerlessV2ScalingConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting serverless_v2_scaling_configuration: %s", err)
	}
	d.Set("storage_encrypted", dbc.StorageEncrypted)
	var securityGroupIDs []string
	for _, v := range dbc.VpcSecurityGroups {
		securityGroupIDs = append(securityGroupIDs, aws.StringValue(v.VpcSecurityGroupId))
	}
	d.Set("vpc_security_group_ids", securityGroupIDs)

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	if d.HasChangesExcept("tags", "tags_all", "iam_roles", "global_cluster_identifier") {
		allowMajorVersionUpgrade := d.Get("allow_major_version_upgrade").(bool)
		input := &neptune.ModifyDBClusterInput{
			AllowMajorVersionUpgrade: aws.Bool(allowMajorVersionUpgrade),
			ApplyImmediately:         aws.Bool(d.Get("apply_immediately").(bool)),
			DBClusterIdentifier:      aws.String(d.Id()),
		}

		if d.HasChange("copy_tags_to_snapshot") {
			input.CopyTagsToSnapshot = aws.Bool(d.Get("copy_tags_to_snapshot").(bool))
		}

		// The DBInstanceParameterGroupName parameter is only valid in combination with the AllowMajorVersionUpgrade parameter.
		if allowMajorVersionUpgrade {
			if v, ok := d.GetOk("neptune_instance_parameter_group_name"); ok {
				input.DBInstanceParameterGroupName = aws.String(v.(string))
			}
		}

		if d.HasChange("vpc_security_group_ids") {
			if v := d.Get("vpc_security_group_ids").(*schema.Set); v.Len() > 0 {
				input.VpcSecurityGroupIds = flex.ExpandStringSet(v)
			} else {
				input.VpcSecurityGroupIds = aws.StringSlice([]string{})
			}
		}

		if d.HasChange("enable_cloudwatch_logs_exports") {
			logs := &neptune.CloudwatchLogsExportConfiguration{}

			old, new := d.GetChange("enable_cloudwatch_logs_exports")

			disableLogTypes := old.(*schema.Set).Difference(new.(*schema.Set))

			if disableLogTypes.Len() > 0 {
				logs.SetDisableLogTypes(flex.ExpandStringSet(disableLogTypes))
			}

			enableLogTypes := new.(*schema.Set).Difference(old.(*schema.Set))

			if enableLogTypes.Len() > 0 {
				logs.SetEnableLogTypes(flex.ExpandStringSet(enableLogTypes))
			}

			input.CloudwatchLogsExportConfiguration = logs
		}

		if d.HasChange("preferred_backup_window") {
			input.PreferredBackupWindow = aws.String(d.Get("preferred_backup_window").(string))
		}

		if d.HasChange("preferred_maintenance_window") {
			input.PreferredMaintenanceWindow = aws.String(d.Get("preferred_maintenance_window").(string))
		}

		if d.HasChange("backup_retention_period") {
			input.BackupRetentionPeriod = aws.Int64(int64(d.Get("backup_retention_period").(int)))
		}

		if d.HasChange("neptune_cluster_parameter_group_name") {
			input.DBClusterParameterGroupName = aws.String(d.Get("neptune_cluster_parameter_group_name").(string))
		}

		if d.HasChange("iam_database_authentication_enabled") {
			input.EnableIAMDatabaseAuthentication = aws.Bool(d.Get("iam_database_authentication_enabled").(bool))
		}

		if d.HasChange("deletion_protection") {
			input.DeletionProtection = aws.Bool(d.Get("deletion_protection").(bool))
		}

		if d.HasChange("engine_version") {
			input.EngineVersion = aws.String(d.Get("engine_version").(string))
			input.DBClusterParameterGroupName = aws.String(d.Get("neptune_cluster_parameter_group_name").(string))
		}

		if d.HasChange("serverless_v2_scaling_configuration") {
			input.ServerlessV2ScalingConfiguration = expandServerlessConfiguration(d.Get("serverless_v2_scaling_configuration").([]interface{}))
		}

		_, err := tfresource.RetryWhen(ctx, 5*time.Minute,
			func() (interface{}, error) {
				return conn.ModifyDBClusterWithContext(ctx, input)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
					return true, err
				}

				if tfawserr.ErrCodeEquals(err, neptune.ErrCodeInvalidDBClusterStateFault) {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Neptune Cluster (%s): %s", d.Id(), err)
		}

		if _, err = waitClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Neptune Cluster (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("global_cluster_identifier") {
		oRaw, nRaw := d.GetChange("global_cluster_identifier")
		o := oRaw.(string)
		n := nRaw.(string)

		if o == "" {
			return sdkdiag.AppendErrorf(diags, "existing Neptune Clusters cannot be added to an existing Neptune Global Cluster")
		}

		if n != "" {
			return sdkdiag.AppendErrorf(diags, "existing Neptune Clusters cannot be migrated between existing Neptune Global Clusters")
		}

		if err := removeClusterFromGlobalCluster(ctx, conn, d.Get("arn").(string), o, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return append(diags, diag.FromErr(err)...)
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
		delRoles := os.Difference(ns)
		addRoles := ns.Difference(os)

		for _, v := range addRoles.List() {
			v := v.(string)

			if err := addIAMRoleToCluster(ctx, conn, d.Id(), v); err != nil {
				return sdkdiag.AppendErrorf(diags, "adding IAM Role (%s) to Neptune Cluster (%s): %s", v, d.Id(), err)
			}
		}

		for _, v := range delRoles.List() {
			v := v.(string)

			if err := removeIAMRoleFromCluster(ctx, conn, d.Id(), v); err != nil {
				return sdkdiag.AppendErrorf(diags, "removing IAM Role (%s) from Neptune Cluster (%s): %s", v, d.Id(), err)
			}
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	skipFinalSnapshot := d.Get("skip_final_snapshot").(bool)
	input := neptune.DeleteDBClusterInput{
		DBClusterIdentifier: aws.String(d.Id()),
		SkipFinalSnapshot:   aws.Bool(skipFinalSnapshot),
	}

	if !skipFinalSnapshot {
		if v, ok := d.GetOk("final_snapshot_identifier"); ok {
			input.FinalDBSnapshotIdentifier = aws.String(v.(string))
		} else {
			return sdkdiag.AppendErrorf(diags, "final_snapshot_identifier is required when skip_final_snapshot is false")
		}
	}

	if v, ok := d.GetOk("global_cluster_identifier"); ok {
		if err := removeClusterFromGlobalCluster(ctx, conn, d.Get("arn").(string), v.(string), d.Timeout(schema.TimeoutDelete)); err != nil {
			return append(diags, diag.FromErr(err)...)
		}
	}

	log.Printf("[DEBUG] Deleting Neptune Cluster: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteDBClusterWithContext(ctx, &input)
	}, neptune.ErrCodeInvalidDBClusterStateFault, "is not currently in the available state")

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterNotFoundFault) {
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Neptune Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func addIAMRoleToCluster(ctx context.Context, conn *neptune.Neptune, clusterID, roleARN string) error {
	_, err := conn.AddRoleToDBClusterWithContext(ctx, &neptune.AddRoleToDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		RoleArn:             aws.String(roleARN),
	})

	return err
}

func removeIAMRoleFromCluster(ctx context.Context, conn *neptune.Neptune, clusterID, roleARN string) error {
	_, err := conn.RemoveRoleFromDBClusterWithContext(ctx, &neptune.RemoveRoleFromDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		RoleArn:             aws.String(roleARN),
	})

	return err
}

func removeClusterFromGlobalCluster(ctx context.Context, conn *neptune.Neptune, clusterARN, globalClusterID string, timeout time.Duration) error {
	input := &neptune.RemoveFromGlobalClusterInput{
		DbClusterIdentifier:     aws.String(clusterARN),
		GlobalClusterIdentifier: aws.String(globalClusterID),
	}

	_, err := conn.RemoveFromGlobalClusterWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterNotFoundFault, neptune.ErrCodeGlobalClusterNotFoundFault) || tfawserr.ErrMessageContains(err, "InvalidParameterValue", "is not found in global cluster") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("removing Neptune Cluster (%s) from Neptune Global Cluster (%s): %w", clusterARN, globalClusterID, err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, timeout, func() (interface{}, error) {
		return findGlobalClusterByClusterARN(ctx, conn, clusterARN)
	})

	if err != nil {
		return fmt.Errorf("waiting for Neptune Cluster (%s) removal from Neptune Global Cluster (%s): %w", clusterARN, globalClusterID, err)
	}

	return nil
}

func FindClusterByID(ctx context.Context, conn *neptune.Neptune, id string) (*neptune.DBCluster, error) {
	input := &neptune.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(id),
	}

	output, err := conn.DescribeDBClustersWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterNotFoundFault) {
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

	dbCluster := output.DBClusters[0]

	// Eventual consistency check.
	if aws.StringValue(dbCluster.DBClusterIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return dbCluster, nil
}

func findClusterByARN(ctx context.Context, conn *neptune.Neptune, arn string) (*neptune.DBCluster, error) {
	input := &neptune.DescribeDBClustersInput{}
	var output *neptune.DBCluster

	err := conn.DescribeDBClustersPagesWithContext(ctx, input, func(page *neptune.DescribeDBClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBClusters {
			if v == nil {
				continue
			}

			if aws.StringValue(v.DBClusterArn) == arn {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func statusCluster(ctx context.Context, conn *neptune.Neptune, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitClusterAvailable(ctx context.Context, conn *neptune.Neptune, id string, timeout time.Duration) (*neptune.DBCluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			"creating",
			"backing-up",
			"modifying",
			"preparing-data-migration",
			"migrating",
			"configuring-iam-database-auth",
			"upgrading",
		},
		Target:     []string{"available"},
		Refresh:    statusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptune.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *neptune.Neptune, id string, timeout time.Duration) (*neptune.DBCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			"available",
			"deleting",
			"backing-up",
			"modifying",
		},
		Target:     []string{},
		Refresh:    statusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptune.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func expandServerlessConfiguration(l []interface{}) *neptune.ServerlessV2ScalingConfiguration {
	if len(l) == 0 {
		return nil
	}

	tfMap := l[0].(map[string]interface{})
	return &neptune.ServerlessV2ScalingConfiguration{
		MinCapacity: aws.Float64(tfMap["min_capacity"].(float64)),
		MaxCapacity: aws.Float64(tfMap["max_capacity"].(float64)),
	}
}

func flattenServerlessV2ScalingConfigurationInfo(serverlessConfig *neptune.ServerlessV2ScalingConfigurationInfo) []map[string]interface{} {
	if serverlessConfig == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"min_capacity": aws.Float64Value(serverlessConfig.MinCapacity),
		"max_capacity": aws.Float64Value(serverlessConfig.MaxCapacity),
	}

	return []map[string]interface{}{m}
}
