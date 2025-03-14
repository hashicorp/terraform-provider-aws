// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandRules(l []any) []awstypes.Rule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]awstypes.Rule, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandRule(rule.(map[string]any)))
	}

	return rules
}

func expandRule(m map[string]any) awstypes.Rule {
	rule := awstypes.Rule{
		Action:           expandRuleAction(m[names.AttrAction].([]any)),
		CaptchaConfig:    expandCaptchaConfig(m["captcha_config"].([]any)),
		Name:             aws.String(m[names.AttrName].(string)),
		Priority:         int32(m[names.AttrPriority].(int)),
		Statement:        expandRuleGroupRootStatement(m["statement"].([]any)),
		VisibilityConfig: expandVisibilityConfig(m["visibility_config"].([]any)),
	}

	if v, ok := m["rule_label"].(*schema.Set); ok && v.Len() > 0 {
		rule.RuleLabels = expandRuleLabels(v.List())
	}

	return rule
}

func expandCaptchaConfig(l []any) *awstypes.CaptchaConfig {
	configuration := &awstypes.CaptchaConfig{}

	if len(l) == 0 || l[0] == nil {
		return configuration
	}

	m := l[0].(map[string]any)
	if v, ok := m["immunity_time_property"]; ok {
		inner := v.([]any)
		if len(inner) == 0 || inner[0] == nil {
			return configuration
		}

		m = inner[0].(map[string]any)

		if v, ok := m["immunity_time"]; ok {
			configuration.ImmunityTimeProperty = &awstypes.ImmunityTimeProperty{
				ImmunityTime: aws.Int64(int64(v.(int))),
			}
		}
	}

	return configuration
}

func expandChallengeConfig(l []any) *awstypes.ChallengeConfig {
	configuration := &awstypes.ChallengeConfig{}

	if len(l) == 0 || l[0] == nil {
		return configuration
	}

	m := l[0].(map[string]any)
	if v, ok := m["immunity_time_property"]; ok {
		inner := v.([]any)
		if len(inner) == 0 || inner[0] == nil {
			return configuration
		}

		m = inner[0].(map[string]any)

		if v, ok := m["immunity_time"]; ok {
			configuration.ImmunityTimeProperty = &awstypes.ImmunityTimeProperty{
				ImmunityTime: aws.Int64(int64(v.(int))),
			}
		}
	}

	return configuration
}

func expandAssociationConfig(l []any) *awstypes.AssociationConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configuration := &awstypes.AssociationConfig{}

	m := l[0].(map[string]any)
	if v, ok := m["request_body"]; ok {
		inner := v.([]any)
		if len(inner) == 0 || inner[0] == nil {
			return configuration
		}

		m = inner[0].(map[string]any)
		if len(m) > 0 {
			configuration.RequestBody = make(map[string]awstypes.RequestBodyAssociatedResourceTypeConfig)
			for _, resourceType := range awstypes.AssociatedResourceType.Values("") {
				if v, ok := m[strings.ToLower(string(resourceType))]; ok {
					m := v.([]any)
					if len(m) > 0 {
						configuration.RequestBody[string(resourceType)] = expandRequestBodyConfigItem(m)
					}
				}
			}
		}
	}

	return configuration
}

func expandRequestBodyConfigItem(l []any) awstypes.RequestBodyAssociatedResourceTypeConfig {
	configuration := awstypes.RequestBodyAssociatedResourceTypeConfig{}

	if len(l) == 0 || l[0] == nil {
		return configuration
	}

	m := l[0].(map[string]any)
	if v, ok := m["default_size_inspection_limit"]; ok {
		if v != "" {
			configuration.DefaultSizeInspectionLimit = awstypes.SizeInspectionLimit(v.(string))
		}
	}

	return configuration
}

func expandRuleLabels(l []any) []awstypes.Label {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	labels := make([]awstypes.Label, 0)

	for _, label := range l {
		if label == nil {
			continue
		}
		m := label.(map[string]any)
		labels = append(labels, awstypes.Label{
			Name: aws.String(m[names.AttrName].(string)),
		})
	}

	return labels
}

func expandCountryCodes(l []any) []awstypes.CountryCode {
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

func expandRuleAction(l []any) *awstypes.RuleAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	action := &awstypes.RuleAction{}

	if v, ok := m["allow"]; ok && len(v.([]any)) > 0 {
		action.Allow = expandAllowAction(v.([]any))
	}

	if v, ok := m["block"]; ok && len(v.([]any)) > 0 {
		action.Block = expandBlockAction(v.([]any))
	}

	if v, ok := m["captcha"]; ok && len(v.([]any)) > 0 {
		action.Captcha = expandCaptchaAction(v.([]any))
	}

	if v, ok := m["challenge"]; ok && len(v.([]any)) > 0 {
		action.Challenge = expandChallengeAction(v.([]any))
	}

	if v, ok := m["count"]; ok && len(v.([]any)) > 0 {
		action.Count = expandCountAction(v.([]any))
	}

	return action
}

func expandAllowAction(l []any) *awstypes.AllowAction {
	action := &awstypes.AllowAction{}

	if len(l) == 0 || l[0] == nil {
		return action
	}

	m, ok := l[0].(map[string]any)
	if !ok {
		return action
	}

	if v, ok := m["custom_request_handling"].([]any); ok && len(v) > 0 {
		action.CustomRequestHandling = expandCustomRequestHandling(v)
	}

	return action
}

func expandBlockAction(l []any) *awstypes.BlockAction {
	action := &awstypes.BlockAction{}

	if len(l) == 0 || l[0] == nil {
		return action
	}

	m, ok := l[0].(map[string]any)
	if !ok {
		return action
	}

	if v, ok := m["custom_response"].([]any); ok && len(v) > 0 {
		action.CustomResponse = expandCustomResponse(v)
	}

	return action
}

func expandCaptchaAction(l []any) *awstypes.CaptchaAction {
	action := &awstypes.CaptchaAction{}

	if len(l) == 0 || l[0] == nil {
		return action
	}

	m, ok := l[0].(map[string]any)
	if !ok {
		return action
	}

	if v, ok := m["custom_request_handling"].([]any); ok && len(v) > 0 {
		action.CustomRequestHandling = expandCustomRequestHandling(v)
	}

	return action
}

func expandChallengeAction(l []any) *awstypes.ChallengeAction {
	action := &awstypes.ChallengeAction{}

	if len(l) == 0 || l[0] == nil {
		return action
	}

	m, ok := l[0].(map[string]any)
	if !ok {
		return action
	}

	if v, ok := m["custom_request_handling"].([]any); ok && len(v) > 0 {
		action.CustomRequestHandling = expandCustomRequestHandling(v)
	}

	return action
}

func expandCountAction(l []any) *awstypes.CountAction {
	action := &awstypes.CountAction{}

	if len(l) == 0 || l[0] == nil {
		return action
	}

	m, ok := l[0].(map[string]any)
	if !ok {
		return action
	}

	if v, ok := m["custom_request_handling"].([]any); ok && len(v) > 0 {
		action.CustomRequestHandling = expandCustomRequestHandling(v)
	}

	return action
}

func expandCustomResponseBodies(m []any) map[string]awstypes.CustomResponseBody {
	if len(m) == 0 {
		return nil
	}

	customResponseBodies := make(map[string]awstypes.CustomResponseBody, len(m))

	for _, v := range m {
		vm := v.(map[string]any)
		key := vm[names.AttrKey].(string)
		customResponseBodies[key] = awstypes.CustomResponseBody{
			Content:     aws.String(vm[names.AttrContent].(string)),
			ContentType: awstypes.ResponseContentType(vm[names.AttrContentType].(string)),
		}
	}

	return customResponseBodies
}

func expandCustomResponse(l []any) *awstypes.CustomResponse {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m, ok := l[0].(map[string]any)
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

func expandCustomRequestHandling(l []any) *awstypes.CustomRequestHandling {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	requestHandling := &awstypes.CustomRequestHandling{}

	if v, ok := m["insert_header"].(*schema.Set); ok && len(v.List()) > 0 {
		requestHandling.InsertHeaders = expandCustomHeaders(v.List())
	}

	return requestHandling
}

func expandCustomHeaders(l []any) []awstypes.CustomHTTPHeader {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	headers := make([]awstypes.CustomHTTPHeader, 0)

	for _, header := range l {
		if header == nil {
			continue
		}
		m := header.(map[string]any)

		headers = append(headers, awstypes.CustomHTTPHeader{
			Name:  aws.String(m[names.AttrName].(string)),
			Value: aws.String(m[names.AttrValue].(string)),
		})
	}

	return headers
}

func expandVisibilityConfig(l []any) *awstypes.VisibilityConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

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

func expandRuleGroupRootStatement(l []any) *awstypes.Statement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return expandStatement(m)
}

func expandStatements(l []any) []awstypes.Statement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	statements := make([]awstypes.Statement, 0)

	for _, statement := range l {
		if statement == nil {
			continue
		}
		statements = append(statements, *expandStatement(statement.(map[string]any)))
	}

	return statements
}

func expandStatement(m map[string]any) *awstypes.Statement {
	if m == nil {
		return nil
	}

	statement := &awstypes.Statement{}

	if v, ok := m["and_statement"]; ok {
		statement.AndStatement = expandAndStatement(v.([]any))
	}

	if v, ok := m["byte_match_statement"]; ok {
		statement.ByteMatchStatement = expandByteMatchStatement(v.([]any))
	}

	if v, ok := m["ip_set_reference_statement"]; ok {
		statement.IPSetReferenceStatement = expandIPSetReferenceStatement(v.([]any))
	}

	if v, ok := m["geo_match_statement"]; ok {
		statement.GeoMatchStatement = expandGeoMatchStatement(v.([]any))
	}

	if v, ok := m["label_match_statement"]; ok {
		statement.LabelMatchStatement = expandLabelMatchStatement(v.([]any))
	}

	if v, ok := m["not_statement"]; ok {
		statement.NotStatement = expandNotStatement(v.([]any))
	}

	if v, ok := m["or_statement"]; ok {
		statement.OrStatement = expandOrStatement(v.([]any))
	}

	if v, ok := m["rate_based_statement"]; ok {
		statement.RateBasedStatement = expandRateBasedStatement(v.([]any))
	}

	if v, ok := m["regex_match_statement"]; ok {
		statement.RegexMatchStatement = expandRegexMatchStatement(v.([]any))
	}

	if v, ok := m["regex_pattern_set_reference_statement"]; ok {
		statement.RegexPatternSetReferenceStatement = expandRegexPatternSetReferenceStatement(v.([]any))
	}

	if v, ok := m["size_constraint_statement"]; ok {
		statement.SizeConstraintStatement = expandSizeConstraintStatement(v.([]any))
	}

	if v, ok := m["sqli_match_statement"]; ok {
		statement.SqliMatchStatement = expandSQLiMatchStatement(v.([]any))
	}

	if v, ok := m["xss_match_statement"]; ok {
		statement.XssMatchStatement = expandXSSMatchStatement(v.([]any))
	}

	return statement
}

func expandAndStatement(l []any) *awstypes.AndStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.AndStatement{
		Statements: expandStatements(m["statement"].([]any)),
	}
}

func expandByteMatchStatement(l []any) *awstypes.ByteMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.ByteMatchStatement{
		FieldToMatch:         expandFieldToMatch(m["field_to_match"].([]any)),
		PositionalConstraint: awstypes.PositionalConstraint(m["positional_constraint"].(string)),
		SearchString:         []byte(m["search_string"].(string)),
		TextTransformations:  expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandFieldToMatch(l []any) *awstypes.FieldToMatch {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	f := &awstypes.FieldToMatch{}

	if v, ok := m["all_query_arguments"]; ok && len(v.([]any)) > 0 {
		f.AllQueryArguments = &awstypes.AllQueryArguments{}
	}

	if v, ok := m["body"]; ok && len(v.([]any)) > 0 {
		f.Body = expandBody(v.([]any))
	}

	if v, ok := m["cookies"]; ok && len(v.([]any)) > 0 {
		f.Cookies = expandCookies(m["cookies"].([]any))
	}

	if v, ok := m["header_order"]; ok && len(v.([]any)) > 0 {
		f.HeaderOrder = expandHeaderOrder(m["header_order"].([]any))
	}

	if v, ok := m["headers"]; ok && len(v.([]any)) > 0 {
		f.Headers = expandHeaders(m["headers"].([]any))
	}

	if v, ok := m["json_body"]; ok && len(v.([]any)) > 0 {
		f.JsonBody = expandJSONBody(v.([]any))
	}

	if v, ok := m["method"]; ok && len(v.([]any)) > 0 {
		f.Method = &awstypes.Method{}
	}

	if v, ok := m["query_string"]; ok && len(v.([]any)) > 0 {
		f.QueryString = &awstypes.QueryString{}
	}

	if v, ok := m["single_header"]; ok && len(v.([]any)) > 0 {
		f.SingleHeader = expandSingleHeader(m["single_header"].([]any))
	}

	if v, ok := m["ja3_fingerprint"]; ok && len(v.([]any)) > 0 {
		f.JA3Fingerprint = expandJA3Fingerprint(v.([]any))
	}

	if v, ok := m["ja4_fingerprint"]; ok && len(v.([]any)) > 0 {
		f.JA4Fingerprint = expandJA4Fingerprint(v.([]any))
	}

	if v, ok := m["single_query_argument"]; ok && len(v.([]any)) > 0 {
		f.SingleQueryArgument = expandSingleQueryArgument(m["single_query_argument"].([]any))
	}

	if v, ok := m["uri_path"]; ok && len(v.([]any)) > 0 {
		f.UriPath = &awstypes.UriPath{}
	}

	return f
}

func expandForwardedIPConfig(l []any) *awstypes.ForwardedIPConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.ForwardedIPConfig{
		FallbackBehavior: awstypes.FallbackBehavior(m["fallback_behavior"].(string)),
		HeaderName:       aws.String(m["header_name"].(string)),
	}
}

func expandIPSetForwardedIPConfig(l []any) *awstypes.IPSetForwardedIPConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.IPSetForwardedIPConfig{
		FallbackBehavior: awstypes.FallbackBehavior(m["fallback_behavior"].(string)),
		HeaderName:       aws.String(m["header_name"].(string)),
		Position:         awstypes.ForwardedIPPosition(m["position"].(string)),
	}
}

func expandCookies(l []any) *awstypes.Cookies {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	cookies := &awstypes.Cookies{
		MatchScope:       awstypes.MapMatchScope(m["match_scope"].(string)),
		OversizeHandling: awstypes.OversizeHandling(m["oversize_handling"].(string)),
	}

	if v, ok := m["match_pattern"]; ok && len(v.([]any)) > 0 {
		cookies.MatchPattern = expandCookieMatchPattern(v.([]any))
	}

	return cookies
}

func expandCookieMatchPattern(l []any) *awstypes.CookieMatchPattern {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	CookieMatchPattern := &awstypes.CookieMatchPattern{}

	if v, ok := m["included_cookies"]; ok && len(v.([]any)) > 0 {
		CookieMatchPattern.IncludedCookies = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := m["excluded_cookies"]; ok && len(v.([]any)) > 0 {
		CookieMatchPattern.ExcludedCookies = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := m["all"].([]any); ok && len(v) > 0 {
		CookieMatchPattern.All = &awstypes.All{}
	}

	return CookieMatchPattern
}

func expandJSONBody(l []any) *awstypes.JsonBody {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	jsonBody := &awstypes.JsonBody{
		MatchScope:       awstypes.JsonMatchScope(m["match_scope"].(string)),
		OversizeHandling: awstypes.OversizeHandling(m["oversize_handling"].(string)),
		MatchPattern:     expandJSONMatchPattern(m["match_pattern"].([]any)),
	}

	if v, ok := m["invalid_fallback_behavior"].(string); ok && v != "" {
		jsonBody.InvalidFallbackBehavior = awstypes.BodyParsingFallbackBehavior(v)
	}

	return jsonBody
}

func expandBody(l []any) *awstypes.Body {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	body := &awstypes.Body{}

	if v, ok := m["oversize_handling"].(string); ok && v != "" {
		body.OversizeHandling = awstypes.OversizeHandling(v)
	}

	return body
}

func expandJA3Fingerprint(l []any) *awstypes.JA3Fingerprint {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	ja3fingerprint := &awstypes.JA3Fingerprint{
		FallbackBehavior: awstypes.FallbackBehavior(m["fallback_behavior"].(string)),
	}

	return ja3fingerprint
}

func expandJA4Fingerprint(l []any) *awstypes.JA4Fingerprint {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	ja4fingerprint := &awstypes.JA4Fingerprint{
		FallbackBehavior: awstypes.FallbackBehavior(m["fallback_behavior"].(string)),
	}

	return ja4fingerprint
}

func expandJSONMatchPattern(l []any) *awstypes.JsonMatchPattern {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	jsonMatchPattern := &awstypes.JsonMatchPattern{}

	if v, ok := m["all"].([]any); ok && len(v) > 0 {
		jsonMatchPattern.All = &awstypes.All{}
	}

	if v, ok := m["included_paths"]; ok && len(v.([]any)) > 0 {
		jsonMatchPattern.IncludedPaths = flex.ExpandStringValueList(v.([]any))
	}

	return jsonMatchPattern
}

func expandSingleHeader(l []any) *awstypes.SingleHeader {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.SingleHeader{
		Name: aws.String(m[names.AttrName].(string)),
	}
}

func expandSingleQueryArgument(l []any) *awstypes.SingleQueryArgument {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.SingleQueryArgument{
		Name: aws.String(m[names.AttrName].(string)),
	}
}

func expandTextTransformations(l []any) []awstypes.TextTransformation {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]awstypes.TextTransformation, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandTextTransformation(rule.(map[string]any)))
	}

	return rules
}

func expandTextTransformation(m map[string]any) awstypes.TextTransformation {
	return awstypes.TextTransformation{
		Priority: int32(m[names.AttrPriority].(int)),
		Type:     awstypes.TextTransformationType(m[names.AttrType].(string)),
	}
}

func expandIPSetReferenceStatement(l []any) *awstypes.IPSetReferenceStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	statement := &awstypes.IPSetReferenceStatement{
		ARN: aws.String(m[names.AttrARN].(string)),
	}

	if v, ok := m["ip_set_forwarded_ip_config"]; ok {
		statement.IPSetForwardedIPConfig = expandIPSetForwardedIPConfig(v.([]any))
	}

	return statement
}

func expandGeoMatchStatement(l []any) *awstypes.GeoMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	statement := &awstypes.GeoMatchStatement{
		CountryCodes: expandCountryCodes(m["country_codes"].([]any)),
	}

	if v, ok := m["forwarded_ip_config"]; ok {
		statement.ForwardedIPConfig = expandForwardedIPConfig(v.([]any))
	}

	return statement
}

func expandLabelMatchStatement(l []any) *awstypes.LabelMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	statement := &awstypes.LabelMatchStatement{
		Key:   aws.String(m[names.AttrKey].(string)),
		Scope: awstypes.LabelMatchScope(m[names.AttrScope].(string)),
	}

	return statement
}

func expandNotStatement(l []any) *awstypes.NotStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	s := m["statement"].([]any)

	if len(s) == 0 || s[0] == nil {
		return nil
	}

	m = s[0].(map[string]any)

	return &awstypes.NotStatement{
		Statement: expandStatement(m),
	}
}

func expandOrStatement(l []any) *awstypes.OrStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.OrStatement{
		Statements: expandStatements(m["statement"].([]any)),
	}
}

func expandRegexMatchStatement(l []any) *awstypes.RegexMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.RegexMatchStatement{
		RegexString:         aws.String(m["regex_string"].(string)),
		FieldToMatch:        expandFieldToMatch(m["field_to_match"].([]any)),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandRegexPatternSetReferenceStatement(l []any) *awstypes.RegexPatternSetReferenceStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.RegexPatternSetReferenceStatement{
		ARN:                 aws.String(m[names.AttrARN].(string)),
		FieldToMatch:        expandFieldToMatch(m["field_to_match"].([]any)),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandSizeConstraintStatement(l []any) *awstypes.SizeConstraintStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.SizeConstraintStatement{
		ComparisonOperator:  awstypes.ComparisonOperator(m["comparison_operator"].(string)),
		FieldToMatch:        expandFieldToMatch(m["field_to_match"].([]any)),
		Size:                int64(m[names.AttrSize].(int)),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandSQLiMatchStatement(l []any) *awstypes.SqliMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.SqliMatchStatement{
		FieldToMatch:        expandFieldToMatch(m["field_to_match"].([]any)),
		SensitivityLevel:    awstypes.SensitivityLevel(m["sensitivity_level"].(string)),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandXSSMatchStatement(l []any) *awstypes.XssMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.XssMatchStatement{
		FieldToMatch:        expandFieldToMatch(m["field_to_match"].([]any)),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandHeaderOrder(l []any) *awstypes.HeaderOrder {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.HeaderOrder{
		OversizeHandling: awstypes.OversizeHandling(m["oversize_handling"].(string)),
	}
}

func expandHeaders(l []any) *awstypes.Headers {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.Headers{
		MatchPattern:     expandHeaderMatchPattern(m["match_pattern"].([]any)),
		MatchScope:       awstypes.MapMatchScope(m["match_scope"].(string)),
		OversizeHandling: awstypes.OversizeHandling(m["oversize_handling"].(string)),
	}
}

func expandHeaderMatchPattern(l []any) *awstypes.HeaderMatchPattern {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	f := &awstypes.HeaderMatchPattern{}

	if v, ok := m["all"]; ok && len(v.([]any)) > 0 {
		f.All = &awstypes.All{}
	}

	if v, ok := m["included_headers"]; ok && len(v.([]any)) > 0 {
		f.IncludedHeaders = flex.ExpandStringValueList(m["included_headers"].([]any))
	}

	if v, ok := m["excluded_headers"]; ok && len(v.([]any)) > 0 {
		f.ExcludedHeaders = flex.ExpandStringValueList(m["excluded_headers"].([]any))
	}

	return f
}

func expandWebACLRulesJSON(rawRules string) ([]awstypes.Rule, error) {
	// Backwards compatibility.
	if rawRules == "" {
		return nil, errors.New("decoding JSON: unexpected end of JSON input")
	}

	var temp []any
	err := tfjson.DecodeFromBytes([]byte(rawRules), &temp)
	if err != nil {
		return nil, fmt.Errorf("decoding JSON: %w", err)
	}

	for _, v := range temp {
		walkWebACLJSON(reflect.ValueOf(v))
	}

	out, err := tfjson.EncodeToBytes(temp)
	if err != nil {
		return nil, err
	}

	var rules []awstypes.Rule
	err = tfjson.DecodeFromBytes(out, &rules)
	if err != nil {
		return nil, err
	}

	for i, r := range rules {
		if reflect.ValueOf(r).IsZero() {
			return nil, fmt.Errorf("invalid ACL Rule supplied at index (%d)", i)
		}
	}
	return rules, nil
}

func walkWebACLJSON(v reflect.Value) {
	m := map[string][]struct {
		key        string
		outputType any
	}{
		"ByteMatchStatement": {
			{key: "SearchString", outputType: []byte{}},
		},
	}

	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		for _, k := range v.MapKeys() {
			if val, ok := m[k.String()]; ok {
				st := v.MapIndex(k).Interface().(map[string]any)
				for _, va := range val {
					if st[va.key] == nil {
						continue
					}
					str := st[va.key]
					switch reflect.ValueOf(va.outputType).Kind() {
					case reflect.Slice, reflect.Array:
						switch reflect.ValueOf(va.outputType).Type().Elem().Kind() {
						case reflect.Uint8:
							base64String := itypes.Base64Encode([]byte(str.(string)))
							st[va.key] = base64String
						default:
						}
					default:
					}
				}
			} else {
				walkWebACLJSON(v.MapIndex(k))
			}
		}
	case reflect.Array, reflect.Slice:
		for i := range v.Len() {
			walkWebACLJSON(v.Index(i))
		}
	default:
	}
}

func expandWebACLRules(l []any) []awstypes.Rule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]awstypes.Rule, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandWebACLRule(rule.(map[string]any)))
	}

	return rules
}

func expandWebACLRule(m map[string]any) awstypes.Rule {
	rule := awstypes.Rule{
		Action:           expandRuleAction(m[names.AttrAction].([]any)),
		CaptchaConfig:    expandCaptchaConfig(m["captcha_config"].([]any)),
		ChallengeConfig:  expandChallengeConfig(m["challenge_config"].([]any)),
		Name:             aws.String(m[names.AttrName].(string)),
		OverrideAction:   expandOverrideAction(m["override_action"].([]any)),
		Priority:         int32(m[names.AttrPriority].(int)),
		Statement:        expandWebACLRootStatement(m["statement"].([]any)),
		VisibilityConfig: expandVisibilityConfig(m["visibility_config"].([]any)),
	}

	if v, ok := m["rule_label"].(*schema.Set); ok && v.Len() > 0 {
		rule.RuleLabels = expandRuleLabels(v.List())
	}

	return rule
}

func expandOverrideAction(l []any) *awstypes.OverrideAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	action := &awstypes.OverrideAction{}

	if v, ok := m["count"]; ok && len(v.([]any)) > 0 {
		action.Count = &awstypes.CountAction{}
	}

	if v, ok := m["none"]; ok && len(v.([]any)) > 0 {
		action.None = &awstypes.NoneAction{}
	}

	return action
}

func expandDefaultAction(l []any) *awstypes.DefaultAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	action := &awstypes.DefaultAction{}

	if v, ok := m["allow"]; ok && len(v.([]any)) > 0 {
		action.Allow = expandAllowAction(v.([]any))
	}

	if v, ok := m["block"]; ok && len(v.([]any)) > 0 {
		action.Block = expandBlockAction(v.([]any))
	}

	return action
}

func expandWebACLRootStatement(l []any) *awstypes.Statement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return expandWebACLStatement(m)
}

func expandWebACLStatement(m map[string]any) *awstypes.Statement {
	if m == nil {
		return nil
	}

	statement := &awstypes.Statement{}

	if v, ok := m["and_statement"]; ok {
		statement.AndStatement = expandAndStatement(v.([]any))
	}

	if v, ok := m["byte_match_statement"]; ok {
		statement.ByteMatchStatement = expandByteMatchStatement(v.([]any))
	}

	if v, ok := m["ip_set_reference_statement"]; ok {
		statement.IPSetReferenceStatement = expandIPSetReferenceStatement(v.([]any))
	}

	if v, ok := m["geo_match_statement"]; ok {
		statement.GeoMatchStatement = expandGeoMatchStatement(v.([]any))
	}

	if v, ok := m["label_match_statement"]; ok {
		statement.LabelMatchStatement = expandLabelMatchStatement(v.([]any))
	}

	if v, ok := m["managed_rule_group_statement"]; ok {
		statement.ManagedRuleGroupStatement = expandManagedRuleGroupStatement(v.([]any))
	}

	if v, ok := m["not_statement"]; ok {
		statement.NotStatement = expandNotStatement(v.([]any))
	}

	if v, ok := m["or_statement"]; ok {
		statement.OrStatement = expandOrStatement(v.([]any))
	}

	if v, ok := m["rate_based_statement"]; ok {
		statement.RateBasedStatement = expandRateBasedStatement(v.([]any))
	}

	if v, ok := m["regex_match_statement"]; ok {
		statement.RegexMatchStatement = expandRegexMatchStatement(v.([]any))
	}

	if v, ok := m["regex_pattern_set_reference_statement"]; ok {
		statement.RegexPatternSetReferenceStatement = expandRegexPatternSetReferenceStatement(v.([]any))
	}

	if v, ok := m["rule_group_reference_statement"]; ok {
		statement.RuleGroupReferenceStatement = expandRuleGroupReferenceStatement(v.([]any))
	}

	if v, ok := m["size_constraint_statement"]; ok {
		statement.SizeConstraintStatement = expandSizeConstraintStatement(v.([]any))
	}

	if v, ok := m["sqli_match_statement"]; ok {
		statement.SqliMatchStatement = expandSQLiMatchStatement(v.([]any))
	}

	if v, ok := m["xss_match_statement"]; ok {
		statement.XssMatchStatement = expandXSSMatchStatement(v.([]any))
	}

	return statement
}

func expandManagedRuleGroupStatement(l []any) *awstypes.ManagedRuleGroupStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	r := &awstypes.ManagedRuleGroupStatement{
		Name:                aws.String(m[names.AttrName].(string)),
		RuleActionOverrides: expandRuleActionOverrides(m["rule_action_override"].([]any)),
		VendorName:          aws.String(m["vendor_name"].(string)),
	}

	if s, ok := m["scope_down_statement"].([]any); ok && len(s) > 0 && s[0] != nil {
		r.ScopeDownStatement = expandStatement(s[0].(map[string]any))
	}

	if v, ok := m[names.AttrVersion]; ok && v != "" {
		r.Version = aws.String(v.(string))
	}
	if v, ok := m["managed_rule_group_configs"].([]any); ok && len(v) > 0 {
		r.ManagedRuleGroupConfigs = expandManagedRuleGroupConfigs(v)
	}

	return r
}

func expandManagedRuleGroupConfigs(tfList []any) []awstypes.ManagedRuleGroupConfig {
	if len(tfList) == 0 {
		return nil
	}

	var out []awstypes.ManagedRuleGroupConfig
	for _, item := range tfList {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		var r awstypes.ManagedRuleGroupConfig
		if v, ok := m["aws_managed_rules_bot_control_rule_set"].([]any); ok && len(v) > 0 {
			r.AWSManagedRulesBotControlRuleSet = expandManagedRulesBotControlRuleSet(v)
		}
		if v, ok := m["aws_managed_rules_acfp_rule_set"].([]any); ok && len(v) > 0 {
			r.AWSManagedRulesACFPRuleSet = expandManagedRulesACFPRuleSet(v)
		}
		if v, ok := m["aws_managed_rules_atp_rule_set"].([]any); ok && len(v) > 0 {
			r.AWSManagedRulesATPRuleSet = expandManagedRulesATPRuleSet(v)
		}
		if v, ok := m["login_path"].(string); ok && v != "" {
			r.LoginPath = aws.String(v)
		}
		if v, ok := m["payload_type"].(string); ok && v != "" {
			r.PayloadType = awstypes.PayloadType(v)
		}
		if v, ok := m["password_field"].([]any); ok && len(v) > 0 {
			r.PasswordField = expandPasswordField(v)
		}
		if v, ok := m["username_field"].([]any); ok && len(v) > 0 {
			r.UsernameField = expandUsernameField(v)
		}

		out = append(out, r)
	}

	return out
}

func expandAddressFields(tfList []any) []awstypes.AddressField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	out := make([]awstypes.AddressField, 0)
	identifiers := tfList[0].(map[string]any)
	for _, v := range identifiers["identifiers"].([]any) {
		r := awstypes.AddressField{
			Identifier: aws.String(v.(string)),
		}

		out = append(out, r)
	}

	return out
}

func expandEmailField(tfList []any) *awstypes.EmailField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]any)
	out := awstypes.EmailField{
		Identifier: aws.String(m[names.AttrIdentifier].(string)),
	}

	return &out
}

func expandPasswordField(tfList []any) *awstypes.PasswordField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]any)
	out := awstypes.PasswordField{
		Identifier: aws.String(m[names.AttrIdentifier].(string)),
	}

	return &out
}

func expandPhoneNumberFields(tfList []any) []awstypes.PhoneNumberField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	out := make([]awstypes.PhoneNumberField, 0)
	identifiers := tfList[0].(map[string]any)
	for _, v := range identifiers["identifiers"].([]any) {
		r := awstypes.PhoneNumberField{
			Identifier: aws.String(v.(string)),
		}

		out = append(out, r)
	}

	return out
}

func expandUsernameField(tfList []any) *awstypes.UsernameField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]any)
	out := awstypes.UsernameField{
		Identifier: aws.String(m[names.AttrIdentifier].(string)),
	}

	return &out
}

func expandManagedRulesBotControlRuleSet(tfList []any) *awstypes.AWSManagedRulesBotControlRuleSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]any)
	out := awstypes.AWSManagedRulesBotControlRuleSet{
		EnableMachineLearning: aws.Bool(m["enable_machine_learning"].(bool)),
		InspectionLevel:       awstypes.InspectionLevel(m["inspection_level"].(string)),
	}

	return &out
}

func expandManagedRulesACFPRuleSet(tfList []any) *awstypes.AWSManagedRulesACFPRuleSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]any)
	out := awstypes.AWSManagedRulesACFPRuleSet{
		CreationPath:         aws.String(m["creation_path"].(string)),
		RegistrationPagePath: aws.String(m["registration_page_path"].(string)),
	}

	if v, ok := m["enable_regex_in_path"].(bool); ok {
		out.EnableRegexInPath = v
	}
	if v, ok := m["request_inspection"].([]any); ok && len(v) > 0 {
		out.RequestInspection = expandRequestInspectionACFP(v)
	}
	if v, ok := m["response_inspection"].([]any); ok && len(v) > 0 {
		out.ResponseInspection = expandResponseInspection(v)
	}

	return &out
}

func expandManagedRulesATPRuleSet(tfList []any) *awstypes.AWSManagedRulesATPRuleSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]any)
	out := awstypes.AWSManagedRulesATPRuleSet{
		LoginPath: aws.String(m["login_path"].(string)),
	}

	if v, ok := m["enable_regex_in_path"].(bool); ok {
		out.EnableRegexInPath = v
	}
	if v, ok := m["request_inspection"].([]any); ok && len(v) > 0 {
		out.RequestInspection = expandRequestInspection(v)
	}
	if v, ok := m["response_inspection"].([]any); ok && len(v) > 0 {
		out.ResponseInspection = expandResponseInspection(v)
	}

	return &out
}

func expandRequestInspection(tfList []any) *awstypes.RequestInspection {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]any)
	out := awstypes.RequestInspection{
		PasswordField: expandPasswordField(m["password_field"].([]any)),
		PayloadType:   awstypes.PayloadType(m["payload_type"].(string)),
		UsernameField: expandUsernameField(m["username_field"].([]any)),
	}

	return &out
}

func expandRequestInspectionACFP(tfList []any) *awstypes.RequestInspectionACFP {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]any)
	out := awstypes.RequestInspectionACFP{
		AddressFields:     expandAddressFields(m["address_fields"].([]any)),
		EmailField:        expandEmailField(m["email_field"].([]any)),
		PasswordField:     expandPasswordField(m["password_field"].([]any)),
		PayloadType:       awstypes.PayloadType(m["payload_type"].(string)),
		PhoneNumberFields: expandPhoneNumberFields(m["phone_number_fields"].([]any)),
		UsernameField:     expandUsernameField(m["username_field"].([]any)),
	}

	return &out
}

func expandResponseInspection(tfList []any) *awstypes.ResponseInspection {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]any)
	out := awstypes.ResponseInspection{}
	if v, ok := m["body_contains"].([]any); ok && len(v) > 0 {
		out.BodyContains = expandBodyContains(v)
	}
	if v, ok := m[names.AttrHeader].([]any); ok && len(v) > 0 {
		out.Header = expandHeader(v)
	}
	if v, ok := m[names.AttrJSON].([]any); ok && len(v) > 0 {
		out.Json = expandResponseInspectionJSON(v)
	}
	if v, ok := m[names.AttrStatusCode].([]any); ok && len(v) > 0 {
		out.StatusCode = expandStatusCode(v)
	}

	return &out
}

func expandBodyContains(tfList []any) *awstypes.ResponseInspectionBodyContains {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]any)
	out := awstypes.ResponseInspectionBodyContains{
		FailureStrings: flex.ExpandStringValueSet(m["failure_strings"].(*schema.Set)),
		SuccessStrings: flex.ExpandStringValueSet(m["success_strings"].(*schema.Set)),
	}

	return &out
}

func expandHeader(tfList []any) *awstypes.ResponseInspectionHeader {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]any)

	out := awstypes.ResponseInspectionHeader{
		Name:          aws.String(m[names.AttrName].(string)),
		FailureValues: flex.ExpandStringValueSet(m["failure_values"].(*schema.Set)),
		SuccessValues: flex.ExpandStringValueSet(m["success_values"].(*schema.Set)),
	}

	return &out
}

func expandResponseInspectionJSON(tfList []any) *awstypes.ResponseInspectionJson {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]any)
	out := awstypes.ResponseInspectionJson{
		FailureValues: flex.ExpandStringValueSet(m["failure_values"].(*schema.Set)),
		Identifier:    aws.String(m[names.AttrIdentifier].(string)),
		SuccessValues: flex.ExpandStringValueSet(m["success_values"].(*schema.Set)),
	}

	return &out
}

func expandStatusCode(tfList []any) *awstypes.ResponseInspectionStatusCode {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	m := tfList[0].(map[string]any)
	out := awstypes.ResponseInspectionStatusCode{
		FailureCodes: flex.ExpandInt32ValueSet(m["failure_codes"].(*schema.Set)),
		SuccessCodes: flex.ExpandInt32ValueSet(m["success_codes"].(*schema.Set)),
	}

	return &out
}

func expandRateLimitCookie(l []any) *awstypes.RateLimitCookie {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]any)

	return &awstypes.RateLimitCookie{
		Name:                aws.String(m[names.AttrName].(string)),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandRateLimitHeader(l []any) *awstypes.RateLimitHeader {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]any)

	return &awstypes.RateLimitHeader{
		Name:                aws.String(m[names.AttrName].(string)),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandRateLimitJa3Fingerprint(l []any) *awstypes.RateLimitJA3Fingerprint {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]any)
	return &awstypes.RateLimitJA3Fingerprint{
		FallbackBehavior: awstypes.FallbackBehavior(m["fallback_behavior"].(string)),
	}
}

func expandRateLimitJa4Fingerprint(l []any) *awstypes.RateLimitJA4Fingerprint {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]any)
	return &awstypes.RateLimitJA4Fingerprint{
		FallbackBehavior: awstypes.FallbackBehavior(m["fallback_behavior"].(string)),
	}
}

func expandRateLimitLabelNamespace(l []any) *awstypes.RateLimitLabelNamespace {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]any)
	return &awstypes.RateLimitLabelNamespace{
		Namespace: aws.String(m[names.AttrNamespace].(string)),
	}
}

func expandRateLimitQueryArgument(l []any) *awstypes.RateLimitQueryArgument {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]any)

	return &awstypes.RateLimitQueryArgument{
		Name:                aws.String(m[names.AttrName].(string)),
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandRateLimitQueryString(l []any) *awstypes.RateLimitQueryString {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]any)
	return &awstypes.RateLimitQueryString{
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandRateLimitURIPath(l []any) *awstypes.RateLimitUriPath {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]any)
	return &awstypes.RateLimitUriPath{
		TextTransformations: expandTextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandRateBasedStatementCustomKeys(l []any) []awstypes.RateBasedStatementCustomKey {
	if len(l) == 0 {
		return nil
	}
	out := make([]awstypes.RateBasedStatementCustomKey, 0)
	for _, ck := range l {
		r := awstypes.RateBasedStatementCustomKey{}
		m := ck.(map[string]any)
		if v, ok := m["cookie"]; ok {
			r.Cookie = expandRateLimitCookie(v.([]any))
		}
		if v, ok := m["forwarded_ip"]; ok && len(v.([]any)) > 0 {
			r.ForwardedIP = &awstypes.RateLimitForwardedIP{}
		}
		if v, ok := m["http_method"]; ok && len(v.([]any)) > 0 {
			r.HTTPMethod = &awstypes.RateLimitHTTPMethod{}
		}
		if v, ok := m[names.AttrHeader]; ok {
			r.Header = expandRateLimitHeader(v.([]any))
		}
		if v, ok := m["ip"]; ok && len(v.([]any)) > 0 {
			r.IP = &awstypes.RateLimitIP{}
		}
		if v, ok := m["ja3_fingerprint"]; ok && len(v.([]any)) > 0 {
			r.JA3Fingerprint = expandRateLimitJa3Fingerprint(v.([]any))
		}
		if v, ok := m["ja4_fingerprint"]; ok && len(v.([]any)) > 0 {
			r.JA4Fingerprint = expandRateLimitJa4Fingerprint(v.([]any))
		}
		if v, ok := m["label_namespace"]; ok {
			r.LabelNamespace = expandRateLimitLabelNamespace(v.([]any))
		}
		if v, ok := m["query_argument"]; ok {
			r.QueryArgument = expandRateLimitQueryArgument(v.([]any))
		}
		if v, ok := m["query_string"]; ok {
			r.QueryString = expandRateLimitQueryString(v.([]any))
		}
		if v, ok := m["uri_path"]; ok {
			r.UriPath = expandRateLimitURIPath(v.([]any))
		}
		out = append(out, r)
	}
	return out
}

func expandRateBasedStatement(l []any) *awstypes.RateBasedStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	r := &awstypes.RateBasedStatement{
		AggregateKeyType:    awstypes.RateBasedStatementAggregateKeyType(m["aggregate_key_type"].(string)),
		EvaluationWindowSec: int64(m["evaluation_window_sec"].(int)),
		Limit:               aws.Int64(int64(m["limit"].(int))),
	}

	if v, ok := m["forwarded_ip_config"]; ok {
		r.ForwardedIPConfig = expandForwardedIPConfig(v.([]any))
	}

	if v, ok := m["custom_key"]; ok {
		r.CustomKeys = expandRateBasedStatementCustomKeys(v.([]any))
	}

	s := m["scope_down_statement"].([]any)
	if len(s) > 0 && s[0] != nil {
		r.ScopeDownStatement = expandStatement(s[0].(map[string]any))
	}

	return r
}

func expandRuleGroupReferenceStatement(l []any) *awstypes.RuleGroupReferenceStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.RuleGroupReferenceStatement{
		ARN:                 aws.String(m[names.AttrARN].(string)),
		RuleActionOverrides: expandRuleActionOverrides(m["rule_action_override"].([]any)),
	}
}

func expandRuleActionOverrides(l []any) []awstypes.RuleActionOverride {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	overrides := make([]awstypes.RuleActionOverride, 0)

	for _, override := range l {
		if override == nil {
			continue
		}
		overrides = append(overrides, expandRuleActionOverride(override.(map[string]any)))
	}

	return overrides
}

func expandRuleActionOverride(m map[string]any) awstypes.RuleActionOverride {
	return awstypes.RuleActionOverride{
		ActionToUse: expandRuleAction(m["action_to_use"].([]any)),
		Name:        aws.String(m[names.AttrName].(string)),
	}
}

func expandRegexPatternSet(l []any) []awstypes.Regex {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	regexPatterns := make([]awstypes.Regex, 0)
	for _, regexPattern := range l {
		regexPatterns = append(regexPatterns, expandRegex(regexPattern.(map[string]any)))
	}

	return regexPatterns
}

func expandRegex(m map[string]any) awstypes.Regex {
	return awstypes.Regex{
		RegexString: aws.String(m["regex_string"].(string)),
	}
}

func flattenRules(r []awstypes.Rule) any {
	out := make([]map[string]any, len(r))
	for i, rule := range r {
		m := make(map[string]any)
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

func flattenRuleAction(a *awstypes.RuleAction) any {
	if a == nil {
		return []any{}
	}

	m := map[string]any{}

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

	return []any{m}
}

func flattenAllow(a *awstypes.AllowAction) []any {
	if a == nil {
		return []any{}
	}
	m := map[string]any{}

	if a.CustomRequestHandling != nil {
		m["custom_request_handling"] = flattenCustomRequestHandling(a.CustomRequestHandling)
	}

	return []any{m}
}

func flattenBlock(a *awstypes.BlockAction) []any {
	if a == nil {
		return []any{}
	}

	m := map[string]any{}

	if a.CustomResponse != nil {
		m["custom_response"] = flattenCustomResponse(a.CustomResponse)
	}

	return []any{m}
}

func flattenCaptcha(a *awstypes.CaptchaAction) []any {
	if a == nil {
		return []any{}
	}

	m := map[string]any{}

	if a.CustomRequestHandling != nil {
		m["custom_request_handling"] = flattenCustomRequestHandling(a.CustomRequestHandling)
	}

	return []any{m}
}

func flattenCaptchaConfig(config *awstypes.CaptchaConfig) any {
	if config == nil {
		return []any{}
	}
	if config.ImmunityTimeProperty == nil {
		return []any{}
	}

	m := map[string]any{
		"immunity_time_property": []any{map[string]any{
			"immunity_time": aws.ToInt64(config.ImmunityTimeProperty.ImmunityTime),
		}},
	}

	return []any{m}
}

func flattenChallengeConfig(config *awstypes.ChallengeConfig) any {
	if config == nil {
		return []any{}
	}
	if config.ImmunityTimeProperty == nil {
		return []any{}
	}

	m := map[string]any{
		"immunity_time_property": []any{map[string]any{
			"immunity_time": aws.ToInt64(config.ImmunityTimeProperty.ImmunityTime),
		}},
	}

	return []any{m}
}

func flattenAssociationConfig(config *awstypes.AssociationConfig) any {
	associationConfig := []any{}
	if config == nil {
		return associationConfig
	}
	if config.RequestBody == nil {
		return associationConfig
	}

	requestBodyConfig := map[string]any{}
	for _, resourceType := range awstypes.AssociatedResourceType.Values("") {
		if requestBodyAssociatedResourceTypeConfig, ok := config.RequestBody[string(resourceType)]; ok {
			requestBodyConfig[strings.ToLower(string(resourceType))] = []map[string]any{{
				"default_size_inspection_limit": requestBodyAssociatedResourceTypeConfig.DefaultSizeInspectionLimit,
			}}
		}
	}
	associationConfig = append(associationConfig, map[string]any{
		"request_body": []map[string]any{
			requestBodyConfig,
		},
	})

	return associationConfig
}

func flattenChallenge(a *awstypes.ChallengeAction) []any {
	if a == nil {
		return []any{}
	}

	m := map[string]any{}

	if a.CustomRequestHandling != nil {
		m["custom_request_handling"] = flattenCustomRequestHandling(a.CustomRequestHandling)
	}

	return []any{m}
}

func flattenCount(a *awstypes.CountAction) []any {
	if a == nil {
		return []any{}
	}
	m := map[string]any{}

	if a.CustomRequestHandling != nil {
		m["custom_request_handling"] = flattenCustomRequestHandling(a.CustomRequestHandling)
	}

	return []any{m}
}

func flattenCustomResponseBodies(b map[string]awstypes.CustomResponseBody) any {
	if len(b) == 0 {
		return make([]map[string]any, 0)
	}

	out := make([]map[string]any, len(b))
	i := 0
	for key, body := range b {
		out[i] = map[string]any{
			names.AttrKey:         key,
			names.AttrContent:     aws.ToString(body.Content),
			names.AttrContentType: body.ContentType,
		}
		i += 1
	}

	return out
}

func flattenCustomRequestHandling(c *awstypes.CustomRequestHandling) []any {
	if c == nil {
		return []any{}
	}

	m := map[string]any{
		"insert_header": flattenCustomHeaders(c.InsertHeaders),
	}

	return []any{m}
}

func flattenCustomResponse(r *awstypes.CustomResponse) []any {
	if r == nil {
		return []any{}
	}

	m := map[string]any{
		"response_code":   aws.ToInt32(r.ResponseCode),
		"response_header": flattenCustomHeaders(r.ResponseHeaders),
	}

	if r.CustomResponseBodyKey != nil {
		m["custom_response_body_key"] = aws.ToString(r.CustomResponseBodyKey)
	}

	return []any{m}
}

func flattenCustomHeaders(h []awstypes.CustomHTTPHeader) []any {
	out := make([]any, len(h))
	for i, header := range h {
		out[i] = flattenCustomHeader(&header)
	}

	return out
}

func flattenCustomHeader(h *awstypes.CustomHTTPHeader) map[string]any {
	if h == nil {
		return map[string]any{}
	}

	m := map[string]any{
		names.AttrName:  aws.ToString(h.Name),
		names.AttrValue: aws.ToString(h.Value),
	}

	return m
}

func flattenRuleLabels(l []awstypes.Label) []any {
	if len(l) == 0 {
		return nil
	}

	out := make([]any, len(l))
	for i, label := range l {
		out[i] = map[string]any{
			names.AttrName: aws.ToString(label.Name),
		}
	}

	return out
}

func flattenRuleGroupRootStatement(s *awstypes.Statement) any {
	if s == nil {
		return []any{}
	}

	return []any{flattenStatement(s)}
}

func flattenStatements(s []awstypes.Statement) any {
	out := make([]any, len(s))
	for i, statement := range s {
		out[i] = flattenStatement(&statement)
	}

	return out
}

func flattenStatement(s *awstypes.Statement) map[string]any {
	if s == nil {
		return map[string]any{}
	}

	m := map[string]any{}

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

func flattenAndStatement(a *awstypes.AndStatement) any {
	if a == nil {
		return []any{}
	}

	m := map[string]any{
		"statement": flattenStatements(a.Statements),
	}

	return []any{m}
}

func flattenByteMatchStatement(b *awstypes.ByteMatchStatement) any {
	if b == nil {
		return []any{}
	}

	m := map[string]any{
		"field_to_match":        flattenFieldToMatch(b.FieldToMatch),
		"positional_constraint": b.PositionalConstraint,
		"search_string":         string(b.SearchString),
		"text_transformation":   flattenTextTransformations(b.TextTransformations),
	}

	return []any{m}
}

func flattenFieldToMatch(f *awstypes.FieldToMatch) any {
	if f == nil {
		return []any{}
	}

	m := map[string]any{}

	if f.AllQueryArguments != nil {
		m["all_query_arguments"] = make([]map[string]any, 1)
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

	if f.JA4Fingerprint != nil {
		m["ja4_fingerprint"] = flattenJA4Fingerprint(f.JA4Fingerprint)
	}

	if f.JsonBody != nil {
		m["json_body"] = flattenJSONBody(f.JsonBody)
	}

	if f.Method != nil {
		m["method"] = make([]map[string]any, 1)
	}

	if f.QueryString != nil {
		m["query_string"] = make([]map[string]any, 1)
	}

	if f.SingleHeader != nil {
		m["single_header"] = flattenSingleHeader(f.SingleHeader)
	}

	if f.SingleQueryArgument != nil {
		m["single_query_argument"] = flattenSingleQueryArgument(f.SingleQueryArgument)
	}

	if f.UriPath != nil {
		m["uri_path"] = make([]map[string]any, 1)
	}

	return []any{m}
}

func flattenForwardedIPConfig(f *awstypes.ForwardedIPConfig) any {
	if f == nil {
		return []any{}
	}

	m := map[string]any{
		"fallback_behavior": f.FallbackBehavior,
		"header_name":       aws.ToString(f.HeaderName),
	}

	return []any{m}
}

func flattenIPSetForwardedIPConfig(i *awstypes.IPSetForwardedIPConfig) any {
	if i == nil {
		return []any{}
	}

	m := map[string]any{
		"fallback_behavior": i.FallbackBehavior,
		"header_name":       aws.ToString(i.HeaderName),
		"position":          i.Position,
	}

	return []any{m}
}

func flattenCookies(c *awstypes.Cookies) any {
	if c == nil {
		return []any{}
	}

	m := map[string]any{
		"match_scope":       c.MatchScope,
		"oversize_handling": c.OversizeHandling,
		"match_pattern":     flattenCookiesMatchPattern(c.MatchPattern),
	}

	return []any{m}
}

func flattenCookiesMatchPattern(c *awstypes.CookieMatchPattern) any {
	if c == nil {
		return []any{}
	}

	m := map[string]any{
		"included_cookies": aws.StringSlice(c.IncludedCookies),
		"excluded_cookies": aws.StringSlice(c.ExcludedCookies),
	}

	if c.All != nil {
		m["all"] = make([]map[string]any, 1)
	}

	return []any{m}
}

func flattenJA3Fingerprint(j *awstypes.JA3Fingerprint) any {
	if j == nil {
		return []any{}
	}

	m := map[string]any{
		"fallback_behavior": j.FallbackBehavior,
	}

	return []any{m}
}

func flattenJA4Fingerprint(j *awstypes.JA4Fingerprint) any {
	if j == nil {
		return []any{}
	}

	m := map[string]any{
		"fallback_behavior": j.FallbackBehavior,
	}

	return []any{m}
}

func flattenJSONBody(b *awstypes.JsonBody) any {
	if b == nil {
		return []any{}
	}

	m := map[string]any{
		"invalid_fallback_behavior": b.InvalidFallbackBehavior,
		"match_pattern":             flattenJSONMatchPattern(b.MatchPattern),
		"match_scope":               b.MatchScope,
		"oversize_handling":         b.OversizeHandling,
	}

	return []any{m}
}

func flattenBody(b *awstypes.Body) any {
	if b == nil {
		return []any{}
	}

	m := map[string]any{
		"oversize_handling": b.OversizeHandling,
	}

	return []any{m}
}

func flattenJSONMatchPattern(p *awstypes.JsonMatchPattern) []any {
	if p == nil {
		return []any{}
	}

	m := map[string]any{
		"included_paths": p.IncludedPaths,
	}

	if p.All != nil {
		m["all"] = make([]map[string]any, 1)
	}

	return []any{m}
}

func flattenSingleHeader(s *awstypes.SingleHeader) any {
	if s == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrName: aws.ToString(s.Name),
	}

	return []any{m}
}

func flattenSingleQueryArgument(s *awstypes.SingleQueryArgument) any {
	if s == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrName: aws.ToString(s.Name),
	}

	return []any{m}
}

func flattenTextTransformations(l []awstypes.TextTransformation) []any {
	out := make([]any, len(l))
	for i, t := range l {
		m := make(map[string]any)
		m[names.AttrPriority] = t.Priority
		m[names.AttrType] = t.Type
		out[i] = m
	}
	return out
}

func flattenIPSetReferenceStatement(i *awstypes.IPSetReferenceStatement) any {
	if i == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrARN:                aws.ToString(i.ARN),
		"ip_set_forwarded_ip_config": flattenIPSetForwardedIPConfig(i.IPSetForwardedIPConfig),
	}

	return []any{m}
}

func flattenGeoMatchStatement(g *awstypes.GeoMatchStatement) any {
	if g == nil {
		return []any{}
	}

	m := map[string]any{
		"country_codes":       g.CountryCodes,
		"forwarded_ip_config": flattenForwardedIPConfig(g.ForwardedIPConfig),
	}

	return []any{m}
}

func flattenLabelMatchStatement(l *awstypes.LabelMatchStatement) any {
	if l == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrKey:   aws.ToString(l.Key),
		names.AttrScope: l.Scope,
	}

	return []any{m}
}

func flattenNotStatement(a *awstypes.NotStatement) any {
	if a == nil {
		return []any{}
	}

	m := map[string]any{
		"statement": []any{flattenStatement(a.Statement)},
	}

	return []any{m}
}

func flattenOrStatement(a *awstypes.OrStatement) any {
	if a == nil {
		return []any{}
	}

	m := map[string]any{
		"statement": flattenStatements(a.Statements),
	}

	return []any{m}
}

func flattenRegexMatchStatement(r *awstypes.RegexMatchStatement) any {
	if r == nil {
		return []any{}
	}

	m := map[string]any{
		"regex_string":        aws.ToString(r.RegexString),
		"field_to_match":      flattenFieldToMatch(r.FieldToMatch),
		"text_transformation": flattenTextTransformations(r.TextTransformations),
	}

	return []any{m}
}

func flattenRegexPatternSetReferenceStatement(r *awstypes.RegexPatternSetReferenceStatement) any {
	if r == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrARN:         aws.ToString(r.ARN),
		"field_to_match":      flattenFieldToMatch(r.FieldToMatch),
		"text_transformation": flattenTextTransformations(r.TextTransformations),
	}

	return []any{m}
}

func flattenSizeConstraintStatement(s *awstypes.SizeConstraintStatement) any {
	if s == nil {
		return []any{}
	}

	m := map[string]any{
		"comparison_operator": s.ComparisonOperator,
		"field_to_match":      flattenFieldToMatch(s.FieldToMatch),
		names.AttrSize:        s.Size,
		"text_transformation": flattenTextTransformations(s.TextTransformations),
	}

	return []any{m}
}

func flattenSQLiMatchStatement(s *awstypes.SqliMatchStatement) any {
	if s == nil {
		return []any{}
	}

	m := map[string]any{
		"field_to_match":      flattenFieldToMatch(s.FieldToMatch),
		"sensitivity_level":   s.SensitivityLevel,
		"text_transformation": flattenTextTransformations(s.TextTransformations),
	}

	return []any{m}
}

func flattenXSSMatchStatement(s *awstypes.XssMatchStatement) any {
	if s == nil {
		return []any{}
	}

	m := map[string]any{
		"field_to_match":      flattenFieldToMatch(s.FieldToMatch),
		"text_transformation": flattenTextTransformations(s.TextTransformations),
	}

	return []any{m}
}

func flattenVisibilityConfig(config *awstypes.VisibilityConfig) any {
	if config == nil {
		return []any{}
	}

	m := map[string]any{
		"cloudwatch_metrics_enabled": aws.Bool(config.CloudWatchMetricsEnabled),
		names.AttrMetricName:         aws.ToString(config.MetricName),
		"sampled_requests_enabled":   aws.Bool(config.SampledRequestsEnabled),
	}

	return []any{m}
}

func flattenHeaderOrder(s *awstypes.HeaderOrder) any {
	if s == nil {
		return []any{}
	}

	m := map[string]any{
		"oversize_handling": s.OversizeHandling,
	}

	return []any{m}
}

func flattenHeaders(s *awstypes.Headers) any {
	if s == nil {
		return []any{}
	}

	m := map[string]any{
		"match_scope":       s.MatchScope,
		"match_pattern":     flattenHeaderMatchPattern(s.MatchPattern),
		"oversize_handling": s.OversizeHandling,
	}

	return []any{m}
}

func flattenHeaderMatchPattern(s *awstypes.HeaderMatchPattern) any {
	if s == nil {
		return []any{}
	}

	m := map[string]any{}

	if s.All != nil {
		m["all"] = make([]map[string]any, 1)
	}

	if s.ExcludedHeaders != nil {
		m["excluded_headers"] = s.ExcludedHeaders
	}

	if s.IncludedHeaders != nil {
		m["included_headers"] = s.IncludedHeaders
	}

	return []any{m}
}

func flattenWebACLRootStatement(s *awstypes.Statement) any {
	if s == nil {
		return []any{}
	}

	return []any{flattenWebACLStatement(s)}
}

func flattenWebACLStatement(s *awstypes.Statement) map[string]any {
	if s == nil {
		return map[string]any{}
	}

	m := map[string]any{}

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

func flattenWebACLRules(r []awstypes.Rule) any {
	out := make([]map[string]any, len(r))
	for i, rule := range r {
		m := make(map[string]any)
		m[names.AttrAction] = flattenRuleAction(rule.Action)
		m["captcha_config"] = flattenCaptchaConfig(rule.CaptchaConfig)
		m["challenge_config"] = flattenChallengeConfig(rule.ChallengeConfig)
		m[names.AttrName] = aws.ToString(rule.Name)
		m["override_action"] = flattenOverrideAction(rule.OverrideAction)
		m[names.AttrPriority] = rule.Priority
		m["rule_label"] = flattenRuleLabels(rule.RuleLabels)
		m["statement"] = flattenWebACLRootStatement(rule.Statement)
		m["visibility_config"] = flattenVisibilityConfig(rule.VisibilityConfig)
		out[i] = m
	}

	return out
}

func flattenOverrideAction(a *awstypes.OverrideAction) any {
	if a == nil {
		return []any{}
	}

	m := map[string]any{}

	if a.Count != nil {
		m["count"] = make([]map[string]any, 1)
	}

	if a.None != nil {
		m["none"] = make([]map[string]any, 1)
	}

	return []any{m}
}

func flattenDefaultAction(a *awstypes.DefaultAction) any {
	if a == nil {
		return []any{}
	}

	m := map[string]any{}

	if a.Allow != nil {
		m["allow"] = flattenAllow(a.Allow)
	}

	if a.Block != nil {
		m["block"] = flattenBlock(a.Block)
	}

	return []any{m}
}

func flattenManagedRuleGroupStatement(apiObject *awstypes.ManagedRuleGroupStatement) any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}

	if apiObject.RuleActionOverrides != nil {
		tfMap["rule_action_override"] = flattenRuleActionOverrides(apiObject.RuleActionOverrides)
	}

	if apiObject.ScopeDownStatement != nil {
		tfMap["scope_down_statement"] = []any{flattenStatement(apiObject.ScopeDownStatement)}
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

	return []any{tfMap}
}

func flattenManagedRuleGroupConfigs(c []awstypes.ManagedRuleGroupConfig) []any {
	if len(c) == 0 {
		return nil
	}

	var out []any

	for _, config := range c {
		m := make(map[string]any)
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

func flattenAddressFields(apiObjects []awstypes.AddressField) []any {
	if apiObjects == nil {
		return nil
	}

	var identifiers []*string
	for _, apiObject := range apiObjects {
		identifiers = append(identifiers, apiObject.Identifier)
	}

	return []any{
		map[string]any{
			"identifiers": aws.ToStringSlice(identifiers),
		},
	}
}

func flattenEmailField(apiObject *awstypes.EmailField) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		names.AttrIdentifier: aws.ToString(apiObject.Identifier),
	}

	return []any{m}
}

func flattenPasswordField(apiObject *awstypes.PasswordField) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		names.AttrIdentifier: aws.ToString(apiObject.Identifier),
	}

	return []any{m}
}

func flattenPhoneNumberFields(apiObjects []awstypes.PhoneNumberField) []any {
	if apiObjects == nil {
		return nil
	}

	var identifiers []*string
	for _, apiObject := range apiObjects {
		identifiers = append(identifiers, apiObject.Identifier)
	}

	return []any{
		map[string]any{
			"identifiers": aws.ToStringSlice(identifiers),
		},
	}
}

func flattenUsernameField(apiObject *awstypes.UsernameField) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		names.AttrIdentifier: aws.ToString(apiObject.Identifier),
	}

	return []any{m}
}

func flattenManagedRulesBotControlRuleSet(apiObject *awstypes.AWSManagedRulesBotControlRuleSet) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		"enable_machine_learning": aws.ToBool(apiObject.EnableMachineLearning),
		"inspection_level":        apiObject.InspectionLevel,
	}

	return []any{m}
}

func flattenManagedRulesACFPRuleSet(apiObject *awstypes.AWSManagedRulesACFPRuleSet) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
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

	return []any{m}
}

func flattenManagedRulesATPRuleSet(apiObject *awstypes.AWSManagedRulesATPRuleSet) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		"enable_regex_in_path": aws.Bool(apiObject.EnableRegexInPath),
		"login_path":           aws.ToString(apiObject.LoginPath),
	}
	if apiObject.RequestInspection != nil {
		m["request_inspection"] = flattenRequestInspection(apiObject.RequestInspection)
	}
	if apiObject.ResponseInspection != nil {
		m["response_inspection"] = flattenResponseInspection(apiObject.ResponseInspection)
	}

	return []any{m}
}

func flattenRequestInspectionACFP(apiObject *awstypes.RequestInspectionACFP) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		"address_fields":      flattenAddressFields(apiObject.AddressFields),
		"email_field":         flattenEmailField(apiObject.EmailField),
		"password_field":      flattenPasswordField(apiObject.PasswordField),
		"payload_type":        apiObject.PayloadType,
		"phone_number_fields": flattenPhoneNumberFields(apiObject.PhoneNumberFields),
		"username_field":      flattenUsernameField(apiObject.UsernameField),
	}

	return []any{m}
}

func flattenRequestInspection(apiObject *awstypes.RequestInspection) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		"password_field": flattenPasswordField(apiObject.PasswordField),
		"payload_type":   apiObject.PayloadType,
		"username_field": flattenUsernameField(apiObject.UsernameField),
	}

	return []any{m}
}

func flattenResponseInspection(apiObject *awstypes.ResponseInspection) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{}
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

	return []any{m}
}

func flattenBodyContains(apiObject *awstypes.ResponseInspectionBodyContains) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		"failure_strings": apiObject.FailureStrings,
		"success_strings": apiObject.SuccessStrings,
	}

	return []any{m}
}

func flattenHeader(apiObject *awstypes.ResponseInspectionHeader) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		"failure_values": apiObject.FailureValues,
		"success_values": apiObject.SuccessValues,
	}

	return []any{m}
}

func flattenResponseInspectionJSON(apiObject *awstypes.ResponseInspectionJson) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		"failure_values":     apiObject.FailureValues,
		names.AttrIdentifier: aws.ToString(apiObject.Identifier),
		"success_values":     apiObject.SuccessValues,
	}

	return []any{m}
}

func flattenStatusCode(apiObject *awstypes.ResponseInspectionStatusCode) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		"failure_codes": flex.FlattenInt32ValueSet(apiObject.FailureCodes),
		"success_codes": flex.FlattenInt32ValueSet(apiObject.SuccessCodes),
	}

	return []any{m}
}

func flattenRateLimitCookie(apiObject *awstypes.RateLimitCookie) []any {
	if apiObject == nil {
		return nil
	}
	return []any{
		map[string]any{
			names.AttrName:        aws.ToString(apiObject.Name),
			"text_transformation": flattenTextTransformations(apiObject.TextTransformations),
		},
	}
}

func flattenRateLimitHeader(apiObject *awstypes.RateLimitHeader) []any {
	if apiObject == nil {
		return nil
	}
	return []any{
		map[string]any{
			names.AttrName:        aws.ToString(apiObject.Name),
			"text_transformation": flattenTextTransformations(apiObject.TextTransformations),
		},
	}
}

func flattenRateLimitJa3Fingerprint(apiObject *awstypes.RateLimitJA3Fingerprint) []any {
	if apiObject == nil {
		return nil
	}
	return []any{
		map[string]any{
			"fallback_behavior": apiObject.FallbackBehavior,
		},
	}
}

func flattenRateLimitJa4Fingerprint(apiObject *awstypes.RateLimitJA4Fingerprint) []any {
	if apiObject == nil {
		return nil
	}
	return []any{
		map[string]any{
			"fallback_behavior": apiObject.FallbackBehavior,
		},
	}
}

func flattenRateLimitLabelNamespace(apiObject *awstypes.RateLimitLabelNamespace) []any {
	if apiObject == nil {
		return nil
	}
	return []any{
		map[string]any{
			names.AttrNamespace: aws.ToString(apiObject.Namespace),
		},
	}
}

func flattenRateLimitQueryArgument(apiObject *awstypes.RateLimitQueryArgument) []any {
	if apiObject == nil {
		return nil
	}
	return []any{
		map[string]any{
			names.AttrName:        aws.ToString(apiObject.Name),
			"text_transformation": flattenTextTransformations(apiObject.TextTransformations),
		},
	}
}

func flattenRateLimitQueryString(apiObject *awstypes.RateLimitQueryString) []any {
	if apiObject == nil {
		return nil
	}
	return []any{
		map[string]any{
			"text_transformation": flattenTextTransformations(apiObject.TextTransformations),
		},
	}
}

func flattenRateLimitURIPath(apiObject *awstypes.RateLimitUriPath) []any {
	if apiObject == nil {
		return nil
	}
	return []any{
		map[string]any{
			"text_transformation": flattenTextTransformations(apiObject.TextTransformations),
		},
	}
}

func flattenRateBasedStatementCustomKeys(apiObject []awstypes.RateBasedStatementCustomKey) []any {
	if apiObject == nil {
		return nil
	}

	out := make([]any, len(apiObject))
	for i, o := range apiObject {
		tfMap := map[string]any{}

		if o.Cookie != nil {
			tfMap["cookie"] = flattenRateLimitCookie(o.Cookie)
		}
		if o.ForwardedIP != nil {
			tfMap["forwarded_ip"] = []any{
				map[string]any{},
			}
		}
		if o.HTTPMethod != nil {
			tfMap["http_method"] = []any{
				map[string]any{},
			}
		}
		if o.Header != nil {
			tfMap[names.AttrHeader] = flattenRateLimitHeader(o.Header)
		}
		if o.IP != nil {
			tfMap["ip"] = []any{
				map[string]any{},
			}
		}
		if o.JA3Fingerprint != nil {
			tfMap["ja3_fingerprint"] = flattenRateLimitJa3Fingerprint(o.JA3Fingerprint)
		}
		if o.JA4Fingerprint != nil {
			tfMap["ja4_fingerprint"] = flattenRateLimitJa4Fingerprint(o.JA4Fingerprint)
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

func flattenRateBasedStatement(apiObject *awstypes.RateBasedStatement) any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
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
		tfMap["scope_down_statement"] = []any{flattenStatement(apiObject.ScopeDownStatement)}
	}

	return []any{tfMap}
}

func flattenRuleGroupReferenceStatement(apiObject *awstypes.RuleGroupReferenceStatement) any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrARN: aws.ToString(apiObject.ARN),
	}

	if apiObject.RuleActionOverrides != nil {
		tfMap["rule_action_override"] = flattenRuleActionOverrides(apiObject.RuleActionOverrides)
	}

	return []any{tfMap}
}

func flattenRuleActionOverrides(r []awstypes.RuleActionOverride) any {
	out := make([]map[string]any, len(r))
	for i, override := range r {
		m := make(map[string]any)
		m["action_to_use"] = flattenRuleAction(override.ActionToUse)
		m[names.AttrName] = aws.ToString(override.Name)
		out[i] = m
	}

	return out
}

func flattenRegexPatternSet(r []awstypes.Regex) any {
	if r == nil {
		return []any{}
	}

	regexPatterns := make([]any, 0)

	for _, regexPattern := range r {
		d := map[string]any{
			"regex_string": aws.ToString(regexPattern.RegexString),
		}
		regexPatterns = append(regexPatterns, d)
	}

	return regexPatterns
}
