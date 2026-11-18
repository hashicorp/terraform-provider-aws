// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
					Deprecated:   "expected_bucket_owner is deprecated. It will be removed in a future verion of the provider.",
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
										"default_event_hold": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"days": {
														Type:          schema.TypeInt,
														Optional:      true,
														ValidateFunc:  validation.IntBetween(1, 36500),
														ConflictsWith: []string{"rule.0.default_retention.0.default_event_hold.0.years"},
													},
													"years": {
														Type:          schema.TypeInt,
														Optional:      true,
														ValidateFunc:  validation.IntBetween(1, 100),
														ConflictsWith: []string{"rule.0.default_retention.0.default_event_hold.0.days"},
													},
												},
											},
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
			}
		},
	}
}

func resourceBucketObjectLockConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}
	expectedBucketOwner := d.Get(names.AttrExpectedBucketOwner).(string)
	input := s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &types.ObjectLockConfiguration{
			// ObjectLockEnabled is required by the API, even if configured directly on the S3 bucket
			// during creation, else a MalformedXML error will be returned.
			ObjectLockEnabled: types.ObjectLockEnabled(d.Get("object_lock_enabled").(string)),
			Rule:              expandObjectLockRule(d.Get(names.AttrRule).([]any)),
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

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func(ctx context.Context) (any, error) {
		return conn.PutObjectLockConfiguration(ctx, &input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotImplemented) {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Object Lock Configuration: %s", bucket, err)
	}

	d.SetId(createResourceID(bucket, expectedBucketOwner))

	_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func(ctx context.Context) (any, error) {
		return findObjectLockConfiguration(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Object Lock Configuration (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBucketObjectLockConfigurationRead(ctx, d, meta)...)
}

func resourceBucketObjectLockConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := parseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	objLockConfig, err := findObjectLockConfiguration(ctx, conn, bucket, expectedBucketOwner)

	if !d.IsNewResource() && retry.NotFound(err) {
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

func resourceBucketObjectLockConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := parseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	input := s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &types.ObjectLockConfiguration{
			// ObjectLockEnabled is required by the API, even if configured directly on the S3 bucket
			// during creation, else a MalformedXML error will be returned.
			ObjectLockEnabled: types.ObjectLockEnabled(d.Get("object_lock_enabled").(string)),
			Rule:              expandObjectLockRule(d.Get(names.AttrRule).([]any)),
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

	_, err = conn.PutObjectLockConfiguration(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Bucket Object Lock Configuration (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBucketObjectLockConfigurationRead(ctx, d, meta)...)
}

func resourceBucketObjectLockConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := parseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if isDirectoryBucket(bucket) {
		conn = meta.(*conns.AWSClient).S3ExpressClient(ctx)
	}

	input := s3.PutObjectLockConfigurationInput{
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

	_, err = conn.PutObjectLockConfiguration(ctx, &input)

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
	input := s3.GetObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetObjectLockConfiguration(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeObjectLockConfigurationNotFoundError) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ObjectLockConfiguration == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.ObjectLockConfiguration, nil
}

func expandObjectLockRule(tfList []any) *types.ObjectLockRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.ObjectLockRule{}

	if v, ok := tfMap["default_retention"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.DefaultRetention = expandDefaultRetention(v)
	}

	return apiObject
}

func expandDefaultRetention(tfList []any) *types.DefaultRetention {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.DefaultRetention{}

	if v, ok := tfMap["days"].(int); ok && v > 0 {
		apiObject.Days = aws.Int32(int32(v))
	}

	if v, ok := tfMap["default_event_hold"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.DefaultEventHold = expandEventHoldDuration(v)
	}

	if v, ok := tfMap[names.AttrMode].(string); ok && v != "" {
		apiObject.Mode = types.ObjectLockRetentionMode(v)
	}

	if v, ok := tfMap["years"].(int); ok && v > 0 {
		apiObject.Years = aws.Int32(int32(v))
	}

	return apiObject
}

func expandEventHoldDuration(tfList []any) *types.EventHoldDuration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.EventHoldDuration{}

	if v, ok := tfMap["days"].(int); ok && v > 0 {
		apiObject.Days = aws.Int32(int32(v))
	}

	if v, ok := tfMap["years"].(int); ok && v > 0 {
		apiObject.Years = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenEventHoldDuration(apiObject *types.EventHoldDuration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"days":  aws.ToInt32(apiObject.Days),
		"years": aws.ToInt32(apiObject.Years),
	}

	return []any{tfMap}
}

func flattenObjectLockRule(apiObject *types.ObjectLockRule) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)

	if apiObject.DefaultRetention != nil {
		tfMap["default_retention"] = flattenDefaultRetention(apiObject.DefaultRetention)
	}

	return []any{tfMap}
}

func flattenDefaultRetention(apiObject *types.DefaultRetention) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"days":               aws.ToInt32(apiObject.Days),
		"default_event_hold": flattenEventHoldDuration(apiObject.DefaultEventHold),
		names.AttrMode:       apiObject.Mode,
		"years":              aws.ToInt32(apiObject.Years),
	}

	return []any{tfMap}
}
