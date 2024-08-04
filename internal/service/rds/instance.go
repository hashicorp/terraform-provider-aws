// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	rds_sdkv2 "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	tfawserr_sdkv2 "github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NOTE ON "ID", "IDENTIFIER":
// ID is overloaded and potentially confusing. Hopefully this clears it up.
// * ID, as in d.Id(), d.SetId(), is:
//    - the same as AWS calls the "dbi-resource-id" a/k/a "database instance resource ID"
//    - unchangeable/immutable
//    - called either "id" or "resource_id" in schema/state (previously was only "resource_id")
// * "identifier" is:
//    - user-defined identifier which AWS calls "identifier"
//    - can be updated
//    - called "identifier" in the schema/state (previously was also "id")

// @SDKResource("aws_db_instance", name="DB Instance")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go/service/rds;rds.DBInstance")
// @Testing(importIgnore="apply_immediately;password")
func ResourceInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceCreate,
		ReadWithoutTimeout:   resourceInstanceRead,
		UpdateWithoutTimeout: resourceInstanceUpdate,
		DeleteWithoutTimeout: resourceInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceInstanceImport,
		},

		SchemaVersion: 2,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceInstanceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: InstanceStateUpgradeV0,
				Version: 0,
			},
			{
				Type:    resourceInstanceResourceV1().CoreConfigSchema().ImpliedType(),
				Upgrade: InstanceStateUpgradeV1,
				Version: 1,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Update: schema.DefaultTimeout(80 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAddress: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAllocatedStorage: {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					mas := d.Get("max_allocated_storage").(int)

					newInt, err := strconv.Atoi(new)
					if err != nil {
						return false
					}

					oldInt, err := strconv.Atoi(old)
					if err != nil {
						return false
					}

					// Allocated is higher than the configuration
					// and autoscaling is enabled
					if oldInt > newInt && mas > newInt {
						return true
					}

					return false
				},
			},
			names.AttrAllowMajorVersionUpgrade: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			// apply_immediately is used to determine when the update modifications
			// take place.
			// See http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html
			names.AttrApplyImmediately: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
				Computed: true,
				ForceNew: true,
			},
			"backup_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 35),
			},
			"backup_target": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(backupTarget_Values(), false),
				ConflictsWith: []string{
					"s3_import",
				},
			},
			"backup_window": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidOnceADayWindowFormat,
			},
			"blue_green_update": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"ca_cert_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"character_set_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ConflictsWith: []string{
					"s3_import",
				},
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					for _, conflictAttr := range []string{
						"replicate_source_db",
						// s3_import is handled by the schema ConflictsWith
						"snapshot_identifier",
						"restore_to_point_in_time",
					} {
						if _, ok := d.GetOk(conflictAttr); ok {
							return true
						}
					}
					return false
				},
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
			"customer_owned_ip_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"db_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ConflictsWith: []string{
					"replicate_source_db",
				},
			},
			"db_subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"dedicated_log_volume": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"domain_fqdn", "domain_ou", "domain_auth_secret_arn", "domain_dns_ips"},
			},
			"domain_auth_secret_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  verify.ValidARN,
				ConflictsWith: []string{names.AttrDomain, "domain_iam_role_name"},
			},
			"domain_dns_ips": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 2,
				MaxItems: 2,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPAddress,
				},
				ConflictsWith: []string{names.AttrDomain, "domain_iam_role_name"},
			},
			"domain_fqdn": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{names.AttrDomain, "domain_iam_role_name"},
			},
			"domain_iam_role_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"domain_fqdn", "domain_ou", "domain_auth_secret_arn", "domain_dns_ips"},
			},
			"domain_ou": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{names.AttrDomain, "domain_iam_role_name"},
			},
			"enabled_cloudwatch_logs_exports": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(InstanceExportableLogType_Values(), false),
				},
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					value := v.(string)
					return strings.ToLower(value)
				},
			},
			"engine_lifecycle_support": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(engineLifecycleSupport_Values(), false),
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
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with alphabetic character"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]+$`), "must only contain alphanumeric characters and hyphens"),
					validation.StringDoesNotMatch(regexache.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexache.MustCompile(`-$`), "cannot end in a hyphen"),
				),
			},
			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_database_authentication_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrIdentifier: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"identifier_prefix"},
				ValidateFunc:  validIdentifier,
			},
			"identifier_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{names.AttrIdentifier},
				ValidateFunc:  validIdentifierPrefix,
			},
			"instance_class": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrIOPS: {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"latest_restorable_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"license_model": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"listener_endpoint": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAddress: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrHostedZoneID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrPort: {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"maintenance_window": {
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
			"manage_master_user_password": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{names.AttrPassword},
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
			"max_allocated_storage": {
				Type:     schema.TypeInt,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "0" && new == fmt.Sprintf("%d", d.Get(names.AttrAllocatedStorage).(int)) {
						return true
					}
					return false
				},
			},
			"monitoring_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntInSlice([]int{0, 1, 5, 10, 15, 30, 60}),
			},
			"monitoring_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"multi_az": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"nchar_character_set_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"network_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(NetworkType_Values(), false),
			},
			"option_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrParameterGroupName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrPassword: {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"manage_master_user_password"},
			},
			"performance_insights_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
					validation.IntDivisibleBy(31),
				),
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrPubliclyAccessible: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"replica_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(rds.ReplicaMode_Values(), false),
			},
			"replicas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"replicate_source_db": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"restore_to_point_in_time": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				ConflictsWith: []string{
					"s3_import",
					"snapshot_identifier",
					"replicate_source_db",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"restore_time": {
							Type:          schema.TypeString,
							Optional:      true,
							ValidateFunc:  verify.ValidUTCTimestamp,
							ConflictsWith: []string{"restore_to_point_in_time.0.use_latest_restorable_time"},
						},
						"source_db_instance_automated_backups_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source_db_instance_identifier": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source_dbi_resource_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"use_latest_restorable_time": {
							Type:          schema.TypeBool,
							Optional:      true,
							ConflictsWith: []string{"restore_to_point_in_time.0.restore_time"},
						},
					},
				},
			},
			"s3_import": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ConflictsWith: []string{
					"snapshot_identifier",
					"replicate_source_db",
				},
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
			},
			"skip_final_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"snapshot_identifier": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"storage_throughput": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrStorageType: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"timezone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ConflictsWith: []string{
					"s3_import",
				},
			},
			"upgrade_storage_config": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrUsername: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"replicate_source_db"},
			},
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
				if !d.Get("blue_green_update.0.enabled").(bool) {
					return nil
				}

				engine := d.Get(names.AttrEngine).(string)
				if !slices.Contains(dbInstanceValidBlueGreenEngines(), engine) {
					return fmt.Errorf(`"blue_green_update.enabled" cannot be set when "engine" is %q.`, engine)
				}
				return nil
			},
			func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
				if !d.Get("blue_green_update.0.enabled").(bool) {
					return nil
				}

				source := d.Get("replicate_source_db").(string)
				if source != "" {
					return errors.New(`"blue_green_update.enabled" cannot be set when "replicate_source_db" is set.`)
				}
				return nil
			},
		),
	}
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	// Some API calls (e.g. CreateDBInstanceReadReplica, RestoreDBInstanceFromDBSnapshot
	// RestoreDBInstanceToPointInTime do not support all parameters to
	// correctly apply all settings in one pass. For missing parameters or
	// unsupported configurations, we may need to call ModifyDBInstance
	// afterwards to prevent Terraform operators from API errors or needing
	// to double apply.
	var requiresModifyDbInstance bool
	modifyDbInstanceInput := &rds.ModifyDBInstanceInput{
		ApplyImmediately: aws.Bool(true),
	}

	// Some ModifyDBInstance parameters (e.g. DBParameterGroupName) require
	// a database instance reboot to take effect. During resource creation,
	// we expect everything to be in sync before returning completion.
	var requiresRebootDbInstance bool

	// See discussion of IDs at the top of file - this is NOT d.Id()
	identifier := create.Name(d.Get(names.AttrIdentifier).(string), d.Get("identifier_prefix").(string))

	var resourceID string // will be assigned depending on how it is created

	if _, ok := d.GetOk("character_set_name"); ok {
		charSetPath := cty.GetAttrPath("character_set_name")
		for _, conflictAttr := range []string{
			"replicate_source_db",
			// s3_import is handled by the schema ConflictsWith
			"snapshot_identifier",
			"restore_to_point_in_time",
		} {
			if _, ok := d.GetOk(conflictAttr); ok {
				diags = append(diags, errs.NewAttributeConflictsWillBeError(
					charSetPath,
					cty.GetAttrPath(conflictAttr),
				))
			}
		}
	}

	if v, ok := d.GetOk("replicate_source_db"); ok {
		sourceDBInstanceID := v.(string)
		input := &rds.CreateDBInstanceReadReplicaInput{
			AutoMinorVersionUpgrade:    aws.Bool(d.Get(names.AttrAutoMinorVersionUpgrade).(bool)),
			CopyTagsToSnapshot:         aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBInstanceClass:            aws.String(d.Get("instance_class").(string)),
			DBInstanceIdentifier:       aws.String(identifier),
			DeletionProtection:         aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			PubliclyAccessible:         aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool)),
			SourceDBInstanceIdentifier: aws.String(sourceDBInstanceID),
			Tags:                       getTagsIn(ctx),
		}

		if _, ok := d.GetOk(names.AttrAllocatedStorage); ok {
			// RDS doesn't allow modifying the storage of a replica within the first 6h of creation.
			// allocated_storage is inherited from the primary so only the same value or no value is correct; a different value would fail the creation.
			// A different value is possible, granted: the value is higher than the current, there has been 6h between
			diags = sdkdiag.AppendWarningf(diags, `"allocated_storage" was ignored for DB Instance (%s) because a replica inherits the primary's allocated_storage and cannot be changed at creation.`, identifier)
		}

		if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
			input.AvailabilityZone = aws.String(v.(string))
		}

		if v, ok := d.GetOk("custom_iam_instance_profile"); ok {
			input.CustomIamInstanceProfile = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("dedicated_log_volume"); ok {
			input.DedicatedLogVolume = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			input.EnableIAMDatabaseAuthentication = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrIOPS); ok {
			input.Iops = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input.KmsKeyId = aws.String(v.(string))
			if arnParts := strings.Split(sourceDBInstanceID, ":"); len(arnParts) >= 4 {
				input.SourceRegion = aws.String(arnParts[3])
			}
		}

		if v, ok := d.GetOk("monitoring_interval"); ok {
			input.MonitoringInterval = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("monitoring_role_arn"); ok {
			input.MonitoringRoleArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("multi_az"); ok {
			input.MultiAZ = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("network_type"); ok {
			input.NetworkType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("option_group_name"); ok {
			input.OptionGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrParameterGroupName); ok {
			crossRegion := false
			if arn.IsARN(sourceDBInstanceID) {
				sourceARN, err := arn.Parse(sourceDBInstanceID)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "creating RDS DB Instance (read replica) (%s): %s", identifier, err)
				}
				crossRegion = sourceARN.Region != meta.(*conns.AWSClient).Region
			}
			if crossRegion {
				input.DBParameterGroupName = aws.String(v.(string))
			}
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

		if v, ok := d.GetOk(names.AttrPort); ok {
			input.Port = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("replica_mode"); ok {
			input.ReplicaMode = aws.String(v.(string))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("storage_throughput"); ok {
			input.StorageThroughput = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("upgrade_storage_config"); ok {
			input.UpgradeStorageConfig = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		}

		output, err := dbInstanceCreateReadReplica(ctx, conn, input)

		// Some engines (e.g. PostgreSQL) you cannot specify a custom parameter group for the read replica during creation.
		// See https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_ReadRepl.html#USER_ReadRepl.XRgn.Cnsdr.
		if input.DBParameterGroupName != nil && tfawserr.ErrMessageContains(err, "InvalidParameterCombination", "A parameter group can't be specified during Read Replica creation for the following DB engine") {
			input.DBParameterGroupName = nil

			output, err = dbInstanceCreateReadReplica(ctx, conn, input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating RDS DB Instance (read replica) (%s): %s", identifier, err)
		}

		resourceID = aws.StringValue(output.DBInstance.DbiResourceId)
		d.SetId(resourceID)

		if v, ok := d.GetOk(names.AttrAllowMajorVersionUpgrade); ok {
			// Having allowing_major_version_upgrade by itself should not trigger ModifyDBInstance
			// "InvalidParameterCombination: No modifications were requested".
			modifyDbInstanceInput.AllowMajorVersionUpgrade = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			if current, desired := aws.Int64Value(output.DBInstance.BackupRetentionPeriod), int64(v.(int)); current != desired {
				modifyDbInstanceInput.BackupRetentionPeriod = aws.Int64(desired)
				requiresModifyDbInstance = true
			}
		}

		if v, ok := d.GetOk("backup_window"); ok {
			if current, desired := aws.StringValue(output.DBInstance.PreferredBackupWindow), v.(string); current != desired {
				modifyDbInstanceInput.PreferredBackupWindow = aws.String(desired)
				requiresModifyDbInstance = true
			}
		}

		if v, ok := d.GetOk("ca_cert_identifier"); ok {
			if current, desired := aws.StringValue(output.DBInstance.CACertificateIdentifier), v.(string); current != desired {
				modifyDbInstanceInput.CACertificateIdentifier = aws.String(desired)
				requiresModifyDbInstance = true
			}
		}

		if v, ok := d.GetOk("maintenance_window"); ok {
			if current, desired := aws.StringValue(output.DBInstance.PreferredMaintenanceWindow), v.(string); current != desired {
				modifyDbInstanceInput.PreferredMaintenanceWindow = aws.String(desired)
				requiresModifyDbInstance = true
			}
		}
		if v, ok := d.GetOk("manage_master_user_password"); ok {
			modifyDbInstanceInput.ManageMasterUserPassword = aws.Bool(v.(bool))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
			modifyDbInstanceInput.MasterUserSecretKmsKeyId = aws.String(v.(string))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("max_allocated_storage"); ok {
			if current, desired := aws.Int64Value(output.DBInstance.MaxAllocatedStorage), int64(v.(int)); current != desired {
				modifyDbInstanceInput.MaxAllocatedStorage = aws.Int64(desired)
				requiresModifyDbInstance = true
			}
		}

		if v, ok := d.GetOk(names.AttrParameterGroupName); ok {
			if len(output.DBInstance.DBParameterGroups) > 0 {
				if current, desired := aws.StringValue(output.DBInstance.DBParameterGroups[0].DBParameterGroupName), v.(string); current != desired {
					modifyDbInstanceInput.DBParameterGroupName = aws.String(desired)
					requiresModifyDbInstance = true
					requiresRebootDbInstance = true
				}
			}
		}

		if v, ok := d.GetOk(names.AttrPassword); ok {
			modifyDbInstanceInput.MasterUserPassword = aws.String(v.(string))
			requiresModifyDbInstance = true
		}
	} else if v, ok := d.GetOk("s3_import"); ok {
		if _, ok := d.GetOk(names.AttrAllocatedStorage); !ok {
			diags = sdkdiag.AppendErrorf(diags, `"allocated_storage": required field is not set`)
		}
		if _, ok := d.GetOk(names.AttrEngine); !ok {
			diags = sdkdiag.AppendErrorf(diags, `"engine": required field is not set`)
		}
		if _, ok := d.GetOk(names.AttrUsername); !ok {
			diags = sdkdiag.AppendErrorf(diags, `"username": required field is not set`)
		}
		if diags.HasError() {
			return diags
		}

		tfMap := v.([]interface{})[0].(map[string]interface{})
		input := &rds.RestoreDBInstanceFromS3Input{
			AllocatedStorage:        aws.Int64(int64(d.Get(names.AttrAllocatedStorage).(int))),
			AutoMinorVersionUpgrade: aws.Bool(d.Get(names.AttrAutoMinorVersionUpgrade).(bool)),
			BackupRetentionPeriod:   aws.Int64(int64(d.Get("backup_retention_period").(int))),
			CopyTagsToSnapshot:      aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBInstanceClass:         aws.String(d.Get("instance_class").(string)),
			DBInstanceIdentifier:    aws.String(identifier),
			DBName:                  aws.String(d.Get("db_name").(string)),
			DeletionProtection:      aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			Engine:                  aws.String(d.Get(names.AttrEngine).(string)),
			EngineVersion:           aws.String(d.Get(names.AttrEngineVersion).(string)),
			MasterUsername:          aws.String(d.Get(names.AttrUsername).(string)),
			PubliclyAccessible:      aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool)),
			S3BucketName:            aws.String(tfMap[names.AttrBucketName].(string)),
			S3IngestionRoleArn:      aws.String(tfMap["ingestion_role"].(string)),
			S3Prefix:                aws.String(tfMap[names.AttrBucketPrefix].(string)),
			SourceEngine:            aws.String(tfMap["source_engine"].(string)),
			SourceEngineVersion:     aws.String(tfMap["source_engine_version"].(string)),
			StorageEncrypted:        aws.Bool(d.Get(names.AttrStorageEncrypted).(bool)),
			Tags:                    getTagsIn(ctx),
		}

		if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
			input.AvailabilityZone = aws.String(v.(string))
		}

		if v, ok := d.GetOk("backup_window"); ok {
			input.PreferredBackupWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("dedicated_log_volume"); ok {
			input.DedicatedLogVolume = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("engine_lifecycle_support"); ok {
			input.EngineLifecycleSupport = aws.String(v.(string))
		}

		if v, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			input.EnableIAMDatabaseAuthentication = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrIOPS); ok {
			input.Iops = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input.KmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("license_model"); ok {
			input.LicenseModel = aws.String(v.(string))
		}

		if v, ok := d.GetOk("maintenance_window"); ok {
			input.PreferredMaintenanceWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk("manage_master_user_password"); ok {
			input.ManageMasterUserPassword = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
			input.MasterUserSecretKmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("monitoring_interval"); ok {
			input.MonitoringInterval = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("monitoring_role_arn"); ok {
			input.MonitoringRoleArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("multi_az"); ok {
			input.MultiAZ = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("network_type"); ok {
			input.NetworkType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("option_group_name"); ok {
			input.OptionGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrParameterGroupName); ok {
			input.DBParameterGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrPassword); ok {
			input.MasterUserPassword = aws.String(v.(string))
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

		if v, ok := d.GetOk(names.AttrPort); ok {
			input.Port = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("storage_throughput"); ok {
			input.StorageThroughput = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		}

		outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.RestoreDBInstanceFromS3WithContext(ctx, input)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "ENHANCED_MONITORING") {
					return true, err
				}
				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "S3_SNAPSHOT_INGESTION") {
					return true, err
				}
				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "S3 bucket cannot be found") {
					return true, err
				}
				// InvalidParameterValue: Files from the specified Amazon S3 bucket cannot be downloaded. Make sure that you have created an AWS Identity and Access Management (IAM) role that lets Amazon RDS access Amazon S3 for you.
				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "Files from the specified Amazon S3 bucket cannot be downloaded") {
					return true, err
				}

				return false, err
			},
		)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating RDS DB Instance (restore from S3) (%s): %s", identifier, err)
		}

		output := outputRaw.(*rds.RestoreDBInstanceFromS3Output)

		resourceID = aws.StringValue(output.DBInstance.DbiResourceId)
		d.SetId(resourceID)
	} else if v, ok := d.GetOk("snapshot_identifier"); ok {
		input := &rds.RestoreDBInstanceFromDBSnapshotInput{
			AutoMinorVersionUpgrade: aws.Bool(d.Get(names.AttrAutoMinorVersionUpgrade).(bool)),
			CopyTagsToSnapshot:      aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBInstanceClass:         aws.String(d.Get("instance_class").(string)),
			DBInstanceIdentifier:    aws.String(identifier),
			DBSnapshotIdentifier:    aws.String(v.(string)),
			DeletionProtection:      aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			PubliclyAccessible:      aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool)),
			Tags:                    getTagsIn(ctx),
		}

		engine := strings.ToLower(d.Get(names.AttrEngine).(string))
		if v, ok := d.GetOk("db_name"); ok {
			// "Note: This parameter [DBName] doesn't apply to the MySQL, PostgreSQL, or MariaDB engines."
			// https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_RestoreDBInstanceFromDBSnapshot.html
			switch engine {
			case InstanceEngineMySQL, InstanceEnginePostgres, InstanceEngineMariaDB:
				// skip
			default:
				input.DBName = aws.String(v.(string))
			}
		}

		if v, ok := d.GetOk(names.AttrAllocatedStorage); ok {
			modifyDbInstanceInput.AllocatedStorage = aws.Int64(int64(v.(int)))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk(names.AttrAllowMajorVersionUpgrade); ok {
			modifyDbInstanceInput.AllowMajorVersionUpgrade = aws.Bool(v.(bool))
			// Having allowing_major_version_upgrade by itself should not trigger ModifyDBInstance
			// InvalidParameterCombination: No modifications were requested
		}

		if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
			input.AvailabilityZone = aws.String(v.(string))
		}

		if v, ok := d.GetOkExists("backup_retention_period"); ok {
			modifyDbInstanceInput.BackupRetentionPeriod = aws.Int64(int64(v.(int)))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("backup_target"); ok {
			input.BackupTarget = aws.String(v.(string))
		}

		if v, ok := d.GetOk("backup_window"); ok {
			modifyDbInstanceInput.PreferredBackupWindow = aws.String(v.(string))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("custom_iam_instance_profile"); ok {
			input.CustomIamInstanceProfile = aws.String(v.(string))
		}

		if v, ok := d.GetOk("customer_owned_ip_enabled"); ok {
			input.EnableCustomerOwnedIp = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("dedicated_log_volume"); ok {
			input.DedicatedLogVolume = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrDomain); ok {
			input.Domain = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_auth_secret_arn"); ok {
			input.DomainAuthSecretArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_dns_ips"); ok && len(v.([]interface{})) > 0 {
			input.DomainDnsIps = flex.ExpandStringList(v.([]interface{}))
		}

		if v, ok := d.GetOk("domain_fqdn"); ok {
			input.DomainFqdn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_iam_role_name"); ok {
			input.DomainIAMRoleName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_ou"); ok {
			input.DomainOu = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringSet(v.(*schema.Set))
		}

		if engine != "" {
			input.Engine = aws.String(engine)
		}

		if v, ok := d.GetOk("engine_lifecycle_support"); ok {
			input.EngineLifecycleSupport = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrEngineVersion); ok {
			modifyDbInstanceInput.EngineVersion = aws.String(v.(string))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			input.EnableIAMDatabaseAuthentication = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrIOPS); ok {
			modifyDbInstanceInput.Iops = aws.Int64(int64(v.(int)))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("license_model"); ok {
			input.LicenseModel = aws.String(v.(string))
		}

		if v, ok := d.GetOk("maintenance_window"); ok {
			modifyDbInstanceInput.PreferredMaintenanceWindow = aws.String(v.(string))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("manage_master_user_password"); ok {
			modifyDbInstanceInput.ManageMasterUserPassword = aws.Bool(v.(bool))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
			modifyDbInstanceInput.MasterUserSecretKmsKeyId = aws.String(v.(string))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("max_allocated_storage"); ok {
			modifyDbInstanceInput.MaxAllocatedStorage = aws.Int64(int64(v.(int)))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("monitoring_interval"); ok {
			modifyDbInstanceInput.MonitoringInterval = aws.Int64(int64(v.(int)))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("monitoring_role_arn"); ok {
			modifyDbInstanceInput.MonitoringRoleArn = aws.String(v.(string))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("multi_az"); ok {
			// When using SQL Server engine with MultiAZ enabled, its not
			// possible to immediately enable mirroring since
			// BackupRetentionPeriod is not available as a parameter to
			// RestoreDBInstanceFromDBSnapshot and you receive an error. e.g.
			// InvalidParameterValue: Mirroring cannot be applied to instances with backup retention set to zero.
			// If we know the engine, prevent the error upfront.
			if strings.HasPrefix(engine, "sqlserver") {
				modifyDbInstanceInput.MultiAZ = aws.Bool(v.(bool))
				requiresModifyDbInstance = true
			} else {
				input.MultiAZ = aws.Bool(v.(bool))
			}
		}

		if v, ok := d.GetOk("network_type"); ok {
			input.NetworkType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("option_group_name"); ok {
			input.OptionGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrParameterGroupName); ok {
			input.DBParameterGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrPassword); ok {
			modifyDbInstanceInput.MasterUserPassword = aws.String(v.(string))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("performance_insights_enabled"); ok {
			modifyDbInstanceInput.EnablePerformanceInsights = aws.Bool(v.(bool))
			requiresModifyDbInstance = true

			if v, ok := d.GetOk("performance_insights_kms_key_id"); ok {
				modifyDbInstanceInput.PerformanceInsightsKMSKeyId = aws.String(v.(string))
			}

			if v, ok := d.GetOk("performance_insights_retention_period"); ok {
				modifyDbInstanceInput.PerformanceInsightsRetentionPeriod = aws.Int64(int64(v.(int)))
			}
		}

		if v, ok := d.GetOk(names.AttrPort); ok {
			input.Port = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("storage_throughput"); ok {
			modifyDbInstanceInput.StorageThroughput = aws.Int64(int64(v.(int)))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			modifyDbInstanceInput.StorageType = aws.String(v.(string))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("tde_credential_arn"); ok {
			input.TdeCredentialArn = aws.String(v.(string))
		}

		if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringSet(v)
		}

		outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.RestoreDBInstanceFromDBSnapshotWithContext(ctx, input)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, errCodeValidationError, "RDS couldn't fetch the role from instance profile") {
					return true, err
				}

				return false, err
			},
		)

		var output *rds.RestoreDBInstanceFromDBSnapshotOutput

		if err == nil {
			output = outputRaw.(*rds.RestoreDBInstanceFromDBSnapshotOutput)
		}

		// When using SQL Server engine with MultiAZ enabled, its not
		// possible to immediately enable mirroring since
		// BackupRetentionPeriod is not available as a parameter to
		// RestoreDBInstanceFromDBSnapshot and you receive an error. e.g.
		// InvalidParameterValue: Mirroring cannot be applied to instances with backup retention set to zero.
		// Since engine is not a required argument when using snapshot_identifier
		// and the RDS API determines this condition, we catch the error
		// and remove the invalid configuration for it to be fixed afterwards.
		if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "Mirroring cannot be applied to instances with backup retention set to zero") {
			input.MultiAZ = aws.Bool(false)
			modifyDbInstanceInput.MultiAZ = aws.Bool(true)
			requiresModifyDbInstance = true
			output, err = conn.RestoreDBInstanceFromDBSnapshotWithContext(ctx, input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating RDS DB Instance (restore from snapshot) (%s): %s", identifier, err)
		}

		resourceID = aws.StringValue(output.DBInstance.DbiResourceId)
		d.SetId(resourceID)
	} else if v, ok := d.GetOk("restore_to_point_in_time"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		input := &rds.RestoreDBInstanceToPointInTimeInput{
			AutoMinorVersionUpgrade:    aws.Bool(d.Get(names.AttrAutoMinorVersionUpgrade).(bool)),
			CopyTagsToSnapshot:         aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBInstanceClass:            aws.String(d.Get("instance_class").(string)),
			DeletionProtection:         aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			PubliclyAccessible:         aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool)),
			Tags:                       getTagsIn(ctx),
			TargetDBInstanceIdentifier: aws.String(identifier),
		}

		if v, ok := tfMap["restore_time"].(string); ok && v != "" {
			v, _ := time.Parse(time.RFC3339, v)

			input.RestoreTime = aws.Time(v)
		}

		if v, ok := tfMap["source_db_instance_automated_backups_arn"].(string); ok && v != "" {
			input.SourceDBInstanceAutomatedBackupsArn = aws.String(v)
		}

		if v, ok := tfMap["source_db_instance_identifier"].(string); ok && v != "" {
			input.SourceDBInstanceIdentifier = aws.String(v)
		}

		if v, ok := tfMap["source_dbi_resource_id"].(string); ok && v != "" {
			input.SourceDbiResourceId = aws.String(v)
		}

		if v, ok := tfMap["use_latest_restorable_time"].(bool); ok && v {
			input.UseLatestRestorableTime = aws.Bool(v)
		}

		if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
			input.AvailabilityZone = aws.String(v.(string))
		}

		if v, ok := d.GetOk("backup_target"); ok {
			input.BackupTarget = aws.String(v.(string))
		}

		if v, ok := d.GetOk("custom_iam_instance_profile"); ok {
			input.CustomIamInstanceProfile = aws.String(v.(string))
		}

		if v, ok := d.GetOk("customer_owned_ip_enabled"); ok {
			input.EnableCustomerOwnedIp = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("db_name"); ok {
			input.DBName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("dedicated_log_volume"); ok {
			input.DedicatedLogVolume = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrDomain); ok {
			input.Domain = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_iam_role_name"); ok {
			input.DomainIAMRoleName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_fqdn"); ok {
			input.DomainFqdn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_ou"); ok {
			input.DomainOu = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_auth_secret_arn"); ok {
			input.DomainAuthSecretArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_dns_ips"); ok && len(v.([]interface{})) > 0 {
			input.DomainDnsIps = flex.ExpandStringList(v.([]interface{}))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk(names.AttrEngine); ok {
			input.Engine = aws.String(v.(string))
		}

		if v, ok := d.GetOk("engine_lifecycle_support"); ok {
			input.EngineLifecycleSupport = aws.String(v.(string))
		}

		if v, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			input.EnableIAMDatabaseAuthentication = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrIOPS); ok {
			input.Iops = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("license_model"); ok {
			input.LicenseModel = aws.String(v.(string))
		}

		if v, ok := d.GetOk("max_allocated_storage"); ok {
			input.MaxAllocatedStorage = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("manage_master_user_password"); ok {
			modifyDbInstanceInput.ManageMasterUserPassword = aws.Bool(v.(bool))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
			modifyDbInstanceInput.MasterUserSecretKmsKeyId = aws.String(v.(string))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("monitoring_interval"); ok {
			modifyDbInstanceInput.MonitoringInterval = aws.Int64(int64(v.(int)))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("monitoring_role_arn"); ok {
			modifyDbInstanceInput.MonitoringRoleArn = aws.String(v.(string))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk("multi_az"); ok {
			input.MultiAZ = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("option_group_name"); ok {
			input.OptionGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrParameterGroupName); ok {
			input.DBParameterGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrPassword); ok {
			modifyDbInstanceInput.MasterUserPassword = aws.String(v.(string))
			requiresModifyDbInstance = true
		}

		if v, ok := d.GetOk(names.AttrPort); ok {
			input.Port = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("tde_credential_arn"); ok {
			input.TdeCredentialArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		}

		outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.RestoreDBInstanceToPointInTimeWithContext(ctx, input)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, errCodeValidationError, "RDS couldn't fetch the role from instance profile") {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating RDS DB Instance (restore to point-in-time) (%s): %s", identifier, err)
		}

		output := outputRaw.(*rds.RestoreDBInstanceToPointInTimeOutput)

		resourceID = aws.StringValue(output.DBInstance.DbiResourceId)
		d.SetId(resourceID)
	} else {
		if _, ok := d.GetOk(names.AttrAllocatedStorage); !ok {
			diags = sdkdiag.AppendErrorf(diags, `"allocated_storage": required field is not set`)
		}
		if _, ok := d.GetOk(names.AttrEngine); !ok {
			diags = sdkdiag.AppendErrorf(diags, `"engine": required field is not set`)
		}
		if _, ok := d.GetOk(names.AttrUsername); !ok {
			diags = sdkdiag.AppendErrorf(diags, `"username": required field is not set`)
		}
		if diags.HasError() {
			return diags
		}

		input := &rds.CreateDBInstanceInput{
			AllocatedStorage:        aws.Int64(int64(d.Get(names.AttrAllocatedStorage).(int))),
			AutoMinorVersionUpgrade: aws.Bool(d.Get(names.AttrAutoMinorVersionUpgrade).(bool)),
			BackupRetentionPeriod:   aws.Int64(int64(d.Get("backup_retention_period").(int))),
			CopyTagsToSnapshot:      aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBInstanceClass:         aws.String(d.Get("instance_class").(string)),
			DBInstanceIdentifier:    aws.String(identifier),
			DBName:                  aws.String(d.Get("db_name").(string)),
			DeletionProtection:      aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			Engine:                  aws.String(d.Get(names.AttrEngine).(string)),
			EngineVersion:           aws.String(d.Get(names.AttrEngineVersion).(string)),
			MasterUsername:          aws.String(d.Get(names.AttrUsername).(string)),
			PubliclyAccessible:      aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool)),
			StorageEncrypted:        aws.Bool(d.Get(names.AttrStorageEncrypted).(bool)),
			Tags:                    getTagsIn(ctx),
		}

		if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
			input.AvailabilityZone = aws.String(v.(string))
		}

		if v, ok := d.GetOk("backup_target"); ok {
			input.BackupTarget = aws.String(v.(string))
		}

		if v, ok := d.GetOk("backup_window"); ok {
			input.PreferredBackupWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk("character_set_name"); ok {
			input.CharacterSetName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("custom_iam_instance_profile"); ok {
			input.CustomIamInstanceProfile = aws.String(v.(string))
		}

		if v, ok := d.GetOk("customer_owned_ip_enabled"); ok {
			input.EnableCustomerOwnedIp = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("dedicated_log_volume"); ok {
			input.DedicatedLogVolume = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrDomain); ok {
			input.Domain = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_auth_secret_arn"); ok {
			input.DomainAuthSecretArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_dns_ips"); ok && len(v.([]interface{})) > 0 {
			input.DomainDnsIps = flex.ExpandStringList(v.([]interface{}))
		}

		if v, ok := d.GetOk("domain_fqdn"); ok {
			input.DomainFqdn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_iam_role_name"); ok {
			input.DomainIAMRoleName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("domain_ou"); ok {
			input.DomainOu = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("engine_lifecycle_support"); ok {
			input.EngineLifecycleSupport = aws.String(v.(string))
		}

		if v, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			input.EnableIAMDatabaseAuthentication = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrIOPS); ok {
			input.Iops = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input.KmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("license_model"); ok {
			input.LicenseModel = aws.String(v.(string))
		}

		if v, ok := d.GetOk("maintenance_window"); ok {
			input.PreferredMaintenanceWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk("manage_master_user_password"); ok {
			input.ManageMasterUserPassword = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
			input.MasterUserSecretKmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("max_allocated_storage"); ok {
			input.MaxAllocatedStorage = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("monitoring_interval"); ok {
			input.MonitoringInterval = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("monitoring_role_arn"); ok {
			input.MonitoringRoleArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("multi_az"); ok {
			input.MultiAZ = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("nchar_character_set_name"); ok {
			input.NcharCharacterSetName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("network_type"); ok {
			input.NetworkType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("option_group_name"); ok {
			input.OptionGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrPassword); ok {
			input.MasterUserPassword = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrParameterGroupName); ok {
			input.DBParameterGroupName = aws.String(v.(string))
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

		if v, ok := d.GetOk(names.AttrPort); ok {
			input.Port = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("storage_throughput"); ok {
			input.StorageThroughput = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("timezone"); ok {
			input.Timezone = aws.String(v.(string))
		}

		if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringSet(v)
		}

		outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.CreateDBInstanceWithContext(ctx, input)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "ENHANCED_MONITORING") {
					return true, err
				}
				if tfawserr.ErrMessageContains(err, errCodeValidationError, "RDS couldn't fetch the role from instance profile") {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating RDS DB Instance (%s): %s", identifier, err)
		}

		output := outputRaw.(*rds.CreateDBInstanceOutput)

		resourceID = aws.StringValue(output.DBInstance.DbiResourceId)
		d.SetId(resourceID)

		// This is added here to avoid unnecessary modification when ca_cert_identifier is the default one
		if v, ok := d.GetOk("ca_cert_identifier"); ok && v.(string) != aws.StringValue(output.DBInstance.CACertificateIdentifier) {
			modifyDbInstanceInput.CACertificateIdentifier = aws.String(v.(string))
			requiresModifyDbInstance = true
		}
	}

	var instance *rds.DBInstance
	var err error
	if instance, err = waitDBInstanceAvailableSDKv1(ctx, conn, identifier, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Instance (%s) create: %s", identifier, err)
	}

	if resourceID == "" {
		resourceID = aws.StringValue(instance.DbiResourceId)
	}

	if d.Id() == "" {
		d.SetId(resourceID)
	}

	if requiresModifyDbInstance {
		modifyDbInstanceInput.DBInstanceIdentifier = aws.String(identifier)

		_, err := conn.ModifyDBInstanceWithContext(ctx, modifyDbInstanceInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): %s", identifier, err)
		}

		if _, err := waitDBInstanceAvailableSDKv1(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Instance (%s) update: %s", identifier, err)
		}
	}

	if requiresRebootDbInstance {
		_, err := conn.RebootDBInstanceWithContext(ctx, &rds.RebootDBInstanceInput{
			DBInstanceIdentifier: aws.String(identifier),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "rebooting RDS DB Instance (%s): %s", identifier, err)
		}

		if _, err := waitDBInstanceAvailableSDKv1(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Instance (%s) update: %s", identifier, err)
		}
	}

	return append(diags, resourceInstanceRead(ctx, d, meta)...)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	var (
		v   *rds.DBInstance
		err error
	)

	if d.IsNewResource() {
		v, err = findDBInstanceByIDSDKv1(ctx, conn, d.Id())
	} else {
		v, err = findDBInstanceByIDSDKv1(ctx, conn, d.Id())
		if tfresource.NotFound(err) { // nosemgrep:ci.semgrep.errors.notfound-without-err-checks
			// Retry with `identifier`
			v, err = findDBInstanceByIDSDKv1(ctx, conn, d.Get(names.AttrIdentifier).(string))
			if tfresource.NotFound(err) { // nosemgrep:ci.semgrep.errors.notfound-without-err-checks
				log.Printf("[WARN] RDS DB Instance (%s) not found, removing from state", d.Get(names.AttrIdentifier).(string))
				d.SetId("")
				return diags
			}
		}
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Instance (%s): %s", d.Get(names.AttrIdentifier).(string), err)
	}

	d.SetId(aws.StringValue(v.DbiResourceId))

	d.Set(names.AttrAllocatedStorage, v.AllocatedStorage)
	d.Set(names.AttrARN, v.DBInstanceArn)
	d.Set(names.AttrAutoMinorVersionUpgrade, v.AutoMinorVersionUpgrade)
	d.Set(names.AttrAvailabilityZone, v.AvailabilityZone)
	d.Set("backup_retention_period", v.BackupRetentionPeriod)
	d.Set("backup_target", v.BackupTarget)
	d.Set("backup_window", v.PreferredBackupWindow)
	d.Set("ca_cert_identifier", v.CACertificateIdentifier)
	d.Set("character_set_name", v.CharacterSetName)
	d.Set("copy_tags_to_snapshot", v.CopyTagsToSnapshot)
	d.Set("custom_iam_instance_profile", v.CustomIamInstanceProfile)
	d.Set("customer_owned_ip_enabled", v.CustomerOwnedIpEnabled)
	d.Set("db_name", v.DBName)
	if v.DBSubnetGroup != nil {
		d.Set("db_subnet_group_name", v.DBSubnetGroup.DBSubnetGroupName)
	}
	d.Set("dedicated_log_volume", v.DedicatedLogVolume)
	d.Set(names.AttrDeletionProtection, v.DeletionProtection)
	if len(v.DomainMemberships) > 0 && v.DomainMemberships[0] != nil {
		v := v.DomainMemberships[0]
		d.Set(names.AttrDomain, v.Domain)
		d.Set("domain_auth_secret_arn", v.AuthSecretArn)
		d.Set("domain_dns_ips", aws.StringValueSlice(v.DnsIps))
		d.Set("domain_fqdn", v.FQDN)
		d.Set("domain_iam_role_name", v.IAMRoleName)
		d.Set("domain_ou", v.OU)
	} else {
		d.Set(names.AttrDomain, nil)
		d.Set("domain_auth_secret_arn", nil)
		d.Set("domain_dns_ips", nil)
		d.Set("domain_fqdn", nil)
		d.Set("domain_iam_role_name", nil)
		d.Set("domain_ou", nil)
	}
	d.Set("enabled_cloudwatch_logs_exports", aws.StringValueSlice(v.EnabledCloudwatchLogsExports))
	d.Set(names.AttrEngine, v.Engine)
	d.Set("engine_lifecycle_support", v.EngineLifecycleSupport)
	d.Set("iam_database_authentication_enabled", v.IAMDatabaseAuthenticationEnabled)
	d.Set(names.AttrIdentifier, v.DBInstanceIdentifier)
	d.Set("identifier_prefix", create.NamePrefixFromName(aws.StringValue(v.DBInstanceIdentifier)))
	d.Set("instance_class", v.DBInstanceClass)
	d.Set(names.AttrIOPS, v.Iops)
	d.Set(names.AttrKMSKeyID, v.KmsKeyId)
	if v.LatestRestorableTime != nil {
		d.Set("latest_restorable_time", aws.TimeValue(v.LatestRestorableTime).Format(time.RFC3339))
	} else {
		d.Set("latest_restorable_time", nil)
	}
	d.Set("license_model", v.LicenseModel)
	d.Set("maintenance_window", v.PreferredMaintenanceWindow)
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
	if v.MasterUserSecret != nil {
		if err := d.Set("master_user_secret", []interface{}{flattenManagedMasterUserSecret(v.MasterUserSecret)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting master_user_secret: %s", err)
		}
	} else {
		d.Set("master_user_secret", nil)
	}

	d.Set("max_allocated_storage", v.MaxAllocatedStorage)
	d.Set("monitoring_interval", v.MonitoringInterval)
	d.Set("monitoring_role_arn", v.MonitoringRoleArn)
	d.Set("multi_az", v.MultiAZ)
	d.Set("nchar_character_set_name", v.NcharCharacterSetName)
	d.Set("network_type", v.NetworkType)
	if len(v.OptionGroupMemberships) > 0 && v.OptionGroupMemberships[0] != nil {
		d.Set("option_group_name", v.OptionGroupMemberships[0].OptionGroupName)
	}
	if len(v.DBParameterGroups) > 0 && v.DBParameterGroups[0] != nil {
		d.Set(names.AttrParameterGroupName, v.DBParameterGroups[0].DBParameterGroupName)
	}
	d.Set("performance_insights_enabled", v.PerformanceInsightsEnabled)
	d.Set("performance_insights_kms_key_id", v.PerformanceInsightsKMSKeyId)
	d.Set("performance_insights_retention_period", v.PerformanceInsightsRetentionPeriod)
	d.Set(names.AttrPort, v.DbInstancePort)
	d.Set(names.AttrPubliclyAccessible, v.PubliclyAccessible)
	d.Set("replica_mode", v.ReplicaMode)
	d.Set("replicas", aws.StringValueSlice(v.ReadReplicaDBInstanceIdentifiers))
	d.Set("replicate_source_db", v.ReadReplicaSourceDBInstanceIdentifier)
	d.Set(names.AttrResourceID, v.DbiResourceId)
	d.Set(names.AttrStatus, v.DBInstanceStatus)
	d.Set(names.AttrStorageEncrypted, v.StorageEncrypted)
	d.Set("storage_throughput", v.StorageThroughput)
	d.Set(names.AttrStorageType, v.StorageType)
	d.Set("timezone", v.Timezone)
	d.Set(names.AttrUsername, v.MasterUsername)
	var vpcSecurityGroupIDs []string
	for _, v := range v.VpcSecurityGroups {
		vpcSecurityGroupIDs = append(vpcSecurityGroupIDs, aws.StringValue(v.VpcSecurityGroupId))
	}
	d.Set(names.AttrVPCSecurityGroupIDs, vpcSecurityGroupIDs)

	if v.Endpoint != nil {
		d.Set(names.AttrAddress, v.Endpoint.Address)
		if v.Endpoint.Address != nil && v.Endpoint.Port != nil {
			d.Set(names.AttrEndpoint, fmt.Sprintf("%s:%d", aws.StringValue(v.Endpoint.Address), aws.Int64Value(v.Endpoint.Port)))
		}
		d.Set(names.AttrHostedZoneID, v.Endpoint.HostedZoneId)
		d.Set(names.AttrPort, v.Endpoint.Port)
	}

	if v.ListenerEndpoint != nil {
		if err := d.Set("listener_endpoint", []interface{}{flattenEndpoint(v.ListenerEndpoint)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting listener_endpoint: %s", err)
		}
	} else {
		d.Set("listener_endpoint", nil)
	}

	dbSetResourceDataEngineVersionFromInstance(d, v)

	setTagsOut(ctx, v.TagList)

	return diags
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)
	deadline := tfresource.NewDeadline(d.Timeout(schema.TimeoutUpdate))

	// Separate request to promote a database.
	if d.HasChange("replicate_source_db") {
		if d.Get("replicate_source_db").(string) == "" {
			input := &rds_sdkv2.PromoteReadReplicaInput{
				BackupRetentionPeriod: aws.Int32(int32(d.Get("backup_retention_period").(int))),
				DBInstanceIdentifier:  aws.String(d.Get(names.AttrIdentifier).(string)),
			}

			if attr, ok := d.GetOk("backup_window"); ok {
				input.PreferredBackupWindow = aws.String(attr.(string))
			}

			_, err := conn.PromoteReadReplica(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "promoting RDS DB Instance (%s): %s", d.Get(names.AttrIdentifier).(string), err)
			}

			if _, err := waitDBInstanceAvailableSDKv2(ctx, conn, d.Id(), deadline.Remaining()); err != nil {
				return sdkdiag.AppendErrorf(diags, "promoting RDS DB Instance (%s): waiting for completion: %s", d.Get(names.AttrIdentifier).(string), err)
			}
		} else {
			return sdkdiag.AppendErrorf(diags, "cannot elect new source database for replication")
		}
	}

	// Having allowing_major_version_upgrade by itself should not trigger ModifyDBInstance
	// as it results in "InvalidParameterCombination: No modifications were requested".
	if d.HasChangesExcept(
		names.AttrAllowMajorVersionUpgrade,
		"blue_green_update",
		"delete_automated_backups",
		names.AttrFinalSnapshotIdentifier,
		"replicate_source_db",
		"skip_final_snapshot",
		names.AttrTags, names.AttrTagsAll,
	) {
		if d.Get("blue_green_update.0.enabled").(bool) && d.HasChangesExcept(
			names.AttrAllowMajorVersionUpgrade,
			"blue_green_update",
			"delete_automated_backups",
			names.AttrFinalSnapshotIdentifier,
			"replicate_source_db",
			"skip_final_snapshot",
			names.AttrTags, names.AttrTagsAll,
			names.AttrDeletionProtection,
			names.AttrPassword,
		) {
			orchestrator := newBlueGreenOrchestrator(conn)
			defer orchestrator.CleanUp(ctx)

			handler := newInstanceHandler(conn)

			err := handler.precondition(ctx, d)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): %s", d.Get(names.AttrIdentifier).(string), err)
			}

			createIn := handler.createBlueGreenInput(d)

			log.Printf("[DEBUG] Updating RDS DB Instance (%s): Creating Blue/Green Deployment", d.Get(names.AttrIdentifier).(string))

			dep, err := orchestrator.CreateDeployment(ctx, createIn)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): %s", d.Get(names.AttrIdentifier).(string), err)
			}

			deploymentIdentifier := dep.BlueGreenDeploymentIdentifier
			defer func() {
				log.Printf("[DEBUG] Updating RDS DB Instance (%s): Deleting Blue/Green Deployment", d.Get(names.AttrIdentifier).(string))

				if dep == nil {
					log.Printf("[DEBUG] Updating RDS DB Instance (%s): Deleting Blue/Green Deployment: deployment disappeared", d.Get(names.AttrIdentifier).(string))
					return
				}

				// Ensure that the Blue/Green Deployment is always cleaned up
				input := &rds_sdkv2.DeleteBlueGreenDeploymentInput{
					BlueGreenDeploymentIdentifier: deploymentIdentifier,
				}
				if aws.StringValue(dep.Status) != "SWITCHOVER_COMPLETED" {
					input.DeleteTarget = aws.Bool(true)
				}
				_, err = conn.DeleteBlueGreenDeployment(ctx, input)
				if err != nil {
					diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): deleting Blue/Green Deployment: %s", d.Get(names.AttrIdentifier).(string), err)
					return
				}

				orchestrator.AddCleanupWaiter(func(ctx context.Context, conn *rds_sdkv2.Client, optFns ...tfresource.OptionsFunc) {
					_, err = waitBlueGreenDeploymentDeleted(ctx, conn, aws.StringValue(deploymentIdentifier), deadline.Remaining(), optFns...)
					if err != nil {
						diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): deleting Blue/Green Deployment: waiting for completion: %s", d.Get(names.AttrIdentifier).(string), err)
					}
				})
			}()

			dep, err = orchestrator.waitForDeploymentAvailable(ctx, aws.StringValue(dep.BlueGreenDeploymentIdentifier), deadline.Remaining())
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): %s", d.Get(names.AttrIdentifier).(string), err)
			}

			targetARN, err := parseDBInstanceARN(aws.StringValue(dep.Target))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): creating Blue/Green Deployment: waiting for Green environment: %s", d.Get(names.AttrIdentifier).(string), err)
			}
			_, err = waitDBInstanceAvailableSDKv2(ctx, conn, targetARN.Identifier, deadline.Remaining())
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): creating Blue/Green Deployment: waiting for Green environment: %s", d.Get(names.AttrIdentifier).(string), err)
			}

			err = handler.modifyTarget(ctx, targetARN.Identifier, d, deadline.Remaining(), fmt.Sprintf("Updating RDS DB Instance (%s)", d.Get(names.AttrIdentifier).(string)))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): %s", d.Get(names.AttrIdentifier).(string), err)
			}

			log.Printf("[DEBUG] Updating RDS DB Instance (%s): Switching over Blue/Green Deployment", d.Get(names.AttrIdentifier).(string))

			dep, err = orchestrator.Switchover(ctx, aws.StringValue(dep.BlueGreenDeploymentIdentifier), deadline.Remaining())
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): %s", d.Get(names.AttrIdentifier).(string), err)
			}

			target, err := findDBInstanceByIDSDKv2(ctx, conn, d.Get(names.AttrIdentifier).(string))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): %s", d.Get(names.AttrIdentifier).(string), err)
			}

			// id changes here
			d.SetId(aws.StringValue(target.DbiResourceId))
			d.Set(names.AttrResourceID, target.DbiResourceId)

			log.Printf("[DEBUG] Updating RDS DB Instance (%s): Deleting Blue/Green Deployment source", d.Get(names.AttrIdentifier).(string))

			sourceARN, err := parseDBInstanceARN(aws.StringValue(dep.Source))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): deleting Blue/Green Deployment source: %s", d.Get(names.AttrIdentifier).(string), err)
			}
			if d.Get(names.AttrDeletionProtection).(bool) {
				input := &rds_sdkv2.ModifyDBInstanceInput{
					ApplyImmediately:     aws.Bool(true),
					DBInstanceIdentifier: aws.String(sourceARN.Identifier),
					DeletionProtection:   aws.Bool(false),
				}
				err := dbInstanceModify(ctx, conn, d.Id(), input, deadline.Remaining())
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): deleting Blue/Green Deployment source: disabling deletion protection: %s", d.Get(names.AttrIdentifier).(string), err)
				}
			}
			deleteInput := &rds_sdkv2.DeleteDBInstanceInput{
				DBInstanceIdentifier: aws.String(sourceARN.Identifier),
				SkipFinalSnapshot:    aws.Bool(true),
			}
			_, err = tfresource.RetryWhen(ctx, 5*time.Minute,
				func() (any, error) {
					return conn.DeleteDBInstance(ctx, deleteInput)
				},
				func(err error) (bool, error) {
					// Retry for IAM eventual consistency.
					if tfawserr_sdkv2.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
						return true, err
					}

					if tfawserr_sdkv2.ErrMessageContains(err, errCodeInvalidParameterCombination, "disable deletion pro") {
						return true, err
					}

					return false, err
				},
			)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): deleting Blue/Green Deployment source: %s", d.Get(names.AttrIdentifier).(string), err)
			}

			orchestrator.AddCleanupWaiter(func(ctx context.Context, conn *rds_sdkv2.Client, optFns ...tfresource.OptionsFunc) {
				_, err = waitDBInstanceDeleted(ctx, meta.(*conns.AWSClient).RDSConn(ctx), sourceARN.Identifier, deadline.Remaining(), optFns...)
				if err != nil {
					diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): deleting Blue/Green Deployment source: waiting for completion: %s", d.Get(names.AttrIdentifier).(string), err)
				}
			})

			if diags.HasError() {
				return diags
			}
		} else {
			oldID := d.Get(names.AttrIdentifier).(string)
			if d.HasChange(names.AttrIdentifier) {
				o, _ := d.GetChange(names.AttrIdentifier)
				oldID = o.(string)
			}

			applyImmediately := d.Get(names.AttrApplyImmediately).(bool)
			input := &rds_sdkv2.ModifyDBInstanceInput{
				ApplyImmediately:     aws.Bool(applyImmediately),
				DBInstanceIdentifier: aws.String(oldID),
			}

			if !applyImmediately {
				log.Println("[INFO] Only settings updating, instance changes will be applied in next maintenance window")
			}

			dbInstancePopulateModify(input, d)

			if d.HasChange(names.AttrEngineVersion) {
				input.EngineVersion = aws.String(d.Get(names.AttrEngineVersion).(string))
				input.AllowMajorVersionUpgrade = aws.Bool(d.Get(names.AttrAllowMajorVersionUpgrade).(bool))
				// if we were to make life easier for practitioners, we could loop through
				// replicas at this point to update them first, prior to dbInstanceModify()
				// for the source
			}

			if d.HasChange(names.AttrParameterGroupName) {
				input.DBParameterGroupName = aws.String(d.Get(names.AttrParameterGroupName).(string))
			}

			err := dbInstanceModify(ctx, conn, d.Id(), input, deadline.Remaining())
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): %s", d.Get(names.AttrIdentifier).(string), err)
			}
		}
	}

	return append(diags, resourceInstanceRead(ctx, d, meta)...)
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier:   aws.String(d.Get(names.AttrIdentifier).(string)),
		DeleteAutomatedBackups: aws.Bool(d.Get("delete_automated_backups").(bool)),
	}

	if d.Get("skip_final_snapshot").(bool) {
		input.SkipFinalSnapshot = aws.Bool(true)
	} else {
		input.SkipFinalSnapshot = aws.Bool(false)

		if v, ok := d.GetOk(names.AttrFinalSnapshotIdentifier); ok {
			input.FinalDBSnapshotIdentifier = aws.String(v.(string))
		} else {
			return sdkdiag.AppendErrorf(diags, "final_snapshot_identifier is required when skip_final_snapshot is false")
		}
	}

	log.Printf("[DEBUG] Deleting RDS DB Instance: %s", d.Get(names.AttrIdentifier).(string))
	_, err := conn.DeleteDBInstanceWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, errCodeInvalidParameterCombination, "disable deletion pro") {
		if v, ok := d.GetOk(names.AttrDeletionProtection); (!ok || !v.(bool)) && d.Get(names.AttrApplyImmediately).(bool) {
			_, ierr := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutUpdate),
				func() (interface{}, error) {
					return conn.ModifyDBInstanceWithContext(ctx, &rds.ModifyDBInstanceInput{
						ApplyImmediately:     aws.Bool(true),
						DBInstanceIdentifier: aws.String(d.Get(names.AttrIdentifier).(string)),
						DeletionProtection:   aws.Bool(false),
					})
				},
				func(err error) (bool, error) {
					// Retry for IAM eventual consistency.
					if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or") {
						return true, err
					}

					// "InvalidDBInstanceState: RDS is configuring Enhanced Monitoring or Performance Insights for this DB instance. Try your request later."
					if tfawserr.ErrMessageContains(err, rds.ErrCodeInvalidDBInstanceStateFault, "your request later") {
						return true, err
					}

					return false, err
				},
			)

			if ierr != nil {
				return sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): %s", d.Get(names.AttrIdentifier).(string), err)
			}

			if _, ierr := waitDBInstanceAvailableSDKv1(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); ierr != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Instance (%s) update: %s", d.Get(names.AttrIdentifier).(string), ierr)
			}

			_, err = conn.DeleteDBInstanceWithContext(ctx, input)
		}
	}

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceNotFoundFault) {
		return diags
	}

	if err != nil && !tfawserr.ErrMessageContains(err, rds.ErrCodeInvalidDBInstanceStateFault, "is already being deleted") {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Instance (%s): %s", d.Get(names.AttrIdentifier).(string), err)
	}

	if _, err := waitDBInstanceDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Instance (%s) delete: %s", d.Get(names.AttrIdentifier).(string), err)
	}

	return diags
}

func resourceInstanceImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Neither skip_final_snapshot nor final_snapshot_identifier can be fetched
	// from any API call, so we need to default skip_final_snapshot to true so
	// that final_snapshot_identifier is not required.
	d.Set("skip_final_snapshot", true)
	d.Set("delete_automated_backups", true)
	return []*schema.ResourceData{d}, nil
}

func dbInstanceCreateReadReplica(ctx context.Context, conn *rds.RDS, input *rds.CreateDBInstanceReadReplicaInput) (*rds.CreateDBInstanceReadReplicaOutput, error) {
	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateDBInstanceReadReplicaWithContext(ctx, input)
		},
		errCodeInvalidParameterValue, "ENHANCED_MONITORING")

	if err != nil {
		return nil, err
	}

	return outputRaw.(*rds.CreateDBInstanceReadReplicaOutput), nil
}

func dbInstancePopulateModify(input *rds_sdkv2.ModifyDBInstanceInput, d *schema.ResourceData) bool {
	needsModify := false

	if d.HasChanges(names.AttrAllocatedStorage, names.AttrIOPS) {
		needsModify = true
		input.AllocatedStorage = aws.Int32(int32(d.Get(names.AttrAllocatedStorage).(int)))

		// Send Iops if it has changed or not (StorageType == "gp3" and AllocatedStorage < threshold).
		if d.HasChange(names.AttrIOPS) || !isStorageTypeGP3BelowAllocatedStorageThreshold(d) {
			input.Iops = aws.Int32(int32(d.Get(names.AttrIOPS).(int)))
		}
	}

	if d.HasChange(names.AttrAutoMinorVersionUpgrade) {
		needsModify = true
		input.AutoMinorVersionUpgrade = aws.Bool(d.Get(names.AttrAutoMinorVersionUpgrade).(bool))
	}

	if d.HasChange("backup_retention_period") {
		needsModify = true
		input.BackupRetentionPeriod = aws.Int32(int32(d.Get("backup_retention_period").(int)))
	}

	if d.HasChange("backup_window") {
		needsModify = true
		input.PreferredBackupWindow = aws.String(d.Get("backup_window").(string))
	}

	if d.HasChange("copy_tags_to_snapshot") {
		needsModify = true
		input.CopyTagsToSnapshot = aws.Bool(d.Get("copy_tags_to_snapshot").(bool))
	}

	if d.HasChange("ca_cert_identifier") {
		needsModify = true
		input.CACertificateIdentifier = aws.String(d.Get("ca_cert_identifier").(string))
	}

	if d.HasChange("customer_owned_ip_enabled") {
		needsModify = true
		input.EnableCustomerOwnedIp = aws.Bool(d.Get("customer_owned_ip_enabled").(bool))
	}

	if d.HasChange("db_subnet_group_name") {
		needsModify = true
		input.DBSubnetGroupName = aws.String(d.Get("db_subnet_group_name").(string))
	}

	if d.HasChange("dedicated_log_volume") {
		needsModify = true
		input.DedicatedLogVolume = aws.Bool(d.Get("dedicated_log_volume").(bool))
	}

	if d.HasChange(names.AttrDeletionProtection) {
		needsModify = true
	}
	// Always set this. Fixes TestAccRDSInstance_BlueGreenDeployment_updateWithDeletionProtection
	input.DeletionProtection = aws.Bool(d.Get(names.AttrDeletionProtection).(bool))

	// "InvalidParameterCombination: Specify the parameters for either AWS Managed Active Directory or self-managed Active Directory".
	if d.HasChanges(names.AttrDomain, "domain_iam_role_name") {
		needsModify = true
		input.Domain = aws.String(d.Get(names.AttrDomain).(string))
		input.DomainIAMRoleName = aws.String(d.Get("domain_iam_role_name").(string))
	} else if d.HasChanges("domain_auth_secret_arn", "domain_dns_ips", "domain_fqdn", "domain_ou") {
		needsModify = true
		input.DomainAuthSecretArn = aws.String(d.Get("domain_auth_secret_arn").(string))
		if v, ok := d.GetOk("domain_dns_ips"); ok && len(v.([]interface{})) > 0 {
			input.DomainDnsIps = flex.ExpandStringValueList(v.([]interface{}))
		}
		input.DomainFqdn = aws.String(d.Get("domain_fqdn").(string))
		input.DomainOu = aws.String(d.Get("domain_ou").(string))
	}

	if d.HasChange("enabled_cloudwatch_logs_exports") {
		needsModify = true
		oraw, nraw := d.GetChange("enabled_cloudwatch_logs_exports")
		o := oraw.(*schema.Set)
		n := nraw.(*schema.Set)

		enable := n.Difference(o)
		disable := o.Difference(n)

		input.CloudwatchLogsExportConfiguration = &types.CloudwatchLogsExportConfiguration{
			EnableLogTypes:  flex.ExpandStringValueSet(enable),
			DisableLogTypes: flex.ExpandStringValueSet(disable),
		}
	}

	if d.HasChange("iam_database_authentication_enabled") {
		needsModify = true
		input.EnableIAMDatabaseAuthentication = aws.Bool(d.Get("iam_database_authentication_enabled").(bool))
	}

	if d.HasChange(names.AttrIdentifier) {
		needsModify = true
		input.NewDBInstanceIdentifier = aws.String(d.Get(names.AttrIdentifier).(string))
	}

	if d.HasChange("instance_class") {
		needsModify = true
		input.DBInstanceClass = aws.String(d.Get("instance_class").(string))
	}

	if d.HasChange("license_model") {
		needsModify = true
		input.LicenseModel = aws.String(d.Get("license_model").(string))
	}

	if d.HasChange("maintenance_window") {
		needsModify = true
		input.PreferredMaintenanceWindow = aws.String(d.Get("maintenance_window").(string))
	}

	if d.HasChange("manage_master_user_password") {
		needsModify = true
		input.ManageMasterUserPassword = aws.Bool(d.Get("manage_master_user_password").(bool))
	}

	if d.HasChange("master_user_secret_kms_key_id") {
		needsModify = true
		if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
			input.MasterUserSecretKmsKeyId = aws.String(v.(string))
			// InvalidParameterValue: A ManageMasterUserPassword value is required when MasterUserSecretKmsKeyId is specified.
			input.ManageMasterUserPassword = aws.Bool(d.Get("manage_master_user_password").(bool))
		}
	}

	if d.HasChange("max_allocated_storage") {
		needsModify = true
		v := d.Get("max_allocated_storage").(int)

		// The API expects the max allocated storage value to be set to the allocated storage
		// value when disabling autoscaling. This check ensures that value is set correctly
		// if the update to the Terraform configuration was removing the argument completely.
		if v == 0 {
			v = d.Get(names.AttrAllocatedStorage).(int)
		}

		input.MaxAllocatedStorage = aws.Int32(int32(v))
	}

	if d.HasChange("monitoring_interval") {
		needsModify = true
		input.MonitoringInterval = aws.Int32(int32(d.Get("monitoring_interval").(int)))
	}

	if d.HasChange("monitoring_role_arn") {
		needsModify = true
		input.MonitoringRoleArn = aws.String(d.Get("monitoring_role_arn").(string))
	}

	if d.HasChange("multi_az") {
		needsModify = true
		input.MultiAZ = aws.Bool(d.Get("multi_az").(bool))
	}

	if d.HasChange("network_type") {
		needsModify = true
		input.NetworkType = aws.String(d.Get("network_type").(string))
	}

	if d.HasChange("option_group_name") {
		needsModify = true
		input.OptionGroupName = aws.String(d.Get("option_group_name").(string))
	}

	if d.HasChange(names.AttrPassword) {
		needsModify = true
		// With ManageMasterUserPassword set to true, the password is no longer needed, so we omit it from the API call.
		if v, ok := d.GetOk(names.AttrPassword); ok {
			input.MasterUserPassword = aws.String(v.(string))
		}
	}

	if d.HasChanges("performance_insights_enabled", "performance_insights_kms_key_id", "performance_insights_retention_period") {
		needsModify = true
		input.EnablePerformanceInsights = aws.Bool(d.Get("performance_insights_enabled").(bool))

		if v, ok := d.GetOk("performance_insights_kms_key_id"); ok {
			input.PerformanceInsightsKMSKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("performance_insights_retention_period"); ok {
			input.PerformanceInsightsRetentionPeriod = aws.Int32(int32(v.(int)))
		}
	}

	if d.HasChange(names.AttrPort) {
		needsModify = true
		input.DBPortNumber = aws.Int32(int32(d.Get(names.AttrPort).(int)))
	}

	if d.HasChange(names.AttrPubliclyAccessible) {
		needsModify = true
		input.PubliclyAccessible = aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool))
	}

	if d.HasChange("replica_mode") {
		needsModify = true
		input.ReplicaMode = types.ReplicaMode(d.Get("replica_mode").(string))
	}

	if d.HasChange("storage_throughput") {
		needsModify = true
		input.StorageThroughput = aws.Int32(int32(d.Get("storage_throughput").(int)))

		if input.Iops == nil {
			input.Iops = aws.Int32(int32(d.Get(names.AttrIOPS).(int)))
		}

		if input.AllocatedStorage == nil {
			input.AllocatedStorage = aws.Int32(int32(d.Get(names.AttrAllocatedStorage).(int)))
		}
	}

	if d.HasChange(names.AttrStorageType) {
		needsModify = true
		input.StorageType = aws.String(d.Get(names.AttrStorageType).(string))

		if slices.Contains([]string{storageTypeIO1, storageTypeIO2}, aws.StringValue(input.StorageType)) {
			input.Iops = aws.Int32(int32(d.Get(names.AttrIOPS).(int)))
		}
	}

	if d.HasChange(names.AttrVPCSecurityGroupIDs) {
		if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
			needsModify = true
			input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v)
		}
	}

	return needsModify
}

func dbInstanceModify(ctx context.Context, conn *rds_sdkv2.Client, resourceID string, input *rds_sdkv2.ModifyDBInstanceInput, timeout time.Duration) error {
	_, err := tfresource.RetryWhen(ctx, timeout,
		func() (interface{}, error) {
			return conn.ModifyDBInstance(ctx, input)
		},
		func(err error) (bool, error) {
			// Retry for IAM eventual consistency.
			if tfawserr_sdkv2.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
				return true, err
			}

			if tfawserr_sdkv2.ErrMessageContains(err, errCodeInvalidParameterCombination, "previous storage change is being optimized") {
				return true, err
			}

			if errs.IsA[*types.InvalidDBClusterStateFault](err) {
				return true, err
			}

			return false, err
		},
	)
	if err != nil {
		return err
	}

	if _, err := waitDBInstanceAvailableSDKv2(ctx, conn, resourceID, timeout); err != nil {
		return fmt.Errorf("waiting for completion: %w", err)
	}
	return nil
}

// See https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html#gp3-storage.
func isStorageTypeGP3BelowAllocatedStorageThreshold(d *schema.ResourceData) bool {
	if storageType := d.Get(names.AttrStorageType).(string); storageType != storageTypeGP3 {
		return false
	}

	switch allocatedStorage, engine := d.Get(names.AttrAllocatedStorage).(int), d.Get(names.AttrEngine).(string); engine {
	case InstanceEngineDB2Advanced, InstanceEngineDB2Standard:
		return allocatedStorage < 100
	case InstanceEngineMariaDB, InstanceEngineMySQL, InstanceEnginePostgres:
		return allocatedStorage < 400
	case InstanceEngineOracleEnterprise, InstanceEngineOracleEnterpriseCDB, InstanceEngineOracleStandard2, InstanceEngineOracleStandard2CDB:
		return allocatedStorage < 200
	}

	return false
}

func dbSetResourceDataEngineVersionFromInstance(d *schema.ResourceData, c *rds.DBInstance) {
	oldVersion := d.Get(names.AttrEngineVersion).(string)
	newVersion := aws.StringValue(c.EngineVersion)
	var pendingVersion string
	if c.PendingModifiedValues != nil && c.PendingModifiedValues.EngineVersion != nil {
		pendingVersion = aws.StringValue(c.PendingModifiedValues.EngineVersion)
	}
	compareActualEngineVersion(d, oldVersion, newVersion, pendingVersion)
}

type dbInstanceARN struct {
	arn.ARN
	Identifier string
}

func parseDBInstanceARN(s string) (dbInstanceARN, error) {
	arn, err := arn.Parse(s)
	if err != nil {
		return dbInstanceARN{}, err
	}

	result := dbInstanceARN{
		ARN: arn,
	}

	re := regexache.MustCompile(`^db:([0-9a-z-]+)$`)
	matches := re.FindStringSubmatch(arn.Resource)
	if matches == nil || len(matches) != 2 {
		return dbInstanceARN{}, errors.New("DB Instance ARN: invalid resource section")
	}
	result.Identifier = matches[1]

	return result, nil
}

// findDBInstanceByIDSDKv1 in general should be called with a DbiResourceId of the form
// "db-BE6UI2KLPQP3OVDYD74ZEV6NUM" rather than a DB identifier. However, in some cases only
// the identifier is available, and can be used.
func findDBInstanceByIDSDKv1(ctx context.Context, conn *rds.RDS, id string) (*rds.DBInstance, error) {
	idLooksLikeDbiResourceId := regexache.MustCompile(`^db-[0-9A-Za-z]{2,255}$`).MatchString(id)
	input := &rds.DescribeDBInstancesInput{}

	if idLooksLikeDbiResourceId {
		input.Filters = []*rds.Filter{
			{
				Name:   aws.String("dbi-resource-id"),
				Values: aws.StringSlice([]string{id}),
			},
		}
	} else {
		input.DBInstanceIdentifier = aws.String(id)
	}

	output, err := findDBInstanceSDKv1(ctx, conn, input, tfslices.PredicateTrue[*rds.DBInstance]())

	// in case a DB has an *identifier* starting with "db-""
	if idLooksLikeDbiResourceId && tfresource.NotFound(err) {
		input := &rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(id),
		}

		output, err = findDBInstanceSDKv1(ctx, conn, input, tfslices.PredicateTrue[*rds.DBInstance]())
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findDBInstanceSDKv1(ctx context.Context, conn *rds.RDS, input *rds.DescribeDBInstancesInput, filter tfslices.Predicate[*rds.DBInstance]) (*rds.DBInstance, error) {
	output, err := findDBInstancesSDKv1(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findDBInstancesSDKv1(ctx context.Context, conn *rds.RDS, input *rds.DescribeDBInstancesInput, filter tfslices.Predicate[*rds.DBInstance]) ([]*rds.DBInstance, error) {
	var output []*rds.DBInstance

	err := conn.DescribeDBInstancesPagesWithContext(ctx, input, func(page *rds.DescribeDBInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBInstances {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

// findDBInstanceByIDSDKv2 in general should be called with a DbiResourceId of the form
// "db-BE6UI2KLPQP3OVDYD74ZEV6NUM" rather than a DB identifier. However, in some cases only
// the identifier is available, and can be used.
func findDBInstanceByIDSDKv2(ctx context.Context, conn *rds_sdkv2.Client, id string, optFns ...func(*rds_sdkv2.Options)) (*types.DBInstance, error) {
	input := &rds_sdkv2.DescribeDBInstancesInput{}

	if regexache.MustCompile(`^db-[0-9A-Za-z]{2,255}$`).MatchString(id) {
		input.Filters = []types.Filter{
			{
				Name:   aws.String("dbi-resource-id"),
				Values: []string{id},
			},
		}
	} else {
		input.DBInstanceIdentifier = aws.String(id)
	}

	output, err := conn.DescribeDBInstances(ctx, input, optFns...)

	// in case a DB has an *identifier* starting with "db-""
	if regexache.MustCompile(`^db-[0-9A-Za-z]{2,255}$`).MatchString(id) && (output == nil || len(output.DBInstances) == 0) {
		input = &rds_sdkv2.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(id),
		}
		output, err = conn.DescribeDBInstances(ctx, input, optFns...)
	}

	if errs.IsA[*types.DBInstanceNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfresource.AssertSingleValueResult(output.DBInstances)
}

func statusDBInstanceSDKv1(ctx context.Context, conn *rds.RDS, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDBInstanceByIDSDKv1(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.DBInstanceStatus), nil
	}
}

func statusDBInstanceSDKv2(ctx context.Context, conn *rds_sdkv2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDBInstanceByIDSDKv2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.DBInstanceStatus), nil
	}
}

func waitDBInstanceAvailableSDKv1(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration, optFns ...tfresource.OptionsFunc) (*rds.DBInstance, error) {
	options := tfresource.Options{
		PollInterval:              10 * time.Second,
		Delay:                     1 * time.Minute,
		ContinuousTargetOccurence: 3,
	}
	for _, fn := range optFns {
		fn(&options)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{
			InstanceStatusBackingUp,
			InstanceStatusConfiguringEnhancedMonitoring,
			InstanceStatusConfiguringIAMDatabaseAuth,
			InstanceStatusConfiguringLogExports,
			InstanceStatusCreating,
			InstanceStatusMaintenance,
			InstanceStatusModifying,
			InstanceStatusMovingToVPC,
			InstanceStatusRebooting,
			InstanceStatusRenaming,
			InstanceStatusResettingMasterCredentials,
			InstanceStatusStarting,
			InstanceStatusStopping,
			InstanceStatusStorageFull,
			InstanceStatusUpgrading,
		},
		Target:  []string{InstanceStatusAvailable, InstanceStatusStorageOptimization},
		Refresh: statusDBInstanceSDKv1(ctx, conn, id),
		Timeout: timeout,
	}
	options.Apply(stateConf)

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBInstance); ok {
		return output, err
	}

	return nil, err
}

func waitDBInstanceAvailableSDKv2(ctx context.Context, conn *rds_sdkv2.Client, id string, timeout time.Duration, optFns ...tfresource.OptionsFunc) (*rds.DBInstance, error) {
	options := tfresource.Options{
		PollInterval:              10 * time.Second,
		Delay:                     1 * time.Minute,
		ContinuousTargetOccurence: 3,
	}
	for _, fn := range optFns {
		fn(&options)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{
			InstanceStatusBackingUp,
			InstanceStatusConfiguringEnhancedMonitoring,
			InstanceStatusConfiguringIAMDatabaseAuth,
			InstanceStatusConfiguringLogExports,
			InstanceStatusCreating,
			InstanceStatusMaintenance,
			InstanceStatusModifying,
			InstanceStatusMovingToVPC,
			InstanceStatusRebooting,
			InstanceStatusRenaming,
			InstanceStatusResettingMasterCredentials,
			InstanceStatusStarting,
			InstanceStatusStopping,
			InstanceStatusStorageFull,
			InstanceStatusUpgrading,
		},
		Target:  []string{InstanceStatusAvailable, InstanceStatusStorageOptimization},
		Refresh: statusDBInstanceSDKv2(ctx, conn, id),
		Timeout: timeout,
	}
	options.Apply(stateConf)

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBInstance); ok {
		return output, err
	}

	return nil, err
}

func waitDBInstanceDeleted(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration, optFns ...tfresource.OptionsFunc) (*rds.DBInstance, error) {
	options := tfresource.Options{
		PollInterval:              10 * time.Second,
		Delay:                     1 * time.Minute,
		ContinuousTargetOccurence: 3,
	}
	for _, fn := range optFns {
		fn(&options)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{
			InstanceStatusAvailable,
			InstanceStatusBackingUp,
			InstanceStatusConfiguringEnhancedMonitoring,
			InstanceStatusConfiguringLogExports,
			InstanceStatusCreating,
			InstanceStatusDeletePreCheck,
			InstanceStatusDeleting,
			InstanceStatusIncompatibleParameters,
			InstanceStatusIncompatibleRestore,
			InstanceStatusModifying,
			InstanceStatusStarting,
			InstanceStatusStopping,
			InstanceStatusStorageFull,
			InstanceStatusStorageOptimization,
		},
		Target:  []string{},
		Refresh: statusDBInstanceSDKv1(ctx, conn, id),
		Timeout: timeout,
	}
	options.Apply(stateConf)

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBInstance); ok {
		return output, err
	}

	return nil, err
}

func findBlueGreenDeploymentByID(ctx context.Context, conn *rds_sdkv2.Client, id string) (*types.BlueGreenDeployment, error) {
	input := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
		BlueGreenDeploymentIdentifier: aws.String(id),
	}

	output, err := conn.DescribeBlueGreenDeployments(ctx, input)

	if errs.IsA[*types.BlueGreenDeploymentNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.BlueGreenDeployments) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	deployment := output.BlueGreenDeployments[0]

	if aws.StringValue(deployment.BlueGreenDeploymentIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return &deployment, nil
}

func statusBlueGreenDeployment(ctx context.Context, conn *rds_sdkv2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findBlueGreenDeploymentByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitBlueGreenDeploymentAvailable(ctx context.Context, conn *rds_sdkv2.Client, id string, timeout time.Duration, optFns ...tfresource.OptionsFunc) (*types.BlueGreenDeployment, error) {
	options := tfresource.Options{
		PollInterval: 10 * time.Second,
		Delay:        1 * time.Minute,
	}
	for _, fn := range optFns {
		fn(&options)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{"PROVISIONING"},
		Target:  []string{"AVAILABLE"},
		Refresh: statusBlueGreenDeployment(ctx, conn, id),
		Timeout: timeout,
	}
	options.Apply(stateConf)

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.BlueGreenDeployment); ok {
		return output, err
	}

	return nil, err
}

func waitBlueGreenDeploymentSwitchoverCompleted(ctx context.Context, conn *rds_sdkv2.Client, id string, timeout time.Duration, optFns ...tfresource.OptionsFunc) (*types.BlueGreenDeployment, error) {
	options := tfresource.Options{
		PollInterval: 10 * time.Second,
		Delay:        1 * time.Minute,
	}
	for _, fn := range optFns {
		fn(&options)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{"AVAILABLE", "SWITCHOVER_IN_PROGRESS"},
		Target:  []string{"SWITCHOVER_COMPLETED"},
		Refresh: statusBlueGreenDeployment(ctx, conn, id),
		Timeout: timeout,
	}
	options.Apply(stateConf)

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.BlueGreenDeployment); ok {
		if status := aws.StringValue(output.Status); status == "INVALID_CONFIGURATION" || status == "SWITCHOVER_FAILED" {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusDetails)))
		}

		return output, err
	}

	return nil, err
}

func waitBlueGreenDeploymentDeleted(ctx context.Context, conn *rds_sdkv2.Client, id string, timeout time.Duration, optFns ...tfresource.OptionsFunc) (*types.BlueGreenDeployment, error) {
	options := tfresource.Options{
		PollInterval: 10 * time.Second,
		Delay:        1 * time.Minute,
	}
	for _, fn := range optFns {
		fn(&options)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{"PROVISIONING", "AVAILABLE", "SWITCHOVER_IN_PROGRESS", "SWITCHOVER_COMPLETED", "INVALID_CONFIGURATION", "SWITCHOVER_FAILED", "DELETING"},
		Target:  []string{},
		Refresh: statusBlueGreenDeployment(ctx, conn, id),
		Timeout: timeout,
	}
	options.Apply(stateConf)

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.BlueGreenDeployment); ok {
		return output, err
	}

	return nil, err
}

func dbInstanceValidBlueGreenEngines() []string {
	return []string{
		InstanceEngineMariaDB,
		InstanceEngineMySQL,
		InstanceEnginePostgres,
	}
}

func flattenEndpoint(apiObject *rds.Endpoint) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Address; v != nil {
		tfMap[names.AttrAddress] = aws.StringValue(v)
	}

	if v := apiObject.HostedZoneId; v != nil {
		tfMap[names.AttrHostedZoneID] = aws.StringValue(v)
	}

	if v := apiObject.Port; v != nil {
		tfMap[names.AttrPort] = aws.Int64Value(v)
	}

	return tfMap
}
