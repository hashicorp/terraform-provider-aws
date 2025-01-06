// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3control_bucket_policy")
func resourceBucketPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketPolicyCreate,
		ReadWithoutTimeout:   resourceBucketPolicyRead,
		UpdateWithoutTimeout: resourceBucketPolicyUpdate,
		DeleteWithoutTimeout: resourceBucketPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceBucketPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	bucket := d.Get(names.AttrBucket).(string)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3control.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policy),
	}

	_, err = conn.PutBucketPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Control Bucket Policy (%s): %s", bucket, err)
	}

	d.SetId(bucket)

	return append(diags, resourceBucketPolicyRead(ctx, d, meta)...)
}

func resourceBucketPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	parsedArn, err := arn.Parse(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if parsedArn.AccountID == "" {
		return sdkdiag.AppendErrorf(diags, "parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	output, err := findBucketPolicyByTwoPartKey(ctx, conn, parsedArn.AccountID, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Control Bucket Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Control Bucket Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, d.Id())

	if output.Policy != nil {
		policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(output.Policy))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set(names.AttrPolicy, policyToSet)
	} else {
		d.Set(names.AttrPolicy, "")
	}

	return diags
}

func resourceBucketPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3control.PutBucketPolicyInput{
		Bucket: aws.String(d.Id()),
		Policy: aws.String(policy),
	}

	_, err = conn.PutBucketPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Control Bucket Policy (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBucketPolicyRead(ctx, d, meta)...)
}

func resourceBucketPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting S3 Control Bucket Policy: %s", d.Id())
	_, err = conn.DeleteBucketPolicy(ctx, &s3control.DeleteBucketPolicyInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchBucketPolicy, errCodeNoSuchOutpost) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Control Bucket Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findBucketPolicyByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, bucket string) (*s3control.GetBucketPolicyOutput, error) {
	input := &s3control.GetBucketPolicyInput{
		AccountId: aws.String(accountID),
		Bucket:    aws.String(bucket),
	}

	output, err := conn.GetBucketPolicy(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchBucketPolicy, errCodeNoSuchOutpost) {
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
