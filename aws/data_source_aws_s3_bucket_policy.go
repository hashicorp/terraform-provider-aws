package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func dataSourceAwsS3BucketPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsS3BucketPolicyRead,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAwsS3BucketPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucketName := d.Get("bucket").(string)
	fmt.Println("bucketName:", bucketName)
	input := &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucketName),
	}

	log.Printf("[DEBUG] Reading S3 bucket policy: %s", input)
	output, err := conn.GetBucketPolicy(input)
	fmt.Printf("output:%v\n", output)
	if err != nil {
		return fmt.Errorf("failed getting S3 bucket policy (%s): %w", bucketName, err)
	}

	fmt.Println("policy:", *output.Policy)
	d.SetId(bucketName)
	d.Set("policy", *output.Policy)
	return nil
}
