// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_website_configuration", name="Bucket Website Configuration")
func resourceBucketWebsiteConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketWebsiteConfigurationCreate,
		ReadWithoutTimeout:   resourceBucketWebsiteConfigurationRead,
		UpdateWithoutTimeout: resourceBucketWebsiteConfigurationUpdate,
		DeleteWithoutTimeout: resourceBucketWebsiteConfigurationDelete,

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
			"error_document": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrExpectedBucketOwner: {
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
						names.AttrProtocol: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.Protocol](),
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
						names.AttrCondition: {
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
									names.AttrProtocol: {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.Protocol](),
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
			"website_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"website_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceBucketWebsiteConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	websiteConfig := &types.WebsiteConfiguration{}

	if v, ok := d.GetOk("error_document"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		websiteConfig.ErrorDocument = expandErrorDocument(v.([]interface{}))
	}

	if v, ok := d.GetOk("index_document"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		websiteConfig.IndexDocument = expandIndexDocument(v.([]interface{}))
	}

	if v, ok := d.GetOk("redirect_all_requests_to"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		websiteConfig.RedirectAllRequestsTo = expandRedirectAllRequestsTo(v.([]interface{}))
	}

	if v, ok := d.GetOk("routing_rule"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		websiteConfig.RoutingRules = expandRoutingRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("routing_rules"); ok {
		var unmarshalledRules []types.RoutingRule
		if err := json.Unmarshal([]byte(v.(string)), &unmarshalledRules); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		websiteConfig.RoutingRules = unmarshalledRules
	}

	bucket := d.Get(names.AttrBucket).(string)
	expectedBucketOwner := d.Get(names.AttrExpectedBucketOwner).(string)
	input := &s3.PutBucketWebsiteInput{
		Bucket:               aws.String(bucket),
		WebsiteConfiguration: websiteConfig,
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutBucketWebsite(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "WebsiteConfiguration is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Website Configuration: %s", bucket, err)
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findBucketWebsite(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Website Configuration (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBucketWebsiteConfigurationRead(ctx, d, meta)...)
}

func resourceBucketWebsiteConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findBucketWebsite(ctx, conn, bucket, expectedBucketOwner)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Website Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Website Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, bucket)
	if err := d.Set("error_document", flattenErrorDocument(output.ErrorDocument)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting error_document: %s", err)
	}
	d.Set(names.AttrExpectedBucketOwner, expectedBucketOwner)
	if err := d.Set("index_document", flattenIndexDocument(output.IndexDocument)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting index_document: %s", err)
	}
	if err := d.Set("redirect_all_requests_to", flattenRedirectAllRequestsTo(output.RedirectAllRequestsTo)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting redirect_all_requests_to: %s", err)
	}
	if err := d.Set("routing_rule", flattenRoutingRules(output.RoutingRules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting routing_rule: %s", err)
	}
	if output.RoutingRules != nil {
		rr, err := normalizeRoutingRules(output.RoutingRules)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		d.Set("routing_rules", rr)
	} else {
		d.Set("routing_rules", nil)
	}

	if output, err := findBucketLocation(ctx, conn, bucket, expectedBucketOwner); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket (%s) Location: %s", d.Id(), err)
	} else {
		endpoint, domain := bucketWebsiteEndpointAndDomain(bucket, string(output.LocationConstraint))
		d.Set("website_domain", domain)
		d.Set("website_endpoint", endpoint)
	}

	return diags
}

func resourceBucketWebsiteConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	websiteConfig := &types.WebsiteConfiguration{}

	if v, ok := d.GetOk("error_document"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		websiteConfig.ErrorDocument = expandErrorDocument(v.([]interface{}))
	}

	if v, ok := d.GetOk("index_document"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		websiteConfig.IndexDocument = expandIndexDocument(v.([]interface{}))
	}

	if v, ok := d.GetOk("redirect_all_requests_to"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		websiteConfig.RedirectAllRequestsTo = expandRedirectAllRequestsTo(v.([]interface{}))
	}

	if d.HasChanges("routing_rule", "routing_rules") {
		if d.HasChange("routing_rule") {
			websiteConfig.RoutingRules = expandRoutingRules(d.Get("routing_rule").([]interface{}))
		} else {
			var unmarshalledRules []types.RoutingRule
			if err := json.Unmarshal([]byte(d.Get("routing_rules").(string)), &unmarshalledRules); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
			websiteConfig.RoutingRules = unmarshalledRules
		}
	} else {
		// Still send the current RoutingRules configuration
		if v, ok := d.GetOk("routing_rule"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			websiteConfig.RoutingRules = expandRoutingRules(v.([]interface{}))
		}

		if v, ok := d.GetOk("routing_rules"); ok {
			var unmarshalledRules []types.RoutingRule
			if err := json.Unmarshal([]byte(v.(string)), &unmarshalledRules); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
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

	_, err = conn.PutBucketWebsite(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Bucket Website Configuration (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBucketWebsiteConfigurationRead(ctx, d, meta)...)
}

func resourceBucketWebsiteConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3.DeleteBucketWebsiteInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.DeleteBucketWebsite(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchWebsiteConfiguration) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Website Configuration (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findBucketWebsite(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Website Configuration (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findBucketWebsite(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) (*s3.GetBucketWebsiteOutput, error) {
	input := &s3.GetBucketWebsiteInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketWebsite(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchWebsiteConfiguration) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findBucketLocation(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) (*s3.GetBucketLocationOutput, error) {
	input := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketLocation(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandErrorDocument(l []interface{}) *types.ErrorDocument {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.ErrorDocument{}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		result.Key = aws.String(v)
	}

	return result
}

func expandIndexDocument(l []interface{}) *types.IndexDocument {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.IndexDocument{}

	if v, ok := tfMap["suffix"].(string); ok && v != "" {
		result.Suffix = aws.String(v)
	}

	return result
}

func expandRedirectAllRequestsTo(l []interface{}) *types.RedirectAllRequestsTo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.RedirectAllRequestsTo{}

	if v, ok := tfMap["host_name"].(string); ok && v != "" {
		result.HostName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		result.Protocol = types.Protocol(v)
	}

	return result
}

func expandRoutingRules(l []interface{}) []types.RoutingRule {
	var results []types.RoutingRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rule := types.RoutingRule{}

		if v, ok := tfMap[names.AttrCondition].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.Condition = expandCondition(v)
		}

		if v, ok := tfMap["redirect"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.Redirect = expandRedirect(v)
		}

		results = append(results, rule)
	}

	return results
}

func expandCondition(l []interface{}) *types.Condition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.Condition{}

	if v, ok := tfMap["http_error_code_returned_equals"].(string); ok && v != "" {
		result.HttpErrorCodeReturnedEquals = aws.String(v)
	}

	if v, ok := tfMap["key_prefix_equals"].(string); ok && v != "" {
		result.KeyPrefixEquals = aws.String(v)
	}

	return result
}

func expandRedirect(l []interface{}) *types.Redirect {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.Redirect{}

	if v, ok := tfMap["host_name"].(string); ok && v != "" {
		result.HostName = aws.String(v)
	}

	if v, ok := tfMap["http_redirect_code"].(string); ok && v != "" {
		result.HttpRedirectCode = aws.String(v)
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok && v != "" {
		result.Protocol = types.Protocol(v)
	}

	if v, ok := tfMap["replace_key_prefix_with"].(string); ok && v != "" {
		result.ReplaceKeyPrefixWith = aws.String(v)
	}

	if v, ok := tfMap["replace_key_with"].(string); ok && v != "" {
		result.ReplaceKeyWith = aws.String(v)
	}

	return result
}

func flattenIndexDocument(i *types.IndexDocument) []interface{} {
	if i == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if i.Suffix != nil {
		m["suffix"] = aws.ToString(i.Suffix)
	}

	return []interface{}{m}
}

func flattenErrorDocument(e *types.ErrorDocument) []interface{} {
	if e == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if e.Key != nil {
		m[names.AttrKey] = aws.ToString(e.Key)
	}

	return []interface{}{m}
}

func flattenRedirectAllRequestsTo(r *types.RedirectAllRequestsTo) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrProtocol: string(r.Protocol),
	}

	if r.HostName != nil {
		m["host_name"] = aws.ToString(r.HostName)
	}

	return []interface{}{m}
}

func flattenRoutingRules(rules []types.RoutingRule) []interface{} {
	var results []interface{}

	for _, rule := range rules {
		m := make(map[string]interface{})

		if rule.Condition != nil {
			m[names.AttrCondition] = flattenCondition(rule.Condition)
		}

		if rule.Redirect != nil {
			m["redirect"] = flattenRedirect(rule.Redirect)
		}

		results = append(results, m)
	}

	return results
}

func flattenCondition(c *types.Condition) []interface{} {
	if c == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if c.KeyPrefixEquals != nil {
		m["key_prefix_equals"] = aws.ToString(c.KeyPrefixEquals)
	}

	if c.HttpErrorCodeReturnedEquals != nil {
		m["http_error_code_returned_equals"] = aws.ToString(c.HttpErrorCodeReturnedEquals)
	}

	return []interface{}{m}
}

func flattenRedirect(r *types.Redirect) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrProtocol: string(r.Protocol),
	}

	if r.HostName != nil {
		m["host_name"] = aws.ToString(r.HostName)
	}

	if r.HttpRedirectCode != nil {
		m["http_redirect_code"] = aws.ToString(r.HttpRedirectCode)
	}

	if r.ReplaceKeyWith != nil {
		m["replace_key_with"] = aws.ToString(r.ReplaceKeyWith)
	}

	if r.ReplaceKeyPrefixWith != nil {
		m["replace_key_prefix_with"] = aws.ToString(r.ReplaceKeyPrefixWith)
	}

	return []interface{}{m}
}

func normalizeRoutingRules(w []types.RoutingRule) (string, error) {
	withNulls, err := json.Marshal(w)
	if err != nil {
		return "", err
	}

	var rules []map[string]interface{}
	if err := json.Unmarshal(withNulls, &rules); err != nil {
		return "", err
	}

	var cleanRules []map[string]interface{}
	for _, rule := range rules {
		cleanRules = append(cleanRules, removeNilOrEmptyProtocol(rule))
	}

	withoutNulls, err := json.Marshal(cleanRules)
	if err != nil {
		return "", err
	}

	return string(withoutNulls), nil
}

// removeNilOrEmptyProtocol removes nils and empty ("") Protocol values from a RoutingRule JSON document.
func removeNilOrEmptyProtocol(data map[string]interface{}) map[string]interface{} {
	withoutNil := make(map[string]interface{})

	for k, v := range data {
		if v == nil {
			continue
		}

		switch v := v.(type) {
		case map[string]interface{}:
			withoutNil[k] = removeNilOrEmptyProtocol(v)
		case string:
			// With AWS SDK for Go v2 Protocol changed type from *string to types.Protocol.
			// An empty ("") value is equivalent to nil.
			if k == "Protocol" && v == "" {
				continue
			}
			withoutNil[k] = v
		default:
			withoutNil[k] = v
		}
	}

	return withoutNil
}
