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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_public_access_block", name="Bucket Public Access Block")
func resourceBucketPublicAccessBlock() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketPublicAccessBlockCreate,
		ReadWithoutTimeout:   resourceBucketPublicAccessBlockRead,
		UpdateWithoutTimeout: resourceBucketPublicAccessBlockUpdate,
		DeleteWithoutTimeout: resourceBucketPublicAccessBlockDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
			names.AttrBucket: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	input := &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(bucket),
		PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(d.Get("block_public_acls").(bool)),
			BlockPublicPolicy:     aws.Bool(d.Get("block_public_policy").(bool)),
			IgnorePublicAcls:      aws.Bool(d.Get("ignore_public_acls").(bool)),
			RestrictPublicBuckets: aws.Bool(d.Get("restrict_public_buckets").(bool)),
		},
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutPublicAccessBlock(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "PublicAccessBlockConfiguration is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Public Access Block: %s", bucket, err)
	}

	d.SetId(bucket)

	_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findPublicAccessBlockConfiguration(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Public Access Block (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBucketPublicAccessBlockRead(ctx, d, meta)...)
}

func resourceBucketPublicAccessBlockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	pabc, err := findPublicAccessBlockConfiguration(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Public Access Block (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Public Access Block (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, d.Id())
	d.Set("block_public_acls", pabc.BlockPublicAcls)
	d.Set("block_public_policy", pabc.BlockPublicPolicy)
	d.Set("ignore_public_acls", pabc.IgnorePublicAcls)
	d.Set("restrict_public_buckets", pabc.RestrictPublicBuckets)

	return diags
}

func resourceBucketPublicAccessBlockUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	input := &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(d.Id()),
		PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(d.Get("block_public_acls").(bool)),
			BlockPublicPolicy:     aws.Bool(d.Get("block_public_policy").(bool)),
			IgnorePublicAcls:      aws.Bool(d.Get("ignore_public_acls").(bool)),
			RestrictPublicBuckets: aws.Bool(d.Get("restrict_public_buckets").(bool)),
		},
	}

	_, err := conn.PutPublicAccessBlock(ctx, input)

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

	// Skip normal Read after Update due to eventual consistency issues.
	return diags
}

func resourceBucketPublicAccessBlockDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	log.Printf("[DEBUG] Deleting S3 Bucket Ownership Controls: %s", d.Id())
	_, err := conn.DeletePublicAccessBlock(ctx, &s3.DeletePublicAccessBlockInput{
		Bucket: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchPublicAccessBlockConfiguration) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Public Access Block (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findPublicAccessBlockConfiguration(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Public Access Block (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findPublicAccessBlockConfiguration(ctx context.Context, conn *s3.Client, bucket string) (*types.PublicAccessBlockConfiguration, error) {
	input := &s3.GetPublicAccessBlockInput{
		Bucket: aws.String(bucket),
	}

	output, err := conn.GetPublicAccessBlock(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchPublicAccessBlockConfiguration) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.PublicAccessBlockConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.PublicAccessBlockConfiguration, nil
}
