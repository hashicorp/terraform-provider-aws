package aws

import (
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsAthenaDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAthenaDatabaseCreate,
		Read:   resourceAwsAthenaDatabaseRead,
		Update: resourceAwsAthenaDatabaseUpdate,
		Delete: resourceAwsAthenaDatabaseDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"bucket": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceAwsAthenaDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	athenaconn := meta.(*AWSClient).athenaconn
	s3conn := meta.(*AWSClient).s3conn

	var bucket string
	if val, ok := d.GetOk("bucket"); ok {
		bucket = val.(string)
	} else {
		bucket = resource.UniqueId()
	}
	d.Set("bucket", bucket)
	var awsRegion string
	if val, ok := d.GetOk("region"); ok {
		awsRegion = val.(string)
	} else {
		awsRegion = meta.(*AWSClient).region
	}

	s3input := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(awsRegion),
		},
	}

	s3resp, err := s3conn.CreateBucket(s3input)
	if err != nil {
		return err
	}

	athenainput := &athena.StartQueryExecutionInput{
		QueryString: aws.String(createDatabaseQueryString(d.Get("name").(string))),
		ResultConfiguration: &athena.ResultConfiguration{
			OutputLocation: aws.String("s3://" + bucket),
		},
	}

	athenaresp, err := athenaconn.StartQueryExecution(athenainput)
	if err != nil {
		return err
	}
	return nil
}

func resourceAwsAthenaDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsAthenaDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsAthenaDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func createDatabaseQueryString(databaseName string) string {
	return fmt.Sprintf("create database %s;", databaseName)
}

func checkCreateDatabaseQueryExecution(qeid string) error {

	return nil
}

func showDatabaseQueryString(databaseName string) string {
	return fmt.Sprint("show databases;")
}

func checkShowDatabaseQueryExecution(qeid string) error {
	return nil
}

func dropDatabaseQueryString(databaseName string) string {
	return fmt.Sprintf("drop database %s;", databaseName)
}

func queryExecutionBody(qeid, bucket string, meta interface{}) ([]byte, error) {
	s3conn := meta.(*AWSClient).s3conn

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(qeid + ".txt"),
	}
	resp, err := s3conn.GetObject(input)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}
