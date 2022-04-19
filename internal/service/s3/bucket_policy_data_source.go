package s3

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceAwsS3BucketPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAwsS3BucketPolicyRead,

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

func dataSourceAwsS3BucketPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucketName := d.Get("bucket").(string)
	input := &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucketName),
	}
	log.Printf("[DEBUG] Reading S3 bucket policy: %s", input)
	output, err := conn.GetBucketPolicy(input)
	if err != nil {
		return diag.Errorf("failed getting S3 bucket policy (%s): %w", bucketName, err)
	}

	policy := *output.Policy
	d.SetId(bucketName)
	d.Set("policy", policy)

	return nil
}
