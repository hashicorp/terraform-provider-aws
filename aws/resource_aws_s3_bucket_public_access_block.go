package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsS3BucketPublicAccessBlock() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketPublicAccessBlockPut,
		Read:   resourceAwsS3BucketPublicAccessBlockRead,
		Update: resourceAwsS3BucketPublicAccessBlockPut,
		Delete: resourceAwsS3BucketPublicAccessBlockDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"block_public_acls": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"block_public_policy": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"ignore_public_acls": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"restrict_public_buckets": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceAwsS3BucketPublicAccessBlockPut(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn
	bucket := d.Get("bucket").(string)

	blockConfig := &s3.PublicAccessBlockConfiguration{}

	if v, ok := d.GetOk("block_public_acls"); ok {
		blockConfig.BlockPublicAcls = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("block_public_policy"); ok {
		blockConfig.BlockPublicPolicy = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("ignore_public_acls"); ok {
		blockConfig.IgnorePublicAcls = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("restrict_public_buckets"); ok {
		blockConfig.RestrictPublicBuckets = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] S3 bucket: %s, public access block: %v", bucket, blockConfig)

	params := &s3.PutPublicAccessBlockInput{
		Bucket:                         aws.String(bucket),
		PublicAccessBlockConfiguration: blockConfig,
	}

	_, err := s3conn.PutPublicAccessBlock(params)
	if err != nil {
		return fmt.Errorf("Error putting public access block policy on S3 bucket (%s): %s", bucket, err)
	}

	d.SetId(bucket)
	return resourceAwsS3BucketPublicAccessBlockRead(d, meta)
}

func resourceAwsS3BucketPublicAccessBlockRead(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	resp, err := s3conn.GetPublicAccessBlock(&s3.GetPublicAccessBlockInput{
		Bucket: aws.String(d.Id()),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok && awsErr.StatusCode() == 404 {
			d.SetId("")
			log.Printf("[WARN] Error Reading bucket public access block config, config not found (HTTP Status 404)")
			return nil
		}
		return err
	}
	config := resp.PublicAccessBlockConfiguration
	log.Printf("[DEBUG] Reading S3 Public Access Block Configuration: %s", resp.PublicAccessBlockConfiguration)

	d.Set("block_public_acls", config.BlockPublicAcls)
	d.Set("block_public_policy", config.BlockPublicPolicy)
	d.Set("ignore_public_acls", config.IgnorePublicAcls)
	d.Set("restrict_public_buckets", config.RestrictPublicBuckets)
	if err := d.Set("bucket", d.Id()); err != nil {
		return err
	}

	return nil
}

func resourceAwsS3BucketPublicAccessBlockDelete(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	log.Printf("[DEBUG] S3 bucket: %s, delete public access block", bucket)
	_, err := s3conn.DeletePublicAccessBlock(&s3.DeletePublicAccessBlockInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.RequestFailure); ok && awsErr.StatusCode() == 404 {
			return nil
		}
		return fmt.Errorf("Error deleting S3 bucket public access block: %s", err)
	}

	return nil
}
