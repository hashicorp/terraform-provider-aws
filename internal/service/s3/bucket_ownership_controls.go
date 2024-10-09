// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"log"
	"time"

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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_ownership_controls", name="Bucket Ownership Controls")
func resourceBucketOwnershipControls() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketOwnershipControlsCreate,
		ReadWithoutTimeout:   resourceBucketOwnershipControlsRead,
		UpdateWithoutTimeout: resourceBucketOwnershipControlsUpdate,
		DeleteWithoutTimeout: resourceBucketOwnershipControlsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			names.AttrRule: {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"object_ownership": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.ObjectOwnership](),
						},
					},
				},
			},
		},
	}
}

func resourceBucketOwnershipControlsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	input := &s3.PutBucketOwnershipControlsInput{
		Bucket: aws.String(bucket),
		OwnershipControls: &types.OwnershipControls{
			Rules: expandOwnershipControlsRules(d.Get(names.AttrRule).([]interface{})),
		},
	}

	_, err := conn.PutBucketOwnershipControls(ctx, input)

	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "OwnershipControls is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Ownership Controls: %s", bucket, err)
	}

	d.SetId(bucket)

	_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findOwnershipControls(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Ownership Controls (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBucketOwnershipControlsRead(ctx, d, meta)...)
}

func resourceBucketOwnershipControlsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	oc, err := findOwnershipControls(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Ownership Controls (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Ownership Controls (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, d.Id())
	if err := d.Set(names.AttrRule, flattenOwnershipControlsRules(oc.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}

	return diags
}

func resourceBucketOwnershipControlsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	input := &s3.PutBucketOwnershipControlsInput{
		Bucket: aws.String(d.Id()),
		OwnershipControls: &types.OwnershipControls{
			Rules: expandOwnershipControlsRules(d.Get(names.AttrRule).([]interface{})),
		},
	}

	_, err := conn.PutBucketOwnershipControls(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Bucket Ownership Controls (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBucketOwnershipControlsRead(ctx, d, meta)...)
}

func resourceBucketOwnershipControlsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	log.Printf("[DEBUG] Deleting S3 Bucket Ownership Controls: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 5*time.Minute, func() (interface{}, error) {
		return conn.DeleteBucketOwnershipControls(ctx, &s3.DeleteBucketOwnershipControlsInput{
			Bucket: aws.String(d.Id()),
		})
	}, errCodeOperationAborted)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeOwnershipControlsNotFoundError) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Ownership Controls (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findOwnershipControls(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Ownership Controls (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findOwnershipControls(ctx context.Context, conn *s3.Client, bucket string) (*types.OwnershipControls, error) {
	input := &s3.GetBucketOwnershipControlsInput{
		Bucket: aws.String(bucket),
	}

	output, err := conn.GetBucketOwnershipControls(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeOwnershipControlsNotFoundError) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.OwnershipControls == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.OwnershipControls, nil
}

func expandOwnershipControlsRules(tfList []interface{}) []types.OwnershipControlsRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []types.OwnershipControlsRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandOwnershipControlsRule(tfMap))
	}

	return apiObjects
}

func expandOwnershipControlsRule(tfMap map[string]interface{}) types.OwnershipControlsRule {
	apiObject := types.OwnershipControlsRule{}

	if v, ok := tfMap["object_ownership"].(string); ok && v != "" {
		apiObject.ObjectOwnership = types.ObjectOwnership(v)
	}

	return apiObject
}

func flattenOwnershipControlsRules(apiObjects []types.OwnershipControlsRule) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenOwnershipControlsRule(apiObject))
	}

	return tfList
}

func flattenOwnershipControlsRule(apiObject types.OwnershipControlsRule) map[string]interface{} {
	tfMap := map[string]interface{}{
		"object_ownership": apiObject.ObjectOwnership,
	}

	return tfMap
}
