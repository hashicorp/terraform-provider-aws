package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsS3BucketPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsS3BucketPolicyRead,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsS3BucketPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)

	input := &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	}

	log.Printf("[DEBUG] Reading S3 bucket policy: %s", input)
	result, err := conn.GetBucketPolicy(input)

	if err != nil {
		return fmt.Errorf("Failed getting S3 bucket policy: %s Bucket: %q", err, bucket)
	}

	d.SetId(bucket)
	d.Set("policy", *result.Policy)

	return err
}
