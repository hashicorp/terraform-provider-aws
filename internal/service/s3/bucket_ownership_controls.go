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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_s3_bucket_ownership_controls")
func ResourceBucketOwnershipControls() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketOwnershipControlsCreate,
		ReadWithoutTimeout:   resourceBucketOwnershipControlsRead,
		UpdateWithoutTimeout: resourceBucketOwnershipControlsUpdate,
		DeleteWithoutTimeout: resourceBucketOwnershipControlsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"object_ownership": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(s3.ObjectOwnership_Values(), false),
						},
					},
				},
			},
		},
	}
}

func resourceBucketOwnershipControlsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	bucket := d.Get("bucket").(string)

	input := &s3.PutBucketOwnershipControlsInput{
		Bucket: aws.String(bucket),
		OwnershipControls: &s3.OwnershipControls{
			Rules: expandOwnershipControlsRules(d.Get("rule").([]interface{})),
		},
	}

	_, err := conn.PutBucketOwnershipControlsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Ownership Controls: %s", bucket, err)
	}

	d.SetId(bucket)

	return append(diags, resourceBucketOwnershipControlsRead(ctx, d, meta)...)
}

func resourceBucketOwnershipControlsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	input := &s3.GetBucketOwnershipControlsInput{
		Bucket: aws.String(d.Id()),
	}

	output, err := conn.GetBucketOwnershipControlsWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket Ownership Controls (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "OwnershipControlsNotFoundError") {
		log.Printf("[WARN] S3 Bucket Ownership Controls (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) Ownership Controls: %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) Ownership Controls: empty response", d.Id())
	}

	d.Set("bucket", d.Id())

	if output.OwnershipControls == nil {
		d.Set("rule", nil)
	} else {
		if err := d.Set("rule", flattenOwnershipControlsRules(output.OwnershipControls.Rules)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
		}
	}

	return diags
}

func resourceBucketOwnershipControlsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	input := &s3.PutBucketOwnershipControlsInput{
		Bucket: aws.String(d.Id()),
		OwnershipControls: &s3.OwnershipControls{
			Rules: expandOwnershipControlsRules(d.Get("rule").([]interface{})),
		},
	}

	_, err := conn.PutBucketOwnershipControlsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Bucket (%s) Ownership Controls: %s", d.Id(), err)
	}

	return append(diags, resourceBucketOwnershipControlsRead(ctx, d, meta)...)
}

func resourceBucketOwnershipControlsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	input := &s3.DeleteBucketOwnershipControlsInput{
		Bucket: aws.String(d.Id()),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 5*time.Minute,
		func() (any, error) {
			return conn.DeleteBucketOwnershipControlsWithContext(ctx, input)
		},
		"OperationAborted",
	)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return diags
	}

	if tfawserr.ErrCodeEquals(err, "OwnershipControlsNotFoundError") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket (%s) Ownership Controls: %s", d.Id(), err)
	}

	return diags
}

func expandOwnershipControlsRules(tfList []interface{}) []*s3.OwnershipControlsRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []*s3.OwnershipControlsRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandOwnershipControlsRule(tfMap))
	}

	return apiObjects
}

func expandOwnershipControlsRule(tfMap map[string]interface{}) *s3.OwnershipControlsRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3.OwnershipControlsRule{}

	if v, ok := tfMap["object_ownership"].(string); ok && v != "" {
		apiObject.ObjectOwnership = aws.String(v)
	}

	return apiObject
}

func flattenOwnershipControlsRules(apiObjects []*s3.OwnershipControlsRule) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenOwnershipControlsRule(apiObject))
	}

	return tfList
}

func flattenOwnershipControlsRule(apiObject *s3.OwnershipControlsRule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ObjectOwnership; v != nil {
		tfMap["object_ownership"] = aws.StringValue(v)
	}

	return tfMap
}
