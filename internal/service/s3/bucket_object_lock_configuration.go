// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"log"
	"net/http"

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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_object_lock_configuration", name="Bucket Object Lock Configuration")
func resourceBucketObjectLockConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketObjectLockConfigurationCreate,
		ReadWithoutTimeout:   resourceBucketObjectLockConfigurationRead,
		UpdateWithoutTimeout: resourceBucketObjectLockConfigurationUpdate,
		DeleteWithoutTimeout: resourceBucketObjectLockConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			names.AttrExpectedBucketOwner: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"object_lock_enabled": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          types.ObjectLockEnabledEnabled,
				ValidateDiagFunc: enum.Validate[types.ObjectLockEnabled](),
			},
			names.AttrRule: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_retention": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:          schema.TypeInt,
										Optional:      true,
										ConflictsWith: []string{"rule.0.default_retention.0.years"},
									},
									names.AttrMode: {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.ObjectLockRetentionMode](),
									},
									"years": {
										Type:          schema.TypeInt,
										Optional:      true,
										ConflictsWith: []string{"rule.0.default_retention.0.days"},
									},
								},
							},
						},
					},
				},
			},
			"token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceBucketObjectLockConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	expectedBucketOwner := d.Get(names.AttrExpectedBucketOwner).(string)
	input := &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &types.ObjectLockConfiguration{
			// ObjectLockEnabled is required by the API, even if configured directly on the S3 bucket
			// during creation, else a MalformedXML error will be returned.
			ObjectLockEnabled: types.ObjectLockEnabled(d.Get("object_lock_enabled").(string)),
			Rule:              expandObjectLockRule(d.Get(names.AttrRule).([]interface{})),
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("request_payer"); ok {
		input.RequestPayer = types.RequestPayer(v.(string))
	}

	if v, ok := d.GetOk("token"); ok {
		input.Token = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutObjectLockConfiguration(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotImplemented) {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Object Lock Configuration: %s", bucket, err)
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findObjectLockConfiguration(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Object Lock Configuration (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBucketObjectLockConfigurationRead(ctx, d, meta)...)
}

func resourceBucketObjectLockConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	objLockConfig, err := findObjectLockConfiguration(ctx, conn, bucket, expectedBucketOwner)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Object Lock Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Object Lock Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, bucket)
	d.Set(names.AttrExpectedBucketOwner, expectedBucketOwner)
	d.Set("object_lock_enabled", objLockConfig.ObjectLockEnabled)
	if err := d.Set(names.AttrRule, flattenObjectLockRule(objLockConfig.Rule)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}

	return diags
}

func resourceBucketObjectLockConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &types.ObjectLockConfiguration{
			// ObjectLockEnabled is required by the API, even if configured directly on the S3 bucket
			// during creation, else a MalformedXML error will be returned.
			ObjectLockEnabled: types.ObjectLockEnabled(d.Get("object_lock_enabled").(string)),
			Rule:              expandObjectLockRule(d.Get(names.AttrRule).([]interface{})),
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("request_payer"); ok {
		input.RequestPayer = types.RequestPayer(v.(string))
	}

	if v, ok := d.GetOk("token"); ok {
		input.Token = aws.String(v.(string))
	}

	_, err = conn.PutObjectLockConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Bucket Object Lock Configuration (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBucketObjectLockConfigurationRead(ctx, d, meta)...)
}

func resourceBucketObjectLockConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &types.ObjectLockConfiguration{
			// ObjectLockEnabled is required by the API, even if configured directly on the S3 bucket
			// during creation, else a MalformedXML error will be returned.
			ObjectLockEnabled: types.ObjectLockEnabled(d.Get("object_lock_enabled").(string)),
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("request_payer"); ok {
		input.RequestPayer = types.RequestPayer(v.(string))
	}

	_, err = conn.PutObjectLockConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeObjectLockConfigurationNotFoundError) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Object Lock Configuration (%s): %s", d.Id(), err)
	}

	// Don't wait for the object lock configuration to disappear as may still exist.

	return diags
}

func findObjectLockConfiguration(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) (*types.ObjectLockConfiguration, error) {
	input := &s3.GetObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetObjectLockConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeObjectLockConfigurationNotFoundError) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ObjectLockConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ObjectLockConfiguration, nil
}

func expandObjectLockRule(l []interface{}) *types.ObjectLockRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	rule := &types.ObjectLockRule{}

	if v, ok := tfMap["default_retention"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		rule.DefaultRetention = expandDefaultRetention(v)
	}

	return rule
}

func expandDefaultRetention(l []interface{}) *types.DefaultRetention {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	dr := &types.DefaultRetention{}

	if v, ok := tfMap["days"].(int); ok && v > 0 {
		dr.Days = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMode].(string); ok && v != "" {
		dr.Mode = types.ObjectLockRetentionMode(v)
	}

	if v, ok := tfMap["years"].(int); ok && v > 0 {
		dr.Years = aws.Int32(int32(v))
	}

	return dr
}

func flattenObjectLockRule(rule *types.ObjectLockRule) []interface{} {
	if rule == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if rule.DefaultRetention != nil {
		m["default_retention"] = flattenDefaultRetention(rule.DefaultRetention)
	}

	return []interface{}{m}
}

func flattenDefaultRetention(dr *types.DefaultRetention) []interface{} {
	if dr == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"days":         dr.Days,
		names.AttrMode: dr.Mode,
		"years":        dr.Years,
	}

	return []interface{}{m}
}
