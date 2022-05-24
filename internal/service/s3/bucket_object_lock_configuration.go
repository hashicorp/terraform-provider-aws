package s3

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBucketObjectLockConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBucketObjectLockConfigurationCreate,
		ReadContext:   resourceBucketObjectLockConfigurationRead,
		UpdateContext: resourceBucketObjectLockConfigurationUpdate,
		DeleteContext: resourceBucketObjectLockConfigurationDelete,
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
				Type:         schema.TypeString,
				Optional:     true,
				Default:      s3.ObjectLockEnabledEnabled,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(s3.ObjectLockEnabled_Values(), false),
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
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
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(s3.ObjectLockRetentionMode_Values(), false),
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
	conn := meta.(*conns.AWSClient).S3Conn

	bucket := d.Get("bucket").(string)
	expectedBucketOwner := d.Get("expected_bucket_owner").(string)

	input := &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &s3.ObjectLockConfiguration{
			// ObjectLockEnabled is required by the API, even if configured directly on the S3 bucket
			// during creation, else a MalformedXML error will be returned.
			ObjectLockEnabled: aws.String(d.Get("object_lock_enabled").(string)),
			Rule:              expandBucketObjectLockConfigurationRule(d.Get("rule").([]interface{})),
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("request_payer"); ok {
		input.RequestPayer = aws.String(v.(string))
	}

	// For existing buckets
	if v, ok := d.GetOk("token"); ok {
		input.Token = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutObjectLockConfigurationWithContext(ctx, input)
	}, s3.ErrCodeNoSuchBucket)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating S3 bucket (%s) Object Lock configuration: %w", bucket, err))
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	return resourceBucketObjectLockConfigurationRead(ctx, d, meta)
}

func resourceBucketObjectLockConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.GetObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetObjectLockConfigurationWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, ErrCodeObjectLockConfigurationNotFound) {
		log.Printf("[WARN] S3 Bucket Object Lock Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading S3 bucket Object Lock configuration (%s): %w", d.Id(), err))
	}

	if output == nil || output.ObjectLockConfiguration == nil {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading S3 bucket Object Lock configuration (%s): empty output", d.Id()))
		}
		log.Printf("[WARN] S3 Bucket Object Lock Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	objLockConfig := output.ObjectLockConfiguration

	d.Set("bucket", bucket)
	d.Set("expected_bucket_owner", expectedBucketOwner)
	d.Set("object_lock_enabled", objLockConfig.ObjectLockEnabled)

	if err := d.Set("rule", flattenBucketObjectLockConfigurationRule(objLockConfig.Rule)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting rule: %w", err))
	}

	return nil
}

func resourceBucketObjectLockConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &s3.ObjectLockConfiguration{
			// ObjectLockEnabled is required by the API, even if configured directly on the S3 bucket
			// during creation, else a MalformedXML error will be returned.
			ObjectLockEnabled: aws.String(d.Get("object_lock_enabled").(string)),
			Rule:              expandBucketObjectLockConfigurationRule(d.Get("rule").([]interface{})),
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("request_payer"); ok {
		input.RequestPayer = aws.String(v.(string))
	}

	if v, ok := d.GetOk("token"); ok {
		input.Token = aws.String(v.(string))
	}

	_, err = conn.PutObjectLockConfigurationWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating S3 bucket Object Lock configuration (%s): %w", d.Id(), err))
	}

	return resourceBucketObjectLockConfigurationRead(ctx, d, meta)
}

func resourceBucketObjectLockConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &s3.ObjectLockConfiguration{
			// ObjectLockEnabled is required by the API, even if configured directly on the S3 bucket
			// during creation, else a MalformedXML error will be returned.
			ObjectLockEnabled: aws.String(d.Get("object_lock_enabled").(string)),
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.PutObjectLockConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, ErrCodeObjectLockConfigurationNotFound) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting S3 bucket Object Lock configuration (%s): %w", d.Id(), err))
	}

	return nil
}

func expandBucketObjectLockConfigurationRule(l []interface{}) *s3.ObjectLockRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	rule := &s3.ObjectLockRule{}

	if v, ok := tfMap["default_retention"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		rule.DefaultRetention = expandBucketObjectLockConfigurationCorsRuleDefaultRetention(v)
	}

	return rule
}

func expandBucketObjectLockConfigurationCorsRuleDefaultRetention(l []interface{}) *s3.DefaultRetention {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	dr := &s3.DefaultRetention{}

	if v, ok := tfMap["days"].(int); ok && v > 0 {
		dr.Days = aws.Int64(int64(v))
	}

	if v, ok := tfMap["mode"].(string); ok && v != "" {
		dr.Mode = aws.String(v)
	}

	if v, ok := tfMap["years"].(int); ok && v > 0 {
		dr.Years = aws.Int64(int64(v))
	}

	return dr
}

func flattenBucketObjectLockConfigurationRule(rule *s3.ObjectLockRule) []interface{} {
	if rule == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if rule.DefaultRetention != nil {
		m["default_retention"] = flattenBucketObjectLockConfigurationRuleDefaultRetention(rule.DefaultRetention)
	}
	return []interface{}{m}
}

func flattenBucketObjectLockConfigurationRuleDefaultRetention(dr *s3.DefaultRetention) []interface{} {
	if dr == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if dr.Days != nil {
		m["days"] = int(aws.Int64Value(dr.Days))
	}

	if dr.Mode != nil {
		m["mode"] = aws.StringValue(dr.Mode)
	}

	if dr.Years != nil {
		m["years"] = int(aws.Int64Value(dr.Years))
	}

	return []interface{}{m}
}
