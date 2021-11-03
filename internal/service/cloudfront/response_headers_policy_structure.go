package cloudfront

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func expandResponseHeadersPolicyAccessControlExposeHeaders(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyAccessControlExposeHeaders {
	if tfMap == nil {
		return nil
	}
	items := flex.ExpandStringSet(tfMap["items"].(*schema.Set))
	apiObject := &cloudfront.ResponseHeadersPolicyAccessControlExposeHeaders{
		Items:    items,
		Quantity: aws.Int64(int64(len(items))),
	}
	return apiObject
}

func expandCloudFrontCorsConfig(tfMap map[string]interface{}) *cloudfront.ResponseHeadersPolicyCorsConfig {
	if tfMap == nil {
		return nil
	}

	var accessControlAllowHeaders *cloudfront.ResponseHeadersPolicyAccessControlAllowHeaders
	var accessControlAllowMethods *cloudfront.ResponseHeadersPolicyAccessControlAllowMethods
	var accessControlAllowOrigins *cloudfront.ResponseHeadersPolicyAccessControlAllowOrigins
	if items, ok := tfMap["access_control_allow_headers"].([]interface{}); ok && len(items) == 1 {
		accessControlAllowHeaders = expandResponseHeadersPolicyAccessControlAllowHeaders(items[0].(map[string]interface{}))
	}
	if items, ok := tfMap["access_control_allow_methods"].([]interface{}); ok && len(items) == 1 {
		accessControlAllowMethods = expandResponseHeadersPolicyAccessControlAllowMethods(items[0].(map[string]interface{}))
	}
	if items, ok := tfMap["access_control_allow_origins"].([]interface{}); ok && len(items) == 1 {
		accessControlAllowOrigins = expandResponseHeadersPolicyAccessControlAllowOrigins(items[0].(map[string]interface{}))
	}
	apiObject := &cloudfront.ResponseHeadersPolicyCorsConfig{
		AccessControlAllowCredentials: aws.Bool(tfMap["access_control_allow_credentials"].(bool)),
		AccessControlAllowHeaders:     accessControlAllowHeaders,
		AccessControlAllowMethods:     accessControlAllowMethods,
		AccessControlAllowOrigins:     accessControlAllowOrigins,
		AccessControlMaxAgeSec:        aws.Int64(int64(tfMap["access_control_max_age_sec"].(int))),
		OriginOverride:                aws.Bool(tfMap["origin_override"].(bool)),
	}
	if items, ok := tfMap["access_control_expose_headers"].([]interface{}); ok && len(items) > 0 {
		apiObject.AccessControlExposeHeaders = expandResponseHeadersPolicyAccessControlExposeHeaders(items[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCloudFrontCustomHeader(m map[string]interface{}) *cloudfront.ResponseHeadersPolicyCustomHeader {
	if m == nil {
		return nil
	}
	apiObject := &cloudfront.ResponseHeadersPolicyCustomHeader{
		Header:   aws.String(m["header"].(string)),
		Override: aws.Bool(m["override"].(bool)),
		Value:    aws.String(m["value"].(string)),
	}
	return apiObject
}

func expandCloudFrontResponseHeadersPolicyCustomHeadersConfig(lst []interface{}) *cloudfront.ResponseHeadersPolicyCustomHeadersConfig {
	var qty int64
	var items []*cloudfront.ResponseHeadersPolicyCustomHeader
	for _, v := range lst[0].(map[string]interface{})["items"].(*schema.Set).List() {
		items = append(items, expandCloudFrontCustomHeader(v.(map[string]interface{})))
		qty++
	}
	return &cloudfront.ResponseHeadersPolicyCustomHeadersConfig{
		Quantity: aws.Int64(qty),
		Items:    items,
	}
}

func expandCloudFrontResponseHeadersPolicyContentSecurityPolicy(m map[string]interface{}) *cloudfront.ResponseHeadersPolicyContentSecurityPolicy {
	if m == nil {
		return nil
	}
	apiObject := &cloudfront.ResponseHeadersPolicyContentSecurityPolicy{
		ContentSecurityPolicy: aws.String(m["content_security_policy"].(string)),
		Override:              aws.Bool(m["override"].(bool)),
	}
	return apiObject
}

func expandCloudFrontResponseHeadersPolicyContentTypeOptions(m map[string]interface{}) *cloudfront.ResponseHeadersPolicyContentTypeOptions {
	if m == nil {
		return nil
	}
	apiObject := &cloudfront.ResponseHeadersPolicyContentTypeOptions{
		Override: aws.Bool(m["override"].(bool)),
	}
	return apiObject
}

func expandCloudFrontResponseHeadersPolicyFrameOptions(m map[string]interface{}) *cloudfront.ResponseHeadersPolicyFrameOptions {
	if m == nil {
		return nil
	}
	apiObject := &cloudfront.ResponseHeadersPolicyFrameOptions{
		FrameOption: aws.String(m["frame_option"].(string)),
		Override:    aws.Bool(m["override"].(bool)),
	}
	return apiObject
}

func expandCloudFrontResponseHeadersPolicyReferrerPolicy(m map[string]interface{}) *cloudfront.ResponseHeadersPolicyReferrerPolicy {
	if m == nil {
		return nil
	}
	apiObject := &cloudfront.ResponseHeadersPolicyReferrerPolicy{
		ReferrerPolicy: aws.String(m["referrer_policy"].(string)),
		Override:       aws.Bool(m["override"].(bool)),
	}
	return apiObject
}

func expandCloudFrontResponseHeadersPolicyStrictTransportSecurity(m map[string]interface{}) *cloudfront.ResponseHeadersPolicyStrictTransportSecurity {
	if m == nil {
		return nil
	}
	apiObject := &cloudfront.ResponseHeadersPolicyStrictTransportSecurity{
		AccessControlMaxAgeSec: aws.Int64(int64(m["access_control_max_age_sec"].(int))),
		IncludeSubdomains:      aws.Bool(m["include_subdomains"].(bool)),
		Override:               aws.Bool(m["override"].(bool)),
		Preload:                aws.Bool(m["preload"].(bool)),
	}
	return apiObject
}

func expandCloudFrontResponseHeadersPolicyXSSProtection(m map[string]interface{}) *cloudfront.ResponseHeadersPolicyXSSProtection {
	if m == nil {
		return nil
	}
	apiObject := &cloudfront.ResponseHeadersPolicyXSSProtection{
		ModeBlock:  aws.Bool(m["mode_block"].(bool)),
		Override:   aws.Bool(m["override"].(bool)),
		Protection: aws.Bool(m["protection"].(bool)),
		ReportUri:  aws.String(m["report_uri"].(string)),
	}
	return apiObject
}

func expandCloudFrontSecurityHeadersPolicySecurityHeadersConfig(m map[string]interface{}) *cloudfront.ResponseHeadersPolicySecurityHeadersConfig {
	if m == nil {
		return nil
	}
	apiObject := &cloudfront.ResponseHeadersPolicySecurityHeadersConfig{
		ContentSecurityPolicy:   expandCloudFrontResponseHeadersPolicyContentSecurityPolicy(m["content_security_policy"].(map[string]interface{})),
		ContentTypeOptions:      expandCloudFrontResponseHeadersPolicyContentTypeOptions(m["content_type_options"].(map[string]interface{})),
		FrameOptions:            expandCloudFrontResponseHeadersPolicyFrameOptions(m["frame_options"].(map[string]interface{})),
		ReferrerPolicy:          expandCloudFrontResponseHeadersPolicyReferrerPolicy(m["referrer_policy"].(map[string]interface{})),
		StrictTransportSecurity: expandCloudFrontResponseHeadersPolicyStrictTransportSecurity(m["strict_transport_security"].(map[string]interface{})),
		XSSProtection:           expandCloudFrontResponseHeadersPolicyXSSProtection(m["xss_protection"].(map[string]interface{})),
	}
	return apiObject
}

func expandCloudFrontResponseHeadersPolicyConfig(d *schema.ResourceData) *cloudfront.ResponseHeadersPolicyConfig {
	responseHeadersPolicy := &cloudfront.ResponseHeadersPolicyConfig{
		Comment:             aws.String(d.Get("comment").(string)),
		CustomHeadersConfig: expandCloudFrontResponseHeadersPolicyCustomHeadersConfig(d.Get("custom_headers_config").([]interface{})),
		Name:                aws.String(d.Get("name").(string)),
	}

	if data, ok := d.GetOk("cors_config"); ok {
		if items := data.([]interface{}); len(items) > 0 {
			responseHeadersPolicy.CorsConfig = expandCloudFrontCorsConfig(items[0].(map[string]interface{}))
		}
	}

	if data, ok := d.GetOk("security_headers_config"); ok {
		if items := data.([]interface{}); len(items) > 0 {
			responseHeadersPolicy.SecurityHeadersConfig = expandCloudFrontSecurityHeadersPolicySecurityHeadersConfig(items[0].(map[string]interface{}))
		}
	}

	return responseHeadersPolicy
}

func setCorsConfig(config *cloudfront.ResponseHeadersPolicyCorsConfig) []map[string]interface{} {
	configFlat := map[string]interface{}{
		"access_control_allow_credentials": aws.BoolValue(config.AccessControlAllowCredentials),
		"origin_override":                  aws.BoolValue(config.OriginOverride),
	}

	if config.AccessControlAllowHeaders != nil && len(config.AccessControlAllowHeaders.Items) > 0 {
		accessControlAllowHeaders := []map[string]interface{}{
			{
				"items": config.AccessControlAllowHeaders.Items,
			},
		}
		configFlat["access_control_allow_headers"] = accessControlAllowHeaders
	}

	if config.AccessControlAllowMethods != nil && len(config.AccessControlAllowMethods.Items) > 0 {
		accessControlAllowMethods := []map[string]interface{}{
			{
				"items": config.AccessControlAllowMethods.Items,
			},
		}
		configFlat["access_control_allow_methods"] = accessControlAllowMethods
	}

	if config.AccessControlAllowOrigins != nil && len(config.AccessControlAllowOrigins.Items) > 0 {
		accessControlAllowOrigins := []map[string]interface{}{
			{
				"items": config.AccessControlAllowOrigins.Items,
			},
		}
		configFlat["access_control_allow_origins"] = accessControlAllowOrigins
	}

	if config.AccessControlExposeHeaders != nil && len(config.AccessControlExposeHeaders.Items) > 0 {
		accessControlExposeHeaders := []map[string]interface{}{
			{
				"items": config.AccessControlExposeHeaders.Items,
			},
		}
		configFlat["access_control_expose_headers"] = accessControlExposeHeaders
	}

	if config.AccessControlMaxAgeSec != nil {
		configFlat["access_control_max_age_sec"] = aws.Int64Value(config.AccessControlMaxAgeSec)
	}
	return []map[string]interface{}{
		configFlat,
	}
}

func flattenCustomHeader(customHeader *cloudfront.ResponseHeadersPolicyCustomHeader) []map[string]interface{} {
	m := map[string]interface{}{}
	m["header"] = aws.StringValue(customHeader.Header)
	m["override"] = aws.BoolValue(customHeader.Override)
	m["value"] = aws.StringValue(customHeader.Value)
	return []map[string]interface{}{
		m,
	}
}

func flattenCustomHeadersConfig(config *cloudfront.ResponseHeadersPolicyCustomHeadersConfig) []interface{} {
	lst := []interface{}{}
	for _, v := range config.Items {
		lst = append(lst, flattenCustomHeader(v))
	}
	return lst
}

func flattenCloudFrontResponseHeadersPolicyContentSecurityPolicy(config *cloudfront.ResponseHeadersPolicyContentSecurityPolicy) []map[string]interface{} {
	m := map[string]interface{}{}
	m["content_security_policy"] = aws.StringValue(config.ContentSecurityPolicy)
	m["override"] = aws.BoolValue(config.Override)
	return []map[string]interface{}{
		m,
	}
}

func flattenCloudFrontResponseHeadersPolicyContentTypeOptions(config *cloudfront.ResponseHeadersPolicyContentTypeOptions) []map[string]interface{} {
	m := map[string]interface{}{}
	m["override"] = aws.BoolValue(config.Override)
	return []map[string]interface{}{
		m,
	}
}

func flattenCloudFrontResponseHeadersPolicyFrameOptions(config *cloudfront.ResponseHeadersPolicyFrameOptions) []map[string]interface{} {
	m := map[string]interface{}{}
	m["frame_option"] = aws.StringValue(config.FrameOption)
	m["override"] = aws.BoolValue(config.Override)
	return []map[string]interface{}{
		m,
	}
}

func flattenCloudFrontResponseHeadersPolicyReferrerPolicy(config *cloudfront.ResponseHeadersPolicyReferrerPolicy) []map[string]interface{} {
	m := map[string]interface{}{}
	m["referrer_policy"] = aws.StringValue(config.ReferrerPolicy)
	m["override"] = aws.BoolValue(config.Override)
	return []map[string]interface{}{
		m,
	}
}

func flattenCloudFrontResponseHeadersPolicyStrictTransportSecurity(config *cloudfront.ResponseHeadersPolicyStrictTransportSecurity) []map[string]interface{} {
	m := map[string]interface{}{}
	m["access_control_max_age_sec"] = aws.Int64Value(config.AccessControlMaxAgeSec)
	m["include_subdomains"] = aws.BoolValue(config.IncludeSubdomains)
	m["override"] = aws.BoolValue(config.Override)
	m["preload"] = aws.BoolValue(config.Preload)
	return []map[string]interface{}{
		m,
	}
}

func flattenCloudFrontResponseHeadersPolicyXSSProtection(config *cloudfront.ResponseHeadersPolicyXSSProtection) []map[string]interface{} {
	m := map[string]interface{}{}
	m["mode_block"] = aws.BoolValue(config.ModeBlock)
	m["override"] = aws.BoolValue(config.Override)
	m["protection"] = aws.BoolValue(config.Protection)
	m["report_uri"] = aws.StringValue(config.ReportUri)
	return []map[string]interface{}{
		m,
	}
}

func setSecurityHeadersConfig(parametersConfig *cloudfront.ResponseHeadersPolicySecurityHeadersConfig) []map[string]interface{} {
	parametersConfigFlat := map[string]interface{}{}
	if parametersConfig.ContentSecurityPolicy != nil {
		parametersConfigFlat["content_security_policy"] = flattenCloudFrontResponseHeadersPolicyContentSecurityPolicy(parametersConfig.ContentSecurityPolicy)
	}
	if parametersConfig.ContentTypeOptions != nil {
		parametersConfigFlat["content_type_options"] = flattenCloudFrontResponseHeadersPolicyContentTypeOptions(parametersConfig.ContentTypeOptions)
	}
	if parametersConfig.ContentTypeOptions != nil {
		parametersConfigFlat["frame_options"] = flattenCloudFrontResponseHeadersPolicyFrameOptions(parametersConfig.FrameOptions)
	}
	if parametersConfig.ReferrerPolicy != nil {
		parametersConfigFlat["referrer_policy"] = flattenCloudFrontResponseHeadersPolicyReferrerPolicy(parametersConfig.ReferrerPolicy)
	}
	if parametersConfig.StrictTransportSecurity != nil {
		parametersConfigFlat["strict_transport_security"] = flattenCloudFrontResponseHeadersPolicyStrictTransportSecurity(parametersConfig.StrictTransportSecurity)
	}
	if parametersConfig.XSSProtection != nil {
		parametersConfigFlat["xss_protection"] = flattenCloudFrontResponseHeadersPolicyXSSProtection(parametersConfig.XSSProtection)
	}
	return []map[string]interface{}{
		parametersConfigFlat,
	}
}
