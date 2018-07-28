package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsNeptuneClusterInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNeptuneClusterInstanceCreate,
		Read:   resourceAwsNeptuneClusterInstanceRead,
		Update: resourceAwsNeptuneClusterInstanceUpdate,
		Delete: resourceAwsNeptuneClusterInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{

			"apply_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"identifier": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"identifier_prefix"},
				ValidateFunc:  validateNeptuneIdentifier,
			},

			"identifier_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateNeptuneIdentifier,
			},

			"neptune_subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"writer": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"port": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  8182,
				ForceNew: true,
			},

			"publicly_accessible": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"instance_class": {
				Type:     schema.TypeString,
				Required: true,
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

			"neptune_parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default.neptune1",
			},

			"kms_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"storage_encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"dbi_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"auto_minor_version_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"monitoring_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

			"monitoring_interval": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},

			"promotion_tier": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsNeptuneClusterInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn
	tags := tagsFromMapNeptune(d.Get("tags").(map[string]interface{}))

	createOpts := &neptune.CreateDBInstanceInput{
		DBInstanceClass:         aws.String(d.Get("instance_class").(string)),
		DBClusterIdentifier:     aws.String(d.Get("cluster_identifier").(string)),
		Engine:                  aws.String(d.Get("engine").(string)),
		PubliclyAccessible:      aws.Bool(d.Get("publicly_accessible").(bool)),
		PromotionTier:           aws.Int64(int64(d.Get("promotion_tier").(int))),
		AutoMinorVersionUpgrade: aws.Bool(d.Get("auto_minor_version_upgrade").(bool)),
		Tags: tags,
	}

	if attr, ok := d.GetOk("availability_zone"); ok {
		createOpts.AvailabilityZone = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("neptune_parameter_group_name"); ok {
		createOpts.DBParameterGroupName = aws.String(attr.(string))
	}

	if v, ok := d.GetOk("identifier"); ok {
		createOpts.DBInstanceIdentifier = aws.String(v.(string))
	} else {
		if v, ok := d.GetOk("identifier_prefix"); ok {
			createOpts.DBInstanceIdentifier = aws.String(resource.PrefixedUniqueId(v.(string)))
		} else {
			createOpts.DBInstanceIdentifier = aws.String(resource.PrefixedUniqueId("tf-"))
		}
	}

	if attr, ok := d.GetOk("neptune_subnet_group_name"); ok {
		createOpts.DBSubnetGroupName = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("engine_version"); ok {
		createOpts.EngineVersion = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("monitoring_role_arn"); ok {
		createOpts.MonitoringRoleArn = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("preferred_backup_window"); ok {
		createOpts.PreferredBackupWindow = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("preferred_maintenance_window"); ok {
		createOpts.PreferredMaintenanceWindow = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("monitoring_interval"); ok {
		createOpts.MonitoringInterval = aws.Int64(int64(attr.(int)))
	}

	log.Printf("[DEBUG] Creating Neptune Instance: %s", createOpts)

	var resp *neptune.CreateDBInstanceOutput
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateDBInstance(createOpts)
		if err != nil {
			if isAWSErr(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error creating Neptune Instance: %s", err)
	}

	d.SetId(aws.StringValue(resp.DBInstance.DBInstanceIdentifier))

	stateConf := &resource.StateChangeConf{
		Pending:    resourceAwsNeptuneClusterInstanceCreateUpdatePendingStates,
		Target:     []string{"available"},
		Refresh:    resourceAwsNeptuneInstanceStateRefreshFunc(d.Id(), conn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return resourceAwsNeptuneClusterInstanceRead(d, meta)
}

var resourceAwsNeptuneClusterInstanceCreateUpdatePendingStates = []string{
	"backing-up",
	"configuring-enhanced-monitoring",
	"creating",
	"maintenance",
	"modifying",
	"rebooting",
	"renaming",
	"starting",
	"upgrading",
}

func resourceAwsNeptuneInstanceStateRefreshFunc(id string, conn *neptune.Neptune) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		v, err := resourceAwsNeptuneInstanceRetrieve(id, conn)

		if err != nil {
			log.Printf("Error on retrieving Neptune Instance when waiting: %s", err)
			return nil, "", err
		}

		if v == nil {
			return nil, "", nil
		}

		if v.DBInstanceStatus != nil {
			log.Printf("[DEBUG] Neptune Instance status for instance %s: %s", id, aws.StringValue(v.DBInstanceStatus))
		}

		return v, aws.StringValue(v.DBInstanceStatus), nil
	}
}

func resourceAwsNeptuneInstanceRetrieve(id string, conn *neptune.Neptune) (*neptune.DBInstance, error) {
	opts := neptune.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(id),
	}

	log.Printf("[DEBUG] Neptune Instance describe configuration: %#v", opts)

	resp, err := conn.DescribeDBInstances(&opts)
	if err != nil {
		if isAWSErr(err, neptune.ErrCodeDBInstanceNotFoundFault, "") {
			return nil, nil
		}
		return nil, fmt.Errorf("Error retrieving Neptune Instances: %s", err)
	}

	if len(resp.DBInstances) != 1 ||
		aws.StringValue(resp.DBInstances[0].DBInstanceIdentifier) != id {
		if err != nil {
			return nil, nil
		}
	}

	return resp.DBInstances[0], nil
}
