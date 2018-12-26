package aws

import (
	"fmt"

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

	params := &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	}

	resp, err := conn.GetBucketPolicy(params)
	if err != nil {
		return err
	}

	d.SetId(bucket)
	if err := d.Set("policy", resp.Policy); err != nil {
		return fmt.Errorf("[WARN] Error setting S3 Bucket (%s) Policy: %s", bucket, err)
	}

	return nil
}
