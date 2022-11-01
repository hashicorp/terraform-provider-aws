package s3

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBucketAccelerateConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBucketAccelerateConfigurationCreate,
		ReadContext:   resourceBucketAccelerateConfigurationRead,
		UpdateContext: resourceBucketAccelerateConfigurationUpdate,
		DeleteContext: resourceBucketAccelerateConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			"expected_bucket_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"status": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(s3.BucketAccelerateStatus_Values(), false),
			},
		},
	}
}

func resourceBucketAccelerateConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket := d.Get("bucket").(string)
	expectedBucketOwner := d.Get("expected_bucket_owner").(string)

	input := &s3.PutBucketAccelerateConfigurationInput{
		Bucket: aws.String(bucket),
		AccelerateConfiguration: &s3.AccelerateConfiguration{
			Status: aws.String(d.Get("status").(string)),
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketAccelerateConfigurationWithContext(ctx, input)
	}, s3.ErrCodeNoSuchBucket)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating S3 bucket (%s) accelerate configuration: %w", bucket, err))
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	return resourceBucketAccelerateConfigurationRead(ctx, d, meta)
}

func resourceBucketAccelerateConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.GetBucketAccelerateConfigurationInput{
		Bucket: aws.String(bucket),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketAccelerateConfigurationWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket Accelerate Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading S3 bucket accelerate configuration (%s): %w", d.Id(), err))
	}

	if output == nil {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading S3 bucket accelerate configuration (%s): empty output", d.Id()))
		}
		log.Printf("[WARN] S3 Bucket Accelerate Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("bucket", bucket)
	d.Set("expected_bucket_owner", expectedBucketOwner)
	d.Set("status", output.Status)

	return nil
}

func resourceBucketAccelerateConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutBucketAccelerateConfigurationInput{
		Bucket: aws.String(bucket),
		AccelerateConfiguration: &s3.AccelerateConfiguration{
			Status: aws.String(d.Get("status").(string)),
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.PutBucketAccelerateConfigurationWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating S3 bucket accelerate configuration (%s): %w", d.Id(), err))
	}

	return resourceBucketAccelerateConfigurationRead(ctx, d, meta)
}

func resourceBucketAccelerateConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutBucketAccelerateConfigurationInput{
		Bucket: aws.String(bucket),
		AccelerateConfiguration: &s3.AccelerateConfiguration{
			Status: aws.String(s3.BucketAccelerateStatusSuspended),
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.PutBucketAccelerateConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting S3 bucket accelerate configuration (%s): %w", d.Id(), err))
	}

	return nil
}
