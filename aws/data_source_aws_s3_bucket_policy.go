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

	bucketName := d.Get("bucket").(string)

	// fails to get policy without this part
	_, err := conn.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("bucket not found: %s", err)
	}

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

	policy := *output.Policy
	fmt.Println("policy:", policy)
	d.SetId(bucketName)
	d.Set("policy", policy)
	return nil
}
