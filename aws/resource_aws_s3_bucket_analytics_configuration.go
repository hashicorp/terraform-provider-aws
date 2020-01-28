package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsS3BucketAnalyticsConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketAnalyticsConfigurationCreate,
		Read:   resourceAwsS3BucketAnalyticsConfigurationRead,
		Update: resourceAwsS3BucketAnalyticsConfigurationUpdate,
		Delete: resourceAwsS3BucketAnalyticsConfigurationDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tags": tagsSchema(),
					},
				},
			},
			"storage_class_analysis": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceAwsS3BucketAnalyticsConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] S3 bucket %q, add analytics configuration %q", bucket, name)

	storageClassAnalysis := &s3.StorageClassAnalysis{}
	analyticsConfiguration := &s3.AnalyticsConfiguration{
		Id:                   aws.String(name),
		StorageClassAnalysis: storageClassAnalysis,
	}

	input := &s3.PutBucketAnalyticsConfigurationInput{
		Bucket:                 aws.String(bucket),
		Id:                     aws.String(name),
		AnalyticsConfiguration: analyticsConfiguration,
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := s3conn.PutBucketAnalyticsConfiguration(input)
		if err != nil {
			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = s3conn.PutBucketAnalyticsConfiguration(input)
	}
	if err != nil {
		return fmt.Errorf("Error adding S3 analytics configuration: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", bucket, name))

	return resourceAwsS3BucketAnalyticsConfigurationRead(d, meta)
}

func resourceAwsS3BucketAnalyticsConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket, name, err := resourceAwsS3BucketAnalyticsConfigurationParseID(d.Id())
	if err != nil {
		return err
	}

	d.Set("bucket", bucket)
	d.Set("name", name)

	input := &s3.GetBucketAnalyticsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Reading S3 bucket analytics configuration: %s", input)
	_, err = conn.GetBucketAnalyticsConfiguration(input)
	if err != nil {
		if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
			log.Printf("[WARN] %s S3 bucket analytics configuration not found, removing from state.", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	return nil
}

func resourceAwsS3BucketAnalyticsConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsS3BucketAnalyticsConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket, name, err := resourceAwsS3BucketAnalyticsConfigurationParseID(d.Id())
	if err != nil {
		return err
	}

	input := &s3.DeleteBucketAnalyticsConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Deleting S3 bucket analytics configuration: %s", input)
	_, err = conn.DeleteBucketAnalyticsConfiguration(input)
	if err != nil {
		if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
			return nil
		}
		return fmt.Errorf("Error deleting S3 analytics configuration: %w", err)
	}

	return waitForDeleteS3BucketAnalyticsConfiguration(conn, bucket, name, 1*time.Minute)
}

func resourceAwsS3BucketAnalyticsConfigurationParseID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("please make sure the ID is in the form BUCKET:NAME (i.e. my-bucket:EntireBucket")
	}
	bucket := idParts[0]
	name := idParts[1]
	return bucket, name, nil
}

func waitForDeleteS3BucketAnalyticsConfiguration(conn *s3.S3, bucket, name string, timeout time.Duration) error {
	err := resource.Retry(timeout, func() *resource.RetryError {
		input := &s3.GetBucketAnalyticsConfigurationInput{
			Bucket: aws.String(bucket),
			Id:     aws.String(name),
		}
		log.Printf("[DEBUG] Reading S3 bucket analytics configuration: %s", input)
		output, err := conn.GetBucketAnalyticsConfiguration(input)
		if err != nil {
			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		if output.AnalyticsConfiguration != nil {
			return resource.RetryableError(fmt.Errorf("S3 bucket analytics configuration exists: %v", output))
		}

		return nil
	})

	return err
}
