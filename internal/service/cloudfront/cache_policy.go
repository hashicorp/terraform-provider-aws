package cloudfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceCachePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCachePolicyCreate,
		ReadWithoutTimeout:   resourceCachePolicyRead,
		UpdateWithoutTimeout: resourceCachePolicyUpdate,
		DeleteWithoutTimeout: resourceCachePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"default_ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  86400,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"max_ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  31536000,
			},
			"min_ttl": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"parameters_in_cache_key_and_forwarded_to_origin": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cookies_config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cookie_behavior": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cloudfront.CachePolicyCookieBehavior_Values(), false),
									},
									"cookies": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
								},
							},
						},
						"enable_accept_encoding_brotli": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"enable_accept_encoding_gzip": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"headers_config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"header_behavior": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(cloudfront.CachePolicyHeaderBehavior_Values(), false),
									},
									"headers": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
								},
							},
						},
						"query_strings_config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"query_string_behavior": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cloudfront.CachePolicyQueryStringBehavior_Values(), false),
									},
									"query_strings": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceCachePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	name := d.Get("name").(string)
	apiObject := &cloudfront.CachePolicyConfig{
		DefaultTTL: aws.Int64(int64(d.Get("default_ttl").(int))),
		MaxTTL:     aws.Int64(int64(d.Get("max_ttl").(int))),
		MinTTL:     aws.Int64(int64(d.Get("min_ttl").(int))),
		Name:       aws.String(name),
	}

	if v, ok := d.GetOk("comment"); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters_in_cache_key_and_forwarded_to_origin"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.ParametersInCacheKeyAndForwardedToOrigin = expandParametersInCacheKeyAndForwardedToOrigin(v.([]interface{})[0].(map[string]interface{}))
	}

	input := &cloudfront.CreateCachePolicyInput{
		CachePolicyConfig: apiObject,
	}

	log.Printf("[DEBUG] Creating CloudFront Cache Policy: (%s)", input)
	output, err := conn.CreateCachePolicyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Cache Policy (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.CachePolicy.Id))

	return append(diags, resourceCachePolicyRead(ctx, d, meta)...)
}

func resourceCachePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	output, err := FindCachePolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Cache Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Cache Policy (%s): %s", d.Id(), err)
	}

	apiObject := output.CachePolicy.CachePolicyConfig
	d.Set("comment", apiObject.Comment)
	d.Set("default_ttl", apiObject.DefaultTTL)
	d.Set("etag", output.ETag)
	d.Set("max_ttl", apiObject.MaxTTL)
	d.Set("min_ttl", apiObject.MinTTL)
	d.Set("name", apiObject.Name)
	if apiObject.ParametersInCacheKeyAndForwardedToOrigin != nil {
		if err := d.Set("parameters_in_cache_key_and_forwarded_to_origin", []interface{}{flattenParametersInCacheKeyAndForwardedToOrigin(apiObject.ParametersInCacheKeyAndForwardedToOrigin)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting parameters_in_cache_key_and_forwarded_to_origin: %s", err)
		}
	} else {
		d.Set("parameters_in_cache_key_and_forwarded_to_origin", nil)
	}

	return diags
}

func resourceCachePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	//
	// https://docs.aws.amazon.com/cloudfront/latest/APIReference/API_UpdateCachePolicy.html:
	// "When you update a cache policy configuration, all the fields are updated with the values provided in the request. You cannot update some fields independent of others."
	//
	apiObject := &cloudfront.CachePolicyConfig{
		DefaultTTL: aws.Int64(int64(d.Get("default_ttl").(int))),
		MaxTTL:     aws.Int64(int64(d.Get("max_ttl").(int))),
		MinTTL:     aws.Int64(int64(d.Get("min_ttl").(int))),
		Name:       aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("comment"); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters_in_cache_key_and_forwarded_to_origin"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.ParametersInCacheKeyAndForwardedToOrigin = expandParametersInCacheKeyAndForwardedToOrigin(v.([]interface{})[0].(map[string]interface{}))
	}

	input := &cloudfront.UpdateCachePolicyInput{
		CachePolicyConfig: apiObject,
		Id:                aws.String(d.Id()),
		IfMatch:           aws.String(d.Get("etag").(string)),
	}

	log.Printf("[DEBUG] Updating CloudFront Cache Policy: (%s)", input)
	_, err := conn.UpdateCachePolicyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Cache Policy (%s): %s", d.Id(), err)
	}

	return append(diags, resourceCachePolicyRead(ctx, d, meta)...)
}

func resourceCachePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	log.Printf("[DEBUG] Deleting CloudFront Cache Policy: (%s)", d.Id())
	_, err := conn.DeleteCachePolicyWithContext(ctx, &cloudfront.DeleteCachePolicyInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	})

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchCachePolicy) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Cache Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func expandParametersInCacheKeyAndForwardedToOrigin(tfMap map[string]interface{}) *cloudfront.ParametersInCacheKeyAndForwardedToOrigin {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ParametersInCacheKeyAndForwardedToOrigin{}

	if v, ok := tfMap["cookies_config"].([]interface{}); ok && len(v) > 0 {
		apiObject.CookiesConfig = expandCachePolicyCookiesConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["enable_accept_encoding_brotli"].(bool); ok {
		apiObject.EnableAcceptEncodingBrotli = aws.Bool(v)
	}

	if v, ok := tfMap["enable_accept_encoding_gzip"].(bool); ok {
		apiObject.EnableAcceptEncodingGzip = aws.Bool(v)
	}

	if v, ok := tfMap["headers_config"].([]interface{}); ok && len(v) > 0 {
		apiObject.HeadersConfig = expandCachePolicyHeadersConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["query_strings_config"].([]interface{}); ok && len(v) > 0 {
		apiObject.QueryStringsConfig = expandCachePolicyQueryStringsConfig(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCachePolicyCookiesConfig(tfMap map[string]interface{}) *cloudfront.CachePolicyCookiesConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.CachePolicyCookiesConfig{}

	if v, ok := tfMap["cookie_behavior"].(string); ok && v != "" {
		apiObject.CookieBehavior = aws.String(v)
	}

	if v, ok := tfMap["cookies"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Cookies = expandCookieNames(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCookieNames(tfMap map[string]interface{}) *cloudfront.CookieNames {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.CookieNames{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := flex.ExpandStringSet(v)
		apiObject.Items = items
		apiObject.Quantity = aws.Int64(int64(len(items)))
	}

	return apiObject
}

func expandCachePolicyHeadersConfig(tfMap map[string]interface{}) *cloudfront.CachePolicyHeadersConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.CachePolicyHeadersConfig{}

	if v, ok := tfMap["header_behavior"].(string); ok && v != "" {
		apiObject.HeaderBehavior = aws.String(v)
	}

	if v, ok := tfMap["headers"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Headers = expandHeaders(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandHeaders(tfMap map[string]interface{}) *cloudfront.Headers {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.Headers{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := flex.ExpandStringSet(v)
		apiObject.Items = items
		apiObject.Quantity = aws.Int64(int64(len(items)))
	}

	return apiObject
}

func expandCachePolicyQueryStringsConfig(tfMap map[string]interface{}) *cloudfront.CachePolicyQueryStringsConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.CachePolicyQueryStringsConfig{}

	if v, ok := tfMap["query_string_behavior"].(string); ok && v != "" {
		apiObject.QueryStringBehavior = aws.String(v)
	}

	if v, ok := tfMap["query_strings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.QueryStrings = expandQueryStringNames(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandQueryStringNames(tfMap map[string]interface{}) *cloudfront.QueryStringNames {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.QueryStringNames{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := flex.ExpandStringSet(v)
		apiObject.Items = items
		apiObject.Quantity = aws.Int64(int64(len(items)))
	}

	return apiObject
}

func flattenParametersInCacheKeyAndForwardedToOrigin(apiObject *cloudfront.ParametersInCacheKeyAndForwardedToOrigin) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := flattenCachePolicyCookiesConfig(apiObject.CookiesConfig); len(v) > 0 {
		tfMap["cookies_config"] = []interface{}{v}
	}

	if v := apiObject.EnableAcceptEncodingBrotli; v != nil {
		tfMap["enable_accept_encoding_brotli"] = aws.BoolValue(v)
	}

	if v := apiObject.EnableAcceptEncodingGzip; v != nil {
		tfMap["enable_accept_encoding_gzip"] = aws.BoolValue(v)
	}

	if v := flattenCachePolicyHeadersConfig(apiObject.HeadersConfig); len(v) > 0 {
		tfMap["headers_config"] = []interface{}{v}
	}

	if v := flattenCachePolicyQueryStringsConfig(apiObject.QueryStringsConfig); len(v) > 0 {
		tfMap["query_strings_config"] = []interface{}{v}
	}

	return tfMap
}

func flattenCachePolicyCookiesConfig(apiObject *cloudfront.CachePolicyCookiesConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CookieBehavior; v != nil {
		tfMap["cookie_behavior"] = aws.StringValue(v)
	}

	if v := flattenCookieNames(apiObject.Cookies); len(v) > 0 {
		tfMap["cookies"] = []interface{}{v}
	}

	return tfMap
}

func flattenCookieNames(apiObject *cloudfront.CookieNames) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = aws.StringValueSlice(v)
	}

	return tfMap
}

func flattenCachePolicyHeadersConfig(apiObject *cloudfront.CachePolicyHeadersConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HeaderBehavior; v != nil {
		tfMap["header_behavior"] = aws.StringValue(v)
	}

	if v := flattenHeaders(apiObject.Headers); len(v) > 0 {
		tfMap["headers"] = []interface{}{v}
	}

	return tfMap
}

func flattenHeaders(apiObject *cloudfront.Headers) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = aws.StringValueSlice(v)
	}

	return tfMap
}

func flattenCachePolicyQueryStringsConfig(apiObject *cloudfront.CachePolicyQueryStringsConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.QueryStringBehavior; v != nil {
		tfMap["query_string_behavior"] = aws.StringValue(v)
	}

	if v := flattenQueryStringNames(apiObject.QueryStrings); len(v) > 0 {
		tfMap["query_strings"] = []interface{}{v}
	}

	return tfMap
}

func flattenQueryStringNames(apiObject *cloudfront.QueryStringNames) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = aws.StringValueSlice(v)
	}

	return tfMap
}
