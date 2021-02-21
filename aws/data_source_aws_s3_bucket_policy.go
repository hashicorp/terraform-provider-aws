package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsS3BucketPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsS3BucketPolicyRead,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsS3BucketPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucketName := d.Get("bucket").(string)
	input := &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucketName),
	}

	log.Printf("[DEBUG] Reading S3 bucket: %s", input)
	output, err := conn.GetBucketPolicy(input)
	if err != nil {
		return fmt.Errorf("Failed getting S3 bucket (%s): %w", bucketName, err)
	}

	d.SetId(bucketName)
	d.Set("policy", output.Policy)
	return nil
}
