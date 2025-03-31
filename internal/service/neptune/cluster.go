// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	awstypes "github.com/aws/aws-sdk-go-v2/service/neptune/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
	// A constant for the supported CloudwatchLogsExports types
	// is not currently available in the AWS sdk-for-go
	// https://docs.aws.amazon.com/sdk-for-go/api/service/neptune/#pkg-constants
	cloudWatchLogsExportsAudit     = "audit"
	cloudWatchLogsExportsSlowQuery = "slowquery"

	DefaultPort = 8182

	oldServerlessMinNCUs = 2.5
	ServerlessMinNCUs    = 1.0
	ServerlessMaxNCUs    = 128.0
)

// @SDKResource("aws_neptune_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
func resourceCluster() *schema.Resource {
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
			names.AttrAllowMajorVersionUpgrade: {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
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
			names.AttrClusterIdentifier: {
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
			names.AttrDeletionProtection: {
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
						cloudWatchLogsExportsSlowQuery,
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
				Default:      engineNeptune,
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
			names.AttrKMSKeyARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"neptune_cluster_parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
			names.AttrPort: {
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
			"serverless_v2_scaling_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrMaxCapacity: {
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
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			names.AttrStorageType: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// https://docs.aws.amazon.com/neptune/latest/userguide/storage-types.html#provisioned-iops-storage:
					// "You can determine whether a cluster is using I/O–Optimized storage using any describe- call. If the I/O–Optimized storage is enabled, the call returns a storage-type field set to iopt1".
					if old == "" && new == storageTypeStandard {
						return true
					}
					return new == old
				},
				ValidateFunc: validation.StringInSlice(storageType_Values(), false),
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
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	clusterID := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrClusterIdentifier).(string)),
		create.WithConfiguredPrefix(d.Get("cluster_identifier_prefix").(string)),
		create.WithDefaultPrefix("tf-"),
	).Generate()

	// Check if any of the parameters that require a cluster modification after creation are set.
	// See https://docs.aws.amazon.com/neptune/latest/userguide/backup-restore-restore-snapshot.html#backup-restore-restore-snapshot-considerations.
	clusterUpdate := false
	restoreDBClusterFromSnapshot := false
	if _, ok := d.GetOk("snapshot_identifier"); ok {
		restoreDBClusterFromSnapshot = true
	}

	serverlessConfiguration := expandServerlessConfiguration(d.Get("serverless_v2_scaling_configuration").([]any))
	inputC := &neptune.CreateDBClusterInput{
		CopyTagsToSnapshot:               aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
		DBClusterIdentifier:              aws.String(clusterID),
		DeletionProtection:               aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
		Engine:                           aws.String(d.Get(names.AttrEngine).(string)),
		Port:                             aws.Int32(int32(d.Get(names.AttrPort).(int))),
		ServerlessV2ScalingConfiguration: serverlessConfiguration,
		StorageEncrypted:                 aws.Bool(d.Get(names.AttrStorageEncrypted).(bool)),
		Tags:                             getTagsIn(ctx),
	}
	inputR := &neptune.RestoreDBClusterFromSnapshotInput{
		CopyTagsToSnapshot:               aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
		DBClusterIdentifier:              aws.String(clusterID),
		DeletionProtection:               aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
		Engine:                           aws.String(d.Get(names.AttrEngine).(string)),
		Port:                             aws.Int32(int32(d.Get(names.AttrPort).(int))),
		ServerlessV2ScalingConfiguration: serverlessConfiguration,
		SnapshotIdentifier:               aws.String(d.Get("snapshot_identifier").(string)),
		Tags:                             getTagsIn(ctx),
	}
	inputM := &neptune.ModifyDBClusterInput{
		ApplyImmediately:    aws.Bool(true),
		DBClusterIdentifier: aws.String(clusterID),
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZones); ok && v.(*schema.Set).Len() > 0 {
		v := v.(*schema.Set)

		inputC.AvailabilityZones = flex.ExpandStringValueSet(v)
		inputR.AvailabilityZones = flex.ExpandStringValueSet(v)
	}

	if v, ok := d.GetOk("backup_retention_period"); ok {
		v := int32(v.(int))

		inputC.BackupRetentionPeriod = aws.Int32(v)
		if restoreDBClusterFromSnapshot {
			clusterUpdate = true
			inputM.BackupRetentionPeriod = aws.Int32(v)
		}
	}

	if v, ok := d.GetOk("enable_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
		v := v.(*schema.Set)

		inputC.EnableCloudwatchLogsExports = flex.ExpandStringValueSet(v)
		inputR.EnableCloudwatchLogsExports = flex.ExpandStringValueSet(v)
	}

	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		v := v.(string)

		inputC.EngineVersion = aws.String(v)
		inputR.EngineVersion = aws.String(v)
	}

	if v, ok := d.GetOk("global_cluster_identifier"); ok {
		v := v.(string)

		inputC.GlobalClusterIdentifier = aws.String(v)
	}

	if v, ok := d.GetOk("iam_database_authentication_enabled"); ok {
		v := v.(bool)

		inputC.EnableIAMDatabaseAuthentication = aws.Bool(v)
		inputR.EnableIAMDatabaseAuthentication = aws.Bool(v)
	}

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		v := v.(string)

		inputC.KmsKeyId = aws.String(v)
		inputR.KmsKeyId = aws.String(v)
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

	if v, ok := d.GetOk(names.AttrPreferredMaintenanceWindow); ok {
		v := v.(string)

		inputC.PreferredMaintenanceWindow = aws.String(v)
	}

	if v, ok := d.GetOk("replication_source_identifier"); ok {
		v := v.(string)

		inputC.ReplicationSourceIdentifier = aws.String(v)
	}

	if v, ok := d.GetOk(names.AttrStorageType); ok {
		v := v.(string)

		inputC.StorageType = aws.String(v)
		inputR.StorageType = aws.String(v)
	}

	if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
		v := v.(*schema.Set)

		inputC.VpcSecurityGroupIds = flex.ExpandStringValueSet(v)
		inputR.VpcSecurityGroupIds = flex.ExpandStringValueSet(v)
		if restoreDBClusterFromSnapshot {
			clusterUpdate = true
			inputM.VpcSecurityGroupIds = flex.ExpandStringValueSet(v)
		}
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (any, error) {
		if restoreDBClusterFromSnapshot {
			return conn.RestoreDBClusterFromSnapshot(ctx, inputR)
		}

		return conn.CreateDBCluster(ctx, inputC)
	}, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Cluster (%s): %s", clusterID, err)
	}

	d.SetId(clusterID)

	if _, err = waitDBClusterAvailable(ctx, conn, d.Id(), false, d.Timeout(schema.TimeoutCreate)); err != nil {
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
		_, err := conn.ModifyDBCluster(ctx, inputM)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Neptune Cluster (%s): %s", d.Id(), err)
		}

		if _, err = waitDBClusterAvailable(ctx, conn, d.Id(), true, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Neptune Cluster (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	dbc, err := findDBClusterByID(ctx, conn, d.Id())

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
	if globalCluster, err := findGlobalClusterByClusterARN(ctx, conn, aws.ToString(dbc.DBClusterArn)); tfresource.NotFound(err) || tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "Access Denied to API Version: APIGlobalDatabases") {
		d.Set("global_cluster_identifier", "")
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Global Cluster information for Neptune Cluster (%s): %s", d.Id(), err)
	} else {
		d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
	}

	arn := aws.ToString(dbc.DBClusterArn)
	d.Set(names.AttrARN, arn)
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
	d.Set("copy_tags_to_snapshot", dbc.CopyTagsToSnapshot)
	d.Set(names.AttrDeletionProtection, dbc.DeletionProtection)
	d.Set("enable_cloudwatch_logs_exports", dbc.EnabledCloudwatchLogsExports)
	d.Set(names.AttrEndpoint, dbc.Endpoint)
	d.Set(names.AttrEngineVersion, dbc.EngineVersion)
	d.Set(names.AttrEngine, dbc.Engine)
	d.Set(names.AttrHostedZoneID, dbc.HostedZoneId)
	d.Set("iam_database_authentication_enabled", dbc.IAMDatabaseAuthenticationEnabled)
	var iamRoles []string
	for _, v := range dbc.AssociatedRoles {
		iamRoles = append(iamRoles, aws.ToString(v.RoleArn))
	}
	d.Set("iam_roles", iamRoles)
	d.Set(names.AttrKMSKeyARN, dbc.KmsKeyId)
	d.Set("neptune_cluster_parameter_group_name", dbc.DBClusterParameterGroup)
	d.Set("neptune_subnet_group_name", dbc.DBSubnetGroup)
	d.Set(names.AttrPort, dbc.Port)
	d.Set("preferred_backup_window", dbc.PreferredBackupWindow)
	d.Set(names.AttrPreferredMaintenanceWindow, dbc.PreferredMaintenanceWindow)
	d.Set("reader_endpoint", dbc.ReaderEndpoint)
	d.Set("replication_source_identifier", dbc.ReplicationSourceIdentifier)
	if err := d.Set("serverless_v2_scaling_configuration", flattenServerlessV2ScalingConfigurationInfo(dbc.ServerlessV2ScalingConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting serverless_v2_scaling_configuration: %s", err)
	}
	d.Set(names.AttrStorageEncrypted, dbc.StorageEncrypted)
	d.Set(names.AttrStorageType, dbc.StorageType)
	var securityGroupIDs []string
	for _, v := range dbc.VpcSecurityGroups {
		securityGroupIDs = append(securityGroupIDs, aws.ToString(v.VpcSecurityGroupId))
	}
	d.Set(names.AttrVPCSecurityGroupIDs, securityGroupIDs)

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, "global_cluster_identifier", "iam_roles", "skip_final_snapshot") {
		allowMajorVersionUpgrade := d.Get(names.AttrAllowMajorVersionUpgrade).(bool)
		input := &neptune.ModifyDBClusterInput{
			AllowMajorVersionUpgrade: aws.Bool(allowMajorVersionUpgrade),
			ApplyImmediately:         aws.Bool(d.Get(names.AttrApplyImmediately).(bool)),
			DBClusterIdentifier:      aws.String(d.Id()),
		}

		if d.HasChange("backup_retention_period") {
			input.BackupRetentionPeriod = aws.Int32(int32(d.Get("backup_retention_period").(int)))
		}

		if d.HasChange("copy_tags_to_snapshot") {
			input.CopyTagsToSnapshot = aws.Bool(d.Get("copy_tags_to_snapshot").(bool))
		}

		if d.HasChange(names.AttrDeletionProtection) {
			input.DeletionProtection = aws.Bool(d.Get(names.AttrDeletionProtection).(bool))
		}

		if d.HasChange("enable_cloudwatch_logs_exports") {
			logs := &awstypes.CloudwatchLogsExportConfiguration{}

			old, new := d.GetChange("enable_cloudwatch_logs_exports")

			disableLogTypes := old.(*schema.Set).Difference(new.(*schema.Set))

			if disableLogTypes.Len() > 0 {
				logs.DisableLogTypes = flex.ExpandStringValueSet(disableLogTypes)
			}

			enableLogTypes := new.(*schema.Set).Difference(old.(*schema.Set))

			if enableLogTypes.Len() > 0 {
				logs.EnableLogTypes = flex.ExpandStringValueSet(enableLogTypes)
			}

			input.CloudwatchLogsExportConfiguration = logs
		}

		if d.HasChange(names.AttrEngineVersion) {
			input.EngineVersion = aws.String(d.Get(names.AttrEngineVersion).(string))
			input.DBClusterParameterGroupName = aws.String(d.Get("neptune_cluster_parameter_group_name").(string))
		}

		if d.HasChange("iam_database_authentication_enabled") {
			input.EnableIAMDatabaseAuthentication = aws.Bool(d.Get("iam_database_authentication_enabled").(bool))
		}

		if d.HasChange("neptune_cluster_parameter_group_name") {
			input.DBClusterParameterGroupName = aws.String(d.Get("neptune_cluster_parameter_group_name").(string))
		}

		// The DBInstanceParameterGroupName parameter is only valid in combination with the AllowMajorVersionUpgrade parameter.
		if allowMajorVersionUpgrade {
			if v, ok := d.GetOk("neptune_instance_parameter_group_name"); ok {
				input.DBInstanceParameterGroupName = aws.String(v.(string))
			}
		}

		if d.HasChange("preferred_backup_window") {
			input.PreferredBackupWindow = aws.String(d.Get("preferred_backup_window").(string))
		}

		if d.HasChange(names.AttrPreferredMaintenanceWindow) {
			input.PreferredMaintenanceWindow = aws.String(d.Get(names.AttrPreferredMaintenanceWindow).(string))
		}

		if d.HasChange("serverless_v2_scaling_configuration") {
			input.ServerlessV2ScalingConfiguration = expandServerlessConfiguration(d.Get("serverless_v2_scaling_configuration").([]any))
		}

		if d.HasChange(names.AttrStorageType) {
			input.StorageType = aws.String(d.Get(names.AttrStorageType).(string))
		}

		if d.HasChange(names.AttrVPCSecurityGroupIDs) {
			if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
				input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v)
			} else {
				input.VpcSecurityGroupIds = []string{}
			}
		}

		_, err := tfresource.RetryWhen(ctx, 5*time.Minute,
			func() (any, error) {
				return conn.ModifyDBCluster(ctx, input)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
					return true, err
				}

				if errs.IsA[*awstypes.InvalidDBClusterStateFault](err) {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Neptune Cluster (%s): %s", d.Id(), err)
		}

		if _, err = waitDBClusterAvailable(ctx, conn, d.Id(), true, d.Timeout(schema.TimeoutUpdate)); err != nil {
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

		if err := removeClusterFromGlobalCluster(ctx, conn, d.Get(names.AttrARN).(string), o, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
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

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	skipFinalSnapshot := d.Get("skip_final_snapshot").(bool)
	input := &neptune.DeleteDBClusterInput{
		DBClusterIdentifier: aws.String(d.Id()),
		SkipFinalSnapshot:   aws.Bool(skipFinalSnapshot),
	}

	if !skipFinalSnapshot {
		if v, ok := d.GetOk(names.AttrFinalSnapshotIdentifier); ok {
			input.FinalDBSnapshotIdentifier = aws.String(v.(string))
		} else {
			return sdkdiag.AppendErrorf(diags, "final_snapshot_identifier is required when skip_final_snapshot is false")
		}
	}

	if v, ok := d.GetOk("global_cluster_identifier"); ok {
		if err := removeClusterFromGlobalCluster(ctx, conn, d.Get(names.AttrARN).(string), v.(string), d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[DEBUG] Deleting Neptune Cluster: %s", d.Id())
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidDBClusterStateFault](ctx, d.Timeout(schema.TimeoutDelete), func() (any, error) {
		return conn.DeleteDBCluster(ctx, input)
	}, "is not currently in the available state")

	if errs.IsA[*awstypes.DBClusterNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Neptune Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitDBClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func addIAMRoleToCluster(ctx context.Context, conn *neptune.Client, clusterID, roleARN string) error {
	_, err := conn.AddRoleToDBCluster(ctx, &neptune.AddRoleToDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		RoleArn:             aws.String(roleARN),
	})

	return err
}

func removeIAMRoleFromCluster(ctx context.Context, conn *neptune.Client, clusterID, roleARN string) error {
	_, err := conn.RemoveRoleFromDBCluster(ctx, &neptune.RemoveRoleFromDBClusterInput{
		DBClusterIdentifier: aws.String(clusterID),
		RoleArn:             aws.String(roleARN),
	})

	return err
}

func removeClusterFromGlobalCluster(ctx context.Context, conn *neptune.Client, clusterARN, globalClusterID string, timeout time.Duration) error {
	input := &neptune.RemoveFromGlobalClusterInput{
		DbClusterIdentifier:     aws.String(clusterARN),
		GlobalClusterIdentifier: aws.String(globalClusterID),
	}

	_, err := conn.RemoveFromGlobalCluster(ctx, input)

	if errs.IsA[*awstypes.DBClusterNotFoundFault](err) || errs.IsA[*awstypes.GlobalClusterNotFoundFault](err) || tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "not found in global cluster") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("removing Neptune Cluster (%s) from Neptune Global Cluster (%s): %w", clusterARN, globalClusterID, err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, timeout, func() (any, error) {
		return findGlobalClusterByClusterARN(ctx, conn, clusterARN)
	})

	if err != nil {
		return fmt.Errorf("waiting for Neptune Cluster (%s) removal from Neptune Global Cluster (%s): %w", clusterARN, globalClusterID, err)
	}

	return nil
}

func findDBClusterByID(ctx context.Context, conn *neptune.Client, id string) (*awstypes.DBCluster, error) {
	input := &neptune.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(id),
	}
	output, err := findDBCluster(ctx, conn, input, tfslices.PredicateTrue[awstypes.DBCluster]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.DBClusterIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findClusterByARN(ctx context.Context, conn *neptune.Client, arn string) (*awstypes.DBCluster, error) {
	input := &neptune.DescribeDBClustersInput{}

	return findDBCluster(ctx, conn, input, func(v awstypes.DBCluster) bool {
		return aws.ToString(v.DBClusterArn) == arn
	})
}

func findDBCluster(ctx context.Context, conn *neptune.Client, input *neptune.DescribeDBClustersInput, filter tfslices.Predicate[awstypes.DBCluster]) (*awstypes.DBCluster, error) {
	output, err := findDBClusters(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBClusters(ctx context.Context, conn *neptune.Client, input *neptune.DescribeDBClustersInput, filter tfslices.Predicate[awstypes.DBCluster]) ([]awstypes.DBCluster, error) {
	var output []awstypes.DBCluster

	pages := neptune.NewDescribeDBClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.DBClusterNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusters {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusDBCluster(ctx context.Context, conn *neptune.Client, id string, waitNoPendingModifiedValues bool) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDBClusterByID(ctx, conn, id)

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

func waitDBClusterAvailable(ctx context.Context, conn *neptune.Client, id string, waitNoPendingModifiedValues bool, timeout time.Duration) (*awstypes.DBCluster, error) { //nolint:unparam
	pendingStatuses := []string{
		clusterStatusCreating,
		clusterStatusBackingUp,
		clusterStatusModifying,
		clusterStatusPreparingDataMigration,
		clusterStatusMigrating,
		clusterStatusConfiguringIAMDatabaseAuth,
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

	if output, ok := outputRaw.(*awstypes.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterDeleted(ctx context.Context, conn *neptune.Client, id string, timeout time.Duration) (*awstypes.DBCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			clusterStatusAvailable,
			clusterStatusDeleting,
			clusterStatusBackingUp,
			clusterStatusModifying,
		},
		Target:     []string{},
		Refresh:    statusDBCluster(ctx, conn, id, false),
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

func expandServerlessConfiguration(l []any) *awstypes.ServerlessV2ScalingConfiguration {
	if len(l) == 0 {
		return nil
	}

	tfMap := l[0].(map[string]any)
	return &awstypes.ServerlessV2ScalingConfiguration{
		MinCapacity: aws.Float64(tfMap["min_capacity"].(float64)),
		MaxCapacity: aws.Float64(tfMap[names.AttrMaxCapacity].(float64)),
	}
}

func flattenServerlessV2ScalingConfigurationInfo(serverlessConfig *awstypes.ServerlessV2ScalingConfigurationInfo) []map[string]any {
	if serverlessConfig == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"min_capacity":        aws.ToFloat64(serverlessConfig.MinCapacity),
		names.AttrMaxCapacity: aws.ToFloat64(serverlessConfig.MaxCapacity),
	}

	return []map[string]any{m}
}
