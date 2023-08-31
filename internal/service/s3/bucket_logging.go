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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_s3_bucket_logging")
func ResourceBucketLogging() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketLoggingCreate,
		ReadWithoutTimeout:   resourceBucketLoggingRead,
		UpdateWithoutTimeout: resourceBucketLoggingUpdate,
		DeleteWithoutTimeout: resourceBucketLoggingDelete,

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
			"target_bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"target_grant": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"grantee": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"display_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"email_address": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(s3.Type_Values(), false),
									},
									"uri": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"permission": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(s3.BucketLogsPermission_Values(), false),
						},
					},
				},
			},
			"target_prefix": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceBucketLoggingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	bucket := d.Get("bucket").(string)
	expectedBucketOwner := d.Get("expected_bucket_owner").(string)
	input := &s3.PutBucketLoggingInput{
		Bucket: aws.String(bucket),
		BucketLoggingStatus: &s3.BucketLoggingStatus{
			LoggingEnabled: &s3.LoggingEnabled{
				TargetBucket: aws.String(d.Get("target_bucket").(string)),
				TargetPrefix: aws.String(d.Get("target_prefix").(string)),
			},
		},
	}

	if v, ok := d.GetOk("target_grant"); ok && v.(*schema.Set).Len() > 0 {
		input.BucketLoggingStatus.LoggingEnabled.TargetGrants = expandBucketLoggingTargetGrants(v.(*schema.Set).List())
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketLoggingWithContext(ctx, input)
	}, s3.ErrCodeNoSuchBucket)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting S3 Bucket (%s) Logging: %s", bucket, err)
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindBucketLogging(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		return diag.Errorf("waiting for S3 Bucket Logging (%s) create: %s", d.Id(), err)
	}

	return resourceBucketLoggingRead(ctx, d, meta)
}

func resourceBucketLoggingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	loggingEnabled, err := FindBucketLogging(ctx, conn, bucket, expectedBucketOwner)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Logging (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Logging (%s): %s", d.Id(), err)
	}

	d.Set("bucket", bucket)
	d.Set("expected_bucket_owner", expectedBucketOwner)
	d.Set("target_bucket", loggingEnabled.TargetBucket)
	if err := d.Set("target_grant", flattenBucketLoggingTargetGrants(loggingEnabled.TargetGrants)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target_grant: %s", err)
	}
	d.Set("target_prefix", loggingEnabled.TargetPrefix)

	return diags
}

func resourceBucketLoggingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3.PutBucketLoggingInput{
		Bucket: aws.String(bucket),
		BucketLoggingStatus: &s3.BucketLoggingStatus{
			LoggingEnabled: &s3.LoggingEnabled{
				TargetBucket: aws.String(d.Get("target_bucket").(string)),
				TargetPrefix: aws.String(d.Get("target_prefix").(string)),
			},
		},
	}

	if v, ok := d.GetOk("target_grant"); ok && v.(*schema.Set).Len() > 0 {
		input.BucketLoggingStatus.LoggingEnabled.TargetGrants = expandBucketLoggingTargetGrants(v.(*schema.Set).List())
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketLoggingWithContext(ctx, input)
	}, s3.ErrCodeNoSuchBucket)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Bucket Logging (%s): %s", d.Id(), err)
	}

	return resourceBucketLoggingRead(ctx, d, meta)
}

func resourceBucketLoggingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3.PutBucketLoggingInput{
		Bucket:              aws.String(bucket),
		BucketLoggingStatus: &s3.BucketLoggingStatus{},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.PutBucketLoggingWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Logging (%s): %s", d.Id(), err)
	}

	return nil
}

func FindBucketLogging(ctx context.Context, conn *s3.S3, bucketName, expectedBucketOwner string) (*s3.LoggingEnabled, error) {
	input := &s3.GetBucketLoggingInput{
		Bucket: aws.String(bucketName),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketLoggingWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.LoggingEnabled == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.LoggingEnabled, nil
}

func expandBucketLoggingTargetGrants(l []interface{}) []*s3.TargetGrant {
	var grants []*s3.TargetGrant

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		grant := &s3.TargetGrant{}

		if v, ok := tfMap["grantee"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			grant.Grantee = expandBucketLoggingTargetGrantGrantee(v)
		}

		if v, ok := tfMap["permission"].(string); ok && v != "" {
			grant.Permission = aws.String(v)
		}

		grants = append(grants, grant)
	}

	return grants
}

func expandBucketLoggingTargetGrantGrantee(l []interface{}) *s3.Grantee {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	grantee := &s3.Grantee{}

	if v, ok := tfMap["display_name"].(string); ok && v != "" {
		grantee.DisplayName = aws.String(v)
	}

	if v, ok := tfMap["email_address"].(string); ok && v != "" {
		grantee.EmailAddress = aws.String(v)
	}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		grantee.ID = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		grantee.Type = aws.String(v)
	}

	if v, ok := tfMap["uri"].(string); ok && v != "" {
		grantee.URI = aws.String(v)
	}

	return grantee
}

func flattenBucketLoggingTargetGrants(grants []*s3.TargetGrant) []interface{} {
	var results []interface{}

	for _, grant := range grants {
		if grant == nil {
			continue
		}

		m := make(map[string]interface{})

		if grant.Grantee != nil {
			m["grantee"] = flattenBucketLoggingTargetGrantGrantee(grant.Grantee)
		}

		if grant.Permission != nil {
			m["permission"] = aws.StringValue(grant.Permission)
		}

		results = append(results, m)
	}

	return results
}

func flattenBucketLoggingTargetGrantGrantee(g *s3.Grantee) []interface{} {
	if g == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if g.DisplayName != nil {
		m["display_name"] = aws.StringValue(g.DisplayName)
	}

	if g.EmailAddress != nil {
		m["email_address"] = aws.StringValue(g.EmailAddress)
	}

	if g.ID != nil {
		m["id"] = aws.StringValue(g.ID)
	}

	if g.Type != nil {
		m["type"] = aws.StringValue(g.Type)
	}

	if g.URI != nil {
		m["uri"] = aws.StringValue(g.URI)
	}

	return []interface{}{m}
}
