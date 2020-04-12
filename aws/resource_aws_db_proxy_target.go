package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsDbProxyTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDbProxyTargetCreate,
		Read:   resourceAwsDbProxyTargetRead,
		Delete: resourceAwsDbProxyTargetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"db_proxy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRdsIdentifier,
			},
			"target_group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRdsIdentifier,
			},
			"db_instance_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"db_instance_identifier",
					"db_cluster_identifier",
				},
				ValidateFunc: validateRdsIdentifier,
			},
			"db_cluster_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"db_instance_identifier",
					"db_cluster_identifier",
				},
				ValidateFunc: validateRdsIdentifier,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"rds_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tracked_cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsDbProxyTargetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	dbProxyName := d.Get("db_proxy_name").(string)
	targetGroupName := d.Get("target_group_name").(string)

	params := rds.RegisterDBProxyTargetsInput{
		DBProxyName:     aws.String(dbProxyName),
		TargetGroupName: aws.String(targetGroupName),
	}

	if v, ok := d.GetOk("db_instance_identifier"); ok {
		params.DBInstanceIdentifiers = []*string{aws.String(v.(string))}
	}

	if v, ok := d.GetOk("db_cluster_identifier"); ok {
		params.DBClusterIdentifiers = []*string{aws.String(v.(string))}
	}

	log.Printf("[DEBUG] Register DB Proxy target: %#v", params)
	resp, err := conn.RegisterDBProxyTargets(&params)
	if err != nil {
		return fmt.Errorf("Error registering DB Proxy target: %s", err)
	}

	dbProxyTarget := resp.DBProxyTargets[0]

	d.SetId(strings.Join([]string{dbProxyName, targetGroupName, *dbProxyTarget.RdsResourceId}, "/"))
	log.Printf("[INFO] DB Proxy target ID: %s", d.Id())

	// stateChangeConf := &resource.StateChangeConf{
	// 	Pending: []string{rds.DBProxyStatusCreating},
	// 	Target:  []string{rds.DBProxyStatusAvailable},
	// 	Refresh: resourceAwsDbProxyTargetRefreshFunc(conn, d.Id()),
	// 	Timeout: d.Timeout(schema.TimeoutCreate),
	// }

	// _, err = stateChangeConf.WaitForState()
	// if err != nil {
	// 	return fmt.Errorf("Error waiting for DB Proxy target registration: %s", err)
	// }

	return resourceAwsDbProxyTargetRead(d, meta)
}

func resourceAwsDbProxyTargetParseID(id string) (string, string, string, error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected db_proxy_name/target_group_name/target_arn", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}

func resourceAwsDbProxyTargetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	dbProxyName, targetGroupName, rdsResourceId, err := resourceAwsDbProxyTargetParseID(d.Id())
	if err != nil {
		return err
	}

	params := rds.DescribeDBProxyTargetsInput{
		DBProxyName:     aws.String(dbProxyName),
		TargetGroupName: aws.String(targetGroupName),
	}

	resp, err := conn.DescribeDBProxyTargets(&params)
	if err != nil {
		if isAWSErr(err, rds.ErrCodeDBProxyNotFoundFault, "") {
			log.Printf("[WARN] DB Proxy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if isAWSErr(err, rds.ErrCodeDBProxyTargetGroupNotFoundFault, "") {
			log.Printf("[WARN] DB Proxy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	var dbProxyTarget *rds.DBProxyTarget
	for _, target := range resp.Targets {
		if *target.RdsResourceId == rdsResourceId {
			dbProxyTarget = target
			break
		}
	}

	if dbProxyTarget == nil {
		return fmt.Errorf("Unable to find DB Proxy target: %#v", params)
	}

	d.Set("endpoint", dbProxyTarget.Endpoint)
	d.Set("port", dbProxyTarget.Port)
	d.Set("rds_resource_id", dbProxyTarget.RdsResourceId)
	d.Set("target_arn", dbProxyTarget.TargetArn)
	d.Set("tracked_cluster_id", dbProxyTarget.TrackedClusterId)
	d.Set("type", dbProxyTarget.Type)

	return nil
}

func resourceAwsDbProxyTargetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	params := rds.DeregisterDBProxyTargetsInput{
		DBProxyName:     aws.String(d.Get("db_proxy_name").(string)),
		TargetGroupName: aws.String(d.Get("target_group_name").(string)),
	}

	if v, ok := d.GetOk("db_instance_identifier"); ok {
		params.DBInstanceIdentifiers = []*string{aws.String(v.(string))}
	}

	if v, ok := d.GetOk("db_cluster_identifier"); ok {
		params.DBClusterIdentifiers = []*string{aws.String(v.(string))}
	}

	log.Printf("[DEBUG] Deregister DB Proxy target: %#v", params)
	_, err := conn.DeregisterDBProxyTargets(&params)
	if err != nil {
		return fmt.Errorf("Error deregistering DB Proxy target: %s", err)
	}

	// stateChangeConf := &resource.StateChangeConf{
	// 	Pending: []string{rds.DBProxyStatusDeleting},
	// 	Target:  []string{""},
	// 	Refresh: resourceAwsDbProxyTargetRefreshFunc(conn, d.Id()),
	// 	Timeout: d.Timeout(schema.TimeoutDelete),
	// }

	// _, err = stateChangeConf.WaitForState()
	// if err != nil {
	// 	return fmt.Errorf("Error waiting for DB Proxy deletion: %s", err)
	// }

	return nil
}
