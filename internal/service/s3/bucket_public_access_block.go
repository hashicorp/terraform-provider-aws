// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_s3_bucket_public_access_block")
func ResourceBucketPublicAccessBlock() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketPublicAccessBlockCreate,
		ReadWithoutTimeout:   resourceBucketPublicAccessBlockRead,
		UpdateWithoutTimeout: resourceBucketPublicAccessBlockUpdate,
		DeleteWithoutTimeout: resourceBucketPublicAccessBlockDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
				Default:  false,
			},

			"block_public_policy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"ignore_public_acls": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"restrict_public_buckets": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceBucketPublicAccessBlockCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)
	bucket := d.Get("bucket").(string)

	input := &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(bucket),
		PublicAccessBlockConfiguration: &s3.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(d.Get("block_public_acls").(bool)),
			BlockPublicPolicy:     aws.Bool(d.Get("block_public_policy").(bool)),
			IgnorePublicAcls:      aws.Bool(d.Get("ignore_public_acls").(bool)),
			RestrictPublicBuckets: aws.Bool(d.Get("restrict_public_buckets").(bool)),
		},
	}

	log.Printf("[DEBUG] S3 bucket: %s, public access block: %v", bucket, input.PublicAccessBlockConfiguration)
	err := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
		_, err := conn.PutPublicAccessBlockWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.PutPublicAccessBlockWithContext(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating public access block policy for S3 bucket (%s): %s", bucket, err)
	}

	d.SetId(bucket)
	return append(diags, resourceBucketPublicAccessBlockRead(ctx, d, meta)...)
}

func resourceBucketPublicAccessBlockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	input := &s3.GetPublicAccessBlockInput{
		Bucket: aws.String(d.Id()),
	}

	// Retry for eventual consistency on creation
	var output *s3.GetPublicAccessBlockOutput
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.GetPublicAccessBlockWithContext(ctx, input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return retry.RetryableError(err)
		}

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeNoSuchPublicAccessBlockConfiguration) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.GetPublicAccessBlockWithContext(ctx, input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeNoSuchPublicAccessBlockConfiguration) {
		log.Printf("[WARN] S3 Bucket Public Access Block (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 bucket Public Access Block (%s): %s", d.Id(), err)
	}

	if output == nil || output.PublicAccessBlockConfiguration == nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Public Access Block (%s): empty response", d.Id())
	}

	d.Set("bucket", d.Id())
	d.Set("block_public_acls", output.PublicAccessBlockConfiguration.BlockPublicAcls)
	d.Set("block_public_policy", output.PublicAccessBlockConfiguration.BlockPublicPolicy)
	d.Set("ignore_public_acls", output.PublicAccessBlockConfiguration.IgnorePublicAcls)
	d.Set("restrict_public_buckets", output.PublicAccessBlockConfiguration.RestrictPublicBuckets)

	return diags
}

func resourceBucketPublicAccessBlockUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	input := &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(d.Id()),
		PublicAccessBlockConfiguration: &s3.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(d.Get("block_public_acls").(bool)),
			BlockPublicPolicy:     aws.Bool(d.Get("block_public_policy").(bool)),
			IgnorePublicAcls:      aws.Bool(d.Get("ignore_public_acls").(bool)),
			RestrictPublicBuckets: aws.Bool(d.Get("restrict_public_buckets").(bool)),
		},
	}

	log.Printf("[DEBUG] Updating S3 bucket Public Access Block: %s", input)
	_, err := conn.PutPublicAccessBlockWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Bucket Public Access Block (%s): %s", d.Id(), err)
	}

	// Workaround API eventual consistency issues. This type of logic should not normally be used.
	// We cannot reliably determine when the Read after Update might be properly updated.
	// Rather than introduce complicated retry logic, we presume that a lack of an update error
	// means our update succeeded with our expected values.
	d.Set("block_public_acls", input.PublicAccessBlockConfiguration.BlockPublicAcls)
	d.Set("block_public_policy", input.PublicAccessBlockConfiguration.BlockPublicPolicy)
	d.Set("ignore_public_acls", input.PublicAccessBlockConfiguration.IgnorePublicAcls)
	d.Set("restrict_public_buckets", input.PublicAccessBlockConfiguration.RestrictPublicBuckets)

	// Skip normal Read after Update due to eventual consistency issues
	return diags
}

func resourceBucketPublicAccessBlockDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	input := &s3.DeletePublicAccessBlockInput{
		Bucket: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] S3 bucket: %s, delete public access block", d.Id())
	_, err := conn.DeletePublicAccessBlockWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchPublicAccessBlockConfiguration) {
		return diags
	}

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Public Access Block (%s): %s", d.Id(), err)
	}

	return diags
}
