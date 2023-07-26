// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"policy": {
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
	conn := meta.(*conns.AWSClient).S3ControlConn(ctx)

	bucket := d.Get("bucket").(string)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return diag.Errorf("policy (%s) is invalid JSON: %s", d.Get("policy").(string), err)
	}

	input := &s3control.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policy),
	}

	_, err = conn.PutBucketPolicyWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating S3 Control Bucket Policy (%s): %s", bucket, err)
	}

	d.SetId(bucket)

	return resourceBucketPolicyRead(ctx, d, meta)
}

func resourceBucketPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn(ctx)

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if parsedArn.AccountID == "" {
		return diag.Errorf("parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	output, err := FindBucketPolicyByTwoPartKey(ctx, conn, parsedArn.AccountID, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Control Bucket Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading S3 Control Bucket Policy (%s): %s", d.Id(), err)
	}

	d.Set("bucket", d.Id())

	if output.Policy != nil {
		policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.StringValue(output.Policy))
		if err != nil {
			return diag.FromErr(err)
		}

		d.Set("policy", policyToSet)
	} else {
		d.Set("policy", "")
	}

	return nil
}

func resourceBucketPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return diag.Errorf("policy (%s) is invalid JSON: %s", d.Get("policy").(string), err)
	}

	input := &s3control.PutBucketPolicyInput{
		Bucket: aws.String(d.Id()),
		Policy: aws.String(policy),
	}

	_, err = conn.PutBucketPolicyWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating S3 Control Bucket Policy (%s): %s", d.Id(), err)
	}

	return resourceBucketPolicyRead(ctx, d, meta)
}

func resourceBucketPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn(ctx)

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting S3 Control Bucket Policy: %s", d.Id())
	_, err = conn.DeleteBucketPolicyWithContext(ctx, &s3control.DeleteBucketPolicyInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchBucketPolicy, errCodeNoSuchOutpost) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting S3 Control Bucket Policy (%s): %s", d.Id(), err)
	}

	return nil
}

func FindBucketPolicyByTwoPartKey(ctx context.Context, conn *s3control.S3Control, accountID, bucket string) (*s3control.GetBucketPolicyOutput, error) {
	input := &s3control.GetBucketPolicyInput{
		AccountId: aws.String(accountID),
		Bucket:    aws.String(bucket),
	}

	output, err := conn.GetBucketPolicyWithContext(ctx, input)

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
