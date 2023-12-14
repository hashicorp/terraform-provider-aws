// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_s3_bucket_accelerate_configuration")
func ResourceBucketAccelerateConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketAccelerateConfigurationCreate,
		ReadWithoutTimeout:   resourceBucketAccelerateConfigurationRead,
		UpdateWithoutTimeout: resourceBucketAccelerateConfigurationUpdate,
		DeleteWithoutTimeout: resourceBucketAccelerateConfigurationDelete,

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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.BucketAccelerateStatus](),
			},
		},
	}
}

func resourceBucketAccelerateConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get("bucket").(string)
	expectedBucketOwner := d.Get("expected_bucket_owner").(string)
	input := &s3.PutBucketAccelerateConfigurationInput{
		Bucket: aws.String(bucket),
		AccelerateConfiguration: &types.AccelerateConfiguration{
			Status: types.BucketAccelerateStatus(d.Get("status").(string)),
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, s3BucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutBucketAccelerateConfiguration(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "AccelerateConfiguration is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return diag.Errorf("creating S3 Bucket (%s) Accelerate Configuration: %s", bucket, err)
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	_, err = tfresource.RetryWhenNotFound(ctx, s3BucketPropagationTimeout, func() (interface{}, error) {
		return findBucketAccelerateConfiguration(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		return diag.Errorf("waiting for S3 Bucket Accelerate Configuration (%s) create: %s", d.Id(), err)
	}

	return resourceBucketAccelerateConfigurationRead(ctx, d, meta)
}

func resourceBucketAccelerateConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	output, err := findBucketAccelerateConfiguration(ctx, conn, bucket, expectedBucketOwner)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Accelerate Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading S3 Bucket Accelerate Configuration (%s): %s", d.Id(), err)
	}

	d.Set("bucket", bucket)
	d.Set("expected_bucket_owner", expectedBucketOwner)
	d.Set("status", output.Status)

	return nil
}

func resourceBucketAccelerateConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutBucketAccelerateConfigurationInput{
		Bucket: aws.String(bucket),
		AccelerateConfiguration: &types.AccelerateConfiguration{
			Status: types.BucketAccelerateStatus(d.Get("status").(string)),
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.PutBucketAccelerateConfiguration(ctx, input)

	if err != nil {
		return diag.Errorf("updating S3 Bucket Accelerate Configuration (%s): %s", d.Id(), err)
	}

	return resourceBucketAccelerateConfigurationRead(ctx, d, meta)
}

func resourceBucketAccelerateConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutBucketAccelerateConfigurationInput{
		Bucket: aws.String(bucket),
		AccelerateConfiguration: &types.AccelerateConfiguration{
			Status: types.BucketAccelerateStatusSuspended,
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	log.Printf("[DEBUG] Deleting S3 Bucket Accelerate Configuration: %s", d.Id())
	_, err = conn.PutBucketAccelerateConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting S3 Bucket Accelerate Configuration (%s): %s", d.Id(), err)
	}

	// Don't wait for the accelerate configuration to disappear as it still exists after suspension.

	return nil
}

func findBucketAccelerateConfiguration(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) (*s3.GetBucketAccelerateConfigurationOutput, error) {
	input := &s3.GetBucketAccelerateConfigurationInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketAccelerateConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
