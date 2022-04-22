package s3

import (
	"context"
	"fmt"
	"log"

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

func ResourceBucketServerSideEncryptionConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBucketServerSideEncryptionConfigurationCreate,
		ReadContext:   resourceBucketServerSideEncryptionConfigurationRead,
		UpdateContext: resourceBucketServerSideEncryptionConfigurationUpdate,
		DeleteContext: resourceBucketServerSideEncryptionConfigurationDelete,
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
			"rule": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"apply_server_side_encryption_by_default": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"kms_master_key_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"sse_algorithm": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(s3.ServerSideEncryption_Values(), false),
									},
								},
							},
						},
						"bucket_key_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceBucketServerSideEncryptionConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket := d.Get("bucket").(string)
	expectedBucketOwner := d.Get("expected_bucket_owner").(string)

	input := &s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucket),
		ServerSideEncryptionConfiguration: &s3.ServerSideEncryptionConfiguration{
			Rules: expandBucketServerSideEncryptionConfigurationRules(d.Get("rule").(*schema.Set).List()),
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(
		propagationTimeout,
		func() (interface{}, error) {
			return conn.PutBucketEncryptionWithContext(ctx, input)
		},
		s3.ErrCodeNoSuchBucket,
		ErrCodeOperationAborted,
	)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating S3 bucket (%s) server-side encryption configuration: %w", bucket, err))
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	return resourceBucketServerSideEncryptionConfigurationRead(ctx, d, meta)
}

func resourceBucketServerSideEncryptionConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucket),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	resp, err := tfresource.RetryWhenAWSErrCodeEquals(
		propagationTimeout,
		func() (interface{}, error) {
			return conn.GetBucketEncryptionWithContext(ctx, input)
		},
		s3.ErrCodeNoSuchBucket,
		ErrCodeServerSideEncryptionConfigurationNotFound,
	)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, ErrCodeServerSideEncryptionConfigurationNotFound) {
		log.Printf("[WARN] S3 Bucket Server-Side Encryption Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading S3 bucket server-side encryption configuration (%s): %w", d.Id(), err))
	}

	output, ok := resp.(*s3.GetBucketEncryptionOutput)
	if !ok || output.ServerSideEncryptionConfiguration == nil {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading S3 bucket server-side encryption configuration (%s): empty output", d.Id()))
		}
		log.Printf("[WARN] S3 Bucket Server-Side Encryption Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	sse := output.ServerSideEncryptionConfiguration

	d.Set("bucket", bucket)
	d.Set("expected_bucket_owner", expectedBucketOwner)
	if err := d.Set("rule", flattenBucketServerSideEncryptionConfigurationRules(sse.Rules)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting rule: %w", err))
	}

	return nil
}

func resourceBucketServerSideEncryptionConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucket),
		ServerSideEncryptionConfiguration: &s3.ServerSideEncryptionConfiguration{
			Rules: expandBucketServerSideEncryptionConfigurationRules(d.Get("rule").(*schema.Set).List()),
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(
		propagationTimeout,
		func() (interface{}, error) {
			return conn.PutBucketEncryptionWithContext(ctx, input)
		},
		s3.ErrCodeNoSuchBucket,
		ErrCodeOperationAborted,
	)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating S3 bucket (%s) server-side encryption configuration: %w", d.Id(), err))
	}

	return resourceBucketServerSideEncryptionConfigurationRead(ctx, d, meta)
}

func resourceBucketServerSideEncryptionConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.DeleteBucketEncryptionInput{
		Bucket: aws.String(bucket),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.DeleteBucketEncryptionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, ErrCodeServerSideEncryptionConfigurationNotFound) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting S3 bucket server-side encryption configuration (%s): %w", d.Id(), err))
	}

	return nil
}

func expandBucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefault(l []interface{}) *s3.ServerSideEncryptionByDefault {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	sse := &s3.ServerSideEncryptionByDefault{}

	if v, ok := tfMap["kms_master_key_id"].(string); ok && v != "" {
		sse.KMSMasterKeyID = aws.String(v)
	}

	if v, ok := tfMap["sse_algorithm"].(string); ok && v != "" {
		sse.SSEAlgorithm = aws.String(v)
	}

	return sse
}

func expandBucketServerSideEncryptionConfigurationRules(l []interface{}) []*s3.ServerSideEncryptionRule {
	var rules []*s3.ServerSideEncryptionRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rule := &s3.ServerSideEncryptionRule{}

		if v, ok := tfMap["apply_server_side_encryption_by_default"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.ApplyServerSideEncryptionByDefault = expandBucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefault(v)
		}

		if v, ok := tfMap["bucket_key_enabled"].(bool); ok {
			rule.BucketKeyEnabled = aws.Bool(v)
		}
		rules = append(rules, rule)
	}

	return rules
}

func flattenBucketServerSideEncryptionConfigurationRules(rules []*s3.ServerSideEncryptionRule) []interface{} {
	var results []interface{}

	for _, rule := range rules {
		if rule == nil {
			continue
		}

		m := make(map[string]interface{})

		if rule.ApplyServerSideEncryptionByDefault != nil {
			m["apply_server_side_encryption_by_default"] = flattenBucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefault(rule.ApplyServerSideEncryptionByDefault)
		}
		if rule.BucketKeyEnabled != nil {
			m["bucket_key_enabled"] = aws.BoolValue(rule.BucketKeyEnabled)
		}

		results = append(results, m)
	}

	return results
}

func flattenBucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefault(sse *s3.ServerSideEncryptionByDefault) []interface{} {
	if sse == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if sse.KMSMasterKeyID != nil {
		m["kms_master_key_id"] = aws.StringValue(sse.KMSMasterKeyID)
	}

	if sse.SSEAlgorithm != nil {
		m["sse_algorithm"] = aws.StringValue(sse.SSEAlgorithm)
	}

	return []interface{}{m}
}
