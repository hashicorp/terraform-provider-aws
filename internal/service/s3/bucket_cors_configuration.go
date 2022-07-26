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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBucketCorsConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBucketCorsConfigurationCreate,
		ReadContext:   resourceBucketCorsConfigurationRead,
		UpdateContext: resourceBucketCorsConfigurationUpdate,
		DeleteContext: resourceBucketCorsConfigurationDelete,
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
						"id": {
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
	conn := meta.(*conns.AWSClient).S3Conn

	bucket := d.Get("bucket").(string)
	expectedBucketOwner := d.Get("expected_bucket_owner").(string)

	input := &s3.PutBucketCorsInput{
		Bucket: aws.String(bucket),
		CORSConfiguration: &s3.CORSConfiguration{
			CORSRules: expandBucketCorsConfigurationCorsRules(d.Get("cors_rule").(*schema.Set).List()),
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketCorsWithContext(ctx, input)
	}, s3.ErrCodeNoSuchBucket)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating S3 bucket (%s) CORS configuration: %w", bucket, err))
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	return resourceBucketCorsConfigurationRead(ctx, d, meta)
}

func resourceBucketCorsConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.GetBucketCorsInput{
		Bucket: aws.String(bucket),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	corsResponse, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.GetBucketCorsWithContext(ctx, input)
	}, ErrCodeNoSuchCORSConfiguration)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, ErrCodeNoSuchCORSConfiguration) {
		log.Printf("[WARN] S3 Bucket CORS Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading S3 bucket CORS configuration (%s): %w", d.Id(), err))
	}

	output, ok := corsResponse.(*s3.GetBucketCorsOutput)
	if !ok || output == nil {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading S3 bucket CORS configuration (%s): empty output", d.Id()))
		}
		log.Printf("[WARN] S3 Bucket CORS Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("bucket", bucket)
	d.Set("expected_bucket_owner", expectedBucketOwner)

	if err := d.Set("cors_rule", flattenBucketCorsConfigurationCorsRules(output.CORSRules)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting cors_rule: %w", err))
	}

	return nil
}

func resourceBucketCorsConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutBucketCorsInput{
		Bucket: aws.String(bucket),
		CORSConfiguration: &s3.CORSConfiguration{
			CORSRules: expandBucketCorsConfigurationCorsRules(d.Get("cors_rule").(*schema.Set).List()),
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.PutBucketCorsWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating S3 bucket CORS configuration (%s): %w", d.Id(), err))
	}

	return resourceBucketCorsConfigurationRead(ctx, d, meta)
}

func resourceBucketCorsConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.DeleteBucketCorsInput{
		Bucket: aws.String(bucket),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.DeleteBucketCorsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting S3 bucket CORS configuration (%s): %w", d.Id(), err))
	}

	return nil
}

func expandBucketCorsConfigurationCorsRules(l []interface{}) []*s3.CORSRule {
	if len(l) == 0 {
		return nil
	}

	var rules []*s3.CORSRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rule := &s3.CORSRule{}

		if v, ok := tfMap["allowed_headers"].(*schema.Set); ok && v.Len() > 0 {
			rule.AllowedHeaders = flex.ExpandStringSet(v)
		}

		if v, ok := tfMap["allowed_methods"].(*schema.Set); ok && v.Len() > 0 {
			rule.AllowedMethods = flex.ExpandStringSet(v)
		}

		if v, ok := tfMap["allowed_origins"].(*schema.Set); ok && v.Len() > 0 {
			rule.AllowedOrigins = flex.ExpandStringSet(v)
		}

		if v, ok := tfMap["expose_headers"].(*schema.Set); ok && v.Len() > 0 {
			rule.ExposeHeaders = flex.ExpandStringSet(v)
		}

		if v, ok := tfMap["id"].(string); ok && v != "" {
			rule.ID = aws.String(v)
		}

		if v, ok := tfMap["max_age_seconds"].(int); ok {
			rule.MaxAgeSeconds = aws.Int64(int64(v))
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenBucketCorsConfigurationCorsRules(rules []*s3.CORSRule) []interface{} {
	var results []interface{}

	for _, rule := range rules {
		if rule == nil {
			continue
		}

		m := make(map[string]interface{})

		if len(rule.AllowedHeaders) > 0 {
			m["allowed_headers"] = flex.FlattenStringSet(rule.AllowedHeaders)
		}

		if len(rule.AllowedMethods) > 0 {
			m["allowed_methods"] = flex.FlattenStringSet(rule.AllowedMethods)
		}

		if len(rule.AllowedOrigins) > 0 {
			m["allowed_origins"] = flex.FlattenStringSet(rule.AllowedOrigins)
		}

		if len(rule.ExposeHeaders) > 0 {
			m["expose_headers"] = flex.FlattenStringSet(rule.ExposeHeaders)
		}

		if rule.ID != nil {
			m["id"] = aws.StringValue(rule.ID)
		}

		if rule.MaxAgeSeconds != nil {
			m["max_age_seconds"] = aws.Int64Value(rule.MaxAgeSeconds)
		}

		results = append(results, m)
	}

	return results
}
