package rds

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstanceCreate,
		Read:   resourceInstanceRead,
		Update: resourceInstanceUpdate,
		Delete: resourceInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceInstanceImport,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceInstanceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: InstanceStateUpgradeV0,
				Version: 0,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Update: schema.DefaultTimeout(80 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"allocated_storage": {
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
			"allow_major_version_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			// apply_immediately is used to determine when the update modifications
			// take place.
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
			"auto_minor_version_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"backup_retention_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"backup_window": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidOnceADayWindowFormat,
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
			},
			"copy_tags_to_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
					"name",
					"replicate_source_db",
				},
			},
			"db_subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"delete_automated_backups": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain_iam_role_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled_cloudwatch_logs_exports": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(InstanceExportableLogType_Values(), false),
				},
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					value := v.(string)
					return strings.ToLower(value)
				},
				ConflictsWith: []string{"replicate_source_db"},
			},
			"engine_version": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"replicate_source_db"},
			},
			"engine_version_actual": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"final_snapshot_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z]`), "must begin with alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[0-9A-Za-z-]+$`), "must only contain alphanumeric characters and hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end in a hyphen"),
				),
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_database_authentication_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"identifier": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"identifier_prefix"},
				ValidateFunc:  validIdentifier,
			},
			"identifier_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifierPrefix,
			},
			"instance_class": {
				Type:     schema.TypeString,
				Required: true,
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
			"latest_restorable_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"license_model": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
			"max_allocated_storage": {
				Type:     schema.TypeInt,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "0" && new == fmt.Sprintf("%d", d.Get("allocated_storage").(int)) {
						return true
					}
					return false
				},
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
			"multi_az": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:       schema.TypeString,
				Optional:   true,
				Computed:   true,
				Deprecated: "Use db_name instead",
				ForceNew:   true,
				ConflictsWith: []string{
					"db_name",
					"replicate_source_db",
				},
			},
			"nchar_character_set_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"option_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
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
			},
			"port": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"publicly_accessible": {
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
			"resource_id": {
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

						"source_db_instance_identifier": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"source_db_instance_automated_backups_arn": {
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
			},
			"security_group_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"storage_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(StorageType_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"timezone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"username": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"replicate_source_db"},
			},
			"vpc_security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	// Some API calls (e.g. CreateDBInstanceReadReplica and
	// RestoreDBInstanceFromDBSnapshot do not support all parameters to
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

	identifier := create.Name(d.Get("identifier").(string), d.Get("identifier_prefix").(string))

	if v, ok := d.GetOk("replicate_source_db"); ok {
		opts := rds.CreateDBInstanceReadReplicaInput{
			AutoMinorVersionUpgrade:    aws.Bool(d.Get("auto_minor_version_upgrade").(bool)),
			CopyTagsToSnapshot:         aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DeletionProtection:         aws.Bool(d.Get("deletion_protection").(bool)),
			DBInstanceClass:            aws.String(d.Get("instance_class").(string)),
			DBInstanceIdentifier:       aws.String(identifier),
			PubliclyAccessible:         aws.Bool(d.Get("publicly_accessible").(bool)),
			SourceDBInstanceIdentifier: aws.String(v.(string)),
			Tags:                       Tags(tags.IgnoreAWS()),
		}

		if _, ok := d.GetOk("allocated_storage"); ok {
			// RDS doesn't allow modifying the storage of a replica within the first 6h of creation.
			// allocated_storage is inherited from the primary so only the same value or no value is correct; a different value would fail the creation.
			// A different value is possible, granted: the value is higher than the current, there has been 6h between
			log.Printf("[INFO] allocated_storage was ignored for DB Instance (%s) because a replica inherits the primary's allocated_storage and this cannot be changed at creation.", d.Id())
		}

		if attr, ok := d.GetOk("availability_zone"); ok {
			opts.AvailabilityZone = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("db_subnet_group_name"); ok {
			opts.DBSubnetGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && attr.(*schema.Set).Len() > 0 {
			opts.EnableCloudwatchLogsExports = flex.ExpandStringSet(attr.(*schema.Set))
		}

		if attr, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			opts.EnableIAMDatabaseAuthentication = aws.Bool(attr.(bool))
		}

		if attr, ok := d.GetOk("iops"); ok {
			opts.Iops = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("kms_key_id"); ok {
			opts.KmsKeyId = aws.String(attr.(string))
			if arnParts := strings.Split(v.(string), ":"); len(arnParts) >= 4 {
				opts.SourceRegion = aws.String(arnParts[3])
			}
		}

		if attr, ok := d.GetOk("monitoring_interval"); ok {
			opts.MonitoringInterval = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("monitoring_role_arn"); ok {
			opts.MonitoringRoleArn = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("multi_az"); ok {
			opts.MultiAZ = aws.Bool(attr.(bool))
		}

		if attr, ok := d.GetOk("option_group_name"); ok {
			opts.OptionGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("port"); ok {
			opts.Port = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("storage_type"); ok {
			opts.StorageType = aws.String(attr.(string))
		}

		if attr := d.Get("vpc_security_group_ids").(*schema.Set); attr.Len() > 0 {
			opts.VpcSecurityGroupIds = flex.ExpandStringSet(attr)
		}

		if attr, ok := d.GetOk("performance_insights_enabled"); ok {
			opts.EnablePerformanceInsights = aws.Bool(attr.(bool))
		}

		if attr, ok := d.GetOk("performance_insights_kms_key_id"); ok {
			opts.PerformanceInsightsKMSKeyId = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("performance_insights_retention_period"); ok {
			opts.PerformanceInsightsRetentionPeriod = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("replica_mode"); ok {
			opts.ReplicaMode = aws.String(attr.(string))
			requiresModifyDbInstance = true
		}

		output, err := conn.CreateDBInstanceReadReplica(&opts)
		if err != nil {
			return fmt.Errorf("Error creating DB Instance: %w", err)
		}

		if attr, ok := d.GetOk("allow_major_version_upgrade"); ok {
			// Having allowing_major_version_upgrade by itself should not trigger ModifyDBInstance
			// InvalidParameterCombination: No modifications were requested
			modifyDbInstanceInput.AllowMajorVersionUpgrade = aws.Bool(attr.(bool))
		}

		if attr, ok := d.GetOk("backup_retention_period"); ok {
			current := aws.Int64Value(output.DBInstance.BackupRetentionPeriod)
			desired := int64(attr.(int))
			if current != desired {
				modifyDbInstanceInput.BackupRetentionPeriod = aws.Int64(desired)
				requiresModifyDbInstance = true
			}
		}

		if attr, ok := d.GetOk("backup_window"); ok {
			current := aws.StringValue(output.DBInstance.PreferredBackupWindow)
			desired := attr.(string)
			if current != desired {
				modifyDbInstanceInput.PreferredBackupWindow = aws.String(desired)
				requiresModifyDbInstance = true
			}
		}

		if attr, ok := d.GetOk("ca_cert_identifier"); ok {
			current := aws.StringValue(output.DBInstance.CACertificateIdentifier)
			desired := attr.(string)
			if current != desired {
				modifyDbInstanceInput.CACertificateIdentifier = aws.String(desired)
				requiresModifyDbInstance = true
			}
		}

		if attr, ok := d.GetOk("maintenance_window"); ok {
			current := aws.StringValue(output.DBInstance.PreferredMaintenanceWindow)
			desired := attr.(string)
			if current != desired {
				modifyDbInstanceInput.PreferredMaintenanceWindow = aws.String(desired)
				requiresModifyDbInstance = true
			}
		}

		if attr, ok := d.GetOk("max_allocated_storage"); ok {
			current := aws.Int64Value(output.DBInstance.MaxAllocatedStorage)
			desired := int64(attr.(int))
			if current != desired {
				modifyDbInstanceInput.MaxAllocatedStorage = aws.Int64(desired)
				requiresModifyDbInstance = true
			}
		}

		if attr, ok := d.GetOk("parameter_group_name"); ok {
			if len(output.DBInstance.DBParameterGroups) > 0 {
				current := aws.StringValue(output.DBInstance.DBParameterGroups[0].DBParameterGroupName)
				desired := attr.(string)
				if current != desired {
					modifyDbInstanceInput.DBParameterGroupName = aws.String(desired)
					requiresModifyDbInstance = true
					requiresRebootDbInstance = true
				}
			}
		}

		if attr, ok := d.GetOk("password"); ok {
			modifyDbInstanceInput.MasterUserPassword = aws.String(attr.(string))
			requiresModifyDbInstance = true
		}

		if attr := d.Get("security_group_names").(*schema.Set); attr.Len() > 0 {
			current := flattenDBSecurityGroups(output.DBInstance.DBSecurityGroups)
			desired := attr
			if !desired.Equal(current) {
				modifyDbInstanceInput.DBSecurityGroups = flex.ExpandStringSet(attr)
				requiresModifyDbInstance = true
			}
		}
	} else if v, ok := d.GetOk("s3_import"); ok {
		dbName := d.Get("db_name").(string)
		if dbName == "" {
			dbName = d.Get("name").(string)
		}

		if _, ok := d.GetOk("allocated_storage"); !ok {

			return fmt.Errorf(`provider.aws: aws_db_instance: %s: "allocated_storage": required field is not set`, dbName)
		}
		if _, ok := d.GetOk("engine"); !ok {
			return fmt.Errorf(`provider.aws: aws_db_instance: %s: "engine": required field is not set`, dbName)
		}
		if _, ok := d.GetOk("password"); !ok {
			return fmt.Errorf(`provider.aws: aws_db_instance: %s: "password": required field is not set`, dbName)
		}
		if _, ok := d.GetOk("username"); !ok {
			return fmt.Errorf(`provider.aws: aws_db_instance: %s: "username": required field is not set`, dbName)
		}

		s3_bucket := v.([]interface{})[0].(map[string]interface{})
		opts := rds.RestoreDBInstanceFromS3Input{
			AllocatedStorage:        aws.Int64(int64(d.Get("allocated_storage").(int))),
			AutoMinorVersionUpgrade: aws.Bool(d.Get("auto_minor_version_upgrade").(bool)),
			CopyTagsToSnapshot:      aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBName:                  aws.String(dbName),
			DBInstanceClass:         aws.String(d.Get("instance_class").(string)),
			DBInstanceIdentifier:    aws.String(identifier),
			DeletionProtection:      aws.Bool(d.Get("deletion_protection").(bool)),
			Engine:                  aws.String(d.Get("engine").(string)),
			EngineVersion:           aws.String(d.Get("engine_version").(string)),
			S3BucketName:            aws.String(s3_bucket["bucket_name"].(string)),
			S3Prefix:                aws.String(s3_bucket["bucket_prefix"].(string)),
			S3IngestionRoleArn:      aws.String(s3_bucket["ingestion_role"].(string)),
			MasterUsername:          aws.String(d.Get("username").(string)),
			MasterUserPassword:      aws.String(d.Get("password").(string)),
			PubliclyAccessible:      aws.Bool(d.Get("publicly_accessible").(bool)),
			StorageEncrypted:        aws.Bool(d.Get("storage_encrypted").(bool)),
			SourceEngine:            aws.String(s3_bucket["source_engine"].(string)),
			SourceEngineVersion:     aws.String(s3_bucket["source_engine_version"].(string)),
			Tags:                    Tags(tags.IgnoreAWS()),
		}

		if attr, ok := d.GetOk("multi_az"); ok {
			opts.MultiAZ = aws.Bool(attr.(bool))
		}

		if _, ok := d.GetOk("character_set_name"); ok {
			return fmt.Errorf(`provider.aws: aws_db_instance: %s: "character_set_name" doesn't work with with restores"`, dbName)
		}
		if _, ok := d.GetOk("timezone"); ok {
			return fmt.Errorf(`provider.aws: aws_db_instance: %s: "timezone" doesn't work with with restores"`, dbName)
		}

		attr := d.Get("backup_retention_period")
		opts.BackupRetentionPeriod = aws.Int64(int64(attr.(int)))

		if attr, ok := d.GetOk("maintenance_window"); ok {
			opts.PreferredMaintenanceWindow = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("backup_window"); ok {
			opts.PreferredBackupWindow = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("license_model"); ok {
			opts.LicenseModel = aws.String(attr.(string))
		}
		if attr, ok := d.GetOk("parameter_group_name"); ok {
			opts.DBParameterGroupName = aws.String(attr.(string))
		}

		if attr := d.Get("vpc_security_group_ids").(*schema.Set); attr.Len() > 0 {
			opts.VpcSecurityGroupIds = flex.ExpandStringSet(attr)
		}

		if attr := d.Get("security_group_names").(*schema.Set); attr.Len() > 0 {
			opts.DBSecurityGroups = flex.ExpandStringSet(attr)
		}
		if attr, ok := d.GetOk("storage_type"); ok {
			opts.StorageType = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("db_subnet_group_name"); ok {
			opts.DBSubnetGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("iops"); ok {
			opts.Iops = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("port"); ok {
			opts.Port = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("availability_zone"); ok {
			opts.AvailabilityZone = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("monitoring_role_arn"); ok {
			opts.MonitoringRoleArn = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("monitoring_interval"); ok {
			opts.MonitoringInterval = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("option_group_name"); ok {
			opts.OptionGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("kms_key_id"); ok {
			opts.KmsKeyId = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			opts.EnableIAMDatabaseAuthentication = aws.Bool(attr.(bool))
		}

		if attr, ok := d.GetOk("performance_insights_enabled"); ok {
			opts.EnablePerformanceInsights = aws.Bool(attr.(bool))
		}

		if attr, ok := d.GetOk("performance_insights_kms_key_id"); ok {
			opts.PerformanceInsightsKMSKeyId = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("performance_insights_retention_period"); ok {
			opts.PerformanceInsightsRetentionPeriod = aws.Int64(int64(attr.(int)))
		}

		log.Printf("[DEBUG] DB Instance S3 Restore configuration: %#v", opts)
		var err error
		// Retry for IAM eventual consistency
		err = resource.Retry(propagationTimeout, func() *resource.RetryError {
			_, err = conn.RestoreDBInstanceFromS3(&opts)
			if err != nil {
				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "ENHANCED_MONITORING") {
					return resource.RetryableError(err)
				}
				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "S3_SNAPSHOT_INGESTION") {
					return resource.RetryableError(err)
				}
				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "S3 bucket cannot be found") {
					return resource.RetryableError(err)
				}
				// InvalidParameterValue: Files from the specified Amazon S3 bucket cannot be downloaded. Make sure that you have created an AWS Identity and Access Management (IAM) role that lets Amazon RDS access Amazon S3 for you.
				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Files from the specified Amazon S3 bucket cannot be downloaded") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if tfresource.TimedOut(err) {
			_, err = conn.RestoreDBInstanceFromS3(&opts)
		}
		if err != nil {
			return fmt.Errorf("Error creating DB Instance: %w", err)
		}

		d.SetId(identifier)

		log.Printf("[INFO] DB Instance ID: %s", d.Id())

		log.Println("[INFO] Waiting for DB Instance to be available")

		stateConf := &resource.StateChangeConf{
			Pending:    resourceInstanceCreatePendingStates,
			Target:     []string{"available", "storage-optimization"},
			Refresh:    resourceDBInstanceStateRefreshFunc(d.Id(), conn),
			Timeout:    d.Timeout(schema.TimeoutCreate),
			MinTimeout: 10 * time.Second,
			Delay:      30 * time.Second, // Wait 30 secs before starting
		}

		// Wait, catching any errors
		_, err = stateConf.WaitForState()
		if err != nil {
			return err
		}

		return resourceInstanceRead(d, meta)
	} else if _, ok := d.GetOk("snapshot_identifier"); ok {
		opts := rds.RestoreDBInstanceFromDBSnapshotInput{
			AutoMinorVersionUpgrade: aws.Bool(d.Get("auto_minor_version_upgrade").(bool)),
			CopyTagsToSnapshot:      aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBInstanceClass:         aws.String(d.Get("instance_class").(string)),
			DBInstanceIdentifier:    aws.String(identifier),
			DBSnapshotIdentifier:    aws.String(d.Get("snapshot_identifier").(string)),
			DeletionProtection:      aws.Bool(d.Get("deletion_protection").(bool)),
			PubliclyAccessible:      aws.Bool(d.Get("publicly_accessible").(bool)),
			Tags:                    Tags(tags.IgnoreAWS()),
		}

		if attr, ok := d.GetOk("db_name"); ok {
			// "Note: This parameter [DBName] doesn't apply to the MySQL, PostgreSQL, or MariaDB engines."
			// https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_RestoreDBInstanceFromDBSnapshot.html
			switch strings.ToLower(d.Get("engine").(string)) {
			case "mysql", "postgres", "mariadb":
				// skip
			default:
				opts.DBName = aws.String(attr.(string))
			}
		} else if attr, ok := d.GetOk("name"); ok {
			// "Note: This parameter [DBName] doesn't apply to the MySQL, PostgreSQL, or MariaDB engines."
			// https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_RestoreDBInstanceFromDBSnapshot.html
			switch strings.ToLower(d.Get("engine").(string)) {
			case "mysql", "postgres", "mariadb":
				// skip
			default:
				opts.DBName = aws.String(attr.(string))
			}
		}

		if attr, ok := d.GetOk("allocated_storage"); ok {
			modifyDbInstanceInput.AllocatedStorage = aws.Int64(int64(attr.(int)))
			requiresModifyDbInstance = true
		}

		if attr, ok := d.GetOk("availability_zone"); ok {
			opts.AvailabilityZone = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("allow_major_version_upgrade"); ok {
			modifyDbInstanceInput.AllowMajorVersionUpgrade = aws.Bool(attr.(bool))
			// Having allowing_major_version_upgrade by itself should not trigger ModifyDBInstance
			// InvalidParameterCombination: No modifications were requested
		}

		if attr, ok := d.GetOkExists("backup_retention_period"); ok {
			modifyDbInstanceInput.BackupRetentionPeriod = aws.Int64(int64(attr.(int)))
			requiresModifyDbInstance = true
		}

		if attr, ok := d.GetOk("backup_window"); ok {
			modifyDbInstanceInput.PreferredBackupWindow = aws.String(attr.(string))
			requiresModifyDbInstance = true
		}

		if attr, ok := d.GetOk("db_subnet_group_name"); ok {
			opts.DBSubnetGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("domain"); ok {
			opts.Domain = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("domain_iam_role_name"); ok {
			opts.DomainIAMRoleName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && attr.(*schema.Set).Len() > 0 {
			opts.EnableCloudwatchLogsExports = flex.ExpandStringSet(attr.(*schema.Set))
		}

		if attr, ok := d.GetOk("engine"); ok {
			opts.Engine = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("engine_version"); ok {
			modifyDbInstanceInput.EngineVersion = aws.String(attr.(string))
			requiresModifyDbInstance = true
		}

		if attr, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			opts.EnableIAMDatabaseAuthentication = aws.Bool(attr.(bool))
		}

		if attr, ok := d.GetOk("iops"); ok {
			modifyDbInstanceInput.Iops = aws.Int64(int64(attr.(int)))
			requiresModifyDbInstance = true
		}

		if attr, ok := d.GetOk("license_model"); ok {
			opts.LicenseModel = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("maintenance_window"); ok {
			modifyDbInstanceInput.PreferredMaintenanceWindow = aws.String(attr.(string))
			requiresModifyDbInstance = true
		}

		if attr, ok := d.GetOk("max_allocated_storage"); ok {
			modifyDbInstanceInput.MaxAllocatedStorage = aws.Int64(int64(attr.(int)))
			requiresModifyDbInstance = true
		}

		if attr, ok := d.GetOk("monitoring_interval"); ok {
			modifyDbInstanceInput.MonitoringInterval = aws.Int64(int64(attr.(int)))
			requiresModifyDbInstance = true
		}

		if attr, ok := d.GetOk("monitoring_role_arn"); ok {
			modifyDbInstanceInput.MonitoringRoleArn = aws.String(attr.(string))
			requiresModifyDbInstance = true
		}

		if attr, ok := d.GetOk("multi_az"); ok {
			// When using SQL Server engine with MultiAZ enabled, its not
			// possible to immediately enable mirroring since
			// BackupRetentionPeriod is not available as a parameter to
			// RestoreDBInstanceFromDBSnapshot and you receive an error. e.g.
			// InvalidParameterValue: Mirroring cannot be applied to instances with backup retention set to zero.
			// If we know the engine, prevent the error upfront.
			if v, ok := d.GetOk("engine"); ok && strings.HasPrefix(strings.ToLower(v.(string)), "sqlserver") {
				modifyDbInstanceInput.MultiAZ = aws.Bool(attr.(bool))
				requiresModifyDbInstance = true
			} else {
				opts.MultiAZ = aws.Bool(attr.(bool))
			}
		}

		if attr, ok := d.GetOk("option_group_name"); ok {
			opts.OptionGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("parameter_group_name"); ok {
			opts.DBParameterGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("password"); ok {
			modifyDbInstanceInput.MasterUserPassword = aws.String(attr.(string))
			requiresModifyDbInstance = true
		}

		if attr, ok := d.GetOk("port"); ok {
			opts.Port = aws.Int64(int64(attr.(int)))
		}

		if attr := d.Get("security_group_names").(*schema.Set); attr.Len() > 0 {
			modifyDbInstanceInput.DBSecurityGroups = flex.ExpandStringSet(attr)
			requiresModifyDbInstance = true
		}

		if attr, ok := d.GetOk("storage_type"); ok {
			modifyDbInstanceInput.StorageType = aws.String(attr.(string))
			requiresModifyDbInstance = true
		}

		if attr, ok := d.GetOk("tde_credential_arn"); ok {
			opts.TdeCredentialArn = aws.String(attr.(string))
		}

		if attr := d.Get("vpc_security_group_ids").(*schema.Set); attr.Len() > 0 {
			opts.VpcSecurityGroupIds = flex.ExpandStringSet(attr)
		}

		if attr, ok := d.GetOk("performance_insights_enabled"); ok {
			modifyDbInstanceInput.EnablePerformanceInsights = aws.Bool(attr.(bool))
			requiresModifyDbInstance = true

			if attr, ok := d.GetOk("performance_insights_kms_key_id"); ok {
				modifyDbInstanceInput.PerformanceInsightsKMSKeyId = aws.String(attr.(string))
			}

			if attr, ok := d.GetOk("performance_insights_retention_period"); ok {
				modifyDbInstanceInput.PerformanceInsightsRetentionPeriod = aws.Int64(int64(attr.(int)))
			}
		}

		if attr, ok := d.GetOk("customer_owned_ip_enabled"); ok {
			opts.EnableCustomerOwnedIp = aws.Bool(attr.(bool))
		}

		log.Printf("[DEBUG] DB Instance restore from snapshot configuration: %s", opts)
		_, err := conn.RestoreDBInstanceFromDBSnapshot(&opts)

		// When using SQL Server engine with MultiAZ enabled, its not
		// possible to immediately enable mirroring since
		// BackupRetentionPeriod is not available as a parameter to
		// RestoreDBInstanceFromDBSnapshot and you receive an error. e.g.
		// InvalidParameterValue: Mirroring cannot be applied to instances with backup retention set to zero.
		// Since engine is not a required argument when using snapshot_identifier
		// and the RDS API determines this condition, we catch the error
		// and remove the invalid configuration for it to be fixed afterwards.
		if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Mirroring cannot be applied to instances with backup retention set to zero") {
			opts.MultiAZ = aws.Bool(false)
			modifyDbInstanceInput.MultiAZ = aws.Bool(true)
			requiresModifyDbInstance = true
			_, err = conn.RestoreDBInstanceFromDBSnapshot(&opts)
		}

		if err != nil {
			return fmt.Errorf("Error creating DB Instance: %w", err)
		}
	} else if v, ok := d.GetOk("restore_to_point_in_time"); ok {
		if input := expandRestoreToPointInTime(v.([]interface{})); input != nil {
			input.AutoMinorVersionUpgrade = aws.Bool(d.Get("auto_minor_version_upgrade").(bool))
			input.CopyTagsToSnapshot = aws.Bool(d.Get("copy_tags_to_snapshot").(bool))
			input.DBInstanceClass = aws.String(d.Get("instance_class").(string))
			input.DeletionProtection = aws.Bool(d.Get("deletion_protection").(bool))
			input.PubliclyAccessible = aws.Bool(d.Get("publicly_accessible").(bool))
			input.Tags = Tags(tags.IgnoreAWS())
			input.TargetDBInstanceIdentifier = aws.String(identifier)

			if v, ok := d.GetOk("availability_zone"); ok {
				input.AvailabilityZone = aws.String(v.(string))
			}

			if v, ok := d.GetOk("db_name"); ok {
				input.DBName = aws.String(v.(string))
			}

			if v, ok := d.GetOk("domain"); ok {
				input.Domain = aws.String(v.(string))
			}

			if v, ok := d.GetOk("domain_iam_role_name"); ok {
				input.DomainIAMRoleName = aws.String(v.(string))
			}

			if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && v.(*schema.Set).Len() > 0 {
				input.EnableCloudwatchLogsExports = flex.ExpandStringSet(v.(*schema.Set))
			}

			if v, ok := d.GetOk("engine"); ok {
				input.Engine = aws.String(v.(string))
			}

			if v, ok := d.GetOk("iam_database_authentication_enabled"); ok {
				input.EnableIAMDatabaseAuthentication = aws.Bool(v.(bool))
			}

			if v, ok := d.GetOk("iops"); ok {
				input.Iops = aws.Int64(int64(v.(int)))
			}

			if v, ok := d.GetOk("license_model"); ok {
				input.LicenseModel = aws.String(v.(string))
			}

			if v, ok := d.GetOk("max_allocated_storage"); ok {
				input.MaxAllocatedStorage = aws.Int64(int64(v.(int)))
			}

			if v, ok := d.GetOk("multi_az"); ok {
				input.MultiAZ = aws.Bool(v.(bool))
			}

			if v, ok := d.GetOk("name"); ok {
				input.DBName = aws.String(v.(string))
			}

			if v, ok := d.GetOk("option_group_name"); ok {
				input.OptionGroupName = aws.String(v.(string))
			}

			if v, ok := d.GetOk("parameter_group_name"); ok {
				input.DBParameterGroupName = aws.String(v.(string))
			}

			if v, ok := d.GetOk("port"); ok {
				input.Port = aws.Int64(int64(v.(int)))
			}

			if v, ok := d.GetOk("storage_type"); ok {
				input.StorageType = aws.String(v.(string))
			}

			if v, ok := d.GetOk("db_subnet_group_name"); ok {
				input.DBSubnetGroupName = aws.String(v.(string))
			}

			if v, ok := d.GetOk("tde_credential_arn"); ok {
				input.TdeCredentialArn = aws.String(v.(string))
			}

			if v, ok := d.GetOk("vpc_security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
				input.VpcSecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
			}

			if attr, ok := d.GetOk("customer_owned_ip_enabled"); ok {
				input.EnableCustomerOwnedIp = aws.Bool(attr.(bool))
			}

			log.Printf("[DEBUG] DB Instance restore to point in time configuration: %s", input)

			_, err := conn.RestoreDBInstanceToPointInTime(input)
			if err != nil {
				return fmt.Errorf("error creating DB Instance: %w", err)
			}
		}
	} else {
		dbName := d.Get("db_name").(string)
		if dbName == "" {
			dbName = d.Get("name").(string)
		}

		if _, ok := d.GetOk("allocated_storage"); !ok {
			return fmt.Errorf(`provider.aws: aws_db_instance: %s: "allocated_storage": required field is not set`, dbName)
		}
		if _, ok := d.GetOk("engine"); !ok {
			return fmt.Errorf(`provider.aws: aws_db_instance: %s: "engine": required field is not set`, dbName)
		}
		if _, ok := d.GetOk("password"); !ok {
			return fmt.Errorf(`provider.aws: aws_db_instance: %s: "password": required field is not set`, dbName)
		}
		if _, ok := d.GetOk("username"); !ok {
			return fmt.Errorf(`provider.aws: aws_db_instance: %s: "username": required field is not set`, dbName)
		}

		opts := rds.CreateDBInstanceInput{
			AllocatedStorage:        aws.Int64(int64(d.Get("allocated_storage").(int))),
			DBName:                  aws.String(dbName),
			DBInstanceClass:         aws.String(d.Get("instance_class").(string)),
			DBInstanceIdentifier:    aws.String(identifier),
			DeletionProtection:      aws.Bool(d.Get("deletion_protection").(bool)),
			MasterUsername:          aws.String(d.Get("username").(string)),
			MasterUserPassword:      aws.String(d.Get("password").(string)),
			Engine:                  aws.String(d.Get("engine").(string)),
			EngineVersion:           aws.String(d.Get("engine_version").(string)),
			StorageEncrypted:        aws.Bool(d.Get("storage_encrypted").(bool)),
			AutoMinorVersionUpgrade: aws.Bool(d.Get("auto_minor_version_upgrade").(bool)),
			PubliclyAccessible:      aws.Bool(d.Get("publicly_accessible").(bool)),
			Tags:                    Tags(tags.IgnoreAWS()),
			CopyTagsToSnapshot:      aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
		}

		attr := d.Get("backup_retention_period")
		opts.BackupRetentionPeriod = aws.Int64(int64(attr.(int)))

		if attr, ok := d.GetOk("multi_az"); ok {
			opts.MultiAZ = aws.Bool(attr.(bool))

		}

		if attr, ok := d.GetOk("character_set_name"); ok {
			opts.CharacterSetName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("nchar_character_set_name"); ok {
			opts.NcharCharacterSetName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("timezone"); ok {
			opts.Timezone = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("maintenance_window"); ok {
			opts.PreferredMaintenanceWindow = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("backup_window"); ok {
			opts.PreferredBackupWindow = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("license_model"); ok {
			opts.LicenseModel = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("max_allocated_storage"); ok {
			opts.MaxAllocatedStorage = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("parameter_group_name"); ok {
			opts.DBParameterGroupName = aws.String(attr.(string))
		}

		if attr := d.Get("vpc_security_group_ids").(*schema.Set); attr.Len() > 0 {
			opts.VpcSecurityGroupIds = flex.ExpandStringSet(attr)
		}

		if attr := d.Get("security_group_names").(*schema.Set); attr.Len() > 0 {
			opts.DBSecurityGroups = flex.ExpandStringSet(attr)
		}
		if attr, ok := d.GetOk("storage_type"); ok {
			opts.StorageType = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("db_subnet_group_name"); ok {
			opts.DBSubnetGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && attr.(*schema.Set).Len() > 0 {
			opts.EnableCloudwatchLogsExports = flex.ExpandStringSet(attr.(*schema.Set))
		}

		if attr, ok := d.GetOk("iops"); ok {
			opts.Iops = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("port"); ok {
			opts.Port = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("availability_zone"); ok {
			opts.AvailabilityZone = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("monitoring_role_arn"); ok {
			opts.MonitoringRoleArn = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("monitoring_interval"); ok {
			opts.MonitoringInterval = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("option_group_name"); ok {
			opts.OptionGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("kms_key_id"); ok {
			opts.KmsKeyId = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			opts.EnableIAMDatabaseAuthentication = aws.Bool(attr.(bool))
		}

		if attr, ok := d.GetOk("domain"); ok {
			opts.Domain = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("domain_iam_role_name"); ok {
			opts.DomainIAMRoleName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("performance_insights_enabled"); ok {
			opts.EnablePerformanceInsights = aws.Bool(attr.(bool))
		}

		if attr, ok := d.GetOk("performance_insights_kms_key_id"); ok {
			opts.PerformanceInsightsKMSKeyId = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("performance_insights_retention_period"); ok {
			opts.PerformanceInsightsRetentionPeriod = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("customer_owned_ip_enabled"); ok {
			opts.EnableCustomerOwnedIp = aws.Bool(attr.(bool))
		}

		log.Printf("[DEBUG] DB Instance create configuration: %#v", opts)
		var err error
		var createdDBInstanceOutput *rds.CreateDBInstanceOutput
		err = resource.Retry(5*time.Minute, func() *resource.RetryError {
			createdDBInstanceOutput, err = conn.CreateDBInstance(&opts)
			if err != nil {
				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "ENHANCED_MONITORING") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if tfresource.TimedOut(err) {
			createdDBInstanceOutput, err = conn.CreateDBInstance(&opts)
		}
		if err != nil {
			if tfawserr.ErrCodeEquals(err, "InvalidParameterValue") {
				opts.MasterUserPassword = aws.String("********")
				return fmt.Errorf("Error creating DB Instance: %w, %+v", err, opts)
			}
			return fmt.Errorf("Error creating DB Instance: %w", err)
		}
		// This is added here to avoid unnecessary modification when ca_cert_identifier is the default one
		if attr, ok := d.GetOk("ca_cert_identifier"); ok && attr.(string) != aws.StringValue(createdDBInstanceOutput.DBInstance.CACertificateIdentifier) {
			modifyDbInstanceInput.CACertificateIdentifier = aws.String(attr.(string))
			requiresModifyDbInstance = true
		}
	}

	d.SetId(identifier)

	stateConf := &resource.StateChangeConf{
		Pending:    resourceInstanceCreatePendingStates,
		Target:     []string{"available", "storage-optimization"},
		Refresh:    resourceDBInstanceStateRefreshFunc(d.Id(), conn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	log.Printf("[INFO] Waiting for DB Instance (%s) to be available", d.Id())
	_, err := stateConf.WaitForState()
	if err != nil {
		return err
	}

	if requiresModifyDbInstance {
		modifyDbInstanceInput.DBInstanceIdentifier = aws.String(d.Id())

		log.Printf("[INFO] DB Instance (%s) configuration requires ModifyDBInstance: %s", d.Id(), modifyDbInstanceInput)
		_, err := conn.ModifyDBInstance(modifyDbInstanceInput)
		if err != nil {
			return fmt.Errorf("error modifying DB Instance (%s): %w", d.Id(), err)
		}

		log.Printf("[INFO] Waiting for DB Instance (%s) to be available", d.Id())
		err = waitUntilDBInstanceAvailableAfterUpdate(d.Id(), conn, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for DB Instance (%s) to be available: %w", d.Id(), err)
		}
	}

	if requiresRebootDbInstance {
		rebootDbInstanceInput := &rds.RebootDBInstanceInput{
			DBInstanceIdentifier: aws.String(d.Id()),
		}

		log.Printf("[INFO] DB Instance (%s) configuration requires RebootDBInstance: %s", d.Id(), rebootDbInstanceInput)
		_, err := conn.RebootDBInstance(rebootDbInstanceInput)
		if err != nil {
			return fmt.Errorf("error rebooting DB Instance (%s): %w", d.Id(), err)
		}

		log.Printf("[INFO] Waiting for DB Instance (%s) to be available", d.Id())
		err = waitUntilDBInstanceAvailableAfterUpdate(d.Id(), conn, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for DB Instance (%s) to be available: %w", d.Id(), err)
		}
	}

	return resourceInstanceRead(d, meta)
}

func resourceInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	v, err := FindDBInstanceByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DB Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DB Instance (%s): %w", d.Id(), err)
	}

	d.Set("db_name", v.DBName)
	d.Set("name", v.DBName)
	d.Set("identifier", v.DBInstanceIdentifier)
	d.Set("identifier_prefix", create.NamePrefixFromName(aws.StringValue(v.DBInstanceIdentifier)))
	d.Set("resource_id", v.DbiResourceId)
	d.Set("username", v.MasterUsername)
	d.Set("deletion_protection", v.DeletionProtection)
	d.Set("engine", v.Engine)
	d.Set("allocated_storage", v.AllocatedStorage)
	d.Set("iops", v.Iops)
	d.Set("copy_tags_to_snapshot", v.CopyTagsToSnapshot)
	d.Set("auto_minor_version_upgrade", v.AutoMinorVersionUpgrade)
	d.Set("storage_type", v.StorageType)
	d.Set("instance_class", v.DBInstanceClass)
	d.Set("availability_zone", v.AvailabilityZone)
	d.Set("backup_retention_period", v.BackupRetentionPeriod)
	d.Set("backup_window", v.PreferredBackupWindow)
	d.Set("latest_restorable_time", aws.TimeValue(v.LatestRestorableTime).Format(time.RFC3339))
	d.Set("license_model", v.LicenseModel)
	d.Set("maintenance_window", v.PreferredMaintenanceWindow)
	d.Set("max_allocated_storage", v.MaxAllocatedStorage)
	d.Set("publicly_accessible", v.PubliclyAccessible)
	d.Set("multi_az", v.MultiAZ)
	d.Set("kms_key_id", v.KmsKeyId)
	d.Set("port", v.DbInstancePort)
	d.Set("iam_database_authentication_enabled", v.IAMDatabaseAuthenticationEnabled)
	d.Set("performance_insights_enabled", v.PerformanceInsightsEnabled)
	d.Set("performance_insights_kms_key_id", v.PerformanceInsightsKMSKeyId)
	d.Set("performance_insights_retention_period", v.PerformanceInsightsRetentionPeriod)
	if v.DBSubnetGroup != nil {
		d.Set("db_subnet_group_name", v.DBSubnetGroup.DBSubnetGroupName)
	}
	d.Set("character_set_name", v.CharacterSetName)
	d.Set("nchar_character_set_name", v.NcharCharacterSetName)
	d.Set("timezone", v.Timezone)

	dbSetResourceDataEngineVersionFromInstance(d, v)

	if len(v.DBParameterGroups) > 0 {
		d.Set("parameter_group_name", v.DBParameterGroups[0].DBParameterGroupName)
	}

	if v.Endpoint != nil {
		d.Set("port", v.Endpoint.Port)
		d.Set("address", v.Endpoint.Address)
		d.Set("hosted_zone_id", v.Endpoint.HostedZoneId)
		if v.Endpoint.Address != nil && v.Endpoint.Port != nil {
			d.Set("endpoint",
				fmt.Sprintf("%s:%d", *v.Endpoint.Address, *v.Endpoint.Port))
		}
	}

	d.Set("status", v.DBInstanceStatus)
	d.Set("storage_encrypted", v.StorageEncrypted)
	if v.OptionGroupMemberships != nil {
		d.Set("option_group_name", v.OptionGroupMemberships[0].OptionGroupName)
	}

	d.Set("monitoring_interval", v.MonitoringInterval)
	d.Set("monitoring_role_arn", v.MonitoringRoleArn)

	if err := d.Set("enabled_cloudwatch_logs_exports", flex.FlattenStringList(v.EnabledCloudwatchLogsExports)); err != nil {
		return fmt.Errorf("error setting enabled_cloudwatch_logs_exports: %w", err)
	}

	d.Set("domain", "")
	d.Set("domain_iam_role_name", "")
	if len(v.DomainMemberships) > 0 && v.DomainMemberships[0] != nil {
		d.Set("domain", v.DomainMemberships[0].Domain)
		d.Set("domain_iam_role_name", v.DomainMemberships[0].IAMRoleName)
	}

	arn := aws.StringValue(v.DBInstanceArn)
	d.Set("arn", arn)

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for RDS DB Instance (%s): %w", d.Get("arn").(string), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	// Create an empty schema.Set to hold all vpc security group ids
	ids := &schema.Set{
		F: schema.HashString,
	}
	for _, v := range v.VpcSecurityGroups {
		ids.Add(*v.VpcSecurityGroupId)
	}
	d.Set("vpc_security_group_ids", ids)

	d.Set("security_group_names", flattenDBSecurityGroups(v.DBSecurityGroups))

	// replica things
	var replicas []string
	for _, v := range v.ReadReplicaDBInstanceIdentifiers {
		replicas = append(replicas, *v)
	}
	if err := d.Set("replicas", replicas); err != nil {
		return fmt.Errorf("Error setting replicas attribute: %#v, error: %w", replicas, err)
	}

	d.Set("replica_mode", v.ReplicaMode)
	d.Set("replicate_source_db", v.ReadReplicaSourceDBInstanceIdentifier)

	d.Set("ca_cert_identifier", v.CACertificateIdentifier)

	d.Set("customer_owned_ip_enabled", v.CustomerOwnedIpEnabled)

	return nil
}

func resourceInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	input := &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier:   aws.String(d.Id()),
		DeleteAutomatedBackups: aws.Bool(d.Get("delete_automated_backups").(bool)),
	}

	if d.Get("skip_final_snapshot").(bool) {
		input.SkipFinalSnapshot = aws.Bool(true)
	} else {
		input.SkipFinalSnapshot = aws.Bool(false)

		if v, ok := d.GetOk("final_snapshot_identifier"); ok {
			input.FinalDBSnapshotIdentifier = aws.String(v.(string))
		} else {
			return fmt.Errorf("final_snapshot_identifier is required when skip_final_snapshot is false")
		}
	}

	log.Printf("[DEBUG] Deleting DB Instance: %s", d.Id())
	_, err := conn.DeleteDBInstance(input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceNotFoundFault) {
		return nil
	}

	if err != nil && !tfawserr.ErrMessageContains(err, rds.ErrCodeInvalidDBInstanceStateFault, "is already being deleted") {
		return fmt.Errorf("error deleting DB Instance (%s): %w", d.Id(), err)
	}

	if _, err := waitDBInstanceDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for DB Instance (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func waitUntilDBInstanceAvailableAfterUpdate(id string, conn *rds.RDS, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    resourceInstanceUpdatePendingStates,
		Target:     []string{"available", "storage-optimization"},
		Refresh:    resourceDBInstanceStateRefreshFunc(id, conn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}
	_, err := stateConf.WaitForState()
	return err
}

func resourceInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	req := &rds.ModifyDBInstanceInput{
		ApplyImmediately:     aws.Bool(d.Get("apply_immediately").(bool)),
		DBInstanceIdentifier: aws.String(d.Id()),
	}

	if !aws.BoolValue(req.ApplyImmediately) {
		log.Println("[INFO] Only settings updating, instance changes will be applied in next maintenance window")
	}

	requestUpdate := false
	if d.HasChanges("allocated_storage", "iops") {
		req.Iops = aws.Int64(int64(d.Get("iops").(int)))
		req.AllocatedStorage = aws.Int64(int64(d.Get("allocated_storage").(int)))
		requestUpdate = true
	}
	if d.HasChange("allow_major_version_upgrade") {
		req.AllowMajorVersionUpgrade = aws.Bool(d.Get("allow_major_version_upgrade").(bool))
		// Having allowing_major_version_upgrade by itself should not trigger ModifyDBInstance
		// as it results in InvalidParameterCombination: No modifications were requested
	}
	if d.HasChange("backup_retention_period") {
		req.BackupRetentionPeriod = aws.Int64(int64(d.Get("backup_retention_period").(int)))
		requestUpdate = true
	}
	if d.HasChange("copy_tags_to_snapshot") {
		req.CopyTagsToSnapshot = aws.Bool(d.Get("copy_tags_to_snapshot").(bool))
		requestUpdate = true
	}
	if d.HasChange("ca_cert_identifier") {
		req.CACertificateIdentifier = aws.String(d.Get("ca_cert_identifier").(string))
		requestUpdate = true
	}
	if d.HasChange("license_model") {
		req.LicenseModel = aws.String(d.Get("license_model").(string))
		requestUpdate = true
	}
	if d.HasChange("deletion_protection") {
		req.DeletionProtection = aws.Bool(d.Get("deletion_protection").(bool))
		requestUpdate = true
	}
	if d.HasChange("instance_class") {
		req.DBInstanceClass = aws.String(d.Get("instance_class").(string))
		requestUpdate = true
	}
	if d.HasChange("parameter_group_name") {
		req.DBParameterGroupName = aws.String(d.Get("parameter_group_name").(string))
		requestUpdate = true
	}
	if d.HasChange("engine_version") {
		req.EngineVersion = aws.String(d.Get("engine_version").(string))
		req.AllowMajorVersionUpgrade = aws.Bool(d.Get("allow_major_version_upgrade").(bool))
		requestUpdate = true
	}
	if d.HasChange("backup_window") {
		req.PreferredBackupWindow = aws.String(d.Get("backup_window").(string))
		requestUpdate = true
	}
	if d.HasChange("maintenance_window") {
		req.PreferredMaintenanceWindow = aws.String(d.Get("maintenance_window").(string))
		requestUpdate = true
	}
	if d.HasChange("max_allocated_storage") {
		mas := d.Get("max_allocated_storage").(int)

		// The API expects the max allocated storage value to be set to the allocated storage
		// value when disabling autoscaling. This check ensures that value is set correctly
		// if the update to the Terraform configuration was removing the argument completely.
		if mas == 0 {
			mas = d.Get("allocated_storage").(int)
		}

		req.MaxAllocatedStorage = aws.Int64(int64(mas))
		requestUpdate = true
	}
	if d.HasChange("password") {
		req.MasterUserPassword = aws.String(d.Get("password").(string))
		requestUpdate = true
	}
	if d.HasChange("multi_az") {
		req.MultiAZ = aws.Bool(d.Get("multi_az").(bool))
		requestUpdate = true
	}
	if d.HasChange("publicly_accessible") {
		req.PubliclyAccessible = aws.Bool(d.Get("publicly_accessible").(bool))
		requestUpdate = true
	}
	if d.HasChange("storage_type") {
		req.StorageType = aws.String(d.Get("storage_type").(string))
		requestUpdate = true

		if aws.StringValue(req.StorageType) == storageTypeIO1 {
			req.Iops = aws.Int64(int64(d.Get("iops").(int)))
		}
	}
	if d.HasChange("auto_minor_version_upgrade") {
		req.AutoMinorVersionUpgrade = aws.Bool(d.Get("auto_minor_version_upgrade").(bool))
		requestUpdate = true
	}

	if d.HasChange("monitoring_role_arn") {
		req.MonitoringRoleArn = aws.String(d.Get("monitoring_role_arn").(string))
		requestUpdate = true
	}

	if d.HasChange("monitoring_interval") {
		req.MonitoringInterval = aws.Int64(int64(d.Get("monitoring_interval").(int)))
		requestUpdate = true
	}

	if d.HasChange("vpc_security_group_ids") {
		if attr := d.Get("vpc_security_group_ids").(*schema.Set); attr.Len() > 0 {
			req.VpcSecurityGroupIds = flex.ExpandStringSet(attr)
		}
		requestUpdate = true
	}

	if d.HasChange("security_group_names") {
		if attr := d.Get("security_group_names").(*schema.Set); attr.Len() > 0 {
			req.DBSecurityGroups = flex.ExpandStringSet(attr)
		}
		requestUpdate = true
	}

	if d.HasChange("option_group_name") {
		req.OptionGroupName = aws.String(d.Get("option_group_name").(string))
		requestUpdate = true
	}

	if d.HasChange("port") {
		req.DBPortNumber = aws.Int64(int64(d.Get("port").(int)))
		requestUpdate = true
	}
	if d.HasChange("db_subnet_group_name") {
		req.DBSubnetGroupName = aws.String(d.Get("db_subnet_group_name").(string))
		requestUpdate = true
	}

	if d.HasChange("enabled_cloudwatch_logs_exports") {
		oraw, nraw := d.GetChange("enabled_cloudwatch_logs_exports")
		o := oraw.(*schema.Set)
		n := nraw.(*schema.Set)

		enable := n.Difference(o)
		disable := o.Difference(n)

		req.CloudwatchLogsExportConfiguration = &rds.CloudwatchLogsExportConfiguration{
			EnableLogTypes:  flex.ExpandStringSet(enable),
			DisableLogTypes: flex.ExpandStringSet(disable),
		}
		requestUpdate = true
	}

	if d.HasChange("iam_database_authentication_enabled") {
		req.EnableIAMDatabaseAuthentication = aws.Bool(d.Get("iam_database_authentication_enabled").(bool))
		requestUpdate = true
	}

	if d.HasChanges("domain", "domain_iam_role_name") {
		req.Domain = aws.String(d.Get("domain").(string))
		req.DomainIAMRoleName = aws.String(d.Get("domain_iam_role_name").(string))
		requestUpdate = true
	}

	if d.HasChanges("performance_insights_enabled", "performance_insights_kms_key_id", "performance_insights_retention_period") {
		req.EnablePerformanceInsights = aws.Bool(d.Get("performance_insights_enabled").(bool))

		if v, ok := d.GetOk("performance_insights_kms_key_id"); ok {
			req.PerformanceInsightsKMSKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("performance_insights_retention_period"); ok {
			req.PerformanceInsightsRetentionPeriod = aws.Int64(int64(v.(int)))
		}

		requestUpdate = true
	}

	if d.HasChange("customer_owned_ip_enabled") {
		req.EnableCustomerOwnedIp = aws.Bool(d.Get("customer_owned_ip_enabled").(bool))
		requestUpdate = true
	}

	if d.HasChange("replica_mode") {
		req.ReplicaMode = aws.String(d.Get("replica_mode").(string))
		requestUpdate = true
	}

	log.Printf("[DEBUG] Send DB Instance Modification request: %t", requestUpdate)
	if requestUpdate {
		log.Printf("[DEBUG] DB Instance Modification request: %s", req)

		err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := conn.ModifyDBInstance(req)

			// Retry for IAM eventual consistency
			if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
				return resource.RetryableError(err)
			}

			// InvalidDBInstanceState: RDS is configuring Enhanced Monitoring or Performance Insights for this DB instance. Try your request later.
			if tfawserr.ErrMessageContains(err, rds.ErrCodeInvalidDBInstanceStateFault, "your request later") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.ModifyDBInstance(req)
		}

		if err != nil {
			return fmt.Errorf("modifying DB Instance %s: %w", d.Id(), err)
		}

		log.Printf("[DEBUG] Waiting for DB Instance (%s) to be available", d.Id())
		err = waitUntilDBInstanceAvailableAfterUpdate(d.Id(), conn, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for DB Instance (%s) to be available: %w", d.Id(), err)
		}
	}

	// separate request to promote a database
	if d.HasChange("replicate_source_db") {
		if d.Get("replicate_source_db").(string) == "" {
			// promote
			opts := rds.PromoteReadReplicaInput{
				DBInstanceIdentifier: aws.String(d.Id()),
			}
			attr := d.Get("backup_retention_period")
			opts.BackupRetentionPeriod = aws.Int64(int64(attr.(int)))
			if attr, ok := d.GetOk("backup_window"); ok {
				opts.PreferredBackupWindow = aws.String(attr.(string))
			}
			_, err := conn.PromoteReadReplica(&opts)
			if err != nil {
				return fmt.Errorf("Error promoting database: %w", err)
			}
			d.Set("replicate_source_db", "")
		} else {
			return fmt.Errorf("cannot elect new source database for replication")
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating RDS DB Instance (%s) tags: %w", d.Get("arn").(string), err)
		}

	}

	return resourceInstanceRead(d, meta)
}

// resourceDBInstanceRetrieve fetches DBInstance information from the AWS
// API. It returns an error if there is a communication problem or unexpected
// error with AWS. When the DBInstance is not found, it returns no error and a
// nil pointer.
func resourceDBInstanceRetrieve(id string, conn *rds.RDS) (*rds.DBInstance, error) {
	opts := rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(id),
	}

	log.Printf("[DEBUG] DB Instance describe configuration: %#v", opts)

	resp, err := conn.DescribeDBInstances(&opts)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceNotFoundFault) {
			return nil, nil
		}
		return nil, fmt.Errorf("Error retrieving DB Instances: %w", err)
	}

	if len(resp.DBInstances) != 1 || resp.DBInstances[0] == nil || aws.StringValue(resp.DBInstances[0].DBInstanceIdentifier) != id {
		return nil, nil
	}

	return resp.DBInstances[0], nil
}

func resourceInstanceImport(
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Neither skip_final_snapshot nor final_snapshot_identifier can be fetched
	// from any API call, so we need to default skip_final_snapshot to true so
	// that final_snapshot_identifier is not required
	d.Set("skip_final_snapshot", true)
	d.Set("delete_automated_backups", true)
	return []*schema.ResourceData{d}, nil
}

func resourceDBInstanceStateRefreshFunc(id string, conn *rds.RDS) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		v, err := resourceDBInstanceRetrieve(id, conn)

		if err != nil {
			return nil, "", err
		}

		if v == nil {
			return nil, "", nil
		}

		return v, aws.StringValue(v.DBInstanceStatus), nil
	}
}

// Database instance status: http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Status.html
var resourceInstanceCreatePendingStates = []string{
	"backing-up",
	"configuring-enhanced-monitoring",
	"configuring-iam-database-auth",
	"configuring-log-exports",
	"creating",
	"maintenance",
	"modifying",
	"rebooting",
	"renaming",
	"resetting-master-credentials",
	"starting",
	"stopping",
	"upgrading",
}

var resourceInstanceUpdatePendingStates = []string{
	"backing-up",
	"configuring-enhanced-monitoring",
	"configuring-iam-database-auth",
	"configuring-log-exports",
	"creating",
	"maintenance",
	"modifying",
	"moving-to-vpc",
	"rebooting",
	"renaming",
	"resetting-master-credentials",
	"starting",
	"stopping",
	"storage-full",
	"upgrading",
}

func expandRestoreToPointInTime(l []interface{}) *rds.RestoreDBInstanceToPointInTimeInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	input := &rds.RestoreDBInstanceToPointInTimeInput{}

	if v, ok := tfMap["restore_time"].(string); ok && v != "" {
		parsedTime, err := time.Parse(time.RFC3339, v)
		if err == nil {
			input.RestoreTime = aws.Time(parsedTime)
		}
	}

	if v, ok := tfMap["source_db_instance_identifier"].(string); ok && v != "" {
		input.SourceDBInstanceIdentifier = aws.String(v)
	}

	if v, ok := tfMap["source_db_instance_automated_backups_arn"].(string); ok && v != "" {
		input.SourceDBInstanceAutomatedBackupsArn = aws.String(v)
	}

	if v, ok := tfMap["source_dbi_resource_id"].(string); ok && v != "" {
		input.SourceDbiResourceId = aws.String(v)
	}

	if v, ok := tfMap["use_latest_restorable_time"].(bool); ok && v {
		input.UseLatestRestorableTime = aws.Bool(v)
	}

	return input
}

func dbSetResourceDataEngineVersionFromInstance(d *schema.ResourceData, c *rds.DBInstance) {
	oldVersion := d.Get("engine_version").(string)
	newVersion := aws.StringValue(c.EngineVersion)
	compareActualEngineVersion(d, oldVersion, newVersion)
}

func flattenDBSecurityGroups(groups []*rds.DBSecurityGroupMembership) *schema.Set {
	result := &schema.Set{
		F: schema.HashString,
	}
	for _, v := range groups {
		result.Add(aws.StringValue(v.DBSecurityGroupName))
	}
	return result
}
