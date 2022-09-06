package rds

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	clusterScalingConfiguration_DefaultMinCapacity = 1
	clusterScalingConfiguration_DefaultMaxCapacity = 16
	clusterTimeoutDelete                           = 2 * time.Minute
)

func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceClusterCreate,
		Read:   resourceClusterRead,
		Update: resourceClusterUpdate,
		Delete: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			State: resourceClusterImport,
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
				Optional:     true,
				ForceNew:     true,
				Default:      ClusterEngineAurora,
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
			"master_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
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
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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

func resourceClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

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
		identifier = resource.PrefixedUniqueId(v.(string))
	} else {
		identifier = resource.PrefixedUniqueId("tf-")
	}

	if _, ok := d.GetOk("snapshot_identifier"); ok {
		opts := rds.RestoreDBClusterFromSnapshotInput{
			CopyTagsToSnapshot:   aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBClusterIdentifier:  aws.String(identifier),
			DeletionProtection:   aws.Bool(d.Get("deletion_protection").(bool)),
			Engine:               aws.String(d.Get("engine").(string)),
			EngineMode:           aws.String(d.Get("engine_mode").(string)),
			ScalingConfiguration: ExpandClusterScalingConfiguration(d.Get("scaling_configuration").([]interface{})),
			SnapshotIdentifier:   aws.String(d.Get("snapshot_identifier").(string)),
			Tags:                 Tags(tags.IgnoreAWS()),
		}

		if attr := d.Get("availability_zones").(*schema.Set); attr.Len() > 0 {
			opts.AvailabilityZones = flex.ExpandStringSet(attr)
		}

		if v, ok := d.GetOk("backtrack_window"); ok {
			opts.BacktrackWindow = aws.Int64(int64(v.(int)))
		}

		if attr, ok := d.GetOk("backup_retention_period"); ok {
			modifyDbClusterInput.BackupRetentionPeriod = aws.Int64(int64(attr.(int)))
			requiresModifyDbCluster = true
		}

		if attr, ok := d.GetOk("database_name"); ok {
			opts.DatabaseName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			opts.DBClusterParameterGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("db_subnet_group_name"); ok {
			opts.DBSubnetGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && attr.(*schema.Set).Len() > 0 {
			opts.EnableCloudwatchLogsExports = flex.ExpandStringSet(attr.(*schema.Set))
		}

		if attr, ok := d.GetOk("engine_version"); ok {
			opts.EngineVersion = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("kms_key_id"); ok {
			opts.KmsKeyId = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("master_password"); ok {
			modifyDbClusterInput.MasterUserPassword = aws.String(attr.(string))
			requiresModifyDbCluster = true
		}

		if attr, ok := d.GetOk("network_type"); ok {
			opts.NetworkType = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("option_group_name"); ok {
			opts.OptionGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("port"); ok {
			opts.Port = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("preferred_backup_window"); ok {
			modifyDbClusterInput.PreferredBackupWindow = aws.String(attr.(string))
			requiresModifyDbCluster = true
		}

		if attr, ok := d.GetOk("preferred_maintenance_window"); ok {
			modifyDbClusterInput.PreferredMaintenanceWindow = aws.String(attr.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("serverlessv2_scaling_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			modifyDbClusterInput.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}

		if attr := d.Get("vpc_security_group_ids").(*schema.Set); attr.Len() > 0 {
			opts.VpcSecurityGroupIds = flex.ExpandStringSet(attr)
		}

		log.Printf("[DEBUG] RDS Cluster restore from snapshot configuration: %s", opts)
		err := resource.Retry(propagationTimeout, func() *resource.RetryError {
			_, err := conn.RestoreDBClusterFromSnapshot(&opts)
			if err != nil {
				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if tfresource.TimedOut(err) {
			_, err = conn.RestoreDBClusterFromSnapshot(&opts)
		}
		if err != nil {
			return fmt.Errorf("Error creating RDS Cluster: %s", err)
		}
	} else if v, ok := d.GetOk("s3_import"); ok {
		if _, ok := d.GetOk("master_password"); !ok {
			return fmt.Errorf(`provider.aws: aws_db_instance: %s: "master_password": required field is not set`, d.Get("name").(string))
		}
		if _, ok := d.GetOk("master_username"); !ok {
			return fmt.Errorf(`provider.aws: aws_db_instance: %s: "master_username": required field is not set`, d.Get("name").(string))
		}
		s3_bucket := v.([]interface{})[0].(map[string]interface{})
		createOpts := &rds.RestoreDBClusterFromS3Input{
			CopyTagsToSnapshot:  aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBClusterIdentifier: aws.String(identifier),
			DeletionProtection:  aws.Bool(d.Get("deletion_protection").(bool)),
			Engine:              aws.String(d.Get("engine").(string)),
			MasterUsername:      aws.String(d.Get("master_username").(string)),
			MasterUserPassword:  aws.String(d.Get("master_password").(string)),
			S3BucketName:        aws.String(s3_bucket["bucket_name"].(string)),
			S3IngestionRoleArn:  aws.String(s3_bucket["ingestion_role"].(string)),
			S3Prefix:            aws.String(s3_bucket["bucket_prefix"].(string)),
			SourceEngine:        aws.String(s3_bucket["source_engine"].(string)),
			SourceEngineVersion: aws.String(s3_bucket["source_engine_version"].(string)),
			Tags:                Tags(tags.IgnoreAWS()),
		}

		if v, ok := d.GetOk("backtrack_window"); ok {
			createOpts.BacktrackWindow = aws.Int64(int64(v.(int)))
		}

		if v := d.Get("database_name"); v.(string) != "" {
			createOpts.DatabaseName = aws.String(v.(string))
		}

		if attr, ok := d.GetOk("port"); ok {
			createOpts.Port = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("db_subnet_group_name"); ok {
			createOpts.DBSubnetGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			createOpts.DBClusterParameterGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("engine_version"); ok {
			createOpts.EngineVersion = aws.String(attr.(string))
		}

		if attr := d.Get("vpc_security_group_ids").(*schema.Set); attr.Len() > 0 {
			createOpts.VpcSecurityGroupIds = flex.ExpandStringSet(attr)
		}

		if attr := d.Get("availability_zones").(*schema.Set); attr.Len() > 0 {
			createOpts.AvailabilityZones = flex.ExpandStringSet(attr)
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			createOpts.BackupRetentionPeriod = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("preferred_backup_window"); ok {
			createOpts.PreferredBackupWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk("preferred_maintenance_window"); ok {
			createOpts.PreferredMaintenanceWindow = aws.String(v.(string))
		}

		if attr, ok := d.GetOk("kms_key_id"); ok {
			createOpts.KmsKeyId = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("network_type"); ok {
			createOpts.NetworkType = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			createOpts.EnableIAMDatabaseAuthentication = aws.Bool(attr.(bool))
		}

		if attr, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && attr.(*schema.Set).Len() > 0 {
			createOpts.EnableCloudwatchLogsExports = flex.ExpandStringSet(attr.(*schema.Set))
		}

		if attr, ok := d.GetOkExists("storage_encrypted"); ok {
			createOpts.StorageEncrypted = aws.Bool(attr.(bool))
		}

		log.Printf("[DEBUG] RDS Cluster restore options: %s", createOpts)
		// Retry for IAM/S3 eventual consistency
		var resp *rds.RestoreDBClusterFromS3Output
		err := resource.Retry(5*time.Minute, func() *resource.RetryError {
			var err error
			resp, err = conn.RestoreDBClusterFromS3(createOpts)
			if err != nil {
				// InvalidParameterValue: Files from the specified Amazon S3 bucket cannot be downloaded.
				// Make sure that you have created an AWS Identity and Access Management (IAM) role that lets Amazon RDS access Amazon S3 for you.
				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Files from the specified Amazon S3 bucket cannot be downloaded") {
					return resource.RetryableError(err)
				}
				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "S3_SNAPSHOT_INGESTION") {
					return resource.RetryableError(err)
				}
				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "S3 bucket cannot be found") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			log.Printf("[DEBUG]: RDS Cluster create response: %s", resp)
			return nil
		})
		if tfresource.TimedOut(err) {
			resp, err = conn.RestoreDBClusterFromS3(createOpts)
		}

		if err != nil {
			log.Printf("[ERROR] Error creating RDS Cluster: %s", err)
			return err
		}

	} else if v, ok := d.GetOk("restore_to_point_in_time"); ok {
		pointInTime := v.([]interface{})[0].(map[string]interface{})
		createOpts := &rds.RestoreDBClusterToPointInTimeInput{
			DBClusterIdentifier:       aws.String(identifier),
			DeletionProtection:        aws.Bool(d.Get("deletion_protection").(bool)),
			SourceDBClusterIdentifier: aws.String(pointInTime["source_cluster_identifier"].(string)),
			Tags:                      Tags(tags.IgnoreAWS()),
		}

		if v, ok := pointInTime["restore_to_time"].(string); ok && v != "" {
			restoreToTime, _ := time.Parse(time.RFC3339, v)
			createOpts.RestoreToTime = aws.Time(restoreToTime)
		}

		if v, ok := pointInTime["use_latest_restorable_time"].(bool); ok && v {
			createOpts.UseLatestRestorableTime = aws.Bool(v)
		}

		if createOpts.RestoreToTime == nil && createOpts.UseLatestRestorableTime == nil {
			return fmt.Errorf(`provider.aws: aws_rds_cluster: %s: Either "restore_to_time" or "use_latest_restorable_time" must be set`, d.Get("database_name").(string))
		}

		if attr, ok := pointInTime["restore_type"].(string); ok {
			createOpts.RestoreType = aws.String(attr)
		}

		if v, ok := d.GetOk("backtrack_window"); ok {
			createOpts.BacktrackWindow = aws.Int64(int64(v.(int)))
		}

		if attr, ok := d.GetOk("db_subnet_group_name"); ok {
			createOpts.DBSubnetGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("port"); ok {
			createOpts.Port = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("option_group_name"); ok {
			createOpts.OptionGroupName = aws.String(attr.(string))
		}

		if attr := d.Get("vpc_security_group_ids").(*schema.Set); attr.Len() > 0 {
			createOpts.VpcSecurityGroupIds = flex.ExpandStringSet(attr)
		}

		if attr, ok := d.GetOk("kms_key_id"); ok {
			createOpts.KmsKeyId = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("network_type"); ok {
			createOpts.NetworkType = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && attr.(*schema.Set).Len() > 0 {
			createOpts.EnableCloudwatchLogsExports = flex.ExpandStringSet(attr.(*schema.Set))
		}

		if attr, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			createOpts.EnableIAMDatabaseAuthentication = aws.Bool(attr.(bool))
		}

		if attr, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			createOpts.DBClusterParameterGroupName = aws.String(attr.(string))
		}

		requireUpdateAttrs := []string{
			"master_password",
			"backup_retention_period",
			"preferred_backup_window",
			"preferred_maintenance_window",
			"scaling_configuration",
		}

		for _, attr := range requireUpdateAttrs {
			if val, ok := d.GetOk(attr); ok {
				requiresModifyDbCluster = true
				switch attr {
				case "master_password":
					modifyDbClusterInput.MasterUserPassword = aws.String(val.(string))
				case "backup_retention_period":
					modifyDbClusterInput.BackupRetentionPeriod = aws.Int64(int64(val.(int)))
				case "preferred_backup_window":
					modifyDbClusterInput.PreferredBackupWindow = aws.String(val.(string))
				case "preferred_maintenance_window":
					modifyDbClusterInput.PreferredMaintenanceWindow = aws.String(val.(string))
				case "scaling_configuration":
					modifyDbClusterInput.ScalingConfiguration = ExpandClusterScalingConfiguration(d.Get("scaling_configuration").([]interface{}))
				case "serverlessv2_scaling_configuration":
					if len(val.([]interface{})) > 0 && val.([]interface{})[0] != nil {
						modifyDbClusterInput.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]interface{})[0].(map[string]interface{}))
					}
				}
			}
		}

		log.Printf("[DEBUG] RDS Cluster restore options: %s", createOpts)

		resp, err := conn.RestoreDBClusterToPointInTime(createOpts)
		if err != nil {
			log.Printf("[ERROR] Error restoring RDS Cluster: %s", err)
			return err
		}

		log.Printf("[DEBUG]: RDS Cluster restore response: %s", resp)
	} else {

		createOpts := &rds.CreateDBClusterInput{
			CopyTagsToSnapshot:   aws.Bool(d.Get("copy_tags_to_snapshot").(bool)),
			DBClusterIdentifier:  aws.String(identifier),
			DeletionProtection:   aws.Bool(d.Get("deletion_protection").(bool)),
			Engine:               aws.String(d.Get("engine").(string)),
			EngineMode:           aws.String(d.Get("engine_mode").(string)),
			ScalingConfiguration: ExpandClusterScalingConfiguration(d.Get("scaling_configuration").([]interface{})),
			Tags:                 Tags(tags.IgnoreAWS()),
		}

		// Note: Username and password credentials are required and valid
		// unless the cluster is a read-replica. This also applies to clusters
		// within a global cluster. Providing a password and/or username for
		// a replica will result in an InvalidParameterValue error.
		if v, ok := d.GetOk("master_password"); ok {
			createOpts.MasterUserPassword = aws.String(v.(string))
		}

		if v, ok := d.GetOk("master_username"); ok {
			createOpts.MasterUsername = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enable_http_endpoint"); ok {
			createOpts.EnableHttpEndpoint = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("backtrack_window"); ok {
			createOpts.BacktrackWindow = aws.Int64(int64(v.(int)))
		}

		if v := d.Get("database_name"); v.(string) != "" {
			createOpts.DatabaseName = aws.String(v.(string))
		}

		if attr, ok := d.GetOk("port"); ok {
			createOpts.Port = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOk("db_subnet_group_name"); ok {
			createOpts.DBSubnetGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			createOpts.DBClusterParameterGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("db_cluster_instance_class"); ok {
			createOpts.DBClusterInstanceClass = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("engine_version"); ok {
			createOpts.EngineVersion = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("global_cluster_identifier"); ok {
			createOpts.GlobalClusterIdentifier = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("enable_global_write_forwarding"); ok {
			createOpts.EnableGlobalWriteForwarding = aws.Bool(attr.(bool))
		}

		if attr := d.Get("vpc_security_group_ids").(*schema.Set); attr.Len() > 0 {
			createOpts.VpcSecurityGroupIds = flex.ExpandStringSet(attr)
		}

		if attr := d.Get("availability_zones").(*schema.Set); attr.Len() > 0 {
			createOpts.AvailabilityZones = flex.ExpandStringSet(attr)
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			createOpts.BackupRetentionPeriod = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("preferred_backup_window"); ok {
			createOpts.PreferredBackupWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk("preferred_maintenance_window"); ok {
			createOpts.PreferredMaintenanceWindow = aws.String(v.(string))
		}

		if attr, ok := d.GetOk("network_type"); ok {
			createOpts.NetworkType = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("kms_key_id"); ok {
			createOpts.KmsKeyId = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("source_region"); ok {
			createOpts.SourceRegion = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			createOpts.EnableIAMDatabaseAuthentication = aws.Bool(attr.(bool))
		}

		if attr, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && attr.(*schema.Set).Len() > 0 {
			createOpts.EnableCloudwatchLogsExports = flex.ExpandStringSet(attr.(*schema.Set))
		}

		if attr, ok := d.GetOk("replication_source_identifier"); ok && createOpts.GlobalClusterIdentifier == nil {
			createOpts.ReplicationSourceIdentifier = aws.String(attr.(string))
		}

		if attr, ok := d.GetOkExists("allocated_storage"); ok {
			createOpts.AllocatedStorage = aws.Int64(int64(attr.(int)))
		}

		if attr, ok := d.GetOkExists("storage_type"); ok {
			createOpts.StorageType = aws.String(attr.(string))
		}

		if attr, ok := d.GetOkExists("iops"); ok {
			createOpts.Iops = aws.Int64(int64(attr.(int)))
		}

		if v, ok := d.GetOk("serverlessv2_scaling_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			createOpts.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}

		if attr, ok := d.GetOkExists("storage_encrypted"); ok {
			createOpts.StorageEncrypted = aws.Bool(attr.(bool))
		}

		log.Printf("[DEBUG] RDS Cluster create options: %s", createOpts)
		var resp *rds.CreateDBClusterOutput
		err := resource.Retry(propagationTimeout, func() *resource.RetryError {
			var err error
			resp, err = conn.CreateDBCluster(createOpts)
			if err != nil {
				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if tfresource.TimedOut(err) {
			resp, err = conn.CreateDBCluster(createOpts)
		}
		if err != nil {
			return fmt.Errorf("error creating RDS cluster: %s", err)
		}

		log.Printf("[DEBUG]: RDS Cluster create response: %s", resp)
	}

	d.SetId(identifier)

	log.Printf("[INFO] RDS Cluster ID: %s", d.Id())

	log.Println("[INFO] Waiting for RDS Cluster to be available")

	stateConf := &resource.StateChangeConf{
		Pending:    resourceClusterCreatePendingStates,
		Target:     []string{"available"},
		Refresh:    resourceClusterStateRefreshFunc(conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for RDS Cluster state to be \"available\": %s", err)
	}

	if v, ok := d.GetOk("iam_roles"); ok {
		for _, role := range v.(*schema.Set).List() {
			err := setIAMRoleToCluster(d.Id(), role.(string), conn)
			if err != nil {
				return err
			}
		}
	}

	if requiresModifyDbCluster {
		modifyDbClusterInput.DBClusterIdentifier = aws.String(d.Id())

		log.Printf("[INFO] RDS Cluster (%s) configuration requires ModifyDBCluster: %s", d.Id(), modifyDbClusterInput)
		_, err := conn.ModifyDBCluster(modifyDbClusterInput)
		if err != nil {
			return fmt.Errorf("error modifying RDS Cluster (%s): %s", d.Id(), err)
		}

		log.Printf("[INFO] Waiting for RDS Cluster (%s) to be available", d.Id())
		err = waitForClusterUpdate(conn, d.Id(), d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return fmt.Errorf("error waiting for RDS Cluster (%s) to be available: %s", d.Id(), err)
		}
	}

	return resourceClusterRead(d, meta)
}

func resourceClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	dbc, err := FindDBClusterByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading RDS Cluster (%s): %w", d.Id(), err)
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
	if dbc.DatabaseName != nil {
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
	d.Set("master_username", dbc.MasterUsername)
	d.Set("network_type", dbc.NetworkType)
	d.Set("port", dbc.Port)
	d.Set("preferred_backup_window", dbc.PreferredBackupWindow)
	d.Set("preferred_maintenance_window", dbc.PreferredMaintenanceWindow)
	d.Set("reader_endpoint", dbc.ReaderEndpoint)
	d.Set("replication_source_identifier", dbc.ReplicationSourceIdentifier)
	if dbc.ScalingConfigurationInfo != nil {
		if err := d.Set("scaling_configuration", []interface{}{flattenScalingConfigurationInfo(dbc.ScalingConfigurationInfo)}); err != nil {
			return fmt.Errorf("setting scaling_configuration: %w", err)
		}
	} else {
		d.Set("scaling_configuration", nil)
	}
	if dbc.ServerlessV2ScalingConfiguration != nil {
		if err := d.Set("serverlessv2_scaling_configuration", []interface{}{flattenServerlessV2ScalingConfigurationInfo(dbc.ServerlessV2ScalingConfiguration)}); err != nil {
			return fmt.Errorf("setting serverlessv2_scaling_configuration: %w", err)
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

	tags, err := ListTags(conn, clusterARN)

	if err != nil {
		return fmt.Errorf("listing tags for RDS Cluster (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	// Fetch and save Global Cluster if engine mode global
	d.Set("global_cluster_identifier", "")

	if aws.StringValue(dbc.EngineMode) == EngineModeGlobal || aws.StringValue(dbc.EngineMode) == EngineModeProvisioned {
		globalCluster, err := FindGlobalClusterByDBClusterARN(conn, aws.StringValue(dbc.DBClusterArn))

		if err == nil {
			d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
		} else if tfresource.NotFound(err) || tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "Access Denied to API Version: APIGlobalDatabases") {
			// Ignore the following API error for regions/partitions that do not support RDS Global Clusters:
			// InvalidParameterValue: Access Denied to API Version: APIGlobalDatabases
		} else {
			return fmt.Errorf("reading RDS Global Cluster for RDS Cluster (%s): %w", d.Id(), err)
		}
	}

	return nil
}

func resourceClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn
	requestUpdate := false

	req := &rds.ModifyDBClusterInput{
		ApplyImmediately:    aws.Bool(d.Get("apply_immediately").(bool)),
		DBClusterIdentifier: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("allow_major_version_upgrade"); ok {
		req.AllowMajorVersionUpgrade = aws.Bool(v.(bool))
	}

	if d.HasChange("backtrack_window") {
		req.BacktrackWindow = aws.Int64(int64(d.Get("backtrack_window").(int)))
		requestUpdate = true
	}

	if d.HasChange("copy_tags_to_snapshot") {
		req.CopyTagsToSnapshot = aws.Bool(d.Get("copy_tags_to_snapshot").(bool))
		requestUpdate = true
	}

	if d.HasChange("db_instance_parameter_group_name") {
		req.DBInstanceParameterGroupName = aws.String(d.Get("db_instance_parameter_group_name").(string))
		requestUpdate = true
	}

	if d.HasChange("master_password") {
		req.MasterUserPassword = aws.String(d.Get("master_password").(string))
		requestUpdate = true
	}

	if d.HasChange("db_cluster_instance_class") {
		req.EngineVersion = aws.String(d.Get("db_cluster_instance_class").(string))
		requestUpdate = true
	}

	if d.HasChange("engine_version") {
		req.EngineVersion = aws.String(d.Get("engine_version").(string))
		requestUpdate = true
	}

	if d.HasChange("vpc_security_group_ids") {
		if attr := d.Get("vpc_security_group_ids").(*schema.Set); attr.Len() > 0 {
			req.VpcSecurityGroupIds = flex.ExpandStringSet(attr)
		} else {
			req.VpcSecurityGroupIds = []*string{}
		}
		requestUpdate = true
	}

	if d.HasChange("port") {
		req.Port = aws.Int64(int64(d.Get("port").(int)))
		requestUpdate = true
	}

	if d.HasChange("storage_type") {
		req.StorageType = aws.String(d.Get("storage_type").(string))
		requestUpdate = true
	}

	if d.HasChange("allocated_storage") {
		req.AllocatedStorage = aws.Int64(int64(d.Get("allocated_storage").(int)))
		requestUpdate = true
	}

	if d.HasChange("iops") {
		req.Iops = aws.Int64(int64(d.Get("iops").(int)))
		requestUpdate = true
	}

	if d.HasChange("preferred_backup_window") {
		req.PreferredBackupWindow = aws.String(d.Get("preferred_backup_window").(string))
		requestUpdate = true
	}

	if d.HasChange("preferred_maintenance_window") {
		req.PreferredMaintenanceWindow = aws.String(d.Get("preferred_maintenance_window").(string))
		requestUpdate = true
	}

	if d.HasChange("backup_retention_period") {
		req.BackupRetentionPeriod = aws.Int64(int64(d.Get("backup_retention_period").(int)))
		requestUpdate = true
	}

	if d.HasChange("db_cluster_parameter_group_name") {
		req.DBClusterParameterGroupName = aws.String(d.Get("db_cluster_parameter_group_name").(string))
		requestUpdate = true
	}

	if d.HasChange("deletion_protection") {
		req.DeletionProtection = aws.Bool(d.Get("deletion_protection").(bool))
		requestUpdate = true
	}

	if d.HasChange("iam_database_authentication_enabled") {
		req.EnableIAMDatabaseAuthentication = aws.Bool(d.Get("iam_database_authentication_enabled").(bool))
		requestUpdate = true
	}

	if d.HasChange("network_type") {
		req.NetworkType = aws.String(d.Get("network_type").(string))
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

	if d.HasChange("scaling_configuration") {
		req.ScalingConfiguration = ExpandClusterScalingConfiguration(d.Get("scaling_configuration").([]interface{}))
		requestUpdate = true
	}

	if d.HasChange("serverlessv2_scaling_configuration") {
		if v, ok := d.GetOk("serverlessv2_scaling_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			req.ServerlessV2ScalingConfiguration = expandServerlessV2ScalingConfiguration(v.([]interface{})[0].(map[string]interface{}))
			requestUpdate = true
		}
	}

	if d.HasChange("enable_http_endpoint") {
		req.EnableHttpEndpoint = aws.Bool(d.Get("enable_http_endpoint").(bool))
		requestUpdate = true
	}

	if d.HasChange("enable_global_write_forwarding") {
		req.EnableGlobalWriteForwarding = aws.Bool(d.Get("enable_global_write_forwarding").(bool))
		requestUpdate = true
	}

	if requestUpdate {
		err := resource.Retry(5*time.Minute, func() *resource.RetryError {
			_, err := conn.ModifyDBCluster(req)
			if err != nil {
				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
					return resource.RetryableError(err)
				}

				if tfawserr.ErrMessageContains(err, rds.ErrCodeInvalidDBClusterStateFault, "Cannot modify engine version without a primary instance in DB cluster") {
					return resource.NonRetryableError(err)
				}

				if tfawserr.ErrCodeEquals(err, rds.ErrCodeInvalidDBClusterStateFault) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if tfresource.TimedOut(err) {
			_, err = conn.ModifyDBCluster(req)
		}
		if err != nil {
			return fmt.Errorf("Failed to modify RDS Cluster (%s): %s", d.Id(), err)
		}

		log.Printf("[INFO] Waiting for RDS Cluster (%s) to be available", d.Id())
		err = waitForClusterUpdate(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for RDS Cluster (%s) to be available: %s", d.Id(), err)
		}
	}

	if d.HasChange("global_cluster_identifier") {
		oRaw, nRaw := d.GetChange("global_cluster_identifier")
		o := oRaw.(string)
		n := nRaw.(string)

		if o == "" {
			return errors.New("Existing RDS Clusters cannot be added to an existing RDS Global Cluster")
		}

		if n != "" {
			return errors.New("Existing RDS Clusters cannot be migrated between existing RDS Global Clusters")
		}

		input := &rds.RemoveFromGlobalClusterInput{
			DbClusterIdentifier:     aws.String(d.Get("arn").(string)),
			GlobalClusterIdentifier: aws.String(o),
		}

		log.Printf("[DEBUG] Removing RDS Cluster from RDS Global Cluster: %s", input)
		_, err := conn.RemoveFromGlobalCluster(input)

		if err != nil && !tfawserr.ErrCodeEquals(err, rds.ErrCodeGlobalClusterNotFoundFault) && !tfawserr.ErrMessageContains(err, "InvalidParameterValue", "is not found in global cluster") {
			return fmt.Errorf("error removing RDS Cluster (%s) from RDS Global Cluster: %s", d.Id(), err)
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
		removeRoles := os.Difference(ns)
		enableRoles := ns.Difference(os)

		for _, role := range enableRoles.List() {
			err := setIAMRoleToCluster(d.Id(), role.(string), conn)
			if err != nil {
				return err
			}
		}

		for _, role := range removeRoles.List() {
			err := removeIAMRoleFromCluster(d.Id(), role.(string), conn)
			if err != nil {
				return err
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceClusterRead(d, meta)
}

func resourceClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	// Automatically remove from global cluster to bypass this error on deletion:
	// InvalidDBClusterStateFault: This cluster is a part of a global cluster, please remove it from globalcluster first
	if d.Get("global_cluster_identifier").(string) != "" {
		globalClusterID := d.Get("global_cluster_identifier").(string)
		input := &rds.RemoveFromGlobalClusterInput{
			DbClusterIdentifier:     aws.String(d.Get("arn").(string)),
			GlobalClusterIdentifier: aws.String(globalClusterID),
		}

		log.Printf("[DEBUG] Removing RDS Cluster from RDS Global Cluster: %s", input)
		_, err := conn.RemoveFromGlobalCluster(input)

		if err != nil && !tfawserr.ErrCodeEquals(err, rds.ErrCodeGlobalClusterNotFoundFault) && !tfawserr.ErrMessageContains(err, "InvalidParameterValue", "is not found in global cluster") {
			return fmt.Errorf("removing RDS Cluster (%s) from RDS Global Cluster (%s): %w", d.Id(), globalClusterID, err)
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
			return fmt.Errorf("RDS Cluster final_snapshot_identifier is required when skip_final_snapshot is false")
		}
	}

	log.Printf("[DEBUG] Deleting RDS Cluster: %s", d.Id())
	_, err := tfresource.RetryWhen(clusterTimeoutDelete,
		func() (interface{}, error) {
			return conn.DeleteDBCluster(input)
		},
		func(err error) (bool, error) {
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
		return fmt.Errorf("deleting RDS Cluster (%s): %w", d.Id(), err)
	}

	if err := WaitForClusterDeletion(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for RDS Cluster (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func resourceClusterImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Neither skip_final_snapshot nor final_snapshot_identifier can be fetched
	// from any API call, so we need to default skip_final_snapshot to true so
	// that final_snapshot_identifier is not required
	d.Set("skip_final_snapshot", true)
	return []*schema.ResourceData{d}, nil
}

func resourceClusterStateRefreshFunc(conn *rds.RDS, dbClusterIdentifier string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeDBClusters(&rds.DescribeDBClustersInput{
			DBClusterIdentifier: aws.String(dbClusterIdentifier),
		})

		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterNotFoundFault) {
			return 42, "destroyed", nil
		}

		if err != nil {
			return nil, "", err
		}

		var dbc *rds.DBCluster

		for _, c := range resp.DBClusters {
			if aws.StringValue(c.DBClusterIdentifier) == dbClusterIdentifier {
				dbc = c
			}
		}

		if dbc == nil {
			return 42, "destroyed", nil
		}

		if dbc.Status != nil {
			log.Printf("[DEBUG] DB Cluster status (%s): %s", dbClusterIdentifier, *dbc.Status)
		}

		return dbc, aws.StringValue(dbc.Status), nil
	}
}

func setIAMRoleToCluster(clusterIdentifier string, roleArn string, conn *rds.RDS) error {
	params := &rds.AddRoleToDBClusterInput{
		DBClusterIdentifier: aws.String(clusterIdentifier),
		RoleArn:             aws.String(roleArn),
	}
	_, err := conn.AddRoleToDBCluster(params)
	return err
}

func removeIAMRoleFromCluster(clusterIdentifier string, roleArn string, conn *rds.RDS) error {
	params := &rds.RemoveRoleFromDBClusterInput{
		DBClusterIdentifier: aws.String(clusterIdentifier),
		RoleArn:             aws.String(roleArn),
	}
	_, err := conn.RemoveRoleFromDBCluster(params)
	return err
}

var resourceClusterCreatePendingStates = []string{
	"creating",
	"backing-up",
	"modifying",
	"preparing-data-migration",
	"migrating",
	"resetting-master-credentials",
	"rebooting",
}

var resourceClusterDeletePendingStates = []string{
	"available",
	"deleting",
	"backing-up",
	"modifying",
}

var resourceClusterUpdatePendingStates = []string{
	"backing-up",
	"configuring-iam-database-auth",
	"modifying",
	"renaming",
	"resetting-master-credentials",
	"upgrading",
}

func waitForClusterUpdate(conn *rds.RDS, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    resourceClusterUpdatePendingStates,
		Target:     []string{"available"},
		Refresh:    resourceClusterStateRefreshFunc(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	_, err := stateConf.WaitForState()
	return err
}

func WaitForClusterDeletion(conn *rds.RDS, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    resourceClusterDeletePendingStates,
		Target:     []string{"destroyed"},
		Refresh:    resourceClusterStateRefreshFunc(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func clusterSetResourceDataEngineVersionFromCluster(d *schema.ResourceData, c *rds.DBCluster) {
	oldVersion := d.Get("engine_version").(string)
	newVersion := aws.StringValue(c.EngineVersion)
	compareActualEngineVersion(d, oldVersion, newVersion)
}
