package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceBucketPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketPolicyPut,
		Read:   resourceBucketPolicyRead,
		Update: resourceAwsS3BucketPolicyPut,
		Delete: resourceBucketPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
		},
	}
}

func resourceAwsS3BucketPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket := d.Get("bucket").(string)
	policy := d.Get("policy").(string)

	log.Printf("[DEBUG] S3 bucket: %s, put policy: %s", bucket, policy)

	params := &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policy),
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.PutBucketPolicy(params)
		if tfawserr.ErrMessageContains(err, "MalformedPolicy", "") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.PutBucketPolicy(params)
	}
	if err != nil {
		return fmt.Errorf("Error putting S3 policy: %s", err)
	}

	d.SetId(bucket)

	return nil
}

func resourceBucketPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	log.Printf("[DEBUG] S3 bucket policy, read for bucket: %s", d.Id())
	pol, err := conn.GetBucketPolicy(&s3.GetBucketPolicyInput{
		Bucket: aws.String(d.Id()),
	})

	v := ""
	if err == nil && pol.Policy != nil {
		v = *pol.Policy
	}
	if err := d.Set("policy", v); err != nil {
		return err
	}
	if err := d.Set("bucket", d.Id()); err != nil {
		return err
	}

	return nil
}

func resourceBucketPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket := d.Get("bucket").(string)

	log.Printf("[DEBUG] S3 bucket: %s, delete policy", bucket)
	_, err := conn.DeleteBucketPolicy(&s3.DeleteBucketPolicyInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchBucket" {
			return nil
		}
		return fmt.Errorf("Error deleting S3 policy: %s", err)
	}

	return nil
}
