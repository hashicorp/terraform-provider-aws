// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudfront_origin_request_policy", name="Origin Request Policy")
func resourceOriginRequestPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOriginRequestPolicyCreate,
		ReadWithoutTimeout:   resourceOriginRequestPolicyRead,
		UpdateWithoutTimeout: resourceOriginRequestPolicyUpdate,
		DeleteWithoutTimeout: resourceOriginRequestPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrComment: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cookies_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cookie_behavior": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.OriginRequestPolicyCookieBehavior](),
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
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"headers_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"header_behavior": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.OriginRequestPolicyHeaderBehavior](),
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"query_strings_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query_string_behavior": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.OriginRequestPolicyQueryStringBehavior](),
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
	}
}

func resourceOriginRequestPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	name := d.Get(names.AttrName).(string)
	apiObject := &awstypes.OriginRequestPolicyConfig{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrComment); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cookies_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CookiesConfig = expandOriginRequestPolicyCookiesConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("headers_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.HeadersConfig = expandOriginRequestPolicyHeadersConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("query_strings_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.QueryStringsConfig = expandOriginRequestPolicyQueryStringsConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	input := &cloudfront.CreateOriginRequestPolicyInput{
		OriginRequestPolicyConfig: apiObject,
	}

	output, err := conn.CreateOriginRequestPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Origin Request Policy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.OriginRequestPolicy.Id))

	return append(diags, resourceOriginRequestPolicyRead(ctx, d, meta)...)
}

func resourceOriginRequestPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	output, err := findOriginRequestPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Origin Request Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Origin Request Policy (%s): %s", d.Id(), err)
	}

	apiObject := output.OriginRequestPolicy.OriginRequestPolicyConfig
	d.Set(names.AttrComment, apiObject.Comment)
	if apiObject.CookiesConfig != nil {
		if err := d.Set("cookies_config", []interface{}{flattenOriginRequestPolicyCookiesConfig(apiObject.CookiesConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting cookies_config: %s", err)
		}
	} else {
		d.Set("cookies_config", nil)
	}
	d.Set("etag", output.ETag)
	if apiObject.HeadersConfig != nil {
		if err := d.Set("headers_config", []interface{}{flattenOriginRequestPolicyHeadersConfig(apiObject.HeadersConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting headers_config: %s", err)
		}
	} else {
		d.Set("headers_config", nil)
	}
	d.Set(names.AttrName, apiObject.Name)
	if apiObject.QueryStringsConfig != nil {
		if err := d.Set("query_strings_config", []interface{}{flattenOriginRequestPolicyQueryStringsConfig(apiObject.QueryStringsConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting query_strings_config: %s", err)
		}
	} else {
		d.Set("query_strings_config", nil)
	}

	return diags
}

func resourceOriginRequestPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	//
	// https://docs.aws.amazon.com/cloudfront/latest/APIReference/API_UpdateOriginRequestPolicy.html:
	// "When you update an origin request policy configuration, all the fields are updated with the values provided in the request. You cannot update some fields independent of others."
	//
	apiObject := &awstypes.OriginRequestPolicyConfig{
		Name: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk(names.AttrComment); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cookies_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CookiesConfig = expandOriginRequestPolicyCookiesConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("headers_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.HeadersConfig = expandOriginRequestPolicyHeadersConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("query_strings_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.QueryStringsConfig = expandOriginRequestPolicyQueryStringsConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	input := &cloudfront.UpdateOriginRequestPolicyInput{
		Id:                        aws.String(d.Id()),
		IfMatch:                   aws.String(d.Get("etag").(string)),
		OriginRequestPolicyConfig: apiObject,
	}

	_, err := conn.UpdateOriginRequestPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Origin Request Policy (%s): %s", d.Id(), err)
	}

	return append(diags, resourceOriginRequestPolicyRead(ctx, d, meta)...)
}

func resourceOriginRequestPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	log.Printf("[DEBUG] Deleting CloudFront Origin Request Policy: %s", d.Id())
	_, err := conn.DeleteOriginRequestPolicy(ctx, &cloudfront.DeleteOriginRequestPolicyInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	})

	if errs.IsA[*awstypes.NoSuchOriginRequestPolicy](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Origin Request Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findOriginRequestPolicyByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetOriginRequestPolicyOutput, error) {
	input := &cloudfront.GetOriginRequestPolicyInput{
		Id: aws.String(id),
	}

	output, err := conn.GetOriginRequestPolicy(ctx, input)

	if errs.IsA[*awstypes.NoSuchOriginRequestPolicy](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.OriginRequestPolicy == nil || output.OriginRequestPolicy.OriginRequestPolicyConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandOriginRequestPolicyCookiesConfig(tfMap map[string]interface{}) *awstypes.OriginRequestPolicyCookiesConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OriginRequestPolicyCookiesConfig{}

	if v, ok := tfMap["cookie_behavior"].(string); ok && v != "" {
		apiObject.CookieBehavior = awstypes.OriginRequestPolicyCookieBehavior(v)
	}

	if v, ok := tfMap["cookies"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Cookies = expandCookieNames(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandOriginRequestPolicyHeadersConfig(tfMap map[string]interface{}) *awstypes.OriginRequestPolicyHeadersConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OriginRequestPolicyHeadersConfig{}

	if v, ok := tfMap["header_behavior"].(string); ok && v != "" {
		apiObject.HeaderBehavior = awstypes.OriginRequestPolicyHeaderBehavior(v)
	}

	if v, ok := tfMap["headers"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Headers = expandHeaders(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandOriginRequestPolicyQueryStringsConfig(tfMap map[string]interface{}) *awstypes.OriginRequestPolicyQueryStringsConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OriginRequestPolicyQueryStringsConfig{}

	if v, ok := tfMap["query_string_behavior"].(string); ok && v != "" {
		apiObject.QueryStringBehavior = awstypes.OriginRequestPolicyQueryStringBehavior(v)
	}

	if v, ok := tfMap["query_strings"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.QueryStrings = expandQueryStringNames(v[0].(map[string]interface{}))
	}

	return apiObject
}

func flattenOriginRequestPolicyCookiesConfig(apiObject *awstypes.OriginRequestPolicyCookiesConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"cookie_behavior": apiObject.CookieBehavior,
	}

	if v := flattenCookieNames(apiObject.Cookies); len(v) > 0 {
		tfMap["cookies"] = []interface{}{v}
	}

	return tfMap
}

func flattenOriginRequestPolicyHeadersConfig(apiObject *awstypes.OriginRequestPolicyHeadersConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"header_behavior": apiObject.HeaderBehavior,
	}

	if v := flattenHeaders(apiObject.Headers); len(v) > 0 {
		tfMap["headers"] = []interface{}{v}
	}

	return tfMap
}

func flattenOriginRequestPolicyQueryStringsConfig(apiObject *awstypes.OriginRequestPolicyQueryStringsConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"query_string_behavior": apiObject.QueryStringBehavior,
	}

	if v := flattenQueryStringNames(apiObject.QueryStrings); len(v) > 0 {
		tfMap["query_strings"] = []interface{}{v}
	}

	return tfMap
}
