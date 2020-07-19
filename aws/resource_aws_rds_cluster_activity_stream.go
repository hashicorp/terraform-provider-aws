package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const (
	AWSRDSClusterActivityStreamRetryDelay      = 5 * time.Second
	AWSRDSClusterActivityStreamRetryMinTimeout = 3 * time.Second
)

func resourceAwsRDSClusterActivityStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRDSClusterActivityStreamCreate,
		Read:   resourceAwsRDSClusterActivityStreamRead,
		Update: resourceAwsRDSClusterActivityStreamUpdate,
		Delete: resourceAwsRDSClusterActivityStreamDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"apply_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mode": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"sync",
					"async",
				}, false),
			},
			"kinesis_stream_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsRDSClusterActivityStreamCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	resourceArn := d.Get("resource_arn").(string)
	applyImmediately := d.Get("apply_immediately").(bool)
	kmsKeyId := d.Get("kms_key_id").(string)
	mode := d.Get("mode").(string)

	startActivityStreamInput := &rds.StartActivityStreamInput{
		ResourceArn:      aws.String(resourceArn),
		ApplyImmediately: aws.Bool(applyImmediately),
		KmsKeyId:         aws.String(kmsKeyId),
		Mode:             aws.String(mode),
	}

	log.Printf("[DEBUG] RDS Cluster start activity stream input: %s", startActivityStreamInput)

	resp, err := conn.StartActivityStream(startActivityStreamInput)
	if err != nil {
		return fmt.Errorf("error creating RDS Cluster Activity Stream: %s", err)
	}

	log.Printf("[DEBUG]: RDS Cluster start activity stream response: %s", resp)

	d.SetId(resourceArn)

	err = resourceAwsRDSClusterActivityStreamWaitForStarted(d.Timeout(schema.TimeoutCreate), d.Id(), conn)
	if err != nil {
		return err
	}

	return resourceAwsRDSClusterActivityStreamRead(d, meta)
}

func resourceAwsRDSClusterActivityStreamRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	input := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Describing RDS Cluster: %s", input)
	resp, err := conn.DescribeDBClusters(input)

	if isAWSErr(err, rds.ErrCodeDBClusterNotFoundFault, "") {
		log.Printf("[WARN] RDS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing RDS Cluster (%s): %s", d.Id(), err)
	}

	if resp == nil {
		return fmt.Errorf("error retrieving RDS cluster: empty response for: %s", input)
	}

	var dbc *rds.DBCluster
	for _, c := range resp.DBClusters {
		if aws.StringValue(c.DBClusterArn) == d.Id() {
			dbc = c
			break
		}
	}

	if dbc == nil {
		log.Printf("[WARN] RDS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("resource_arn", dbc.DBClusterArn)
	d.Set("kms_key_id", dbc.ActivityStreamKmsKeyId)
	d.Set("kinesis_stream_name", dbc.ActivityStreamKinesisStreamName)
	d.Set("mode", dbc.ActivityStreamMode)

	return nil
}

func resourceAwsRDSClusterActivityStreamUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChanges("resource_arn", "kms_key_id", "mode") {
		log.Printf("[DEBUG] Stopping RDS Cluster Activity Stream before updating")
		if err := resourceAwsRDSClusterActivityStreamDelete(d, meta); err != nil {
			return err
		}

		log.Printf("[DEBUG] Starting RDS Cluster Activity Stream")
		if err := resourceAwsRDSClusterActivityStreamCreate(d, meta); err != nil {
			return err
		}
	}

	return resourceAwsRDSClusterActivityStreamRead(d, meta)
}

func resourceAwsRDSClusterActivityStreamDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	stopActivityStreamInput := &rds.StopActivityStreamInput{
		ApplyImmediately: aws.Bool(true),
		ResourceArn:      aws.String(d.Id()),
	}

	log.Printf("[DEBUG] RDS Cluster stop activity stream input: %s", stopActivityStreamInput)

	resp, err := conn.StopActivityStream(stopActivityStreamInput)
	if err != nil {
		return fmt.Errorf("error stopping RDS Cluster Activity Stream: %s", err)
	}

	log.Printf("[DEBUG] RDS Cluster stop activity stream response: %s", resp)

	if err := resourceAwsRDSClusterActivityStreamWaitForStopped(d.Timeout(schema.TimeoutDelete), d.Id(), conn); err != nil {
		return err
	}

	return nil
}

func resourceAwsRDSClusterActivityStreamWaitForStarted(timeout time.Duration, id string, conn *rds.RDS) error {
	log.Printf("[DEBUG] Waiting for RDS Cluster Activity Stream %s to become started...", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{rds.ActivityStreamStatusStarting},
		Target:     []string{rds.ActivityStreamStatusStarted},
		Refresh:    resourceAwsRDSClusterActivityStreamStateRefreshFunc(conn, id),
		Timeout:    timeout,
		Delay:      AWSRDSClusterActivityStreamRetryDelay,
		MinTimeout: AWSRDSClusterActivityStreamRetryMinTimeout,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for RDS Cluster Activity Stream (%s) to be started: %v", id, err)
	}
	return nil
}

func resourceAwsRDSClusterActivityStreamWaitForStopped(timeout time.Duration, id string, conn *rds.RDS) error {
	log.Printf("[DEBUG] Waiting for RDS Cluster Activity Stream %s to become stopped...", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{rds.ActivityStreamStatusStopping},
		Target:     []string{rds.ActivityStreamStatusStopped},
		Refresh:    resourceAwsRDSClusterActivityStreamStateRefreshFunc(conn, id),
		Timeout:    timeout,
		Delay:      AWSRDSClusterActivityStreamRetryDelay,
		MinTimeout: AWSRDSClusterActivityStreamRetryMinTimeout,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for RDS Cluster Activity Stream (%s) to be stopped: %v", id, err)
	}
	return nil
}

func resourceAwsRDSClusterActivityStreamStateRefreshFunc(conn *rds.RDS, dbClusterIdentifier string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		emptyResp := &rds.DescribeDBClustersInput{}

		resp, err := conn.DescribeDBClusters(&rds.DescribeDBClustersInput{
			DBClusterIdentifier: aws.String(dbClusterIdentifier),
		})

		if err != nil {
			log.Printf("[DEBUG] Refreshing RDS Cluster Activity Stream State. Occur error: %s", err)
			if isAWSErr(err, rds.ErrCodeDBClusterNotFoundFault, "") {
				return emptyResp, rds.ActivityStreamStatusStopped, nil
			} else if resp != nil && len(resp.DBClusters) == 0 {
				return emptyResp, rds.ActivityStreamStatusStopped, nil
			} else {
				return emptyResp, "", fmt.Errorf("error on refresh: %+v", err)
			}
		}

		if resp == nil || resp.DBClusters == nil || len(resp.DBClusters) == 0 {
			log.Printf("[DEBUG] Refreshing RDS Cluster Activity Stream State. Invalid resp: %s", resp)
			return emptyResp, rds.ActivityStreamStatusStopped, nil
		}

		cluster := resp.DBClusters[0]
		status := aws.StringValue(cluster.ActivityStreamStatus)
		log.Printf("[DEBUG] Refreshing RDS Cluster Activity Stream State... %s", status)
		return cluster, status, nil
	}
}
