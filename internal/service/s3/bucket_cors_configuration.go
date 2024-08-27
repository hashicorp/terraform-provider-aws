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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_cors_configuration", name="Bucket CORS Configuration")
func resourceBucketCorsConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketCorsConfigurationCreate,
		ReadWithoutTimeout:   resourceBucketCorsConfigurationRead,
		UpdateWithoutTimeout: resourceBucketCorsConfigurationUpdate,
		DeleteWithoutTimeout: resourceBucketCorsConfigurationDelete,

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
			"cors_rule": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_headers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_methods": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_origins": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"expose_headers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrID: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"max_age_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceBucketCorsConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	expectedBucketOwner := d.Get(names.AttrExpectedBucketOwner).(string)
	input := &s3.PutBucketCorsInput{
		Bucket: aws.String(bucket),
		CORSConfiguration: &types.CORSConfiguration{
			CORSRules: expandCORSRules(d.Get("cors_rule").(*schema.Set).List()),
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutBucketCors(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "CORSConfiguration is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) CORS Configuration: %s", bucket, err)
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findCORSRules(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket CORS Configuration (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBucketCorsConfigurationRead(ctx, d, meta)...)
}

func resourceBucketCorsConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	corsRules, err := findCORSRules(ctx, conn, bucket, expectedBucketOwner)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket CORS Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket CORS Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, bucket)
	if err := d.Set("cors_rule", flattenCORSRules(corsRules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cors_rule: %s", err)
	}
	d.Set(names.AttrExpectedBucketOwner, expectedBucketOwner)

	return diags
}

func resourceBucketCorsConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3.PutBucketCorsInput{
		Bucket: aws.String(bucket),
		CORSConfiguration: &types.CORSConfiguration{
			CORSRules: expandCORSRules(d.Get("cors_rule").(*schema.Set).List()),
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.PutBucketCors(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Bucket CORS Configuration (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBucketCorsConfigurationRead(ctx, d, meta)...)
}

func resourceBucketCorsConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3.DeleteBucketCorsInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.DeleteBucketCors(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchCORSConfiguration) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket CORS Configuration (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findCORSRules(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket CORS Configuration (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findCORSRules(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) ([]types.CORSRule, error) {
	input := &s3.GetBucketCorsInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketCors(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchCORSConfiguration) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CORSRules) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.CORSRules, nil
}

func expandCORSRules(l []interface{}) []types.CORSRule {
	if len(l) == 0 {
		return nil
	}

	var rules []types.CORSRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rule := types.CORSRule{}

		if v, ok := tfMap["allowed_headers"].(*schema.Set); ok && v.Len() > 0 {
			rule.AllowedHeaders = flex.ExpandStringValueSet(v)
		}

		if v, ok := tfMap["allowed_methods"].(*schema.Set); ok && v.Len() > 0 {
			rule.AllowedMethods = flex.ExpandStringValueSet(v)
		}

		if v, ok := tfMap["allowed_origins"].(*schema.Set); ok && v.Len() > 0 {
			rule.AllowedOrigins = flex.ExpandStringValueSet(v)
		}

		if v, ok := tfMap["expose_headers"].(*schema.Set); ok && v.Len() > 0 {
			rule.ExposeHeaders = flex.ExpandStringValueSet(v)
		}

		if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
			rule.ID = aws.String(v)
		}

		if v, ok := tfMap["max_age_seconds"].(int); ok {
			rule.MaxAgeSeconds = aws.Int32(int32(v))
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenCORSRules(rules []types.CORSRule) []interface{} {
	var results []interface{}

	for _, rule := range rules {
		m := map[string]interface{}{
			"max_age_seconds": rule.MaxAgeSeconds,
		}

		if len(rule.AllowedHeaders) > 0 {
			m["allowed_headers"] = rule.AllowedHeaders
		}

		if len(rule.AllowedMethods) > 0 {
			m["allowed_methods"] = rule.AllowedMethods
		}

		if len(rule.AllowedOrigins) > 0 {
			m["allowed_origins"] = rule.AllowedOrigins
		}

		if len(rule.ExposeHeaders) > 0 {
			m["expose_headers"] = rule.ExposeHeaders
		}

		if rule.ID != nil {
			m[names.AttrID] = aws.ToString(rule.ID)
		}

		results = append(results, m)
	}

	return results
}
