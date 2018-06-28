package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsNeptuneCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNeptuneClusterCreate,
		Read:   resourceAwsNeptuneClusterRead,
		Update: resourceAwsNeptuneClusterUpdate,
		Delete: resourceAwsNeptuneClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"availability_zones": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				ForceNew: true,
				Computed: true,
				Set:      schema.HashString,
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
				ValidateFunc:  validateNeptuneIdentifier,
			},

			"cluster_identifier_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateNeptuneIdentifierPrefix,
			},

			"cluster_members": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
				Set:      schema.HashString,
			},

			"database_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"db_subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"db_cluster_parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"reader_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"engine": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "neptune",
				ForceNew:     true,
				ValidateFunc: validateNeptuneEngine(),
			},

			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"storage_encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
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

			"skip_final_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"master_username": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"master_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			"snapshot_identifier": {
				Type:     schema.TypeString,
				Computed: false,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"port": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			// apply_immediately is used to determine when the update modifications
			// take place.
			"apply_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"vpc_security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"preferred_backup_window": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateOnceADayWindowFormat,
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
				ValidateFunc: validateOnceAWeekWindowFormat,
			},

			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"replication_source_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"iam_roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"iam_database_authentication_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"cluster_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsNeptuneClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn
	tags := tagsFromMapNeptune(d.Get("tags").(map[string]interface{}))

	var identifier string
	if v, ok := d.GetOk("cluster_identifier"); ok {
		identifier = v.(string)
	} else {
		if v, ok := d.GetOk("cluster_identifier_prefix"); ok {
			identifier = resource.PrefixedUniqueId(v.(string))
		} else {
			identifier = resource.PrefixedUniqueId("tf-")
		}

		d.Set("cluster_identifier", identifier)
	}

	if _, ok := d.GetOk("snapshot_identifier"); ok {
		opts := neptune.RestoreDBClusterFromSnapshotInput{
			DBClusterIdentifier: aws.String(d.Get("cluster_identifier").(string)),
			Engine:              aws.String(d.Get("engine").(string)),
			SnapshotIdentifier:  aws.String(d.Get("snapshot_identifier").(string)),
			Tags:                tags,
		}

		if attr, ok := d.GetOk("engine_version"); ok {
			opts.EngineVersion = aws.String(attr.(string))
		}

		if attr := d.Get("availability_zones").(*schema.Set); attr.Len() > 0 {
			opts.AvailabilityZones = expandStringList(attr.List())
		}

		if attr, ok := d.GetOk("db_subnet_group_name"); ok {
			opts.DBSubnetGroupName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("database_name"); ok {
			opts.DatabaseName = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("port"); ok {
			opts.Port = aws.Int64(int64(attr.(int)))
		}

		// Check if any of the parameters that require a cluster modification after creation are set
		var clusterUpdate bool
		if attr := d.Get("vpc_security_group_ids").(*schema.Set); attr.Len() > 0 {
			clusterUpdate = true
			opts.VpcSecurityGroupIds = expandStringList(attr.List())
		}

		if _, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			clusterUpdate = true
		}

		if _, ok := d.GetOk("backup_retention_period"); ok {
			clusterUpdate = true
		}

		log.Printf("[DEBUG] Neptune Cluster restore from snapshot configuration: %s", opts)
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.RestoreDBClusterFromSnapshot(&opts)
			if err != nil {
				if isAWSErr(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("Error creating Neptune Cluster: %s", err)
		}

		if clusterUpdate {
			log.Printf("[INFO] Neptune Cluster is restoring from snapshot with default db_cluster_parameter_group_name, backup_retention_period and vpc_security_group_ids" +
				"but custom values should be set, will now update after snapshot is restored!")

			d.SetId(d.Get("cluster_identifier").(string))

			log.Printf("[INFO] Neptune Cluster ID: %s", d.Id())

			log.Println("[INFO] Waiting for Neptune Cluster to be available")

			stateConf := &resource.StateChangeConf{
				Pending:    resourceAwsNeptuneClusterCreatePendingStates,
				Target:     []string{"available"},
				Refresh:    resourceAwsNeptuneClusterStateRefreshFunc(d, meta),
				Timeout:    d.Timeout(schema.TimeoutCreate),
				MinTimeout: 10 * time.Second,
				Delay:      30 * time.Second,
			}

			// Wait, catching any errors
			_, err := stateConf.WaitForState()
			if err != nil {
				return err
			}

			err = resourceAwsNeptuneClusterUpdate(d, meta)
			if err != nil {
				return err
			}
		}
	} else if _, ok := d.GetOk("replication_source_identifier"); ok {
		createOpts := &neptune.CreateDBClusterInput{
			DBClusterIdentifier:         aws.String(d.Get("cluster_identifier").(string)),
			Engine:                      aws.String(d.Get("engine").(string)),
			StorageEncrypted:            aws.Bool(d.Get("storage_encrypted").(bool)),
			ReplicationSourceIdentifier: aws.String(d.Get("replication_source_identifier").(string)),
			Tags: tags,
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
			createOpts.VpcSecurityGroupIds = expandStringList(attr.List())
		}

		if attr := d.Get("availability_zones").(*schema.Set); attr.Len() > 0 {
			createOpts.AvailabilityZones = expandStringList(attr.List())
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

		if attr, ok := d.GetOk("kms_key_arn"); ok {
			createOpts.KmsKeyId = aws.String(attr.(string))
		}

		log.Printf("[DEBUG] Create Neptune Cluster as read replica: %s", createOpts)
		var resp *neptune.CreateDBClusterOutput
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			var err error
			resp, err = conn.CreateDBCluster(createOpts)
			if err != nil {
				if isAWSErr(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("error creating Neptune cluster: %s", err)
		}

		log.Printf("[DEBUG]: Neptune Cluster create response: %s", resp)

	} else {
		if _, ok := d.GetOk("master_password"); !ok {
			return fmt.Errorf(`provider.aws: aws_neptune_cluster: %s: "master_password": required field is not set`, d.Get("database_name").(string))
		}

		if _, ok := d.GetOk("master_username"); !ok {
			return fmt.Errorf(`provider.aws: aws_neptune_cluster: %s: "master_username": required field is not set`, d.Get("database_name").(string))
		}

		createOpts := &neptune.CreateDBClusterInput{
			DBClusterIdentifier: aws.String(d.Get("cluster_identifier").(string)),
			Engine:              aws.String(d.Get("engine").(string)),
			MasterUserPassword:  aws.String(d.Get("master_password").(string)),
			MasterUsername:      aws.String(d.Get("master_username").(string)),
			StorageEncrypted:    aws.Bool(d.Get("storage_encrypted").(bool)),
			Tags:                tags,
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
			createOpts.VpcSecurityGroupIds = expandStringList(attr.List())
		}

		if attr := d.Get("availability_zones").(*schema.Set); attr.Len() > 0 {
			createOpts.AvailabilityZones = expandStringList(attr.List())
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

		if attr, ok := d.GetOk("kms_key_arn"); ok {
			createOpts.KmsKeyId = aws.String(attr.(string))
		}

		if attr, ok := d.GetOk("iam_database_authentication_enabled"); ok {
			createOpts.EnableIAMDatabaseAuthentication = aws.Bool(attr.(bool))
		}

		log.Printf("[DEBUG] Neptune Cluster create options: %s", createOpts)
		var resp *neptune.CreateDBClusterOutput
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			var err error
			resp, err = conn.CreateDBCluster(createOpts)
			if err != nil {
				if isAWSErr(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("error creating Neptune cluster: %s", err)
		}

		log.Printf("[DEBUG]: Neptune Cluster create response: %s", resp)
	}

	d.SetId(d.Get("cluster_identifier").(string))

	log.Printf("[INFO] Neptune Cluster ID: %s", d.Id())

	log.Println(
		"[INFO] Waiting for Neptune Cluster to be available")

	stateConf := &resource.StateChangeConf{
		Pending:    resourceAwsNeptuneClusterCreatePendingStates,
		Target:     []string{"available"},
		Refresh:    resourceAwsNeptuneClusterStateRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("[WARN] Error waiting for Neptune Cluster state to be \"available\": %s", err)
	}

	if v, ok := d.GetOk("iam_roles"); ok {
		for _, role := range v.(*schema.Set).List() {
			err := setIamRoleToNeptuneCluster(d.Id(), role.(string), conn)
			if err != nil {
				return err
			}
		}
	}

	return resourceAwsNeptuneClusterRead(d, meta)

}

func resourceAwsNeptuneClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn

	resp, err := conn.DescribeDBClusters(&neptune.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, neptune.ErrCodeDBClusterNotFoundFault, "") {
			d.SetId("")
			log.Printf("[DEBUG] Neptune Cluster (%s) not found", d.Id())
			return nil
		}
		log.Printf("[DEBUG] Error describing Neptune Cluster (%s) when waiting: %s", d.Id(), err)
		return err
	}

	var dbc *neptune.DBCluster
	for _, v := range resp.DBClusters {
		if aws.StringValue(v.DBClusterIdentifier) == d.Id() {
			dbc = v
		}
	}

	if dbc == nil {
		log.Printf("[WARN] Neptune Cluster (%s) not found", d.Id())
		d.SetId("")
		return nil
	}

	return flattenAwsNeptuneClusterResource(d, meta, dbc)
}

func flattenAwsNeptuneClusterResource(d *schema.ResourceData, meta interface{}, dbc *neptune.DBCluster) error {
	conn := meta.(*AWSClient).neptuneconn

	if err := d.Set("availability_zones", aws.StringValueSlice(dbc.AvailabilityZones)); err != nil {
		return fmt.Errorf("[DEBUG] Error saving AvailabilityZones to state for Neptune Cluster (%s): %s", d.Id(), err)
	}

	if dbc.DatabaseName != nil {
		d.Set("database_name", dbc.DatabaseName)
	}

	d.Set("cluster_identifier", dbc.DBClusterIdentifier)
	d.Set("cluster_resource_id", dbc.DbClusterResourceId)
	d.Set("db_subnet_group_name", dbc.DBSubnetGroup)
	d.Set("db_cluster_parameter_group_name", dbc.DBClusterParameterGroup)
	d.Set("endpoint", dbc.Endpoint)
	d.Set("engine", dbc.Engine)
	d.Set("engine_version", dbc.EngineVersion)
	d.Set("master_username", dbc.MasterUsername)
	d.Set("port", dbc.Port)
	d.Set("storage_encrypted", dbc.StorageEncrypted)
	d.Set("backup_retention_period", dbc.BackupRetentionPeriod)
	d.Set("preferred_backup_window", dbc.PreferredBackupWindow)
	d.Set("preferred_maintenance_window", dbc.PreferredMaintenanceWindow)
	d.Set("kms_key_arn", dbc.KmsKeyId)
	d.Set("reader_endpoint", dbc.ReaderEndpoint)
	d.Set("replication_source_identifier", dbc.ReplicationSourceIdentifier)
	d.Set("iam_database_authentication_enabled", dbc.IAMDatabaseAuthenticationEnabled)
	d.Set("hosted_zone_id", dbc.HostedZoneId)

	var sg []string
	for _, g := range dbc.VpcSecurityGroups {
		sg = append(sg, aws.StringValue(g.VpcSecurityGroupId))
	}
	if err := d.Set("vpc_security_group_ids", sg); err != nil {
		return fmt.Errorf("[DEBUG] Error saving VPC Security Group IDs to state for Neptune Cluster (%s): %s", d.Id(), err)
	}

	var cm []string
	for _, m := range dbc.DBClusterMembers {
		cm = append(cm, aws.StringValue(m.DBInstanceIdentifier))
	}
	if err := d.Set("cluster_members", cm); err != nil {
		return fmt.Errorf("[DEBUG] Error saving Neptune Cluster Members to state for Neptune Cluster (%s): %s", d.Id(), err)
	}

	var roles []string
	for _, r := range dbc.AssociatedRoles {
		roles = append(roles, aws.StringValue(r.RoleArn))
	}

	if err := d.Set("iam_roles", roles); err != nil {
		return fmt.Errorf("[DEBUG] Error saving IAM Roles to state for Neptune Cluster (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(dbc.DBClusterArn)
	d.Set("arn", arn)

	if err := saveTagsNeptune(conn, d, arn); err != nil {
		log.Printf("[WARN] Failed to save tags for Neptune Cluster (%s): %s", aws.StringValue(dbc.DBClusterIdentifier), err)
	}

	return nil
}

func resourceAwsNeptuneClusterStateRefreshFunc(
	d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		conn := meta.(*AWSClient).neptuneconn

		resp, err := conn.DescribeDBClusters(&neptune.DescribeDBClustersInput{
			DBClusterIdentifier: aws.String(d.Id()),
		})

		if err != nil {
			if isAWSErr(err, neptune.ErrCodeDBClusterNotFoundFault, "") {
				log.Printf("[DEBUG] Neptune Cluster (%s) not found", d.Id())
				return 42, "destroyed", nil
			}
			log.Printf("[DEBUG] Error on retrieving DB Cluster (%s) when waiting: %s", d.Id(), err)
			return nil, "", err
		}

		var dbc *neptune.DBCluster

		for _, v := range resp.DBClusters {
			if aws.StringValue(v.DBClusterIdentifier) == d.Id() {
				dbc = v
			}
		}

		if dbc == nil {
			return 42, "destroyed", nil
		}

		if dbc.Status != nil {
			log.Printf("[DEBUG] Neptune Cluster status (%s): %s", d.Id(), aws.StringValue(dbc.Status))
		}

		return dbc, aws.StringValue(dbc.Status), nil
	}
}

func setIamRoleToNeptuneCluster(clusterIdentifier string, roleArn string, conn *neptune.Neptune) error {
	params := &neptune.AddRoleToDBClusterInput{
		DBClusterIdentifier: aws.String(clusterIdentifier),
		RoleArn:             aws.String(roleArn),
	}
	_, err := conn.AddRoleToDBCluster(params)
	if err != nil {
		return err
	}

	return nil
}

var resourceAwsNeptuneClusterCreatePendingStates = []string{
	"creating",
	"backing-up",
	"modifying",
	"preparing-data-migration",
	"migrating",
	"resetting-master-credentials",
}
