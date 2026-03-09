// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_policy", name="Bucket Policy")
// @IdentityAttribute("bucket")
// @Testing(preIdentityVersion="v6.9.0")
func resourceBucketPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketPolicyPut,
		ReadWithoutTimeout:   resourceBucketPolicyRead,
		UpdateWithoutTimeout: resourceBucketPolicyPut,
		DeleteWithoutTimeout: resourceBucketPolicyDelete,

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrPolicy: sdkv2.IAMPolicyDocumentSchemaRequired(),
		},
	}
}

func resourceBucketPolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	bucket := d.Get(names.AttrBucket).(string)
	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}
	input := &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policy),
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func(ctx context.Context) (any, error) {
		return conn.PutBucketPolicy(ctx, input)
	}, errCodeMalformedPolicy, errCodeNoSuchBucket)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) Policy: %s", bucket, err)
	}

	if d.IsNewResource() {
		d.SetId(bucket)

		_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func(ctx context.Context) (any, error) {
			return findBucketPolicy(ctx, conn, bucket)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Policy (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceBucketPolicyRead(ctx, d, meta)...)
}

func resourceBucketPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Id()
	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	policy, err := findBucketPolicy(ctx, conn, bucket)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Policy (%s): %s", d.Id(), err)
	}

	policy, err = verify.PolicyToSet(d.Get(names.AttrPolicy).(string), policy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = resourceBucketPolicyFlatten(ctx, policy, d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourceBucketPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Id()
	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	log.Printf("[DEBUG] Deleting S3 Bucket Policy: %s", d.Id())
	_, err := conn.DeleteBucketPolicy(ctx, &s3.DeleteBucketPolicyInput{
		Bucket: aws.String(bucket),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Policy (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, bucketPropagationTimeout, func(ctx context.Context) (any, error) {
		return findBucketPolicy(ctx, conn, bucket)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Policy (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findBucketPolicy(ctx context.Context, conn *s3.Client, bucket string) (string, error) {
	input := &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	}

	output, err := conn.GetBucketPolicy(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchBucketPolicy) {
		return "", &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return "", err
	}

	if output == nil || output.Policy == nil {
		return "", tfresource.NewEmptyResultError()
	}

	return aws.ToString(output.Policy), nil
}

func resourceBucketPolicyFlatten(_ context.Context, policy string, d *schema.ResourceData) error {
	policy, err := structure.NormalizeJsonString(policy)
	if err != nil {
		return fmt.Errorf("could not normalize policy JSON: %w", err)
	}
	d.Set(names.AttrPolicy, policy)
	return nil
}
