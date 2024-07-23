// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandRules(l []interface{}) []awstypes.Rule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]awstypes.Rule, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandRule(rule.(map[string]interface{})))
	}

	return rules
}

func expandRule(m map[string]interface{}) awstypes.Rule {
	rule := awstypes.Rule{
		Action:           expandRuleAction(m[names.AttrAction].([]interface{})),
		CaptchaConfig:    expandCaptchaConfig(m["captcha_config"].([]interface{})),
		Name:             aws.String(m[names.AttrName].(string)),
		Priority:         int32(m[names.AttrPriority].(int)),
		Statement:        expandRuleGroupRootStatement(m["statement"].([]interface{})),
		VisibilityConfig: expandVisibilityConfig(m["visibility_config"].([]interface{})),
	}

	if v, ok := m["rule_label"].(*schema.Set); ok && v.Len() > 0 {
		rule.RuleLabels = expandRuleLabels(v.List())
	}

	return rule
}

func expandCaptchaConfig(l []interface{}) *awstypes.CaptchaConfig {
	configuration := &awstypes.CaptchaConfig{}

	if len(l) == 0 || l[0] == nil {
		return configuration
	}

	m := l[0].(map[string]interface{})
	if v, ok := m["immunity_time_property"]; ok {
		inner := v.([]interface{})
		if len(inner) == 0 || inner[0] == nil {
			return configuration
		}

		m = inner[0].(map[string]interface{})

		if v, ok := m["immunity_time"]; ok {
			configuration.ImmunityTimeProperty = &awstypes.ImmunityTimeProperty{
				ImmunityTime: aws.Int64(int64(v.(int))),
			}
		}
	}

	return configuration
}

func expandChallengeConfig(l []interface{}) *awstypes.ChallengeConfig {
	configuration := &awstypes.ChallengeConfig{}

	if len(l) == 0 || l[0] == nil {
		return configuration
	}

	m := l[0].(map[string]interface{})
	if v, ok := m["immunity_time_property"]; ok {
		inner := v.([]interface{})
		if len(inner) == 0 || inner[0] == nil {
			return configuration
		}

		m = inner[0].(map[string]interface{})

		if v, ok := m["immunity_time"]; ok {
			configuration.ImmunityTimeProperty = &awstypes.ImmunityTimeProperty{
				ImmunityTime: aws.Int64(int64(v.(int))),
			}
		}
	}

	return configuration
}

func expandAssociationConfig(l []interface{}) *awstypes.AssociationConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configuration := &awstypes.AssociationConfig{}

	m := l[0].(map[string]interface{})
	if v, ok := m["request_body"]; ok {
		inner := v.([]interface{})
		if len(inner) == 0 || inner[0] == nil {
			return configuration
		}

		m = inner[0].(map[string]interface{})
		if len(m) > 0 {
			configuration.RequestBody = make(map[string]awstypes.RequestBodyAssociatedResourceTypeConfig)
			for _, resourceType := range wafv2.AssociatedResourceType_Values() {
				if v, ok := m[strings.ToLower(resourceType)]; ok {
					m := v.([]interface{})
					if len(m) > 0 {
						configuration.RequestBody[resourceType] = expandRequestBodyConfigItem(m)
					}
				}
			}
		}
	}

	return configuration
}

func expandRequestBodyConfigItem(l []interface{}) awstypes.RequestBodyAssociatedResourceTypeConfig {
	configuration := awstypes.RequestBodyAssociatedResourceTypeConfig{}

	if len(l) == 0 || l[0] == nil {
		return configuration
	}

	m := l[0].(map[string]interface{})
	if v, ok := m["default_size_inspection_limit"]; ok {
		if v != "" {
			configuration.DefaultSizeInspectionLimit = awstypes.SizeInspectionLimit(v.(string))
		}
	}

	return configuration
}

func expandRuleLabels(l []interface{}) []awstypes.Label {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	labels := make([]awstypes.Label, 0)

	for _, label := range l {
		if label == nil {
			continue
		}
		m := label.(map[string]interface{})
		labels = append(labels, awstypes.Label{
			Name: aws.String(m[names.AttrName].(string)),
		})
	}

	return labels
}

func expandCountryCodes(l []interface{}) []awstypes.CountryCode {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	countryCodes := make([]awstypes.CountryCode, 0)
	for _, countryCode := range l {
		if countryCode == nil {
			continue
		}

		if v, ok := countryCode.(string); ok {
			countryCodes = append(countryCodes, awstypes.CountryCode(v))
		}
	}

	return countryCodes
}

func expandRuleAction(l []interface{}) *awstypes.RuleAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	action := &awstypes.RuleAction{}

	if v, ok := m["allow"]; ok && len(v.([]interface{})) > 0 {
		action.Allow = expandAllowAction(v.([]interface{}))
	}

	if v, ok := m["block"]; ok && len(v.([]interface{})) > 0 {
		action.Block = expandBlockAction(v.([]interface{}))
	}

	if v, ok := m["captcha"]; ok && len(v.([]interface{})) > 0 {
		action.Captcha = expandCaptchaAction(v.([]interface{}))
	}

	if v, ok := m["challenge"]; ok && len(v.([]interface{})) > 0 {
		action.Challenge = expandChallengeAction(v.([]interface{}))
	}

	if v, ok := m["count"]; ok && len(v.([]interface{})) > 0 {
		action.Count = expandCountAction(v.([]interface{}))
	}

	return action
}

func expandAllowAction(l []interface{}) *awstypes.AllowAction {
	action := &awstypes.AllowAction{}

	if len(l) == 0 || l[0] == nil {
		return action
	}

	m, ok := l[0].(map[string]interface{})
	if !ok {
		return action
	}

	if v, ok := m["custom_request_handling"].([]interface{}); ok && len(v) > 0 {
		action.CustomRequestHandling = expandCustomRequestHandling(v)
	}

	return action
}

func expandBlockAction(l []interface{}) *awstypes.BlockAction {
	action := &awstypes.BlockAction{}

	if len(l) == 0 || l[0] == nil {
		return action
	}

	m, ok := l[0].(map[string]interface{})
	if !ok {
		return action
	}

	if v, ok := m["custom_response"].([]interface{}); ok && len(v) > 0 {
		action.CustomResponse = expandCustomResponse(v)
	}

	return action
}

func expandCaptchaAction(l []interface{}) *awstypes.CaptchaAction {
	action := &awstypes.CaptchaAction{}

	if len(l) == 0 || l[0] == nil {
		return action
	}

	m, ok := l[0].(map[string]interface{})
	if !ok {
		return action
	}

	if v, ok := m["custom_request_handling"].([]interface{}); ok && len(v) > 0 {
		action.CustomRequestHandling = expandCustomRequestHandling(v)
	}

	return action
}

func expandChallengeAction(l []interface{}) *awstypes.ChallengeAction {
	action := &awstypes.ChallengeAction{}

	if len(l) == 0 || l[0] == nil {
		return action
	}

	m, ok := l[0].(map[string]interface{})
	if !ok {
		return action
	}

	if v, ok := m["custom_request_handling"].([]interface{}); ok && len(v) > 0 {
		action.CustomRequestHandling = expandCustomRequestHandling(v)
	}

	return action
}

func expandCountAction(l []interface{}) *awstypes.CountAction {
	action := &awstypes.CountAction{}

	if len(l) == 0 || l[0] == nil {
		return action
	}

	m, ok := l[0].(map[string]interface{})
	if !ok {
		return action
	}

	if v, ok := m["custom_request_handling"].([]interface{}); ok && len(v) > 0 {
		action.CustomRequestHandling = expandCustomRequestHandling(v)
	}

	return action
}

func expandCustomResponseBodies(m []interface{}) map[string]awstypes.CustomResponseBody {
	if len(m) == 0 {
		return nil
	}

	customResponseBodies := make(map[string]awstypes.CustomResponseBody, len(m))

	for _, v := range m {
		vm := v.(map[string]interface{})
		key := vm[names.AttrKey].(string)
		customResponseBodies[key] = awstypes.CustomResponseBody{
			Content:     aws.String(vm[names.AttrContent].(string)),
			ContentType: awstypes.ResponseContentType(vm[names.AttrContentType].(string)),
		}
	}

	return customResponseBodies
}

func expandCustomResponse(l []interface{}) *awstypes.CustomResponse {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	customResponse := &awstypes.CustomResponse{}

	if v, ok := m["custom_response_body_key"].(string); ok && v != "" {
		customResponse.CustomResponseBodyKey = aws.String(v)
	}
	if v, ok := m["response_code"].(int); ok && v > 0 {
		customResponse.ResponseCode = aws.Int32(int32(v))
	}
	if v, ok := m["response_header"].(*schema.Set); ok && len(v.List()) > 0 {
		customResponse.ResponseHeaders = expandCustomHeaders(v.List())
	}

	return customResponse
}

func expandCustomRequestHandling(l []interface{}) *awstypes.CustomRequestHandling {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	requestHandling := &awstypes.CustomRequestHandling{}

	if v, ok := m["insert_header"].(*schema.Set); ok && len(v.List()) > 0 {
		requestHandling.InsertHeaders = expandCustomHeaders(v.List())
	}

	return requestHandling
}

func expandCustomHeaders(l []interface{}) []awstypes.CustomHTTPHeader {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	headers := make([]awstypes.CustomHTTPHeader, 0)

	for _, header := range l {
		if header == nil {
			continue
		}
		m := header.(map[string]interface{})

		headers = append(headers, awstypes.CustomHTTPHeader{
			Name:  aws.String(m[names.AttrName].(string)),
			Value: aws.String(m[names.AttrValue].(string)),
		})
	}

	return headers
}

func expandVisibilityConfig(l []interface{}) *awstypes.VisibilityConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	configuration := &awstypes.VisibilityConfig{}

	if v, ok := m["cloudwatch_metrics_enabled"]; ok {
		configuration.CloudWatchMetricsEnabled = v.(bool)
	}

	if v, ok := m[names.AttrMetricName]; ok && len(v.(string)) > 0 {
		configuration.MetricName = aws.String(v.(string))
	}

	if v, ok := m["sampled_requests_enabled"]; ok {
		configuration.SampledRequestsEnabled = v.(bool)
	}

	return configuration
}

func expandRuleGroupRootStatement(l []interface{}) *awstypes.Statement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return expandStatement(m)
}

func expandStatements(l []interface{}) []awstypes.Statement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	statements := make([]awstypes.Statement, 0)

	for _, statement := range l {
		if statement == nil {
			continue
		}
		statements = append(statements, *expandStatement(statement.(map[string]interface{})))
	}

	return statements
}

func expandStatement(m map[string]interface{}) *awstypes.Statement {
	if m == nil {
		return nil
	}

	statement := &awstypes.Statement{}

	if v, ok := m["and_statement"]; ok {
		statement.AndStatement = expandAndStatement(v.([]interface{}))
	}

	if v, ok := m["byte_match_statement"]; ok {
		statement.ByteMatchStatement = expandByteMatchStatement(v.([]interface{}))
	}

	if v, ok := m["ip_set_reference_statement"]; ok {
		statement.IPSetReferenceStatement = expandIPSetReferenceStatement(v.([]interface{}))
	}

	if v, ok := m["geo_match_statement"]; ok {
		statement.GeoMatchStatement = expandGeoMatchStatement(v.([]interface{}))
	}

	if v, ok := m["label_match_statement"]; ok {
		statement.LabelMatchStatement = expandLabelMatchStatement(v.([]interface{}))
	}

	if v, ok := m["not_statement"]; ok {
		statement.NotStatement = expandNotStatement(v.([]interface{}))
	}

	if v, ok := m["or_statement"]; ok {
		statement.OrStatement = expandOrStatement(v.([]interface{}))
	}

	if v, ok := m["rate_based_statement"]; ok {
		statement.RateBasedStatement = expandRateBasedStatement(v.([]interface{}))
	}

	if v, ok := m["regex_match_statement"]; ok {
		statement.RegexMatchStatement = expandRegexMatchStatement(v.([]interface{}))
	}

	if v, ok := m["regex_pattern_set_reference_statement"]; ok {
		statement.RegexPatternSetReferenceStatement = expandRegexPatternSetReferenceStatement(v.([]interface{}))
	}

	if v, ok := m["size_constraint_statement"]; ok {
		statement.SizeConstraintStatement = expandSizeConstraintStatement(v.([]interface{}))
	}

	if v, ok := m["sqli_match_statement"]; ok {
		statement.SqliMatchStatement = expandSQLiMatchStatement(v.([]interface{}))
	}

	if v, ok := m["xss_match_statement"]; ok {
		statement.XssMatchStatement = expandXSSMatchStatement(v.([]interface{}))
	}

	return statement
}

func expandAndStatement(l []interface{}) *awstypes.AndStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.AndStatement{
		Statements: expandStatements(m["statement"].([]interface{})),
	}
}

func expandByteMatchStatement(l []interface{}) *awstypes.ByteMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.ByteMatchStatement{
		FieldToMatch:         expandFieldToMatch(m["field_to_match"].([]interface{})),
		PositionalConstraint: awstypes.PositionalConstraint(m["positional_constraint"].(string)),
		SearchString:         []byte(m["search_string"].(string)),
		TextTransformations:  expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandFieldToMatch(l []interface{}) *awstypes.FieldToMatch {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	f := &awstypes.FieldToMatch{}

	if v, ok := m["all_query_arguments"]; ok && len(v.([]interface{})) > 0 {
		f.AllQueryArguments = &awstypes.AllQueryArguments{}
	}

	if v, ok := m["body"]; ok && len(v.([]interface{})) > 0 {
		f.Body = expandBody(v.([]interface{}))
	}

	if v, ok := m["cookies"]; ok && len(v.([]interface{})) > 0 {
		f.Cookies = expandCookies(m["cookies"].([]interface{}))
	}

	if v, ok := m["header_order"]; ok && len(v.([]interface{})) > 0 {
		f.HeaderOrder = expandHeaderOrder(m["header_order"].([]interface{}))
	}

	if v, ok := m["headers"]; ok && len(v.([]interface{})) > 0 {
		f.Headers = expandHeaders(m["headers"].([]interface{}))
	}

	if v, ok := m["json_body"]; ok && len(v.([]interface{})) > 0 {
		f.JsonBody = expandJSONBody(v.([]interface{}))
	}

	if v, ok := m["method"]; ok && len(v.([]interface{})) > 0 {
		f.Method = &awstypes.Method{}
	}

	if v, ok := m["query_string"]; ok && len(v.([]interface{})) > 0 {
		f.QueryString = &awstypes.QueryString{}
	}

	if v, ok := m["single_header"]; ok && len(v.([]interface{})) > 0 {
		f.SingleHeader = expandSingleHeader(m["single_header"].([]interface{}))
	}

	if v, ok := m["ja3_fingerprint"]; ok && len(v.([]interface{})) > 0 {
		f.JA3Fingerprint = expandJA3Fingerprint(v.([]interface{}))
	}

	if v, ok := m["single_query_argument"]; ok && len(v.([]interface{})) > 0 {
		f.SingleQueryArgument = expandSingleQueryArgument(m["single_query_argument"].([]interface{}))
	}

	if v, ok := m["uri_path"]; ok && len(v.([]interface{})) > 0 {
		f.UriPath = &awstypes.UriPath{}
	}

	return f
}

func expandForwardedIPConfig(l []interface{}) *awstypes.ForwardedIPConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.ForwardedIPConfig{
		FallbackBehavior: awstypes.FallbackBehavior(m["fallback_behavior"].(string)),
		HeaderName:       aws.String(m["header_name"].(string)),
	}
}

func expandIPSetForwardedIPConfig(l []interface{}) *awstypes.IPSetForwardedIPConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.IPSetForwardedIPConfig{
		FallbackBehavior: awstypes.FallbackBehavior(m["fallback_behavior"].(string)),
		HeaderName:       aws.String(m["header_name"].(string)),
		Position:         awstypes.ForwardedIPPosition(m["position"].(string)),
	}
}

func expandCookies(l []interface{}) *awstypes.Cookies {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	cookies := &awstypes.Cookies{
		MatchScope:       awstypes.MapMatchScope(m["match_scope"].(string)),
		OversizeHandling: awstypes.OversizeHandling(m["oversize_handling"].(string)),
	}

	if v, ok := m["match_pattern"]; ok && len(v.([]interface{})) > 0 {
		cookies.MatchPattern = expandCookieMatchPattern(v.([]interface{}))
	}

	return cookies
}

func expandCookieMatchPattern(l []interface{}) *awstypes.CookieMatchPattern {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	CookieMatchPattern := &awstypes.CookieMatchPattern{}

	if v, ok := m["included_cookies"]; ok && len(v.([]interface{})) > 0 {
		CookieMatchPattern.IncludedCookies = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := m["excluded_cookies"]; ok && len(v.([]interface{})) > 0 {
		CookieMatchPattern.ExcludedCookies = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := m["all"].([]interface{}); ok && len(v) > 0 {
		CookieMatchPattern.All = &awstypes.All{}
	}

	return CookieMatchPattern
}

func expandJSONBody(l []interface{}) *awstypes.JsonBody {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	jsonBody := &awstypes.JsonBody{
		MatchScope:       awstypes.JsonMatchScope(m["match_scope"].(string)),
		OversizeHandling: awstypes.OversizeHandling(m["oversize_handling"].(string)),
		MatchPattern:     expandJSONMatchPattern(m["match_pattern"].([]interface{})),
	}

	if v, ok := m["invalid_fallback_behavior"].(string); ok && v != "" {
		jsonBody.InvalidFallbackBehavior = awstypes.BodyParsingFallbackBehavior(v)
	}

	return jsonBody
}

func expandBody(l []interface{}) *awstypes.Body {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	body := &awstypes.Body{}

	if v, ok := m["oversize_handling"].(string); ok && v != "" {
		body.OversizeHandling = awstypes.OversizeHandling(v)
	}

	return body
}

func expandJA3Fingerprint(l []interface{}) *awstypes.JA3Fingerprint {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	ja3fingerprint := &awstypes.JA3Fingerprint{
		FallbackBehavior: awstypes.FallbackBehavior(m["fallback_behavior"].(string)),
	}

	return ja3fingerprint
}

func expandJSONMatchPattern(l []interface{}) *awstypes.JsonMatchPattern {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	jsonMatchPattern := &awstypes.JsonMatchPattern{}

	if v, ok := m["all"].([]interface{}); ok && len(v) > 0 {
		jsonMatchPattern.All = &awstypes.All{}
	}

	if v, ok := m["included_paths"]; ok && len(v.([]interface{})) > 0 {
		jsonMatchPattern.IncludedPaths = flex.ExpandStringValueList(v.([]interface{}))
	}

	return jsonMatchPattern
}

func expandSingleHeader(l []interface{}) *awstypes.SingleHeader {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.SingleHeader{
		Name: aws.String(m[names.AttrName].(string)),
	}
}

func expandSingleQueryArgument(l []interface{}) *awstypes.SingleQueryArgument {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.SingleQueryArgument{
		Name: aws.String(m[names.AttrName].(string)),
	}
}

func expandTextTransformations(l []interface{}) []awstypes.TextTransformation {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]awstypes.TextTransformation, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandTextTransformation(rule.(map[string]interface{})))
	}

	return rules
}

func expandTextTransformation(m map[string]interface{}) awstypes.TextTransformation {
	return awstypes.TextTransformation{
		Priority: int32(m[names.AttrPriority].(int)),
		Type:     awstypes.TextTransformationType(m[names.AttrType].(string)),
	}
}

func expandIPSetReferenceStatement(l []interface{}) *awstypes.IPSetReferenceStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	statement := &awstypes.IPSetReferenceStatement{
		ARN: aws.String(m[names.AttrARN].(string)),
	}

	if v, ok := m["ip_set_forwarded_ip_config"]; ok {
		statement.IPSetForwardedIPConfig = expandIPSetForwardedIPConfig(v.([]interface{}))
	}

	return statement
}

func expandGeoMatchStatement(l []interface{}) *awstypes.GeoMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	statement := &awstypes.GeoMatchStatement{
		CountryCodes: expandCountryCodes(m["country_codes"].([]interface{})),
	}

	if v, ok := m["forwarded_ip_config"]; ok {
		statement.ForwardedIPConfig = expandForwardedIPConfig(v.([]interface{}))
	}

	return statement
}

func expandLabelMatchStatement(l []interface{}) *awstypes.LabelMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	statement := &awstypes.LabelMatchStatement{
		Key:   aws.String(m[names.AttrKey].(string)),
		Scope: awstypes.LabelMatchScope(m[names.AttrScope].(string)),
	}

	return statement
}

func expandNotStatement(l []interface{}) *awstypes.NotStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	s := m["statement"].([]interface{})

	if len(s) == 0 || s[0] == nil {
		return nil
	}

	m = s[0].(map[string]interface{})

	return &awstypes.NotStatement{
		Statement: expandStatement(m),
	}
}

func expandOrStatement(l []interface{}) *awstypes.OrStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.OrStatement{
		Statements: expandStatements(m["statement"].([]interface{})),
	}
}

func expandRegexMatchStatement(l []interface{}) *awstypes.RegexMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.RegexMatchStatement{
		RegexString:         aws.String(m["regex_string"].(string)),
		FieldToMatch:        expandFieldToMatch(m["field_to_match"].([]interface{})),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandRegexPatternSetReferenceStatement(l []interface{}) *awstypes.RegexPatternSetReferenceStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.RegexPatternSetReferenceStatement{
		ARN:                 aws.String(m[names.AttrARN].(string)),
		FieldToMatch:        expandFieldToMatch(m["field_to_match"].([]interface{})),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandSizeConstraintStatement(l []interface{}) *awstypes.SizeConstraintStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.SizeConstraintStatement{
		ComparisonOperator:  awstypes.ComparisonOperator(m["comparison_operator"].(string)),
		FieldToMatch:        expandFieldToMatch(m["field_to_match"].([]interface{})),
		Size:                int64(m[names.AttrSize].(int)),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandSQLiMatchStatement(l []interface{}) *awstypes.SqliMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.SqliMatchStatement{
		FieldToMatch:        expandFieldToMatch(m["field_to_match"].([]interface{})),
		SensitivityLevel:    awstypes.SensitivityLevel(m["sensitivity_level"].(string)),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandXSSMatchStatement(l []interface{}) *awstypes.XssMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.XssMatchStatement{
		FieldToMatch:        expandFieldToMatch(m["field_to_match"].([]interface{})),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandHeaderOrder(l []interface{}) *awstypes.HeaderOrder {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.HeaderOrder{
		OversizeHandling: awstypes.OversizeHandling(m["oversize_handling"].(string)),
	}
}

func expandHeaders(l []interface{}) *awstypes.Headers {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.Headers{
		MatchPattern:     expandHeaderMatchPattern(m["match_pattern"].([]interface{})),
		MatchScope:       awstypes.MapMatchScope(m["match_scope"].(string)),
		OversizeHandling: awstypes.OversizeHandling(m["oversize_handling"].(string)),
	}
}

func expandHeaderMatchPattern(l []interface{}) *awstypes.HeaderMatchPattern {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	f := &awstypes.HeaderMatchPattern{}

	if v, ok := m["all"]; ok && len(v.([]interface{})) > 0 {
		f.All = &awstypes.All{}
	}

	if v, ok := m["included_headers"]; ok && len(v.([]interface{})) > 0 {
		f.IncludedHeaders = flex.ExpandStringValueList(m["included_headers"].([]interface{}))
	}

	if v, ok := m["excluded_headers"]; ok && len(v.([]interface{})) > 0 {
		f.ExcludedHeaders = flex.ExpandStringValueList(m["excluded_headers"].([]interface{}))
	}

	return f
}

func expandWebACLRules(l []interface{}) []awstypes.Rule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]awstypes.Rule, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandWebACLRule(rule.(map[string]interface{})))
	}

	return rules
}

func expandWebACLRule(m map[string]interface{}) awstypes.Rule {
	rule := awstypes.Rule{
		Action:           expandRuleAction(m[names.AttrAction].([]interface{})),
		CaptchaConfig:    expandCaptchaConfig(m["captcha_config"].([]interface{})),
		Name:             aws.String(m[names.AttrName].(string)),
		OverrideAction:   expandOverrideAction(m["override_action"].([]interface{})),
		Priority:         int32(m[names.AttrPriority].(int)),
		Statement:        expandWebACLRootStatement(m["statement"].([]interface{})),
		VisibilityConfig: expandVisibilityConfig(m["visibility_config"].([]interface{})),
	}

	if v, ok := m["rule_label"].(*schema.Set); ok && v.Len() > 0 {
		rule.RuleLabels = expandRuleLabels(v.List())
	}

	return rule
}

func expandOverrideAction(l []interface{}) *awstypes.OverrideAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	action := &awstypes.OverrideAction{}

	if v, ok := m["count"]; ok && len(v.([]interface{})) > 0 {
		action.Count = &awstypes.CountAction{}
	}

	if v, ok := m["none"]; ok && len(v.([]interface{})) > 0 {
		action.None = &awstypes.NoneAction{}
	}

	return action
}

func expandDefaultAction(l []interface{}) *awstypes.DefaultAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	action := &awstypes.DefaultAction{}

	if v, ok := m["allow"]; ok && len(v.([]interface{})) > 0 {
		action.Allow = expandAllowAction(v.([]interface{}))
	}

	if v, ok := m["block"]; ok && len(v.([]interface{})) > 0 {
		action.Block = expandBlockAction(v.([]interface{}))
	}

	return action
}

func expandWebACLRootStatement(l []interface{}) *awstypes.Statement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return expandWebACLStatement(m)
}

func expandWebACLStatement(m map[string]interface{}) *awstypes.Statement {
	if m == nil {
		return nil
	}

	statement := &awstypes.Statement{}

	if v, ok := m["and_statement"]; ok {
		statement.AndStatement = expandAndStatement(v.([]interface{}))
	}

	if v, ok := m["byte_match_statement"]; ok {
		statement.ByteMatchStatement = expandByteMatchStatement(v.([]interface{}))
	}

	if v, ok := m["ip_set_reference_statement"]; ok {
		statement.IPSetReferenceStatement = expandIPSetReferenceStatement(v.([]interface{}))
	}

	if v, ok := m["geo_match_statement"]; ok {
		statement.GeoMatchStatement = expandGeoMatchStatement(v.([]interface{}))
	}

	if v, ok := m["label_match_statement"]; ok {
		statement.LabelMatchStatement = expandLabelMatchStatement(v.([]interface{}))
	}

	if v, ok := m["managed_rule_group_statement"]; ok {
		statement.ManagedRuleGroupStatement = expandManagedRuleGroupStatement(v.([]interface{}))
	}

	if v, ok := m["not_statement"]; ok {
		statement.NotStatement = expandNotStatement(v.([]interface{}))
	}

	if v, ok := m["or_statement"]; ok {
		statement.OrStatement = expandOrStatement(v.([]interface{}))
	}

	if v, ok := m["rate_based_statement"]; ok {
		statement.RateBasedStatement = expandRateBasedStatement(v.([]interface{}))
	}

	if v, ok := m["regex_match_statement"]; ok {
		statement.RegexMatchStatement = expandRegexMatchStatement(v.([]interface{}))
	}

	if v, ok := m["regex_pattern_set_reference_statement"]; ok {
		statement.RegexPatternSetReferenceStatement = expandRegexPatternSetReferenceStatement(v.([]interface{}))
	}

	if v, ok := m["rule_group_reference_statement"]; ok {
		statement.RuleGroupReferenceStatement = expandRuleGroupReferenceStatement(v.([]interface{}))
	}

	if v, ok := m["size_constraint_statement"]; ok {
		statement.SizeConstraintStatement = expandSizeConstraintStatement(v.([]interface{}))
	}

	if v, ok := m["sqli_match_statement"]; ok {
		statement.SqliMatchStatement = expandSQLiMatchStatement(v.([]interface{}))
	}

	if v, ok := m["xss_match_statement"]; ok {
		statement.XssMatchStatement = expandXSSMatchStatement(v.([]interface{}))
	}

	return statement
}

func expandManagedRuleGroupStatement(l []interface{}) *awstypes.ManagedRuleGroupStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	r := &awstypes.ManagedRuleGroupStatement{
		Name:                aws.String(m[names.AttrName].(string)),
		RuleActionOverrides: expandRuleActionOverrides(m["rule_action_override"].([]interface{})),
		VendorName:          aws.String(m["vendor_name"].(string)),
	}

	if s, ok := m["scope_down_statement"].([]interface{}); ok && len(s) > 0 && s[0] != nil {
		r.ScopeDownStatement = expandStatement(s[0].(map[string]interface{}))
	}

	if v, ok := m[names.AttrVersion]; ok && v != "" {
		r.Version = aws.String(v.(string))
	}
	if v, ok := m["managed_rule_group_configs"].([]interface{}); ok && len(v) > 0 {
		r.ManagedRuleGroupConfigs = expandManagedRuleGroupConfigs(v)
	}

	return r
}

func expandManagedRuleGroupConfigs(tfList []interface{}) []awstypes.ManagedRuleGroupConfig {
	if len(tfList) == 0 {
		return nil
	}

	var out []awstypes.ManagedRuleGroupConfig
	for _, item := range tfList {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var r awstypes.ManagedRuleGroupConfig
		if v, ok := m["aws_managed_rules_bot_control_rule_set"].([]interface{}); ok && len(v) > 0 {
			r.AWSManagedRulesBotControlRuleSet = expandManagedRulesBotControlRuleSet(v)
		}
		if v, ok := m["aws_managed_rules_acfp_rule_set"].([]interface{}); ok && len(v) > 0 {
			r.AWSManagedRulesACFPRuleSet = expandManagedRulesACFPRuleSet(v)
		}
		if v, ok := m["aws_managed_rules_atp_rule_set"].([]interface{}); ok && len(v) > 0 {
			r.AWSManagedRulesATPRuleSet = expandManagedRulesATPRuleSet(v)
		}
		if v, ok := m["login_path"].(string); ok && v != "" {
			r.LoginPath = aws.String(v)
		}
		if v, ok := m["payload_type"].(string); ok && v != "" {
			r.PayloadType = awstypes.PayloadType(v)
		}
		if v, ok := m["password_field"].([]interface{}); ok && len(v) > 0 {
			r.PasswordField = expandPasswordField(v)
		}
		if v, ok := m["username_field"].([]interface{}); ok && len(v) > 0 {
			r.UsernameField = expandUsernameField(v)
		}

		out = append(out, r)
	}

	return out
}

func expandAddressFields(tfList []interface{}) []awstypes.AddressField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	out := make([]awstypes.AddressField, 0)
	identifiers := tfList[0].(map[string]interface{})
	for _, v := range identifiers["identifiers"].([]interface{}) {
		r := awstypes.AddressField{
			Identifier: aws.String(v.(string)),
		}

		out = append(out, r)
	}

	return out
}

func expandEmailField(tfList []interface{}) *awstypes.EmailField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})
	out := awstypes.EmailField{
		Identifier: aws.String(m[names.AttrIdentifier].(string)),
	}

	return &out
}

func expandPasswordField(tfList []interface{}) *awstypes.PasswordField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})
	out := awstypes.PasswordField{
		Identifier: aws.String(m[names.AttrIdentifier].(string)),
	}

	return &out
}

func expandPhoneNumberFields(tfList []interface{}) []awstypes.PhoneNumberField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	out := make([]awstypes.PhoneNumberField, 0)
	identifiers := tfList[0].(map[string]interface{})
	for _, v := range identifiers["identifiers"].([]interface{}) {
		r := awstypes.PhoneNumberField{
			Identifier: aws.String(v.(string)),
		}

		out = append(out, r)
	}

	return out
}

func expandUsernameField(tfList []interface{}) *awstypes.UsernameField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})
	out := awstypes.UsernameField{
		Identifier: aws.String(m[names.AttrIdentifier].(string)),
	}

	return &out
}

func expandManagedRulesBotControlRuleSet(tfList []interface{}) *awstypes.AWSManagedRulesBotControlRuleSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})
	out := awstypes.AWSManagedRulesBotControlRuleSet{
		EnableMachineLearning: aws.Bool(m["enable_machine_learning"].(bool)),
		InspectionLevel:       awstypes.InspectionLevel(m["inspection_level"].(string)),
	}

	return &out
}

func expandManagedRulesACFPRuleSet(tfList []interface{}) *awstypes.AWSManagedRulesACFPRuleSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})
	out := awstypes.AWSManagedRulesACFPRuleSet{
		CreationPath:         aws.String(m["creation_path"].(string)),
		RegistrationPagePath: aws.String(m["registration_page_path"].(string)),
	}

	if v, ok := m["enable_regex_in_path"].(bool); ok {
		out.EnableRegexInPath = v
	}
	if v, ok := m["request_inspection"].([]interface{}); ok && len(v) > 0 {
		out.RequestInspection = expandRequestInspectionACFP(v)
	}
	if v, ok := m["response_inspection"].([]interface{}); ok && len(v) > 0 {
		out.ResponseInspection = expandResponseInspection(v)
	}

	return &out
}

func expandManagedRulesATPRuleSet(tfList []interface{}) *awstypes.AWSManagedRulesATPRuleSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})
	out := awstypes.AWSManagedRulesATPRuleSet{
		LoginPath: aws.String(m["login_path"].(string)),
	}

	if v, ok := m["enable_regex_in_path"].(bool); ok {
		out.EnableRegexInPath = v
	}
	if v, ok := m["request_inspection"].([]interface{}); ok && len(v) > 0 {
		out.RequestInspection = expandRequestInspection(v)
	}
	if v, ok := m["response_inspection"].([]interface{}); ok && len(v) > 0 {
		out.ResponseInspection = expandResponseInspection(v)
	}

	return &out
}

func expandRequestInspection(tfList []interface{}) *awstypes.RequestInspection {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})
	out := awstypes.RequestInspection{
		PasswordField: expandPasswordField(m["password_field"].([]interface{})),
		PayloadType:   awstypes.PayloadType(m["payload_type"].(string)),
		UsernameField: expandUsernameField(m["username_field"].([]interface{})),
	}

	return &out
}

func expandRequestInspectionACFP(tfList []interface{}) *awstypes.RequestInspectionACFP {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})
	out := awstypes.RequestInspectionACFP{
		AddressFields:     expandAddressFields(m["address_fields"].([]interface{})),
		EmailField:        expandEmailField(m["email_field"].([]interface{})),
		PasswordField:     expandPasswordField(m["password_field"].([]interface{})),
		PayloadType:       awstypes.PayloadType(m["payload_type"].(string)),
		PhoneNumberFields: expandPhoneNumberFields(m["phone_number_fields"].([]interface{})),
		UsernameField:     expandUsernameField(m["username_field"].([]interface{})),
	}

	return &out
}

func expandResponseInspection(tfList []interface{}) *awstypes.ResponseInspection {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})
	out := awstypes.ResponseInspection{}
	if v, ok := m["body_contains"].([]interface{}); ok && len(v) > 0 {
		out.BodyContains = expandBodyContains(v)
	}
	if v, ok := m[names.AttrHeader].([]interface{}); ok && len(v) > 0 {
		out.Header = expandHeader(v)
	}
	if v, ok := m[names.AttrJSON].([]interface{}); ok && len(v) > 0 {
		out.Json = expandResponseInspectionJSON(v)
	}
	if v, ok := m[names.AttrStatusCode].([]interface{}); ok && len(v) > 0 {
		out.StatusCode = expandStatusCode(v)
	}

	return &out
}

func expandBodyContains(tfList []interface{}) *awstypes.ResponseInspectionBodyContains {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})
	out := awstypes.ResponseInspectionBodyContains{
		FailureStrings: flex.ExpandStringValueSet(m["failure_strings"].(*schema.Set)),
		SuccessStrings: flex.ExpandStringValueSet(m["success_strings"].(*schema.Set)),
	}

	return &out
}

func expandHeader(tfList []interface{}) *awstypes.ResponseInspectionHeader {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	out := awstypes.ResponseInspectionHeader{
		Name:          aws.String(m[names.AttrName].(string)),
		FailureValues: flex.ExpandStringValueSet(m["failure_values"].(*schema.Set)),
		SuccessValues: flex.ExpandStringValueSet(m["success_values"].(*schema.Set)),
	}

	return &out
}

func expandResponseInspectionJSON(tfList []interface{}) *awstypes.ResponseInspectionJson {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})
	out := awstypes.ResponseInspectionJson{
		FailureValues: flex.ExpandStringValueSet(m["failure_values"].(*schema.Set)),
		Identifier:    aws.String(m[names.AttrIdentifier].(string)),
		SuccessValues: flex.ExpandStringValueSet(m["success_values"].(*schema.Set)),
	}

	return &out
}

func expandStatusCode(tfList []interface{}) *awstypes.ResponseInspectionStatusCode {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]interface{})
	out := awstypes.ResponseInspectionStatusCode{
		FailureCodes: flex.ExpandInt32ValueSet(m["failure_codes"].(*schema.Set)),
		SuccessCodes: flex.ExpandInt32ValueSet(m["success_codes"].(*schema.Set)),
	}

	return &out
}

func expandRateLimitCookie(l []interface{}) *awstypes.RateLimitCookie {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]interface{})

	return &awstypes.RateLimitCookie{
		Name:                aws.String(m[names.AttrName].(string)),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandRateLimitHeader(l []interface{}) *awstypes.RateLimitHeader {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]interface{})

	return &awstypes.RateLimitHeader{
		Name:                aws.String(m[names.AttrName].(string)),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandRateLimitLabelNamespace(l []interface{}) *awstypes.RateLimitLabelNamespace {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]interface{})
	return &awstypes.RateLimitLabelNamespace{
		Namespace: aws.String(m[names.AttrNamespace].(string)),
	}
}

func expandRateLimitQueryArgument(l []interface{}) *awstypes.RateLimitQueryArgument {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]interface{})

	return &awstypes.RateLimitQueryArgument{
		Name:                aws.String(m[names.AttrName].(string)),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandRateLimitQueryString(l []interface{}) *awstypes.RateLimitQueryString {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]interface{})
	return &awstypes.RateLimitQueryString{
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandRateLimitURIPath(l []interface{}) *awstypes.RateLimitUriPath {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]interface{})
	return &awstypes.RateLimitUriPath{
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandRateBasedStatementCustomKeys(l []interface{}) []awstypes.RateBasedStatementCustomKey {
	if len(l) == 0 {
		return nil
	}
	out := make([]awstypes.RateBasedStatementCustomKey, 0)
	for _, ck := range l {
		r := awstypes.RateBasedStatementCustomKey{}
		m := ck.(map[string]interface{})
		if v, ok := m["cookie"]; ok {
			r.Cookie = expandRateLimitCookie(v.([]interface{}))
		}
		if v, ok := m["forwarded_ip"]; ok && len(v.([]interface{})) > 0 {
			r.ForwardedIP = &awstypes.RateLimitForwardedIP{}
		}
		if v, ok := m["http_method"]; ok && len(v.([]interface{})) > 0 {
			r.HTTPMethod = &awstypes.RateLimitHTTPMethod{}
		}
		if v, ok := m[names.AttrHeader]; ok {
			r.Header = expandRateLimitHeader(v.([]interface{}))
		}
		if v, ok := m["ip"]; ok && len(v.([]interface{})) > 0 {
			r.IP = &awstypes.RateLimitIP{}
		}
		if v, ok := m["label_namespace"]; ok {
			r.LabelNamespace = expandRateLimitLabelNamespace(v.([]interface{}))
		}
		if v, ok := m["query_argument"]; ok {
			r.QueryArgument = expandRateLimitQueryArgument(v.([]interface{}))
		}
		if v, ok := m["query_string"]; ok {
			r.QueryString = expandRateLimitQueryString(v.([]interface{}))
		}
		if v, ok := m["uri_path"]; ok {
			r.UriPath = expandRateLimitURIPath(v.([]interface{}))
		}
		out = append(out, r)
	}
	return out
}

func expandRateBasedStatement(l []interface{}) *awstypes.RateBasedStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	r := &awstypes.RateBasedStatement{
		AggregateKeyType:    awstypes.RateBasedStatementAggregateKeyType(m["aggregate_key_type"].(string)),
		EvaluationWindowSec: int64(m["evaluation_window_sec"].(int)),
		Limit:               aws.Int64(int64(m["limit"].(int))),
	}

	if v, ok := m["forwarded_ip_config"]; ok {
		r.ForwardedIPConfig = expandForwardedIPConfig(v.([]interface{}))
	}

	if v, ok := m["custom_key"]; ok {
		r.CustomKeys = expandRateBasedStatementCustomKeys(v.([]interface{}))
	}

	s := m["scope_down_statement"].([]interface{})
	if len(s) > 0 && s[0] != nil {
		r.ScopeDownStatement = expandStatement(s[0].(map[string]interface{}))
	}

	return r
}

func expandRuleGroupReferenceStatement(l []interface{}) *awstypes.RuleGroupReferenceStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.RuleGroupReferenceStatement{
		ARN:                 aws.String(m[names.AttrARN].(string)),
		RuleActionOverrides: expandRuleActionOverrides(m["rule_action_override"].([]interface{})),
	}
}

func expandRuleActionOverrides(l []interface{}) []awstypes.RuleActionOverride {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	overrides := make([]awstypes.RuleActionOverride, 0)

	for _, override := range l {
		if override == nil {
			continue
		}
		overrides = append(overrides, expandRuleActionOverride(override.(map[string]interface{})))
	}

	return overrides
}

func expandRuleActionOverride(m map[string]interface{}) awstypes.RuleActionOverride {
	return awstypes.RuleActionOverride{
		ActionToUse: expandRuleAction(m["action_to_use"].([]interface{})),
		Name:        aws.String(m[names.AttrName].(string)),
	}
}

func expandRegexPatternSet(l []interface{}) []awstypes.Regex {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	regexPatterns := make([]awstypes.Regex, 0)
	for _, regexPattern := range l {
		regexPatterns = append(regexPatterns, expandRegex(regexPattern.(map[string]interface{})))
	}

	return regexPatterns
}

func expandRegex(m map[string]interface{}) awstypes.Regex {
	return awstypes.Regex{
		RegexString: aws.String(m["regex_string"].(string)),
	}
}

func flattenRules(r []awstypes.Rule) interface{} {
	out := make([]map[string]interface{}, len(r))
	for i, rule := range r {
		m := make(map[string]interface{})
		m[names.AttrAction] = flattenRuleAction(rule.Action)
		m["captcha_config"] = flattenCaptchaConfig(rule.CaptchaConfig)
		m[names.AttrName] = aws.ToString(rule.Name)
		m[names.AttrPriority] = rule.Priority
		m["rule_label"] = flattenRuleLabels(rule.RuleLabels)
		m["statement"] = flattenRuleGroupRootStatement(rule.Statement)
		m["visibility_config"] = flattenVisibilityConfig(rule.VisibilityConfig)
		out[i] = m
	}

	return out
}

func flattenRuleAction(a *awstypes.RuleAction) interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if a.Allow != nil {
		m["allow"] = flattenAllow(a.Allow)
	}

	if a.Block != nil {
		m["block"] = flattenBlock(a.Block)
	}

	if a.Captcha != nil {
		m["captcha"] = flattenCaptcha(a.Captcha)
	}

	if a.Challenge != nil {
		m["challenge"] = flattenChallenge(a.Challenge)
	}

	if a.Count != nil {
		m["count"] = flattenCount(a.Count)
	}

	return []interface{}{m}
}

func flattenAllow(a *awstypes.AllowAction) []interface{} {
	if a == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{}

	if a.CustomRequestHandling != nil {
		m["custom_request_handling"] = flattenCustomRequestHandling(a.CustomRequestHandling)
	}

	return []interface{}{m}
}

func flattenBlock(a *awstypes.BlockAction) []interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if a.CustomResponse != nil {
		m["custom_response"] = flattenCustomResponse(a.CustomResponse)
	}

	return []interface{}{m}
}

func flattenCaptcha(a *awstypes.CaptchaAction) []interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if a.CustomRequestHandling != nil {
		m["custom_request_handling"] = flattenCustomRequestHandling(a.CustomRequestHandling)
	}

	return []interface{}{m}
}

func flattenCaptchaConfig(config *awstypes.CaptchaConfig) interface{} {
	if config == nil {
		return []interface{}{}
	}
	if config.ImmunityTimeProperty == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"immunity_time_property": []interface{}{map[string]interface{}{
			"immunity_time": aws.ToInt64(config.ImmunityTimeProperty.ImmunityTime),
		}},
	}

	return []interface{}{m}
}

func flattenChallengeConfig(config *awstypes.ChallengeConfig) interface{} {
	if config == nil {
		return []interface{}{}
	}
	if config.ImmunityTimeProperty == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"immunity_time_property": []interface{}{map[string]interface{}{
			"immunity_time": aws.ToInt64(config.ImmunityTimeProperty.ImmunityTime),
		}},
	}

	return []interface{}{m}
}

func flattenAssociationConfig(config *awstypes.AssociationConfig) interface{} {
	associationConfig := []interface{}{}
	if config == nil {
		return associationConfig
	}
	if config.RequestBody == nil {
		return associationConfig
	}

	requestBodyConfig := map[string]interface{}{}
	for _, resourceType := range wafv2.AssociatedResourceType_Values() {
		if requestBodyAssociatedResourceTypeConfig, ok := config.RequestBody[resourceType]; ok {
			requestBodyConfig[strings.ToLower(resourceType)] = []map[string]interface{}{{
				"default_size_inspection_limit": requestBodyAssociatedResourceTypeConfig.DefaultSizeInspectionLimit,
			}}
		}
	}
	associationConfig = append(associationConfig, map[string]interface{}{
		"request_body": []map[string]interface{}{
			requestBodyConfig,
		},
	})

	return associationConfig
}

func flattenChallenge(a *awstypes.ChallengeAction) []interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if a.CustomRequestHandling != nil {
		m["custom_request_handling"] = flattenCustomRequestHandling(a.CustomRequestHandling)
	}

	return []interface{}{m}
}

func flattenCount(a *awstypes.CountAction) []interface{} {
	if a == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{}

	if a.CustomRequestHandling != nil {
		m["custom_request_handling"] = flattenCustomRequestHandling(a.CustomRequestHandling)
	}

	return []interface{}{m}
}

func flattenCustomResponseBodies(b map[string]awstypes.CustomResponseBody) interface{} {
	if len(b) == 0 {
		return make([]map[string]interface{}, 0)
	}

	out := make([]map[string]interface{}, len(b))
	i := 0
	for key, body := range b {
		out[i] = map[string]interface{}{
			names.AttrKey:         key,
			names.AttrContent:     aws.ToString(body.Content),
			names.AttrContentType: body.ContentType,
		}
		i += 1
	}

	return out
}

func flattenCustomRequestHandling(c *awstypes.CustomRequestHandling) []interface{} {
	if c == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"insert_header": flattenCustomHeaders(c.InsertHeaders),
	}

	return []interface{}{m}
}

func flattenCustomResponse(r *awstypes.CustomResponse) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"response_code":   aws.ToInt32(r.ResponseCode),
		"response_header": flattenCustomHeaders(r.ResponseHeaders),
	}

	if r.CustomResponseBodyKey != nil {
		m["custom_response_body_key"] = aws.ToString(r.CustomResponseBodyKey)
	}

	return []interface{}{m}
}

func flattenCustomHeaders(h []awstypes.CustomHTTPHeader) []interface{} {
	out := make([]interface{}, len(h))
	for i, header := range h {
		out[i] = flattenCustomHeader(&header)
	}

	return out
}

func flattenCustomHeader(h *awstypes.CustomHTTPHeader) map[string]interface{} {
	if h == nil {
		return map[string]interface{}{}
	}

	m := map[string]interface{}{
		names.AttrName:  aws.ToString(h.Name),
		names.AttrValue: aws.ToString(h.Value),
	}

	return m
}

func flattenRuleLabels(l []awstypes.Label) []interface{} {
	if len(l) == 0 {
		return nil
	}

	out := make([]interface{}, len(l))
	for i, label := range l {
		out[i] = map[string]interface{}{
			names.AttrName: aws.ToString(label.Name),
		}
	}

	return out
}

func flattenRuleGroupRootStatement(s *awstypes.Statement) interface{} {
	if s == nil {
		return []interface{}{}
	}

	return []interface{}{flattenStatement(s)}
}

func flattenStatements(s []awstypes.Statement) interface{} {
	out := make([]interface{}, len(s))
	for i, statement := range s {
		out[i] = flattenStatement(&statement)
	}

	return out
}

func flattenStatement(s *awstypes.Statement) map[string]interface{} {
	if s == nil {
		return map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if s.AndStatement != nil {
		m["and_statement"] = flattenAndStatement(s.AndStatement)
	}

	if s.ByteMatchStatement != nil {
		m["byte_match_statement"] = flattenByteMatchStatement(s.ByteMatchStatement)
	}

	if s.IPSetReferenceStatement != nil {
		m["ip_set_reference_statement"] = flattenIPSetReferenceStatement(s.IPSetReferenceStatement)
	}

	if s.GeoMatchStatement != nil {
		m["geo_match_statement"] = flattenGeoMatchStatement(s.GeoMatchStatement)
	}

	if s.LabelMatchStatement != nil {
		m["label_match_statement"] = flattenLabelMatchStatement(s.LabelMatchStatement)
	}

	if s.NotStatement != nil {
		m["not_statement"] = flattenNotStatement(s.NotStatement)
	}

	if s.OrStatement != nil {
		m["or_statement"] = flattenOrStatement(s.OrStatement)
	}

	if s.RateBasedStatement != nil {
		m["rate_based_statement"] = flattenRateBasedStatement(s.RateBasedStatement)
	}

	if s.RegexMatchStatement != nil {
		m["regex_match_statement"] = flattenRegexMatchStatement(s.RegexMatchStatement)
	}

	if s.RegexPatternSetReferenceStatement != nil {
		m["regex_pattern_set_reference_statement"] = flattenRegexPatternSetReferenceStatement(s.RegexPatternSetReferenceStatement)
	}

	if s.SizeConstraintStatement != nil {
		m["size_constraint_statement"] = flattenSizeConstraintStatement(s.SizeConstraintStatement)
	}

	if s.SqliMatchStatement != nil {
		m["sqli_match_statement"] = flattenSQLiMatchStatement(s.SqliMatchStatement)
	}

	if s.XssMatchStatement != nil {
		m["xss_match_statement"] = flattenXSSMatchStatement(s.XssMatchStatement)
	}

	return m
}

func flattenAndStatement(a *awstypes.AndStatement) interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"statement": flattenStatements(a.Statements),
	}

	return []interface{}{m}
}

func flattenByteMatchStatement(b *awstypes.ByteMatchStatement) interface{} {
	if b == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"field_to_match":        flattenFieldToMatch(b.FieldToMatch),
		"positional_constraint": b.PositionalConstraint,
		"search_string":         string(b.SearchString),
		"text_transformation":   flattenTextTransformations(b.TextTransformations),
	}

	return []interface{}{m}
}

func flattenFieldToMatch(f *awstypes.FieldToMatch) interface{} {
	if f == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if f.AllQueryArguments != nil {
		m["all_query_arguments"] = make([]map[string]interface{}, 1)
	}

	if f.Body != nil {
		m["body"] = flattenBody(f.Body)
	}

	if f.Cookies != nil {
		m["cookies"] = flattenCookies(f.Cookies)
	}

	if f.HeaderOrder != nil {
		m["header_order"] = flattenHeaderOrder(f.HeaderOrder)
	}

	if f.Headers != nil {
		m["headers"] = flattenHeaders(f.Headers)
	}

	if f.JA3Fingerprint != nil {
		m["ja3_fingerprint"] = flattenJA3Fingerprint(f.JA3Fingerprint)
	}

	if f.JsonBody != nil {
		m["json_body"] = flattenJSONBody(f.JsonBody)
	}

	if f.Method != nil {
		m["method"] = make([]map[string]interface{}, 1)
	}

	if f.QueryString != nil {
		m["query_string"] = make([]map[string]interface{}, 1)
	}

	if f.SingleHeader != nil {
		m["single_header"] = flattenSingleHeader(f.SingleHeader)
	}

	if f.SingleQueryArgument != nil {
		m["single_query_argument"] = flattenSingleQueryArgument(f.SingleQueryArgument)
	}

	if f.UriPath != nil {
		m["uri_path"] = make([]map[string]interface{}, 1)
	}

	return []interface{}{m}
}

func flattenForwardedIPConfig(f *awstypes.ForwardedIPConfig) interface{} {
	if f == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"fallback_behavior": f.FallbackBehavior,
		"header_name":       aws.ToString(f.HeaderName),
	}

	return []interface{}{m}
}

func flattenIPSetForwardedIPConfig(i *awstypes.IPSetForwardedIPConfig) interface{} {
	if i == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"fallback_behavior": i.FallbackBehavior,
		"header_name":       aws.ToString(i.HeaderName),
		"position":          i.Position,
	}

	return []interface{}{m}
}

func flattenCookies(c *awstypes.Cookies) interface{} {
	if c == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"match_scope":       c.MatchScope,
		"oversize_handling": c.OversizeHandling,
		"match_pattern":     flattenCookiesMatchPattern(c.MatchPattern),
	}

	return []interface{}{m}
}

func flattenCookiesMatchPattern(c *awstypes.CookieMatchPattern) interface{} {
	if c == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"included_cookies": aws.StringSlice(c.IncludedCookies),
		"excluded_cookies": aws.StringSlice(c.ExcludedCookies),
	}

	if c.All != nil {
		m["all"] = make([]map[string]interface{}, 1)
	}

	return []interface{}{m}
}

func flattenJA3Fingerprint(j *awstypes.JA3Fingerprint) interface{} {
	if j == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"fallback_behavior": j.FallbackBehavior,
	}

	return []interface{}{m}
}

func flattenJSONBody(b *awstypes.JsonBody) interface{} {
	if b == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"invalid_fallback_behavior": b.InvalidFallbackBehavior,
		"match_pattern":             flattenJSONMatchPattern(b.MatchPattern),
		"match_scope":               b.MatchScope,
		"oversize_handling":         b.OversizeHandling,
	}

	return []interface{}{m}
}

func flattenBody(b *awstypes.Body) interface{} {
	if b == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"oversize_handling": b.OversizeHandling,
	}

	return []interface{}{m}
}

func flattenJSONMatchPattern(p *awstypes.JsonMatchPattern) []interface{} {
	if p == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"included_paths": p.IncludedPaths,
	}

	if p.All != nil {
		m["all"] = make([]map[string]interface{}, 1)
	}

	return []interface{}{m}
}

func flattenSingleHeader(s *awstypes.SingleHeader) interface{} {
	if s == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrName: aws.ToString(s.Name),
	}

	return []interface{}{m}
}

func flattenSingleQueryArgument(s *awstypes.SingleQueryArgument) interface{} {
	if s == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrName: aws.ToString(s.Name),
	}

	return []interface{}{m}
}

func flattenTextTransformations(l []awstypes.TextTransformation) []interface{} {
	out := make([]interface{}, len(l))
	for i, t := range l {
		m := make(map[string]interface{})
		m[names.AttrPriority] = t.Priority
		m[names.AttrType] = t.Type
		out[i] = m
	}
	return out
}

func flattenIPSetReferenceStatement(i *awstypes.IPSetReferenceStatement) interface{} {
	if i == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrARN:                aws.ToString(i.ARN),
		"ip_set_forwarded_ip_config": flattenIPSetForwardedIPConfig(i.IPSetForwardedIPConfig),
	}

	return []interface{}{m}
}

func flattenGeoMatchStatement(g *awstypes.GeoMatchStatement) interface{} {
	if g == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"country_codes":       g.CountryCodes,
		"forwarded_ip_config": flattenForwardedIPConfig(g.ForwardedIPConfig),
	}

	return []interface{}{m}
}

func flattenLabelMatchStatement(l *awstypes.LabelMatchStatement) interface{} {
	if l == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrKey:   aws.ToString(l.Key),
		names.AttrScope: l.Scope,
	}

	return []interface{}{m}
}

func flattenNotStatement(a *awstypes.NotStatement) interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"statement": []interface{}{flattenStatement(a.Statement)},
	}

	return []interface{}{m}
}

func flattenOrStatement(a *awstypes.OrStatement) interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"statement": flattenStatements(a.Statements),
	}

	return []interface{}{m}
}

func flattenRegexMatchStatement(r *awstypes.RegexMatchStatement) interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"regex_string":        aws.ToString(r.RegexString),
		"field_to_match":      flattenFieldToMatch(r.FieldToMatch),
		"text_transformation": flattenTextTransformations(r.TextTransformations),
	}

	return []interface{}{m}
}

func flattenRegexPatternSetReferenceStatement(r *awstypes.RegexPatternSetReferenceStatement) interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrARN:         aws.ToString(r.ARN),
		"field_to_match":      flattenFieldToMatch(r.FieldToMatch),
		"text_transformation": flattenTextTransformations(r.TextTransformations),
	}

	return []interface{}{m}
}

func flattenSizeConstraintStatement(s *awstypes.SizeConstraintStatement) interface{} {
	if s == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"comparison_operator": s.ComparisonOperator,
		"field_to_match":      flattenFieldToMatch(s.FieldToMatch),
		names.AttrSize:        s.Size,
		"text_transformation": flattenTextTransformations(s.TextTransformations),
	}

	return []interface{}{m}
}

func flattenSQLiMatchStatement(s *awstypes.SqliMatchStatement) interface{} {
	if s == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"field_to_match":      flattenFieldToMatch(s.FieldToMatch),
		"sensitivity_level":   s.SensitivityLevel,
		"text_transformation": flattenTextTransformations(s.TextTransformations),
	}

	return []interface{}{m}
}

func flattenXSSMatchStatement(s *awstypes.XssMatchStatement) interface{} {
	if s == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"field_to_match":      flattenFieldToMatch(s.FieldToMatch),
		"text_transformation": flattenTextTransformations(s.TextTransformations),
	}

	return []interface{}{m}
}

func flattenVisibilityConfig(config *awstypes.VisibilityConfig) interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_metrics_enabled": aws.Bool(config.CloudWatchMetricsEnabled),
		names.AttrMetricName:         aws.ToString(config.MetricName),
		"sampled_requests_enabled":   aws.Bool(config.SampledRequestsEnabled),
	}

	return []interface{}{m}
}

func flattenHeaderOrder(s *awstypes.HeaderOrder) interface{} {
	if s == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"oversize_handling": s.OversizeHandling,
	}

	return []interface{}{m}
}

func flattenHeaders(s *awstypes.Headers) interface{} {
	if s == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"match_scope":       s.MatchScope,
		"match_pattern":     flattenHeaderMatchPattern(s.MatchPattern),
		"oversize_handling": s.OversizeHandling,
	}

	return []interface{}{m}
}

func flattenHeaderMatchPattern(s *awstypes.HeaderMatchPattern) interface{} {
	if s == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if s.All != nil {
		m["all"] = make([]map[string]interface{}, 1)
	}

	if s.ExcludedHeaders != nil {
		m["excluded_headers"] = s.ExcludedHeaders
	}

	if s.IncludedHeaders != nil {
		m["included_headers"] = s.IncludedHeaders
	}

	return []interface{}{m}
}

func flattenWebACLRootStatement(s *awstypes.Statement) interface{} {
	if s == nil {
		return []interface{}{}
	}

	return []interface{}{flattenWebACLStatement(s)}
}

func flattenWebACLStatement(s *awstypes.Statement) map[string]interface{} {
	if s == nil {
		return map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if s.AndStatement != nil {
		m["and_statement"] = flattenAndStatement(s.AndStatement)
	}

	if s.ByteMatchStatement != nil {
		m["byte_match_statement"] = flattenByteMatchStatement(s.ByteMatchStatement)
	}

	if s.IPSetReferenceStatement != nil {
		m["ip_set_reference_statement"] = flattenIPSetReferenceStatement(s.IPSetReferenceStatement)
	}

	if s.GeoMatchStatement != nil {
		m["geo_match_statement"] = flattenGeoMatchStatement(s.GeoMatchStatement)
	}

	if s.LabelMatchStatement != nil {
		m["label_match_statement"] = flattenLabelMatchStatement(s.LabelMatchStatement)
	}

	if s.ManagedRuleGroupStatement != nil {
		m["managed_rule_group_statement"] = flattenManagedRuleGroupStatement(s.ManagedRuleGroupStatement)
	}

	if s.NotStatement != nil {
		m["not_statement"] = flattenNotStatement(s.NotStatement)
	}

	if s.OrStatement != nil {
		m["or_statement"] = flattenOrStatement(s.OrStatement)
	}

	if s.RateBasedStatement != nil {
		m["rate_based_statement"] = flattenRateBasedStatement(s.RateBasedStatement)
	}

	if s.RegexMatchStatement != nil {
		m["regex_match_statement"] = flattenRegexMatchStatement(s.RegexMatchStatement)
	}

	if s.RegexPatternSetReferenceStatement != nil {
		m["regex_pattern_set_reference_statement"] = flattenRegexPatternSetReferenceStatement(s.RegexPatternSetReferenceStatement)
	}

	if s.RuleGroupReferenceStatement != nil {
		m["rule_group_reference_statement"] = flattenRuleGroupReferenceStatement(s.RuleGroupReferenceStatement)
	}

	if s.SizeConstraintStatement != nil {
		m["size_constraint_statement"] = flattenSizeConstraintStatement(s.SizeConstraintStatement)
	}

	if s.SqliMatchStatement != nil {
		m["sqli_match_statement"] = flattenSQLiMatchStatement(s.SqliMatchStatement)
	}

	if s.XssMatchStatement != nil {
		m["xss_match_statement"] = flattenXSSMatchStatement(s.XssMatchStatement)
	}

	return m
}

func flattenWebACLRules(r []awstypes.Rule) interface{} {
	out := make([]map[string]interface{}, len(r))
	for i, rule := range r {
		m := make(map[string]interface{})
		m[names.AttrAction] = flattenRuleAction(rule.Action)
		m["captcha_config"] = flattenCaptchaConfig(rule.CaptchaConfig)
		m["override_action"] = flattenOverrideAction(rule.OverrideAction)
		m[names.AttrName] = aws.ToString(rule.Name)
		m[names.AttrPriority] = rule.Priority
		m["rule_label"] = flattenRuleLabels(rule.RuleLabels)
		m["statement"] = flattenWebACLRootStatement(rule.Statement)
		m["visibility_config"] = flattenVisibilityConfig(rule.VisibilityConfig)
		out[i] = m
	}

	return out
}

func flattenOverrideAction(a *awstypes.OverrideAction) interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if a.Count != nil {
		m["count"] = make([]map[string]interface{}, 1)
	}

	if a.None != nil {
		m["none"] = make([]map[string]interface{}, 1)
	}

	return []interface{}{m}
}

func flattenDefaultAction(a *awstypes.DefaultAction) interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if a.Allow != nil {
		m["allow"] = flattenAllow(a.Allow)
	}

	if a.Block != nil {
		m["block"] = flattenBlock(a.Block)
	}

	return []interface{}{m}
}

func flattenManagedRuleGroupStatement(apiObject *awstypes.ManagedRuleGroupStatement) interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}

	if apiObject.RuleActionOverrides != nil {
		tfMap["rule_action_override"] = flattenRuleActionOverrides(apiObject.RuleActionOverrides)
	}

	if apiObject.ScopeDownStatement != nil {
		tfMap["scope_down_statement"] = []interface{}{flattenStatement(apiObject.ScopeDownStatement)}
	}

	if apiObject.VendorName != nil {
		tfMap["vendor_name"] = aws.ToString(apiObject.VendorName)
	}

	if apiObject.Version != nil {
		tfMap[names.AttrVersion] = aws.ToString(apiObject.Version)
	}

	if apiObject.ManagedRuleGroupConfigs != nil {
		tfMap["managed_rule_group_configs"] = flattenManagedRuleGroupConfigs(apiObject.ManagedRuleGroupConfigs)
	}

	return []interface{}{tfMap}
}

func flattenManagedRuleGroupConfigs(c []awstypes.ManagedRuleGroupConfig) []interface{} {
	if len(c) == 0 {
		return nil
	}

	var out []interface{}

	for _, config := range c {
		m := make(map[string]interface{})
		if config.AWSManagedRulesACFPRuleSet != nil {
			m["aws_managed_rules_acfp_rule_set"] = flattenManagedRulesACFPRuleSet(config.AWSManagedRulesACFPRuleSet)
		}
		if config.AWSManagedRulesBotControlRuleSet != nil {
			m["aws_managed_rules_bot_control_rule_set"] = flattenManagedRulesBotControlRuleSet(config.AWSManagedRulesBotControlRuleSet)
		}
		if config.AWSManagedRulesATPRuleSet != nil {
			m["aws_managed_rules_atp_rule_set"] = flattenManagedRulesATPRuleSet(config.AWSManagedRulesATPRuleSet)
		}
		if config.LoginPath != nil {
			m["login_path"] = aws.ToString(config.LoginPath)
		}

		m["payload_type"] = config.PayloadType

		if config.PasswordField != nil {
			m["password_field"] = flattenPasswordField(config.PasswordField)
		}
		if config.UsernameField != nil {
			m["username_field"] = flattenUsernameField(config.UsernameField)
		}

		out = append(out, m)
	}

	return out
}

func flattenAddressFields(apiObjects []awstypes.AddressField) []interface{} {
	if apiObjects == nil {
		return nil
	}

	var identifiers []*string
	for _, apiObject := range apiObjects {
		identifiers = append(identifiers, apiObject.Identifier)
	}

	return []interface{}{
		map[string]interface{}{
			"identifiers": aws.ToStringSlice(identifiers),
		},
	}
}

func flattenEmailField(apiObject *awstypes.EmailField) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrIdentifier: aws.ToString(apiObject.Identifier),
	}

	return []interface{}{m}
}

func flattenPasswordField(apiObject *awstypes.PasswordField) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrIdentifier: aws.ToString(apiObject.Identifier),
	}

	return []interface{}{m}
}

func flattenPhoneNumberFields(apiObjects []awstypes.PhoneNumberField) []interface{} {
	if apiObjects == nil {
		return nil
	}

	var identifiers []*string
	for _, apiObject := range apiObjects {
		identifiers = append(identifiers, apiObject.Identifier)
	}

	return []interface{}{
		map[string]interface{}{
			"identifiers": aws.ToStringSlice(identifiers),
		},
	}
}

func flattenUsernameField(apiObject *awstypes.UsernameField) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrIdentifier: aws.ToString(apiObject.Identifier),
	}

	return []interface{}{m}
}

func flattenManagedRulesBotControlRuleSet(apiObject *awstypes.AWSManagedRulesBotControlRuleSet) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"enable_machine_learning": aws.ToBool(apiObject.EnableMachineLearning),
		"inspection_level":        apiObject.InspectionLevel,
	}

	return []interface{}{m}
}

func flattenManagedRulesACFPRuleSet(apiObject *awstypes.AWSManagedRulesACFPRuleSet) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"enable_regex_in_path":   aws.Bool(apiObject.EnableRegexInPath),
		"creation_path":          aws.ToString(apiObject.CreationPath),
		"registration_page_path": aws.ToString(apiObject.RegistrationPagePath),
	}
	if apiObject.RequestInspection != nil {
		m["request_inspection"] = flattenRequestInspectionACFP(apiObject.RequestInspection)
	}
	if apiObject.ResponseInspection != nil {
		m["response_inspection"] = flattenResponseInspection(apiObject.ResponseInspection)
	}

	return []interface{}{m}
}

func flattenManagedRulesATPRuleSet(apiObject *awstypes.AWSManagedRulesATPRuleSet) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"enable_regex_in_path": aws.Bool(apiObject.EnableRegexInPath),
		"login_path":           aws.ToString(apiObject.LoginPath),
	}
	if apiObject.RequestInspection != nil {
		m["request_inspection"] = flattenRequestInspection(apiObject.RequestInspection)
	}
	if apiObject.ResponseInspection != nil {
		m["response_inspection"] = flattenResponseInspection(apiObject.ResponseInspection)
	}

	return []interface{}{m}
}

func flattenRequestInspectionACFP(apiObject *awstypes.RequestInspectionACFP) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"address_fields":      flattenAddressFields(apiObject.AddressFields),
		"email_field":         flattenEmailField(apiObject.EmailField),
		"password_field":      flattenPasswordField(apiObject.PasswordField),
		"payload_type":        apiObject.PayloadType,
		"phone_number_fields": flattenPhoneNumberFields(apiObject.PhoneNumberFields),
		"username_field":      flattenUsernameField(apiObject.UsernameField),
	}

	return []interface{}{m}
}

func flattenRequestInspection(apiObject *awstypes.RequestInspection) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"password_field": flattenPasswordField(apiObject.PasswordField),
		"payload_type":   apiObject.PayloadType,
		"username_field": flattenUsernameField(apiObject.UsernameField),
	}

	return []interface{}{m}
}

func flattenResponseInspection(apiObject *awstypes.ResponseInspection) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}
	if apiObject.BodyContains != nil {
		m["body_contains"] = flattenBodyContains(apiObject.BodyContains)
	}
	if apiObject.Header != nil {
		m[names.AttrHeader] = flattenHeader(apiObject.Header)
	}
	if apiObject.Json != nil {
		m[names.AttrJSON] = flattenResponseInspectionJSON(apiObject.Json)
	}
	if apiObject.StatusCode != nil {
		m[names.AttrStatusCode] = flattenStatusCode(apiObject.StatusCode)
	}

	return []interface{}{m}
}

func flattenBodyContains(apiObject *awstypes.ResponseInspectionBodyContains) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"failure_strings": apiObject.FailureStrings,
		"success_strings": apiObject.SuccessStrings,
	}

	return []interface{}{m}
}

func flattenHeader(apiObject *awstypes.ResponseInspectionHeader) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"failure_values": apiObject.FailureValues,
		"success_values": apiObject.SuccessValues,
	}

	return []interface{}{m}
}

func flattenResponseInspectionJSON(apiObject *awstypes.ResponseInspectionJson) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"failure_values":     apiObject.FailureValues,
		names.AttrIdentifier: aws.ToString(apiObject.Identifier),
		"success_values":     apiObject.SuccessValues,
	}

	return []interface{}{m}
}

func flattenStatusCode(apiObject *awstypes.ResponseInspectionStatusCode) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"failure_codes": flex.FlattenInt32ValueSet(apiObject.FailureCodes),
		"success_codes": flex.FlattenInt32ValueSet(apiObject.SuccessCodes),
	}

	return []interface{}{m}
}

func flattenRateLimitCookie(apiObject *awstypes.RateLimitCookie) []interface{} {
	if apiObject == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			names.AttrName:        aws.ToString(apiObject.Name),
			"text_transformation": flattenTextTransformations(apiObject.TextTransformations),
		},
	}
}

func flattenRateLimitHeader(apiObject *awstypes.RateLimitHeader) []interface{} {
	if apiObject == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			names.AttrName:        aws.ToString(apiObject.Name),
			"text_transformation": flattenTextTransformations(apiObject.TextTransformations),
		},
	}
}

func flattenRateLimitLabelNamespace(apiObject *awstypes.RateLimitLabelNamespace) []interface{} {
	if apiObject == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			names.AttrNamespace: aws.ToString(apiObject.Namespace),
		},
	}
}

func flattenRateLimitQueryArgument(apiObject *awstypes.RateLimitQueryArgument) []interface{} {
	if apiObject == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			names.AttrName:        aws.ToString(apiObject.Name),
			"text_transformation": flattenTextTransformations(apiObject.TextTransformations),
		},
	}
}

func flattenRateLimitQueryString(apiObject *awstypes.RateLimitQueryString) []interface{} {
	if apiObject == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"text_transformation": flattenTextTransformations(apiObject.TextTransformations),
		},
	}
}

func flattenRateLimitURIPath(apiObject *awstypes.RateLimitUriPath) []interface{} {
	if apiObject == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"text_transformation": flattenTextTransformations(apiObject.TextTransformations),
		},
	}
}

func flattenRateBasedStatementCustomKeys(apiObject []awstypes.RateBasedStatementCustomKey) []interface{} {
	if apiObject == nil {
		return nil
	}

	out := make([]interface{}, len(apiObject))
	for i, o := range apiObject {
		tfMap := map[string]interface{}{}

		if o.Cookie != nil {
			tfMap["cookie"] = flattenRateLimitCookie(o.Cookie)
		}
		if o.ForwardedIP != nil {
			tfMap["forwarded_ip"] = []interface{}{
				map[string]interface{}{},
			}
		}
		if o.HTTPMethod != nil {
			tfMap["http_method"] = []interface{}{
				map[string]interface{}{},
			}
		}
		if o.Header != nil {
			tfMap[names.AttrHeader] = flattenRateLimitHeader(o.Header)
		}
		if o.IP != nil {
			tfMap["ip"] = []interface{}{
				map[string]interface{}{},
			}
		}
		if o.LabelNamespace != nil {
			tfMap["label_namespace"] = flattenRateLimitLabelNamespace(o.LabelNamespace)
		}
		if o.QueryArgument != nil {
			tfMap["query_argument"] = flattenRateLimitQueryArgument(o.QueryArgument)
		}
		if o.QueryString != nil {
			tfMap["query_string"] = flattenRateLimitQueryString(o.QueryString)
		}
		if o.UriPath != nil {
			tfMap["uri_path"] = flattenRateLimitURIPath(o.UriPath)
		}
		out[i] = tfMap
	}
	return out
}

func flattenRateBasedStatement(apiObject *awstypes.RateBasedStatement) interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"aggregate_key_type":    apiObject.AggregateKeyType,
		"evaluation_window_sec": apiObject.EvaluationWindowSec,
	}

	if apiObject.ForwardedIPConfig != nil {
		tfMap["forwarded_ip_config"] = flattenForwardedIPConfig(apiObject.ForwardedIPConfig)
	}

	if apiObject.CustomKeys != nil {
		tfMap["custom_key"] = flattenRateBasedStatementCustomKeys(apiObject.CustomKeys)
	}

	if apiObject.Limit != nil {
		tfMap["limit"] = int(aws.ToInt64(apiObject.Limit))
	}

	if apiObject.ScopeDownStatement != nil {
		tfMap["scope_down_statement"] = []interface{}{flattenStatement(apiObject.ScopeDownStatement)}
	}

	return []interface{}{tfMap}
}

func flattenRuleGroupReferenceStatement(apiObject *awstypes.RuleGroupReferenceStatement) interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		names.AttrARN: aws.ToString(apiObject.ARN),
	}

	if apiObject.RuleActionOverrides != nil {
		tfMap["rule_action_override"] = flattenRuleActionOverrides(apiObject.RuleActionOverrides)
	}

	return []interface{}{tfMap}
}

func flattenRuleActionOverrides(r []awstypes.RuleActionOverride) interface{} {
	out := make([]map[string]interface{}, len(r))
	for i, override := range r {
		m := make(map[string]interface{})
		m["action_to_use"] = flattenRuleAction(override.ActionToUse)
		m[names.AttrName] = aws.ToString(override.Name)
		out[i] = m
	}

	return out
}

func flattenRegexPatternSet(r []awstypes.Regex) interface{} {
	if r == nil {
		return []interface{}{}
	}

	regexPatterns := make([]interface{}, 0)

	for _, regexPattern := range r {
		d := map[string]interface{}{
			"regex_string": aws.ToString(regexPattern.RegexString),
		}
		regexPatterns = append(regexPatterns, d)
	}

	return regexPatterns
}
