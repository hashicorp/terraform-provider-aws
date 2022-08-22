package s3

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBucketWebsiteConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBucketWebsiteConfigurationCreate,
		ReadContext:   resourceBucketWebsiteConfigurationRead,
		UpdateContext: resourceBucketWebsiteConfigurationUpdate,
		DeleteContext: resourceBucketWebsiteConfigurationDelete,
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
			"error_document": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"expected_bucket_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"index_document": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"suffix": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"redirect_all_requests_to": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ConflictsWith: []string{
					"error_document",
					"index_document",
					"routing_rule",
					"routing_rules",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"protocol": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(s3.Protocol_Values(), false),
						},
					},
				},
			},
			"routing_rule": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"routing_rules"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"condition": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"http_error_code_returned_equals": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"key_prefix_equals": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"redirect": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"http_redirect_code": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"protocol": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(s3.Protocol_Values(), false),
									},
									"replace_key_prefix_with": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"replace_key_with": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"routing_rules": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"routing_rule"},
				ValidateFunc:  validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"website_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"website_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceBucketWebsiteConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket := d.Get("bucket").(string)
	expectedBucketOwner := d.Get("expected_bucket_owner").(string)

	websiteConfig := &s3.WebsiteConfiguration{}

	if v, ok := d.GetOk("error_document"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		websiteConfig.ErrorDocument = expandBucketWebsiteConfigurationErrorDocument(v.([]interface{}))
	}

	if v, ok := d.GetOk("index_document"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		websiteConfig.IndexDocument = expandBucketWebsiteConfigurationIndexDocument(v.([]interface{}))
	}

	if v, ok := d.GetOk("redirect_all_requests_to"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		websiteConfig.RedirectAllRequestsTo = expandBucketWebsiteConfigurationRedirectAllRequestsTo(v.([]interface{}))
	}

	if v, ok := d.GetOk("routing_rule"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		websiteConfig.RoutingRules = expandBucketWebsiteConfigurationRoutingRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("routing_rules"); ok {
		var unmarshalledRules []*s3.RoutingRule
		if err := json.Unmarshal([]byte(v.(string)), &unmarshalledRules); err != nil {
			return diag.FromErr(fmt.Errorf("error creating S3 Bucket (%s) website configuration: %w", bucket, err))
		}
		websiteConfig.RoutingRules = unmarshalledRules
	}

	input := &s3.PutBucketWebsiteInput{
		Bucket:               aws.String(bucket),
		WebsiteConfiguration: websiteConfig,
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketWebsiteWithContext(ctx, input)
	}, s3.ErrCodeNoSuchBucket)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating S3 bucket (%s) website configuration: %w", bucket, err))
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	return resourceBucketWebsiteConfigurationRead(ctx, d, meta)
}

func resourceBucketWebsiteConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.GetBucketWebsiteInput{
		Bucket: aws.String(bucket),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketWebsiteWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, ErrCodeNoSuchWebsiteConfiguration) {
		log.Printf("[WARN] S3 Bucket Website Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if output == nil {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading S3 bucket website configuration (%s): empty output", d.Id()))
		}
		log.Printf("[WARN] S3 Bucket Website Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("bucket", bucket)
	d.Set("expected_bucket_owner", expectedBucketOwner)

	if err := d.Set("error_document", flattenBucketWebsiteConfigurationErrorDocument(output.ErrorDocument)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting error_document: %w", err))
	}

	if err := d.Set("index_document", flattenBucketWebsiteConfigurationIndexDocument(output.IndexDocument)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting index_document: %w", err))
	}

	if err := d.Set("redirect_all_requests_to", flattenBucketWebsiteConfigurationRedirectAllRequestsTo(output.RedirectAllRequestsTo)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting redirect_all_requests_to: %w", err))
	}

	if err := d.Set("routing_rule", flattenBucketWebsiteConfigurationRoutingRules(output.RoutingRules)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting routing_rule: %w", err))
	}

	if output.RoutingRules != nil {
		rr, err := normalizeRoutingRules(output.RoutingRules)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error while marshaling routing rules: %w", err))
		}
		d.Set("routing_rules", rr)
	} else {
		d.Set("routing_rules", nil)
	}

	// Add website_endpoint and website_domain as attributes
	websiteEndpoint, err := resourceBucketWebsiteConfigurationWebsiteEndpoint(ctx, meta.(*conns.AWSClient), bucket, expectedBucketOwner)
	if err != nil {
		return diag.FromErr(err)
	}

	if websiteEndpoint != nil {
		d.Set("website_endpoint", websiteEndpoint.Endpoint)
		d.Set("website_domain", websiteEndpoint.Domain)
	}

	return nil
}

func resourceBucketWebsiteConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	websiteConfig := &s3.WebsiteConfiguration{}

	if v, ok := d.GetOk("error_document"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		websiteConfig.ErrorDocument = expandBucketWebsiteConfigurationErrorDocument(v.([]interface{}))
	}

	if v, ok := d.GetOk("index_document"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		websiteConfig.IndexDocument = expandBucketWebsiteConfigurationIndexDocument(v.([]interface{}))
	}

	if v, ok := d.GetOk("redirect_all_requests_to"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		websiteConfig.RedirectAllRequestsTo = expandBucketWebsiteConfigurationRedirectAllRequestsTo(v.([]interface{}))
	}

	if d.HasChanges("routing_rule", "routing_rules") {
		if d.HasChange("routing_rule") {
			websiteConfig.RoutingRules = expandBucketWebsiteConfigurationRoutingRules(d.Get("routing_rule").([]interface{}))
		} else {
			var unmarshalledRules []*s3.RoutingRule
			if err := json.Unmarshal([]byte(d.Get("routing_rules").(string)), &unmarshalledRules); err != nil {
				return diag.FromErr(fmt.Errorf("error updating S3 Bucket (%s) website configuration: %w", bucket, err))
			}
			websiteConfig.RoutingRules = unmarshalledRules
		}
	} else {
		// Still send the current RoutingRules configuration
		if v, ok := d.GetOk("routing_rule"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			websiteConfig.RoutingRules = expandBucketWebsiteConfigurationRoutingRules(v.([]interface{}))
		}

		if v, ok := d.GetOk("routing_rules"); ok {
			var unmarshalledRules []*s3.RoutingRule
			if err := json.Unmarshal([]byte(v.(string)), &unmarshalledRules); err != nil {
				return diag.FromErr(fmt.Errorf("error updating S3 Bucket (%s) website configuration: %w", bucket, err))
			}
			websiteConfig.RoutingRules = unmarshalledRules
		}
	}

	input := &s3.PutBucketWebsiteInput{
		Bucket:               aws.String(bucket),
		WebsiteConfiguration: websiteConfig,
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.PutBucketWebsiteWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating S3 bucket website configuration (%s): %w", d.Id(), err))
	}

	return resourceBucketWebsiteConfigurationRead(ctx, d, meta)
}

func resourceBucketWebsiteConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.DeleteBucketWebsiteInput{
		Bucket: aws.String(bucket),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.DeleteBucketWebsiteWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, ErrCodeNoSuchWebsiteConfiguration) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting S3 bucket website configuration (%s): %w", d.Id(), err))
	}

	return nil
}

func resourceBucketWebsiteConfigurationWebsiteEndpoint(ctx context.Context, client *conns.AWSClient, bucket, expectedBucketOwner string) (*S3Website, error) {
	conn := client.S3Conn

	input := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketLocationWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error getting S3 Bucket (%s) Location: %w", bucket, err)
	}

	var region string
	if output.LocationConstraint != nil {
		region = aws.StringValue(output.LocationConstraint)
	}

	return WebsiteEndpoint(client, bucket, region), nil
}

func expandBucketWebsiteConfigurationErrorDocument(l []interface{}) *s3.ErrorDocument {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &s3.ErrorDocument{}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		result.Key = aws.String(v)
	}

	return result
}

func expandBucketWebsiteConfigurationIndexDocument(l []interface{}) *s3.IndexDocument {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &s3.IndexDocument{}

	if v, ok := tfMap["suffix"].(string); ok && v != "" {
		result.Suffix = aws.String(v)
	}

	return result
}

func expandBucketWebsiteConfigurationRedirectAllRequestsTo(l []interface{}) *s3.RedirectAllRequestsTo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &s3.RedirectAllRequestsTo{}

	if v, ok := tfMap["host_name"].(string); ok && v != "" {
		result.HostName = aws.String(v)
	}

	if v, ok := tfMap["protocol"].(string); ok && v != "" {
		result.Protocol = aws.String(v)
	}

	return result
}

func expandBucketWebsiteConfigurationRoutingRules(l []interface{}) []*s3.RoutingRule {
	var results []*s3.RoutingRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rule := &s3.RoutingRule{}

		if v, ok := tfMap["condition"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.Condition = expandBucketWebsiteConfigurationRoutingRuleCondition(v)
		}

		if v, ok := tfMap["redirect"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.Redirect = expandBucketWebsiteConfigurationRoutingRuleRedirect(v)
		}

		results = append(results, rule)
	}

	return results
}

func expandBucketWebsiteConfigurationRoutingRuleCondition(l []interface{}) *s3.Condition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &s3.Condition{}

	if v, ok := tfMap["http_error_code_returned_equals"].(string); ok && v != "" {
		result.HttpErrorCodeReturnedEquals = aws.String(v)
	}

	if v, ok := tfMap["key_prefix_equals"].(string); ok && v != "" {
		result.KeyPrefixEquals = aws.String(v)
	}

	return result
}

func expandBucketWebsiteConfigurationRoutingRuleRedirect(l []interface{}) *s3.Redirect {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &s3.Redirect{}

	if v, ok := tfMap["host_name"].(string); ok && v != "" {
		result.HostName = aws.String(v)
	}

	if v, ok := tfMap["http_redirect_code"].(string); ok && v != "" {
		result.HttpRedirectCode = aws.String(v)
	}

	if v, ok := tfMap["protocol"].(string); ok && v != "" {
		result.Protocol = aws.String(v)
	}

	if v, ok := tfMap["replace_key_prefix_with"].(string); ok && v != "" {
		result.ReplaceKeyPrefixWith = aws.String(v)
	}

	if v, ok := tfMap["replace_key_with"].(string); ok && v != "" {
		result.ReplaceKeyWith = aws.String(v)
	}

	return result
}

func flattenBucketWebsiteConfigurationIndexDocument(i *s3.IndexDocument) []interface{} {
	if i == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if i.Suffix != nil {
		m["suffix"] = aws.StringValue(i.Suffix)
	}

	return []interface{}{m}
}

func flattenBucketWebsiteConfigurationErrorDocument(e *s3.ErrorDocument) []interface{} {
	if e == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if e.Key != nil {
		m["key"] = aws.StringValue(e.Key)
	}

	return []interface{}{m}
}

func flattenBucketWebsiteConfigurationRedirectAllRequestsTo(r *s3.RedirectAllRequestsTo) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if r.HostName != nil {
		m["host_name"] = aws.StringValue(r.HostName)
	}

	if r.Protocol != nil {
		m["protocol"] = aws.StringValue(r.Protocol)
	}

	return []interface{}{m}
}

func flattenBucketWebsiteConfigurationRoutingRules(rules []*s3.RoutingRule) []interface{} {
	var results []interface{}

	for _, rule := range rules {
		if rule == nil {
			continue
		}

		m := make(map[string]interface{})

		if rule.Condition != nil {
			m["condition"] = flattenBucketWebsiteConfigurationRoutingRuleCondition(rule.Condition)
		}

		if rule.Redirect != nil {
			m["redirect"] = flattenBucketWebsiteConfigurationRoutingRuleRedirect(rule.Redirect)
		}

		results = append(results, m)
	}

	return results
}

func flattenBucketWebsiteConfigurationRoutingRuleCondition(c *s3.Condition) []interface{} {
	if c == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if c.KeyPrefixEquals != nil {
		m["key_prefix_equals"] = aws.StringValue(c.KeyPrefixEquals)
	}

	if c.HttpErrorCodeReturnedEquals != nil {
		m["http_error_code_returned_equals"] = aws.StringValue(c.HttpErrorCodeReturnedEquals)
	}

	return []interface{}{m}
}

func flattenBucketWebsiteConfigurationRoutingRuleRedirect(r *s3.Redirect) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if r.HostName != nil {
		m["host_name"] = aws.StringValue(r.HostName)
	}

	if r.HttpRedirectCode != nil {
		m["http_redirect_code"] = aws.StringValue(r.HttpRedirectCode)
	}

	if r.Protocol != nil {
		m["protocol"] = aws.StringValue(r.Protocol)
	}

	if r.ReplaceKeyWith != nil {
		m["replace_key_with"] = aws.StringValue(r.ReplaceKeyWith)
	}

	if r.ReplaceKeyPrefixWith != nil {
		m["replace_key_prefix_with"] = aws.StringValue(r.ReplaceKeyPrefixWith)
	}

	return []interface{}{m}
}
