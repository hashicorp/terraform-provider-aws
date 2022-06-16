package cloudfront

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceResponseHeadersPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceResponseHeadersPolicyCreate,
		Read:   resourceResponseHeadersPolicyRead,
		Update: resourceResponseHeadersPolicyUpdate,
		Delete: resourceResponseHeadersPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"comment": {
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
				AtLeastOneOf: []string{"cors_config", "custom_headers_config", "security_headers_config", "server_timing_headers_config"},
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
									"header": {
										Type:     schema.TypeString,
										Required: true,
									},
									"override": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
				AtLeastOneOf: []string{"cors_config", "custom_headers_config", "security_headers_config", "server_timing_headers_config"},
			},
			"etag": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
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
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cloudfront.FrameOptionsList_Values(), false),
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
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(cloudfront.ReferrerPolicyList_Values(), false),
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
				AtLeastOneOf: []string{"cors_config", "custom_headers_config", "security_headers_config", "server_timing_headers_config"},
			},
			"server_timing_headers_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
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
				AtLeastOneOf: []string{"cors_config", "custom_headers_config", "security_headers_config", "server_timing_headers_config"},
			},
		},
	}
}

func resourceResponseHeadersPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	name := d.Get("name").(string)
	apiObject := &cloudfront.ResponseHeadersPolicyConfig{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("comment"); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cors_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CorsConfig = expandResponseHeadersPolicyCorsConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("custom_headers_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CustomHeadersConfig = expandResponseHeadersPolicyCustomHeadersConfig(v.([]interface{})[0].(map[string]interface{}))
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

	log.Printf("[DEBUG] Creating CloudFront Response Headers Policy: (%s)", input)
	output, err := conn.CreateResponseHeadersPolicy(input)

	if err != nil {
		return fmt.Errorf("error creating CloudFront Response Headers Policy (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.ResponseHeadersPolicy.Id))

	return resourceResponseHeadersPolicyRead(d, meta)
}

func resourceResponseHeadersPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	output, err := FindResponseHeadersPolicyByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Response Headers Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudFront Response Headers Policy (%s): %w", d.Id(), err)
	}

	apiObject := output.ResponseHeadersPolicy.ResponseHeadersPolicyConfig
	d.Set("comment", apiObject.Comment)
	if apiObject.CorsConfig != nil {
		if err := d.Set("cors_config", []interface{}{flattenResponseHeadersPolicyCorsConfig(apiObject.CorsConfig)}); err != nil {
			return fmt.Errorf("error setting cors_config: %w", err)
		}
	} else {
		d.Set("cors_config", nil)
	}
	if apiObject.CustomHeadersConfig != nil {
		if err := d.Set("custom_headers_config", []interface{}{flattenResponseHeadersPolicyCustomHeadersConfig(apiObject.CustomHeadersConfig)}); err != nil {
			return fmt.Errorf("error setting custom_headers_config: %w", err)
		}
	} else {
		d.Set("custom_headers_config", nil)
	}
	d.Set("etag", output.ETag)
	d.Set("name", apiObject.Name)
	if apiObject.SecurityHeadersConfig != nil {
		if err := d.Set("security_headers_config", []interface{}{flattenResponseHeadersPolicySecurityHeadersConfig(apiObject.SecurityHeadersConfig)}); err != nil {
			return fmt.Errorf("error setting security_headers_config: %w", err)
		}
	} else {
		d.Set("security_headers_config", nil)
	}

	if apiObject.ServerTimingHeadersConfig != nil {
		if err := d.Set("server_timing_headers_config", []interface{}{flattenResponseHeadersPolicyServerTimingHeadersConfig(apiObject.ServerTimingHeadersConfig)}); err != nil {
			return fmt.Errorf("error setting server_timing_headers_config: %w", err)
		}
	} else {
		d.Set("server_timing_headers_config", nil)
	}

	return nil
}

func resourceResponseHeadersPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	//
	// https://docs.aws.amazon.com/cloudfront/latest/APIReference/API_UpdateResponseHeadersPolicy.html:
	// "When you update a response headers policy, the entire policy is replaced. You cannot update some policy fields independent of others."
	//
	apiObject := &cloudfront.ResponseHeadersPolicyConfig{
		Name: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("comment"); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cors_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CorsConfig = expandResponseHeadersPolicyCorsConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("custom_headers_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CustomHeadersConfig = expandResponseHeadersPolicyCustomHeadersConfig(v.([]interface{})[0].(map[string]interface{}))
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

	log.Printf("[DEBUG] Updating CloudFront Response Headers Policy: (%s)", input)
	_, err := conn.UpdateResponseHeadersPolicy(input)

	if err != nil {
		return fmt.Errorf("error updating CloudFront Response Headers Policy (%s): %w", d.Id(), err)
	}

	return resourceResponseHeadersPolicyRead(d, meta)
}

func resourceResponseHeadersPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	log.Printf("[DEBUG] Deleting CloudFront Response Headers Policy: (%s)", d.Id())
	_, err := conn.DeleteResponseHeadersPolicy(&cloudfront.DeleteResponseHeadersPolicyInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	})

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchResponseHeadersPolicy) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudFront Response Headers Policy (%s): %w", d.Id(), err)
	}

	return nil
}

//
// cors_config:
//

func expandResponseHeadersPolicyCorsConfig(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyCorsConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicyCorsConfig{}

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
		apiObject.AccessControlMaxAgeSec = aws.Int64(int64(v))
	}

	if v, ok := tfMap["origin_override"].(bool); ok {
		apiObject.OriginOverride = aws.Bool(v)
	}

	return apiObject
}

func expandResponseHeadersPolicyAccessControlAllowHeaders(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyAccessControlAllowHeaders {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicyAccessControlAllowHeaders{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := flex.ExpandStringSet(v)
		apiObject.Items = items
		apiObject.Quantity = aws.Int64(int64(len(items)))
	}

	return apiObject
}

func expandResponseHeadersPolicyAccessControlAllowMethods(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyAccessControlAllowMethods {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicyAccessControlAllowMethods{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := flex.ExpandStringSet(v)
		apiObject.Items = items
		apiObject.Quantity = aws.Int64(int64(len(items)))
	}

	return apiObject
}

func expandResponseHeadersPolicyAccessControlAllowOrigins(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyAccessControlAllowOrigins {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicyAccessControlAllowOrigins{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := flex.ExpandStringSet(v)
		apiObject.Items = items
		apiObject.Quantity = aws.Int64(int64(len(items)))
	}

	return apiObject
}

func expandResponseHeadersPolicyAccessControlExposeHeaders(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyAccessControlExposeHeaders {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicyAccessControlExposeHeaders{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := flex.ExpandStringSet(v)
		apiObject.Items = items
		apiObject.Quantity = aws.Int64(int64(len(items)))
	}

	return apiObject
}

func flattenResponseHeadersPolicyCorsConfig(apiObject *cloudfront.ResponseHeadersPolicyCorsConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AccessControlAllowCredentials; v != nil {
		tfMap["access_control_allow_credentials"] = aws.BoolValue(v)
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
		tfMap["access_control_max_age_sec"] = aws.Int64Value(v)
	}

	if v := apiObject.OriginOverride; v != nil {
		tfMap["origin_override"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyAccessControlAllowHeaders(apiObject *cloudfront.ResponseHeadersPolicyAccessControlAllowHeaders) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = aws.StringValueSlice(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyAccessControlAllowMethods(apiObject *cloudfront.ResponseHeadersPolicyAccessControlAllowMethods) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = aws.StringValueSlice(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyAccessControlAllowOrigins(apiObject *cloudfront.ResponseHeadersPolicyAccessControlAllowOrigins) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = aws.StringValueSlice(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyAccessControlExposeHeaders(apiObject *cloudfront.ResponseHeadersPolicyAccessControlExposeHeaders) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = aws.StringValueSlice(v)
	}

	return tfMap
}

//
// custom_headers_config:
//

func expandResponseHeadersPolicyCustomHeadersConfig(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyCustomHeadersConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicyCustomHeadersConfig{}

	if v, ok := tfMap["items"].(*schema.Set); ok && v.Len() > 0 {
		items := expandResponseHeadersPolicyCustomHeaders(v.List())
		apiObject.Items = items
		apiObject.Quantity = aws.Int64(int64(len(items)))
	}

	return apiObject
}

func expandResponseHeadersPolicyCustomHeader(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyCustomHeader {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicyCustomHeader{}

	if v, ok := tfMap["header"].(string); ok && v != "" {
		apiObject.Header = aws.String(v)
	}

	if v, ok := tfMap["override"].(bool); ok {
		apiObject.Override = aws.Bool(v)
	}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandResponseHeadersPolicyCustomHeaders(tfList []interface{}) []*cloudfront.ResponseHeadersPolicyCustomHeader {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*cloudfront.ResponseHeadersPolicyCustomHeader

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandResponseHeadersPolicyCustomHeader(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenResponseHeadersPolicyCustomHeadersConfig(apiObject *cloudfront.ResponseHeadersPolicyCustomHeadersConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Items; len(v) > 0 {
		tfMap["items"] = flattenResponseHeadersPolicyCustomHeaders(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyCustomHeader(apiObject *cloudfront.ResponseHeadersPolicyCustomHeader) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Header; v != nil {
		tfMap["header"] = aws.StringValue(v)
	}

	if v := apiObject.Override; v != nil {
		tfMap["override"] = aws.BoolValue(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap["value"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyCustomHeaders(apiObjects []*cloudfront.ResponseHeadersPolicyCustomHeader) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if v := flattenResponseHeadersPolicyCustomHeader(apiObject); len(v) > 0 {
			tfList = append(tfList, v)
		}
	}

	return tfList
}

//
// security_headers_config:
//

func expandResponseHeadersPolicySecurityHeadersConfig(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicySecurityHeadersConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicySecurityHeadersConfig{}

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

func expandResponseHeadersPolicyContentSecurityPolicy(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyContentSecurityPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicyContentSecurityPolicy{}

	if v, ok := tfMap["content_security_policy"].(string); ok && v != "" {
		apiObject.ContentSecurityPolicy = aws.String(v)
	}

	if v, ok := tfMap["override"].(bool); ok {
		apiObject.Override = aws.Bool(v)
	}

	return apiObject
}

func expandResponseHeadersPolicyContentTypeOptions(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyContentTypeOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicyContentTypeOptions{}

	if v, ok := tfMap["override"].(bool); ok {
		apiObject.Override = aws.Bool(v)
	}

	return apiObject
}

func expandResponseHeadersPolicyFrameOptions(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyFrameOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicyFrameOptions{}

	if v, ok := tfMap["frame_option"].(string); ok && v != "" {
		apiObject.FrameOption = aws.String(v)
	}

	if v, ok := tfMap["override"].(bool); ok {
		apiObject.Override = aws.Bool(v)
	}

	return apiObject
}

func expandResponseHeadersPolicyReferrerPolicy(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyReferrerPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicyReferrerPolicy{}

	if v, ok := tfMap["override"].(bool); ok {
		apiObject.Override = aws.Bool(v)
	}

	if v, ok := tfMap["referrer_policy"].(string); ok && v != "" {
		apiObject.ReferrerPolicy = aws.String(v)
	}

	return apiObject
}

func expandResponseHeadersPolicyStrictTransportSecurity(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyStrictTransportSecurity {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicyStrictTransportSecurity{}

	if v, ok := tfMap["access_control_max_age_sec"].(int); ok && v != 0 {
		apiObject.AccessControlMaxAgeSec = aws.Int64(int64(v))
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

func expandResponseHeadersPolicyXSSProtection(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyXSSProtection {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicyXSSProtection{}

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

func flattenResponseHeadersPolicySecurityHeadersConfig(apiObject *cloudfront.ResponseHeadersPolicySecurityHeadersConfig) map[string]interface{} {
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

func flattenResponseHeadersPolicyContentSecurityPolicy(apiObject *cloudfront.ResponseHeadersPolicyContentSecurityPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ContentSecurityPolicy; v != nil {
		tfMap["content_security_policy"] = aws.StringValue(v)
	}

	if v := apiObject.Override; v != nil {
		tfMap["override"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyContentTypeOptions(apiObject *cloudfront.ResponseHeadersPolicyContentTypeOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Override; v != nil {
		tfMap["override"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyFrameOptions(apiObject *cloudfront.ResponseHeadersPolicyFrameOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FrameOption; v != nil {
		tfMap["frame_option"] = aws.StringValue(v)
	}

	if v := apiObject.Override; v != nil {
		tfMap["override"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyReferrerPolicy(apiObject *cloudfront.ResponseHeadersPolicyReferrerPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Override; v != nil {
		tfMap["override"] = aws.BoolValue(v)
	}

	if v := apiObject.ReferrerPolicy; v != nil {
		tfMap["referrer_policy"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyStrictTransportSecurity(apiObject *cloudfront.ResponseHeadersPolicyStrictTransportSecurity) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AccessControlMaxAgeSec; v != nil {
		tfMap["access_control_max_age_sec"] = aws.Int64Value(v)
	}

	if v := apiObject.IncludeSubdomains; v != nil {
		tfMap["include_subdomains"] = aws.BoolValue(v)
	}

	if v := apiObject.Override; v != nil {
		tfMap["override"] = aws.BoolValue(v)
	}

	if v := apiObject.Preload; v != nil {
		tfMap["preload"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenResponseHeadersPolicyXSSProtection(apiObject *cloudfront.ResponseHeadersPolicyXSSProtection) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ModeBlock; v != nil {
		tfMap["mode_block"] = aws.BoolValue(v)
	}

	if v := apiObject.Override; v != nil {
		tfMap["override"] = aws.BoolValue(v)
	}

	if v := apiObject.Protection; v != nil {
		tfMap["protection"] = aws.BoolValue(v)
	}

	if v := apiObject.ReportUri; v != nil {
		tfMap["report_uri"] = aws.StringValue(v)
	}

	return tfMap
}

//
// server_timing_headers_config:
//

func expandResponseHeadersPolicyServerTimingHeadersConfig(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyServerTimingHeadersConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.ResponseHeadersPolicyServerTimingHeadersConfig{}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["sampling_rate"].(float64); ok && v != 0 {
		apiObject.SamplingRate = aws.Float64(v)
	}

	return apiObject
}

func flattenResponseHeadersPolicyServerTimingHeadersConfig(apiObject *cloudfront.ResponseHeadersPolicyServerTimingHeadersConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap["enabled"] = aws.BoolValue(v)
	}

	if v := apiObject.SamplingRate; v != nil {
		tfMap["sampling_rate"] = aws.Float64Value(v)
	}

	return tfMap
}
