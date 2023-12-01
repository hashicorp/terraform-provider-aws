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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_s3_bucket_object_lock_configuration")
func ResourceBucketObjectLockConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketObjectLockConfigurationCreate,
		ReadWithoutTimeout:   resourceBucketObjectLockConfigurationRead,
		UpdateWithoutTimeout: resourceBucketObjectLockConfigurationUpdate,
		DeleteWithoutTimeout: resourceBucketObjectLockConfigurationDelete,

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
			"object_lock_enabled": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          types.ObjectLockEnabledEnabled,
				ValidateDiagFunc: enum.Validate[types.ObjectLockEnabled](),
			},
			"rule": {
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
									"mode": {
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
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get("bucket").(string)
	expectedBucketOwner := d.Get("expected_bucket_owner").(string)
	input := &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &types.ObjectLockConfiguration{
			// ObjectLockEnabled is required by the API, even if configured directly on the S3 bucket
			// during creation, else a MalformedXML error will be returned.
			ObjectLockEnabled: types.ObjectLockEnabled(d.Get("object_lock_enabled").(string)),
			Rule:              expandBucketObjectLockConfigurationRule(d.Get("rule").([]interface{})),
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

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, s3BucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutObjectLockConfiguration(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotImplemented) {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return diag.Errorf("creating S3 Bucket (%s) Object Lock Configuration: %s", bucket, err)
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	_, err = tfresource.RetryWhenNotFound(ctx, s3BucketPropagationTimeout, func() (interface{}, error) {
		return findObjectLockConfiguration(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		return diag.Errorf("waiting for S3 Bucket Object Lock Configuration (%s) create: %s", d.Id(), err)
	}

	return resourceBucketObjectLockConfigurationRead(ctx, d, meta)
}

func resourceBucketObjectLockConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	objLockConfig, err := findObjectLockConfiguration(ctx, conn, bucket, expectedBucketOwner)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Object Lock Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading S3 Bucket Object Lock Configuration (%s): %s", d.Id(), err)
	}

	d.Set("bucket", bucket)
	d.Set("expected_bucket_owner", expectedBucketOwner)
	d.Set("object_lock_enabled", objLockConfig.ObjectLockEnabled)
	if err := d.Set("rule", flattenBucketObjectLockConfigurationRule(objLockConfig.Rule)); err != nil {
		return diag.Errorf("setting rule: %s", err)
	}

	return nil
}

func resourceBucketObjectLockConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &types.ObjectLockConfiguration{
			// ObjectLockEnabled is required by the API, even if configured directly on the S3 bucket
			// during creation, else a MalformedXML error will be returned.
			ObjectLockEnabled: types.ObjectLockEnabled(d.Get("object_lock_enabled").(string)),
			Rule:              expandBucketObjectLockConfigurationRule(d.Get("rule").([]interface{})),
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
		return diag.Errorf("updating S3 Bucket Object Lock Configuration (%s): %s", d.Id(), err)
	}

	return resourceBucketObjectLockConfigurationRead(ctx, d, meta)
}

func resourceBucketObjectLockConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
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
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting S3 Bucket Object Lock Configuration (%s): %s", d.Id(), err)
	}

	// Don't wait for the object lock configuration to disappear as may still exist.

	return nil
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

func expandBucketObjectLockConfigurationRule(l []interface{}) *types.ObjectLockRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	rule := &types.ObjectLockRule{}

	if v, ok := tfMap["default_retention"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		rule.DefaultRetention = expandBucketObjectLockConfigurationCorsRuleDefaultRetention(v)
	}

	return rule
}

func expandBucketObjectLockConfigurationCorsRuleDefaultRetention(l []interface{}) *types.DefaultRetention {
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

	if v, ok := tfMap["mode"].(string); ok && v != "" {
		dr.Mode = types.ObjectLockRetentionMode(v)
	}

	if v, ok := tfMap["years"].(int); ok && v > 0 {
		dr.Years = aws.Int32(int32(v))
	}

	return dr
}

func flattenBucketObjectLockConfigurationRule(rule *types.ObjectLockRule) []interface{} {
	if rule == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if rule.DefaultRetention != nil {
		m["default_retention"] = flattenBucketObjectLockConfigurationRuleDefaultRetention(rule.DefaultRetention)
	}

	return []interface{}{m}
}

func flattenBucketObjectLockConfigurationRuleDefaultRetention(dr *types.DefaultRetention) []interface{} {
	if dr == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"days":  dr.Days,
		"mode":  dr.Mode,
		"years": dr.Years,
	}

	return []interface{}{m}
}
