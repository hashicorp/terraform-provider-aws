package rds

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceClusterActivityStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRDSClusterActivityStreamCreate,
		Read:   resourceAwsRDSClusterActivityStreamRead,
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
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
					rds.ActivityStreamModeSync,
					rds.ActivityStreamModeAsync,
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
	conn := meta.(*conns.AWSClient).RDSConn

	resourceArn := d.Get("resource_arn").(string)
	kmsKeyId := d.Get("kms_key_id").(string)
	mode := d.Get("mode").(string)

	startActivityStreamInput := &rds.StartActivityStreamInput{
		ResourceArn:      aws.String(resourceArn),
		ApplyImmediately: aws.Bool(true),
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

	err = waitActivityStreamStarted(conn, d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return err
	}

	return resourceAwsRDSClusterActivityStreamRead(d, meta)
}

func resourceAwsRDSClusterActivityStreamRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	input := &rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Describing RDS Cluster: %s", input)
	resp, err := conn.DescribeDBClusters(input)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == rds.ErrCodeDBClusterNotFoundFault {
			log.Printf("[WARN] RDS Cluster (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

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

	if aws.StringValue(dbc.ActivityStreamStatus) == rds.ActivityStreamStatusStopped {
		log.Printf("[WARN] RDS Cluster (%s) Activity Stream already stopped, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("resource_arn", dbc.DBClusterArn)
	d.Set("kms_key_id", dbc.ActivityStreamKmsKeyId)
	d.Set("kinesis_stream_name", dbc.ActivityStreamKinesisStreamName)
	d.Set("mode", dbc.ActivityStreamMode)

	return nil
}

func resourceAwsRDSClusterActivityStreamDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

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

	err = waitActivityStreamStopped(conn, d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return err
	}

	return nil
}
