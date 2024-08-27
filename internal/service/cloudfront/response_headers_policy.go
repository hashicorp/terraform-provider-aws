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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudfront_response_headers_policy", name="Response Headers Policy")
func resourceResponseHeadersPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResponseHeadersPolicyCreate,
		ReadWithoutTimeout:   resourceResponseHeadersPolicyRead,
		UpdateWithoutTimeout: resourceResponseHeadersPolicyUpdate,
		DeleteWithoutTimeout: resourceResponseHeadersPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrComment: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cors_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_control_allow_credentials": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"access_control_allow_headers": {
							Type:     schema.TypeList,
							Required: true,
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
						"access_control_allow_methods": {
							Type:     schema.TypeList,
							Required: true,
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
						"access_control_allow_origins": {
							Type:     schema.TypeList,
							Required: true,
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
						"access_control_expose_headers": {
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
						"access_control_max_age_sec": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"origin_override": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
				AtLeastOneOf: []string{"cors_config", "custom_headers_config", "remove_headers_config", "security_headers_config", "server_timing_headers_config"},
			},
			"custom_headers_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"items": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrHeader: {
										Type:     schema.TypeString,
										Required: true,
									},
									"override": {
										Type:     schema.TypeBool,
										Required: true,
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
				AtLeastOneOf: []string{"cors_config", "custom_headers_config", "remove_headers_config", "security_headers_config", "server_timing_headers_config"},
			},
			"etag": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"remove_headers_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"items": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrHeader: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
				AtLeastOneOf: []string{"cors_config", "custom_headers_config", "remove_headers_config", "security_headers_config", "server_timing_headers_config"},
			},
			"security_headers_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content_security_policy": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"content_security_policy": {
										Type:     schema.TypeString,
										Required: true,
									},
									"override": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"content_type_options": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"override": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"frame_options": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"frame_option": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.FrameOptionsList](),
									},
									"override": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"referrer_policy": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"override": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"referrer_policy": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ReferrerPolicyList](),
									},
								},
							},
						},
						"strict_transport_security": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_control_max_age_sec": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"include_subdomains": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"override": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"preload": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"xss_protection": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode_block": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"override": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"protection": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"report_uri": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
				AtLeastOneOf: []string{"cors_config", "custom_headers_config", "remove_headers_config", "security_headers_config", "server_timing_headers_config"},
			},
			"server_timing_headers_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Required: true,
						},
						"sampling_rate": {
							Type:         schema.TypeFloat,
							Required:     true,
							ValidateFunc: validation.FloatBetween(0.0, 100.0),
						},
					},
				},
				AtLeastOneOf: []string{"cors_config", "custom_headers_config", "remove_headers_config", "security_headers_config", "server_timing_headers_config"},
			},
		},
	}
}

func resourceResponseHeadersPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	name := d.Get(names.AttrName).(string)
	apiObject := &awstypes.ResponseHeadersPolicyConfig{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrComment); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cors_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CorsConfig = expandResponseHeadersPolicyCorsConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("custom_headers_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CustomHeadersConfig = expandResponseHeadersPolicyCustomHeadersConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("remove_headers_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.RemoveHeadersConfig = expandResponseHeadersPolicyRemoveHeadersConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("security_headers_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.SecurityHeadersConfig = expandResponseHeadersPolicySecurityHeadersConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("server_timing_headers_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.ServerTimingHeadersConfig = expandResponseHeadersPolicyServerTimingHeadersConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	input := &cloudfront.CreateResponseHeadersPolicyInput{
		ResponseHeadersPolicyConfig: apiObject,
	}

	output, err := conn.CreateResponseHeadersPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Response Headers Policy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ResponseHeadersPolicy.Id))

	return append(diags, resourceResponseHeadersPolicyRead(ctx, d, meta)...)
}

func resourceResponseHeadersPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	output, err := findResponseHeadersPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Response Headers Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Response Headers Policy (%s): %s", d.Id(), err)
	}

	apiObject := output.ResponseHeadersPolicy.ResponseHeadersPolicyConfig
	d.Set(names.AttrComment, apiObject.Comment)
	if apiObject.CorsConfig != nil {
		if err := d.Set("cors_config", []interface{}{flattenResponseHeadersPolicyCorsConfig(apiObject.CorsConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting cors_config: %s", err)
		}
	} else {
		d.Set("cors_config", nil)
	}
	if apiObject.CustomHeadersConfig != nil {
		if err := d.Set("custom_headers_config", []interface{}{flattenResponseHeadersPolicyCustomHeadersConfig(apiObject.CustomHeadersConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting custom_headers_config: %s", err)
		}
	} else {
		d.Set("custom_headers_config", nil)
	}
	d.Set("etag", output.ETag)
	d.Set(names.AttrName, apiObject.Name)
	if apiObject.RemoveHeadersConfig != nil {
		if err := d.Set("remove_headers_config", []interface{}{flattenResponseHeadersPolicyRemoveHeadersConfig(apiObject.RemoveHeadersConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting remove_headers_config: %s", err)
		}
	} else {
		d.Set("remove_headers_config", nil)
	}
	if apiObject.SecurityHeadersConfig != nil {
		if err := d.Set("security_headers_config", []interface{}{flattenResponseHeadersPolicySecurityHeadersConfig(apiObject.SecurityHeadersConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting security_headers_config: %s", err)
		}
	} else {
		d.Set("security_headers_config", nil)
	}
	if apiObject.ServerTimingHeadersConfig != nil {
		if err := d.Set("server_timing_headers_config", []interface{}{flattenResponseHeadersPolicyServerTimingHeadersConfig(apiObject.ServerTimingHeadersConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting server_timing_headers_config: %s", err)
		}
	} else {
		d.Set("server_timing_headers_config", nil)
	}

	return diags
}

func resourceResponseHeadersPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	//
	// https://docs.aws.amazon.com/cloudfront/latest/APIReference/API_UpdateResponseHeadersPolicy.html:
	// "When you update a response headers policy, the entire policy is replaced. You cannot update some policy fields independent of others."
	//
	apiObject := &awstypes.ResponseHeadersPolicyConfig{
		Name: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk(names.AttrComment); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cors_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CorsConfig = expandResponseHeadersPolicyCorsConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("custom_headers_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CustomHeadersConfig = expandResponseHeadersPolicyCustomHeadersConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("remove_headers_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.RemoveHeadersConfig = expandResponseHeadersPolicyRemoveHeadersConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("security_headers_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.SecurityHeadersConfig = expandResponseHeadersPolicySecurityHeadersConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("server_timing_headers_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.ServerTimingHeadersConfig = expandResponseHeadersPolicyServerTimingHeadersConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	input := &cloudfront.UpdateResponseHeadersPolicyInput{
		Id:                          aws.String(d.Id()),
		IfMatch:                     aws.String(d.Get("etag").(string)),
		ResponseHeadersPolicyConfig: apiObject,
	}

	_, err := conn.UpdateResponseHeadersPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Response Headers Policy (%s): %s", d.Id(), err)
	}

	return append(diags, resourceResponseHeadersPolicyRead(ctx, d, meta)...)
}

func resourceResponseHeadersPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	log.Printf("[DEBUG] Deleting CloudFront Response Headers Policy: %s", d.Id())
	_, err := conn.DeleteResponseHeadersPolicy(ctx, &cloudfront.DeleteResponseHeadersPolicyInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	})

	if errs.IsA[*awstypes.NoSuchResponseHeadersPolicy](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Response Headers Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findResponseHeadersPolicyByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetResponseHeadersPolicyOutput, error) {
	input := &cloudfront.GetResponseHeadersPolicyInput{
		Id: aws.String(id),
	}

	output, err := conn.GetResponseHeadersPolicy(ctx, input)

	if errs.IsA[*awstypes.NoSuchResponseHeadersPolicy](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResponseHeadersPolicy == nil || output.ResponseHeadersPolicy.ResponseHeadersPolicyConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

//
// cors_config:
//

func expandResponseHeadersPolicyCorsConfig(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyCorsConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyCorsConfig{}

	if v, ok := tfMap["access_control_allow_credentials"].(bool); ok {
		apiObject.AccessControlAllowCredentials = aws.Bool(v)
	}

	if v, ok := tfMap["access_control_allow_headers"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AccessControlAllowHeaders = expandResponseHeadersPolicyAccessControlAllowHeaders(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["access_control_allow_methods"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AccessControlAllowMethods = expandResponseHeadersPolicyAccessControlAllowMethods(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["access_control_allow_origins"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AccessControlAllowOrigins = expandResponseHeadersPolicyAccessControlAllowOrigins(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["access_control_expose_headers"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AccessControlExposeHeaders = expandResponseHeadersPolicyAccessControlExposeHeaders(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["access_control_max_age_sec"].(int); ok && v != 0 {
		apiObject.AccessControlMaxAgeSec = aws.Int32(int32(v))
	}

	if v, ok := tfMap["origin_override"].(bool); ok {
		apiObject.OriginOverride = aws.Bool(v)
	}

	return apiObject
}

func expandResponseHeadersPolicyAccessControlAllowHeaders(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyAccessControlAllowHeaders {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyAccessControlAllowHeaders{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := flex.ExpandStringValueSet(v)
		apiObject.Items = items
		apiObject.Quantity = aws.Int32(int32(len(items)))
	}

	return apiObject
}

func expandResponseHeadersPolicyAccessControlAllowMethods(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyAccessControlAllowMethods {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyAccessControlAllowMethods{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := flex.ExpandStringyValueSet[awstypes.ResponseHeadersPolicyAccessControlAllowMethodsValues](v)
		apiObject.Items = items
		apiObject.Quantity = aws.Int32(int32(len(items)))
	}

	return apiObject
}

func expandResponseHeadersPolicyAccessControlAllowOrigins(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyAccessControlAllowOrigins {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyAccessControlAllowOrigins{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := flex.ExpandStringValueSet(v)
		apiObject.Items = items
		apiObject.Quantity = aws.Int32(int32(len(items)))
	}

	return apiObject
}

func expandResponseHeadersPolicyAccessControlExposeHeaders(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyAccessControlExposeHeaders {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyAccessControlExposeHeaders{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := flex.ExpandStringValueSet(v)
		apiObject.Items = items
		apiObject.Quantity = aws.Int32(int32(len(items)))
	}

	return apiObject
}

func flattenResponseHeadersPolicyCorsConfig(apiObject *awstypes.ResponseHeadersPolicyCorsConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AccessControlAllowCredentials; v != nil {
		tfMap["access_control_allow_credentials"] = aws.ToBool(v)
	}

	if v := flattenResponseHeadersPolicyAccessControlAllowHeaders(apiObject.AccessControlAllowHeaders); len(v) > 0 {
		tfMap["access_control_allow_headers"] = []interface{}{v}
	}

	if v := flattenResponseHeadersPolicyAccessControlAllowMethods(apiObject.AccessControlAllowMethods); len(v) > 0 {
		tfMap["access_control_allow_methods"] = []interface{}{v}
	}

	if v := flattenResponseHeadersPolicyAccessControlAllowOrigins(apiObject.AccessControlAllowOrigins); len(v) > 0 {
		tfMap["access_control_allow_origins"] = []interface{}{v}
	}

	if v := flattenResponseHeadersPolicyAccessControlExposeHeaders(apiObject.AccessControlExposeHeaders); len(v) > 0 {
		tfMap["access_control_expose_headers"] = []interface{}{v}
	}

	if v := apiObject.AccessControlMaxAgeSec; v != nil {
		tfMap["access_control_max_age_sec"] = aws.ToInt32(v)
	}

	if v := apiObject.OriginOverride; v != nil {
		tfMap["origin_override"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyAccessControlAllowHeaders(apiObject *awstypes.ResponseHeadersPolicyAccessControlAllowHeaders) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = v
	}

	return tfMap
}

func flattenResponseHeadersPolicyAccessControlAllowMethods(apiObject *awstypes.ResponseHeadersPolicyAccessControlAllowMethods) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = v
	}

	return tfMap
}

func flattenResponseHeadersPolicyAccessControlAllowOrigins(apiObject *awstypes.ResponseHeadersPolicyAccessControlAllowOrigins) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = v
	}

	return tfMap
}

func flattenResponseHeadersPolicyAccessControlExposeHeaders(apiObject *awstypes.ResponseHeadersPolicyAccessControlExposeHeaders) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = v
	}

	return tfMap
}

//
// custom_headers_config:
//

func expandResponseHeadersPolicyCustomHeadersConfig(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyCustomHeadersConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyCustomHeadersConfig{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := expandResponseHeadersPolicyCustomHeaders(v.List())
		apiObject.Items = items
		apiObject.Quantity = aws.Int32(int32(len(items)))
	}

	return apiObject
}

func expandResponseHeadersPolicyCustomHeader(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyCustomHeader {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyCustomHeader{}

	if v, ok := tfMap[names.AttrHeader].(string); ok && v != "" {
		apiObject.Header = aws.String(v)
	}

	if v, ok := tfMap["override"].(bool); ok {
		apiObject.Override = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandResponseHeadersPolicyCustomHeaders(tfList []interface{}) []awstypes.ResponseHeadersPolicyCustomHeader {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ResponseHeadersPolicyCustomHeader

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandResponseHeadersPolicyCustomHeader(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenResponseHeadersPolicyCustomHeadersConfig(apiObject *awstypes.ResponseHeadersPolicyCustomHeadersConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = flattenResponseHeadersPolicyCustomHeaders(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyCustomHeader(apiObject *awstypes.ResponseHeadersPolicyCustomHeader) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Header; v != nil {
		tfMap[names.AttrHeader] = aws.ToString(v)
	}

	if v := apiObject.Override; v != nil {
		tfMap["override"] = aws.ToBool(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyCustomHeaders(apiObjects []awstypes.ResponseHeadersPolicyCustomHeader) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if v := flattenResponseHeadersPolicyCustomHeader(&apiObject); len(v) > 0 {
			tfList = append(tfList, v)
		}
	}

	return tfList
}

//
// remove_headers_config:
//

func expandResponseHeadersPolicyRemoveHeadersConfig(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyRemoveHeadersConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyRemoveHeadersConfig{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := expandResponseHeadersPolicyRemoveHeaders(v.List())
		apiObject.Items = items
		apiObject.Quantity = aws.Int32(int32(len(items)))
	}

	return apiObject
}

func expandResponseHeadersPolicyRemoveHeader(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyRemoveHeader {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyRemoveHeader{}

	if v, ok := tfMap[names.AttrHeader].(string); ok && v != "" {
		apiObject.Header = aws.String(v)
	}

	return apiObject
}

func expandResponseHeadersPolicyRemoveHeaders(tfList []interface{}) []awstypes.ResponseHeadersPolicyRemoveHeader {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ResponseHeadersPolicyRemoveHeader

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandResponseHeadersPolicyRemoveHeader(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenResponseHeadersPolicyRemoveHeadersConfig(apiObject *awstypes.ResponseHeadersPolicyRemoveHeadersConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = flattenResponseHeadersPolicyRemoveHeaders(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyRemoveHeader(apiObject *awstypes.ResponseHeadersPolicyRemoveHeader) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Header; v != nil {
		tfMap[names.AttrHeader] = aws.ToString(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyRemoveHeaders(apiObjects []awstypes.ResponseHeadersPolicyRemoveHeader) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if v := flattenResponseHeadersPolicyRemoveHeader(&apiObject); len(v) > 0 {
			tfList = append(tfList, v)
		}
	}

	return tfList
}

//
// security_headers_config:
//

func expandResponseHeadersPolicySecurityHeadersConfig(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicySecurityHeadersConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicySecurityHeadersConfig{}

	if v, ok := tfMap["content_security_policy"].([]interface{}); ok && len(v) > 0 {
		apiObject.ContentSecurityPolicy = expandResponseHeadersPolicyContentSecurityPolicy(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["content_type_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.ContentTypeOptions = expandResponseHeadersPolicyContentTypeOptions(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["frame_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.FrameOptions = expandResponseHeadersPolicyFrameOptions(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["referrer_policy"].([]interface{}); ok && len(v) > 0 {
		apiObject.ReferrerPolicy = expandResponseHeadersPolicyReferrerPolicy(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["strict_transport_security"].([]interface{}); ok && len(v) > 0 {
		apiObject.StrictTransportSecurity = expandResponseHeadersPolicyStrictTransportSecurity(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["xss_protection"].([]interface{}); ok && len(v) > 0 {
		apiObject.XSSProtection = expandResponseHeadersPolicyXSSProtection(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandResponseHeadersPolicyContentSecurityPolicy(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyContentSecurityPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyContentSecurityPolicy{}

	if v, ok := tfMap["content_security_policy"].(string); ok && v != "" {
		apiObject.ContentSecurityPolicy = aws.String(v)
	}

	if v, ok := tfMap["override"].(bool); ok {
		apiObject.Override = aws.Bool(v)
	}

	return apiObject
}

func expandResponseHeadersPolicyContentTypeOptions(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyContentTypeOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyContentTypeOptions{}

	if v, ok := tfMap["override"].(bool); ok {
		apiObject.Override = aws.Bool(v)
	}

	return apiObject
}

func expandResponseHeadersPolicyFrameOptions(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyFrameOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyFrameOptions{}

	if v, ok := tfMap["frame_option"].(string); ok && v != "" {
		apiObject.FrameOption = awstypes.FrameOptionsList(v)
	}

	if v, ok := tfMap["override"].(bool); ok {
		apiObject.Override = aws.Bool(v)
	}

	return apiObject
}

func expandResponseHeadersPolicyReferrerPolicy(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyReferrerPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyReferrerPolicy{}

	if v, ok := tfMap["override"].(bool); ok {
		apiObject.Override = aws.Bool(v)
	}

	if v, ok := tfMap["referrer_policy"].(string); ok && v != "" {
		apiObject.ReferrerPolicy = awstypes.ReferrerPolicyList(v)
	}

	return apiObject
}

func expandResponseHeadersPolicyStrictTransportSecurity(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyStrictTransportSecurity {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyStrictTransportSecurity{}

	if v, ok := tfMap["access_control_max_age_sec"].(int); ok && v != 0 {
		apiObject.AccessControlMaxAgeSec = aws.Int32(int32(v))
	}

	if v, ok := tfMap["include_subdomains"].(bool); ok {
		apiObject.IncludeSubdomains = aws.Bool(v)
	}

	if v, ok := tfMap["override"].(bool); ok {
		apiObject.Override = aws.Bool(v)
	}

	if v, ok := tfMap["preload"].(bool); ok {
		apiObject.Preload = aws.Bool(v)
	}

	return apiObject
}

func expandResponseHeadersPolicyXSSProtection(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyXSSProtection {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyXSSProtection{}

	if v, ok := tfMap["mode_block"].(bool); ok {
		apiObject.ModeBlock = aws.Bool(v)
	}

	if v, ok := tfMap["override"].(bool); ok {
		apiObject.Override = aws.Bool(v)
	}

	if v, ok := tfMap["protection"].(bool); ok {
		apiObject.Protection = aws.Bool(v)
	}

	if v, ok := tfMap["report_uri"].(string); ok && v != "" {
		apiObject.ReportUri = aws.String(v)
	}

	return apiObject
}

func flattenResponseHeadersPolicySecurityHeadersConfig(apiObject *awstypes.ResponseHeadersPolicySecurityHeadersConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := flattenResponseHeadersPolicyContentSecurityPolicy(apiObject.ContentSecurityPolicy); len(v) > 0 {
		tfMap["content_security_policy"] = []interface{}{v}
	}

	if v := flattenResponseHeadersPolicyContentTypeOptions(apiObject.ContentTypeOptions); len(v) > 0 {
		tfMap["content_type_options"] = []interface{}{v}
	}

	if v := flattenResponseHeadersPolicyFrameOptions(apiObject.FrameOptions); len(v) > 0 {
		tfMap["frame_options"] = []interface{}{v}
	}

	if v := flattenResponseHeadersPolicyReferrerPolicy(apiObject.ReferrerPolicy); len(v) > 0 {
		tfMap["referrer_policy"] = []interface{}{v}
	}

	if v := flattenResponseHeadersPolicyStrictTransportSecurity(apiObject.StrictTransportSecurity); len(v) > 0 {
		tfMap["strict_transport_security"] = []interface{}{v}
	}

	if v := flattenResponseHeadersPolicyXSSProtection(apiObject.XSSProtection); len(v) > 0 {
		tfMap["xss_protection"] = []interface{}{v}
	}

	return tfMap
}

func flattenResponseHeadersPolicyContentSecurityPolicy(apiObject *awstypes.ResponseHeadersPolicyContentSecurityPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ContentSecurityPolicy; v != nil {
		tfMap["content_security_policy"] = aws.ToString(v)
	}

	if v := apiObject.Override; v != nil {
		tfMap["override"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyContentTypeOptions(apiObject *awstypes.ResponseHeadersPolicyContentTypeOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Override; v != nil {
		tfMap["override"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyFrameOptions(apiObject *awstypes.ResponseHeadersPolicyFrameOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FrameOption; v != "" {
		tfMap["frame_option"] = v
	}

	if v := apiObject.Override; v != nil {
		tfMap["override"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyReferrerPolicy(apiObject *awstypes.ResponseHeadersPolicyReferrerPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Override; v != nil {
		tfMap["override"] = aws.ToBool(v)
	}

	if v := apiObject.ReferrerPolicy; v != "" {
		tfMap["referrer_policy"] = v
	}

	return tfMap
}

func flattenResponseHeadersPolicyStrictTransportSecurity(apiObject *awstypes.ResponseHeadersPolicyStrictTransportSecurity) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AccessControlMaxAgeSec; v != nil {
		tfMap["access_control_max_age_sec"] = aws.ToInt32(v)
	}

	if v := apiObject.IncludeSubdomains; v != nil {
		tfMap["include_subdomains"] = aws.ToBool(v)
	}

	if v := apiObject.Override; v != nil {
		tfMap["override"] = aws.ToBool(v)
	}

	if v := apiObject.Preload; v != nil {
		tfMap["preload"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyXSSProtection(apiObject *awstypes.ResponseHeadersPolicyXSSProtection) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ModeBlock; v != nil {
		tfMap["mode_block"] = aws.ToBool(v)
	}

	if v := apiObject.Override; v != nil {
		tfMap["override"] = aws.ToBool(v)
	}

	if v := apiObject.Protection; v != nil {
		tfMap["protection"] = aws.ToBool(v)
	}

	if v := apiObject.ReportUri; v != nil {
		tfMap["report_uri"] = aws.ToString(v)
	}

	return tfMap
}

//
// server_timing_headers_config:
//

func expandResponseHeadersPolicyServerTimingHeadersConfig(tfMap map[string]interface{}) *awstypes.ResponseHeadersPolicyServerTimingHeadersConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ResponseHeadersPolicyServerTimingHeadersConfig{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["sampling_rate"].(float64); ok {
		apiObject.SamplingRate = aws.Float64(v)
	}

	return apiObject
}

func flattenResponseHeadersPolicyServerTimingHeadersConfig(apiObject *awstypes.ResponseHeadersPolicyServerTimingHeadersConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	if v := apiObject.SamplingRate; v != nil {
		tfMap["sampling_rate"] = aws.ToFloat64(v)
	}

	return tfMap
}
