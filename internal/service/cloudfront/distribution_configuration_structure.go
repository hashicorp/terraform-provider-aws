// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func expandDistributionConfig(d *schema.ResourceData) *awstypes.DistributionConfig {
	apiObject := &awstypes.DistributionConfig{
		CacheBehaviors:               expandCacheBehaviors(d.Get("ordered_cache_behavior").([]interface{})),
		CallerReference:              aws.String(id.UniqueId()),
		Comment:                      aws.String(d.Get("comment").(string)),
		ContinuousDeploymentPolicyId: aws.String(d.Get("continuous_deployment_policy_id").(string)),
		CustomErrorResponses:         expandCustomErrorResponses(d.Get("custom_error_response").(*schema.Set).List()),
		DefaultCacheBehavior:         expandDefaultCacheBehavior(d.Get("default_cache_behavior").([]interface{})[0].(map[string]interface{})),
		DefaultRootObject:            aws.String(d.Get("default_root_object").(string)),
		Enabled:                      aws.Bool(d.Get("enabled").(bool)),
		IsIPV6Enabled:                aws.Bool(d.Get("is_ipv6_enabled").(bool)),
		HttpVersion:                  awstypes.HttpVersion(d.Get("http_version").(string)),
		Origins:                      expandOrigins(d.Get("origin").(*schema.Set).List()),
		PriceClass:                   awstypes.PriceClass(d.Get("price_class").(string)),
		Staging:                      aws.Bool(d.Get("staging").(bool)),
		WebACLId:                     aws.String(d.Get("web_acl_id").(string)),
	}

	if v, ok := d.GetOk("aliases"); ok {
		apiObject.Aliases = expandAliases(v.(*schema.Set).List())
	} else {
		apiObject.Aliases = expandAliases([]interface{}{})
	}

	if v, ok := d.GetOk("caller_reference"); ok {
		apiObject.CallerReference = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.Logging = expandLoggingConfig(v.([]interface{})[0].(map[string]interface{}))
	} else {
		apiObject.Logging = expandLoggingConfig(nil)
	}

	if v, ok := d.GetOk("origin_group"); ok {
		apiObject.OriginGroups = expandOriginGroups(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("restrictions"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.Restrictions = expandRestrictions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("viewer_certificate"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.ViewerCertificate = expandViewerCertificate(v.([]interface{})[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCacheBehavior(tfMap map[string]interface{}) *awstypes.CacheBehavior {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CacheBehavior{
		CachePolicyId:           aws.String(tfMap["cache_policy_id"].(string)),
		Compress:                aws.Bool(tfMap["compress"].(bool)),
		FieldLevelEncryptionId:  aws.String(tfMap["field_level_encryption_id"].(string)),
		OriginRequestPolicyId:   aws.String(tfMap["origin_request_policy_id"].(string)),
		ResponseHeadersPolicyId: aws.String(tfMap["response_headers_policy_id"].(string)),
		TargetOriginId:          aws.String(tfMap["target_origin_id"].(string)),
		ViewerProtocolPolicy:    awstypes.ViewerProtocolPolicy(tfMap["viewer_protocol_policy"].(string)),
	}

	if v, ok := tfMap["allowed_methods"]; ok {
		apiObject.AllowedMethods = expandAllowedMethods(v.(*schema.Set).List())
	}

	if tfMap["cache_policy_id"].(string) == "" {
		apiObject.DefaultTTL = aws.Int64(int64(tfMap["default_ttl"].(int)))
		apiObject.MaxTTL = aws.Int64(int64(tfMap["max_ttl"].(int)))
		apiObject.MinTTL = aws.Int64(int64(tfMap["min_ttl"].(int)))
	}

	if v, ok := tfMap["cached_methods"]; ok {
		apiObject.AllowedMethods.CachedMethods = expandCachedMethods(v.(*schema.Set).List())
	}

	if v, ok := tfMap["forwarded_values"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ForwardedValues = expandForwardedValues(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["function_association"]; ok {
		apiObject.FunctionAssociations = expandFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := tfMap["lambda_function_association"]; ok {
		apiObject.LambdaFunctionAssociations = expandLambdaFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := tfMap["path_pattern"]; ok {
		apiObject.PathPattern = aws.String(v.(string))
	}

	if v, ok := tfMap["realtime_log_config_arn"]; ok && v.(string) != "" {
		apiObject.RealtimeLogConfigArn = aws.String(v.(string))
	}

	if v, ok := tfMap["smooth_streaming"]; ok {
		apiObject.SmoothStreaming = aws.Bool(v.(bool))
	}

	if v, ok := tfMap["trusted_key_groups"]; ok {
		apiObject.TrustedKeyGroups = expandTrustedKeyGroups(v.([]interface{}))
	} else {
		apiObject.TrustedKeyGroups = expandTrustedKeyGroups([]interface{}{})
	}

	if v, ok := tfMap["trusted_signers"]; ok {
		apiObject.TrustedSigners = expandTrustedSigners(v.([]interface{}))
	} else {
		apiObject.TrustedSigners = expandTrustedSigners([]interface{}{})
	}

	return apiObject
}

func expandCacheBehaviors(tfList []interface{}) *awstypes.CacheBehaviors {
	if len(tfList) == 0 {
		return nil
	}

	var items []awstypes.CacheBehavior

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandCacheBehavior(tfMap)

		if apiObject == nil {
			continue
		}

		items = append(items, *apiObject)
	}

	return &awstypes.CacheBehaviors{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenCacheBehavior(apiObject *awstypes.CacheBehavior) map[string]interface{} {
	tfMap := make(map[string]interface{})

	tfMap["cache_policy_id"] = aws.ToString(apiObject.CachePolicyId)
	tfMap["compress"] = aws.ToBool(apiObject.Compress)
	tfMap["field_level_encryption_id"] = aws.ToString(apiObject.FieldLevelEncryptionId)
	tfMap["viewer_protocol_policy"] = apiObject.ViewerProtocolPolicy
	tfMap["target_origin_id"] = aws.ToString(apiObject.TargetOriginId)
	tfMap["min_ttl"] = aws.ToInt64(apiObject.MinTTL)
	tfMap["origin_request_policy_id"] = aws.ToString(apiObject.OriginRequestPolicyId)
	tfMap["realtime_log_config_arn"] = aws.ToString(apiObject.RealtimeLogConfigArn)
	tfMap["response_headers_policy_id"] = aws.ToString(apiObject.ResponseHeadersPolicyId)

	if apiObject.AllowedMethods != nil {
		tfMap["allowed_methods"] = flattenAllowedMethods(apiObject.AllowedMethods)
	}

	if apiObject.AllowedMethods.CachedMethods != nil {
		tfMap["cached_methods"] = flattenCachedMethods(apiObject.AllowedMethods.CachedMethods)
	}

	if apiObject.DefaultTTL != nil {
		tfMap["default_ttl"] = aws.ToInt64(apiObject.DefaultTTL)
	}

	if apiObject.ForwardedValues != nil {
		tfMap["forwarded_values"] = []interface{}{flattenForwardedValues(apiObject.ForwardedValues)}
	}

	if len(apiObject.FunctionAssociations.Items) > 0 {
		tfMap["function_association"] = flattenFunctionAssociations(apiObject.FunctionAssociations)
	}

	if len(apiObject.LambdaFunctionAssociations.Items) > 0 {
		tfMap["lambda_function_association"] = flattenLambdaFunctionAssociations(apiObject.LambdaFunctionAssociations)
	}

	if apiObject.MaxTTL != nil {
		tfMap["max_ttl"] = aws.ToInt64(apiObject.MaxTTL)
	}

	if apiObject.PathPattern != nil {
		tfMap["path_pattern"] = aws.ToString(apiObject.PathPattern)
	}

	if apiObject.SmoothStreaming != nil {
		tfMap["smooth_streaming"] = aws.ToBool(apiObject.SmoothStreaming)
	}

	if len(apiObject.TrustedKeyGroups.Items) > 0 {
		tfMap["trusted_key_groups"] = flattenTrustedKeyGroups(apiObject.TrustedKeyGroups)
	}

	if len(apiObject.TrustedSigners.Items) > 0 {
		tfMap["trusted_signers"] = flattenTrustedSigners(apiObject.TrustedSigners)
	}

	return tfMap
}

func flattenCacheBehaviors(apiObject *awstypes.CacheBehaviors) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfList := []interface{}{}

	for _, v := range apiObject.Items {
		tfList = append(tfList, flattenCacheBehavior(&v))
	}

	return tfList
}

func expandDefaultCacheBehavior(tfMap map[string]interface{}) *awstypes.DefaultCacheBehavior {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DefaultCacheBehavior{
		CachePolicyId:           aws.String(tfMap["cache_policy_id"].(string)),
		Compress:                aws.Bool(tfMap["compress"].(bool)),
		FieldLevelEncryptionId:  aws.String(tfMap["field_level_encryption_id"].(string)),
		OriginRequestPolicyId:   aws.String(tfMap["origin_request_policy_id"].(string)),
		ResponseHeadersPolicyId: aws.String(tfMap["response_headers_policy_id"].(string)),
		TargetOriginId:          aws.String(tfMap["target_origin_id"].(string)),
		ViewerProtocolPolicy:    awstypes.ViewerProtocolPolicy(tfMap["viewer_protocol_policy"].(string)),
	}

	if v, ok := tfMap["allowed_methods"]; ok {
		apiObject.AllowedMethods = expandAllowedMethods(v.(*schema.Set).List())
	}

	if tfMap["cache_policy_id"].(string) == "" {
		apiObject.MinTTL = aws.Int64(int64(tfMap["min_ttl"].(int)))
		apiObject.MaxTTL = aws.Int64(int64(tfMap["max_ttl"].(int)))
		apiObject.DefaultTTL = aws.Int64(int64(tfMap["default_ttl"].(int)))
	}

	if v, ok := tfMap["cached_methods"]; ok {
		apiObject.AllowedMethods.CachedMethods = expandCachedMethods(v.(*schema.Set).List())
	}

	if forwardedValuesFlat, ok := tfMap["forwarded_values"].([]interface{}); ok && len(forwardedValuesFlat) == 1 {
		apiObject.ForwardedValues = expandForwardedValues(tfMap["forwarded_values"].([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["function_association"]; ok {
		apiObject.FunctionAssociations = expandFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := tfMap["lambda_function_association"]; ok {
		apiObject.LambdaFunctionAssociations = expandLambdaFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := tfMap["realtime_log_config_arn"]; ok && v.(string) != "" {
		apiObject.RealtimeLogConfigArn = aws.String(v.(string))
	}

	if v, ok := tfMap["smooth_streaming"]; ok {
		apiObject.SmoothStreaming = aws.Bool(v.(bool))
	}

	if v, ok := tfMap["trusted_key_groups"]; ok {
		apiObject.TrustedKeyGroups = expandTrustedKeyGroups(v.([]interface{}))
	} else {
		apiObject.TrustedKeyGroups = expandTrustedKeyGroups([]interface{}{})
	}

	if v, ok := tfMap["trusted_signers"]; ok {
		apiObject.TrustedSigners = expandTrustedSigners(v.([]interface{}))
	} else {
		apiObject.TrustedSigners = expandTrustedSigners([]interface{}{})
	}

	return apiObject
}

func flattenDefaultCacheBehavior(apiObject *awstypes.DefaultCacheBehavior) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"cache_policy_id":            aws.ToString(apiObject.CachePolicyId),
		"compress":                   aws.ToBool(apiObject.Compress),
		"field_level_encryption_id":  aws.ToString(apiObject.FieldLevelEncryptionId),
		"viewer_protocol_policy":     apiObject.ViewerProtocolPolicy,
		"target_origin_id":           aws.ToString(apiObject.TargetOriginId),
		"min_ttl":                    aws.ToInt64(apiObject.MinTTL),
		"origin_request_policy_id":   aws.ToString(apiObject.OriginRequestPolicyId),
		"realtime_log_config_arn":    aws.ToString(apiObject.RealtimeLogConfigArn),
		"response_headers_policy_id": aws.ToString(apiObject.ResponseHeadersPolicyId),
	}

	if apiObject.AllowedMethods != nil {
		tfMap["allowed_methods"] = flattenAllowedMethods(apiObject.AllowedMethods)
	}

	if apiObject.AllowedMethods.CachedMethods != nil {
		tfMap["cached_methods"] = flattenCachedMethods(apiObject.AllowedMethods.CachedMethods)
	}

	if apiObject.DefaultTTL != nil {
		tfMap["default_ttl"] = aws.ToInt64(apiObject.DefaultTTL)
	}

	if apiObject.ForwardedValues != nil {
		tfMap["forwarded_values"] = []interface{}{flattenForwardedValues(apiObject.ForwardedValues)}
	}

	if len(apiObject.FunctionAssociations.Items) > 0 {
		tfMap["function_association"] = flattenFunctionAssociations(apiObject.FunctionAssociations)
	}

	if len(apiObject.LambdaFunctionAssociations.Items) > 0 {
		tfMap["lambda_function_association"] = flattenLambdaFunctionAssociations(apiObject.LambdaFunctionAssociations)
	}

	if apiObject.MaxTTL != nil {
		tfMap["max_ttl"] = aws.ToInt64(apiObject.MaxTTL)
	}

	if apiObject.SmoothStreaming != nil {
		tfMap["smooth_streaming"] = aws.ToBool(apiObject.SmoothStreaming)
	}

	if len(apiObject.TrustedKeyGroups.Items) > 0 {
		tfMap["trusted_key_groups"] = flattenTrustedKeyGroups(apiObject.TrustedKeyGroups)
	}

	if len(apiObject.TrustedSigners.Items) > 0 {
		tfMap["trusted_signers"] = flattenTrustedSigners(apiObject.TrustedSigners)
	}

	return tfMap
}

func expandTrustedKeyGroups(tfList []interface{}) *awstypes.TrustedKeyGroups {
	apiObject := &awstypes.TrustedKeyGroups{}

	if len(tfList) > 0 {
		apiObject.Enabled = aws.Bool(true)
		apiObject.Items = flex.ExpandStringValueList(tfList)
		apiObject.Quantity = aws.Int32(int32(len(tfList)))
	} else {
		apiObject.Enabled = aws.Bool(false)
		apiObject.Quantity = aws.Int32(0)
	}

	return apiObject
}

func flattenTrustedKeyGroups(apiObject *awstypes.TrustedKeyGroups) []interface{} {
	if apiObject.Items != nil {
		return flex.FlattenStringValueList(apiObject.Items)
	}

	return []interface{}{}
}

func expandTrustedSigners(tfList []interface{}) *awstypes.TrustedSigners {
	apiObject := &awstypes.TrustedSigners{}

	if len(tfList) > 0 {
		apiObject.Enabled = aws.Bool(true)
		apiObject.Items = flex.ExpandStringValueList(tfList)
		apiObject.Quantity = aws.Int32(int32(len(tfList)))
	} else {
		apiObject.Enabled = aws.Bool(false)
		apiObject.Quantity = aws.Int32(0)
	}

	return apiObject
}

func flattenTrustedSigners(apiObject *awstypes.TrustedSigners) []interface{} {
	if apiObject.Items != nil {
		return flex.FlattenStringValueList(apiObject.Items)
	}

	return []interface{}{}
}

func expandLambdaFunctionAssociation(tfMap map[string]interface{}) *awstypes.LambdaFunctionAssociation {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LambdaFunctionAssociation{}

	if v, ok := tfMap["event_type"]; ok {
		apiObject.EventType = awstypes.EventType(v.(string))
	}

	if v, ok := tfMap["include_body"]; ok {
		apiObject.IncludeBody = aws.Bool(v.(bool))
	}

	if v, ok := tfMap["lambda_arn"]; ok {
		apiObject.LambdaFunctionARN = aws.String(v.(string))
	}

	return apiObject
}

func expandLambdaFunctionAssociations(v interface{}) *awstypes.LambdaFunctionAssociations {
	if v == nil {
		return &awstypes.LambdaFunctionAssociations{
			Quantity: aws.Int32(0),
		}
	}

	tfList := v.([]interface{})

	var items []awstypes.LambdaFunctionAssociation

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := expandLambdaFunctionAssociation(tfMap)

		if item == nil {
			continue
		}

		items = append(items, *item)
	}

	return &awstypes.LambdaFunctionAssociations{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func expandFunctionAssociation(tfMap map[string]interface{}) *awstypes.FunctionAssociation {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FunctionAssociation{}

	if v, ok := tfMap["event_type"]; ok {
		apiObject.EventType = awstypes.EventType(v.(string))
	}

	if v, ok := tfMap["function_arn"]; ok {
		apiObject.FunctionARN = aws.String(v.(string))
	}

	return apiObject
}

func expandFunctionAssociations(v interface{}) *awstypes.FunctionAssociations {
	if v == nil {
		return &awstypes.FunctionAssociations{
			Quantity: aws.Int32(0),
		}
	}

	tfList := v.([]interface{})

	var items []awstypes.FunctionAssociation

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := expandFunctionAssociation(tfMap)

		if item == nil {
			continue
		}

		items = append(items, *item)
	}

	return &awstypes.FunctionAssociations{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenLambdaFunctionAssociation(apiObject *awstypes.LambdaFunctionAssociation) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if apiObject != nil {
		tfMap["event_type"] = apiObject.EventType
		tfMap["include_body"] = aws.ToBool(apiObject.IncludeBody)
		tfMap["lambda_arn"] = aws.ToString(apiObject.LambdaFunctionARN)
	}

	return tfMap
}

func flattenLambdaFunctionAssociations(apiObject *awstypes.LambdaFunctionAssociations) []interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []interface{}

	for _, v := range apiObject.Items {
		tfList = append(tfList, flattenLambdaFunctionAssociation(&v))
	}

	return tfList
}

func flattenFunctionAssociation(apiObject *awstypes.FunctionAssociation) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if apiObject != nil {
		tfMap["event_type"] = apiObject.EventType
		tfMap["function_arn"] = aws.ToString(apiObject.FunctionARN)
	}

	return tfMap
}

func flattenFunctionAssociations(apiObject *awstypes.FunctionAssociations) []interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []interface{}

	for _, v := range apiObject.Items {
		tfList = append(tfList, flattenFunctionAssociation(&v))
	}

	return tfList
}

func expandForwardedValues(tfMap map[string]interface{}) *awstypes.ForwardedValues {
	if len(tfMap) < 1 {
		return nil
	}

	apiObject := &awstypes.ForwardedValues{
		QueryString: aws.Bool(tfMap["query_string"].(bool)),
	}

	if v, ok := tfMap["cookies"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.Cookies = expandCookiePreference(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["headers"]; ok {
		apiObject.Headers = expandForwardedValuesHeaders(v.(*schema.Set).List())
	}

	if v, ok := tfMap["query_string_cache_keys"]; ok {
		apiObject.QueryStringCacheKeys = expandQueryStringCacheKeys(v.([]interface{}))
	}

	return apiObject
}

func flattenForwardedValues(apiObject *awstypes.ForwardedValues) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	tfMap["query_string"] = aws.ToBool(apiObject.QueryString)

	if apiObject.Cookies != nil {
		tfMap["cookies"] = []interface{}{flattenCookiePreference(apiObject.Cookies)}
	}

	if apiObject.Headers != nil {
		tfMap["headers"] = flattenForwardedValuesHeaders(apiObject.Headers)
	}

	if apiObject.QueryStringCacheKeys != nil {
		tfMap["query_string_cache_keys"] = flattenQueryStringCacheKeys(apiObject.QueryStringCacheKeys)
	}

	return tfMap
}

func expandForwardedValuesHeaders(tfList []interface{}) *awstypes.Headers {
	return &awstypes.Headers{
		Items:    flex.ExpandStringValueList(tfList),
		Quantity: aws.Int32(int32(len(tfList))),
	}
}

func flattenForwardedValuesHeaders(apiObject *awstypes.Headers) []interface{} {
	if apiObject.Items != nil {
		return flex.FlattenStringValueList(apiObject.Items)
	}

	return []interface{}{}
}

func expandQueryStringCacheKeys(tfList []interface{}) *awstypes.QueryStringCacheKeys {
	return &awstypes.QueryStringCacheKeys{
		Items:    flex.ExpandStringValueList(tfList),
		Quantity: aws.Int32(int32(len(tfList))),
	}
}

func flattenQueryStringCacheKeys(apiObject *awstypes.QueryStringCacheKeys) []interface{} {
	if apiObject.Items != nil {
		return flex.FlattenStringValueList(apiObject.Items)
	}

	return []interface{}{}
}

func expandCookiePreference(tfMap map[string]interface{}) *awstypes.CookiePreference {
	apiObject := &awstypes.CookiePreference{
		Forward: awstypes.ItemSelection(tfMap["forward"].(string)),
	}

	if v, ok := tfMap["whitelisted_names"]; ok {
		apiObject.WhitelistedNames = expandCookiePreferenceCookieNames(v.(*schema.Set).List())
	}

	return apiObject
}

func flattenCookiePreference(apiObject *awstypes.CookiePreference) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	tfMap["forward"] = apiObject.Forward

	if apiObject.WhitelistedNames != nil {
		tfMap["whitelisted_names"] = flattenCookiePreferenceCookieNames(apiObject.WhitelistedNames)
	}

	return tfMap
}

func expandCookiePreferenceCookieNames(tfList []interface{}) *awstypes.CookieNames {
	return &awstypes.CookieNames{
		Items:    flex.ExpandStringValueList(tfList),
		Quantity: aws.Int32(int32(len(tfList))),
	}
}

func flattenCookiePreferenceCookieNames(apiObject *awstypes.CookieNames) []interface{} {
	if apiObject.Items != nil {
		return flex.FlattenStringValueList(apiObject.Items)
	}

	return []interface{}{}
}

func expandAllowedMethods(tfList []interface{}) *awstypes.AllowedMethods {
	return &awstypes.AllowedMethods{
		Items:    flex.ExpandStringyValueList[awstypes.Method](tfList),
		Quantity: aws.Int32(int32(len(tfList))),
	}
}

func flattenAllowedMethods(apiObject *awstypes.AllowedMethods) []interface{} {
	if apiObject.Items != nil {
		return flex.FlattenStringyValueList(apiObject.Items)
	}

	return nil
}

func expandCachedMethods(tfList []interface{}) *awstypes.CachedMethods {
	return &awstypes.CachedMethods{
		Items:    flex.ExpandStringyValueList[awstypes.Method](tfList),
		Quantity: aws.Int32(int32(len(tfList))),
	}
}

func flattenCachedMethods(apiObject *awstypes.CachedMethods) []interface{} {
	if apiObject.Items != nil {
		return flex.FlattenStringyValueList(apiObject.Items)
	}

	return nil
}

func expandOrigins(tfList []interface{}) *awstypes.Origins {
	var items []awstypes.Origin

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := expandOrigin(tfMap)

		if item == nil {
			continue
		}

		items = append(items, *item)
	}

	return &awstypes.Origins{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenOrigins(apiObject *awstypes.Origins) []interface{} {
	if apiObject.Items == nil {
		return nil
	}

	tfList := []interface{}{}

	for _, v := range apiObject.Items {
		tfList = append(tfList, flattenOrigin(&v))
	}

	return tfList
}

func expandOrigin(tfMap map[string]interface{}) *awstypes.Origin {
	apiObject := &awstypes.Origin{
		DomainName: aws.String(tfMap["domain_name"].(string)),
		Id:         aws.String(tfMap["origin_id"].(string)),
	}

	if v, ok := tfMap["connection_attempts"]; ok {
		apiObject.ConnectionAttempts = aws.Int32(int32(v.(int)))
	}

	if v, ok := tfMap["connection_timeout"]; ok {
		apiObject.ConnectionTimeout = aws.Int32(int32(v.(int)))
	}

	if v, ok := tfMap["custom_header"]; ok {
		apiObject.CustomHeaders = expandCustomHeaders(v.(*schema.Set).List())
	}

	if v, ok := tfMap["custom_origin_config"]; ok {
		if v := v.([]interface{}); len(v) > 0 {
			apiObject.CustomOriginConfig = expandCustomOriginConfig(v[0].(map[string]interface{}))
		}
	}

	if v, ok := tfMap["origin_access_control_id"]; ok {
		apiObject.OriginAccessControlId = aws.String(v.(string))
	}

	if v, ok := tfMap["origin_path"]; ok {
		apiObject.OriginPath = aws.String(v.(string))
	}

	if v, ok := tfMap["origin_shield"]; ok {
		if v := v.([]interface{}); len(v) > 0 {
			apiObject.OriginShield = expandOriginShield(v[0].(map[string]interface{}))
		}
	}

	if v, ok := tfMap["s3_origin_config"]; ok {
		if v := v.([]interface{}); len(v) > 0 {
			apiObject.S3OriginConfig = expandS3OriginConfig(v[0].(map[string]interface{}))
		}
	}

	// if both custom and s3 origin are missing, add an empty s3 origin
	// One or the other must be specified, but the S3 origin can be "empty"
	if apiObject.S3OriginConfig == nil && apiObject.CustomOriginConfig == nil {
		apiObject.S3OriginConfig = &awstypes.S3OriginConfig{
			OriginAccessIdentity: aws.String(""),
		}
	}

	return apiObject
}

func flattenOrigin(apiObject *awstypes.Origin) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})
	tfMap["domain_name"] = aws.ToString(apiObject.DomainName)
	tfMap["origin_id"] = aws.ToString(apiObject.Id)

	if apiObject.ConnectionAttempts != nil {
		tfMap["connection_attempts"] = aws.ToInt32(apiObject.ConnectionAttempts)
	}

	if apiObject.ConnectionTimeout != nil {
		tfMap["connection_timeout"] = aws.ToInt32(apiObject.ConnectionTimeout)
	}

	if apiObject.CustomHeaders != nil {
		tfMap["custom_header"] = flattenCustomHeaders(apiObject.CustomHeaders)
	}

	if apiObject.CustomOriginConfig != nil {
		tfMap["custom_origin_config"] = []interface{}{flattenCustomOriginConfig(apiObject.CustomOriginConfig)}
	}

	if apiObject.OriginAccessControlId != nil {
		tfMap["origin_access_control_id"] = aws.ToString(apiObject.OriginAccessControlId)
	}

	if apiObject.OriginPath != nil {
		tfMap["origin_path"] = aws.ToString(apiObject.OriginPath)
	}

	if apiObject.OriginShield != nil && aws.ToBool(apiObject.OriginShield.Enabled) {
		tfMap["origin_shield"] = []interface{}{flattenOriginShield(apiObject.OriginShield)}
	}

	if apiObject.S3OriginConfig != nil && aws.ToString(apiObject.S3OriginConfig.OriginAccessIdentity) != "" {
		tfMap["s3_origin_config"] = []interface{}{flattenS3OriginConfig(apiObject.S3OriginConfig)}
	}

	return tfMap
}

func expandOriginGroups(tfList []interface{}) *awstypes.OriginGroups {
	var items []awstypes.OriginGroup

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := expandOriginGroup(tfMap)

		if item == nil {
			continue
		}

		items = append(items, *item)
	}

	return &awstypes.OriginGroups{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenOriginGroups(apiObject *awstypes.OriginGroups) []interface{} {
	if apiObject.Items == nil {
		return nil
	}

	var tfList []interface{}

	for _, v := range apiObject.Items {
		tfList = append(tfList, flattenOriginGroup(&v))
	}

	return tfList
}

func expandOriginGroup(tfMap map[string]interface{}) *awstypes.OriginGroup {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OriginGroup{
		FailoverCriteria: expandOriginGroupFailoverCriteria(tfMap["failover_criteria"].([]interface{})[0].(map[string]interface{})),
		Id:               aws.String(tfMap["origin_id"].(string)),
		Members:          expandMembers(tfMap["member"].([]interface{})),
	}

	return apiObject
}

func flattenOriginGroup(apiObject *awstypes.OriginGroup) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})
	tfMap["origin_id"] = aws.ToString(apiObject.Id)

	if apiObject.FailoverCriteria != nil {
		tfMap["failover_criteria"] = flattenOriginGroupFailoverCriteria(apiObject.FailoverCriteria)
	}

	if apiObject.Members != nil {
		tfMap["member"] = flattenOriginGroupMembers(apiObject.Members)
	}

	return tfMap
}

func expandOriginGroupFailoverCriteria(tfMap map[string]interface{}) *awstypes.OriginGroupFailoverCriteria {
	apiObject := &awstypes.OriginGroupFailoverCriteria{}

	if v, ok := tfMap["status_codes"]; ok {
		codes := flex.ExpandInt32ValueList(v.(*schema.Set).List())

		apiObject.StatusCodes = &awstypes.StatusCodes{
			Items:    codes,
			Quantity: aws.Int32(int32(len(codes))),
		}
	}

	return apiObject
}

func flattenOriginGroupFailoverCriteria(apiObject *awstypes.OriginGroupFailoverCriteria) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.StatusCodes.Items; v != nil {
		tfMap["status_codes"] = flex.FlattenInt32ValueList(apiObject.StatusCodes.Items)
	}

	return []interface{}{tfMap}
}

func expandMembers(tfList []interface{}) *awstypes.OriginGroupMembers {
	var items []awstypes.OriginGroupMember

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := awstypes.OriginGroupMember{
			OriginId: aws.String(tfMap["origin_id"].(string)),
		}

		items = append(items, item)
	}

	return &awstypes.OriginGroupMembers{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenOriginGroupMembers(apiObject *awstypes.OriginGroupMembers) []interface{} {
	if apiObject.Items == nil {
		return nil
	}

	tfList := []interface{}{}

	for _, apiObject := range apiObject.Items {
		tfMap := map[string]interface{}{
			"origin_id": aws.ToString(apiObject.OriginId),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandCustomHeaders(tfList []interface{}) *awstypes.CustomHeaders {
	var items []awstypes.OriginCustomHeader

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := expandOriginCustomHeader(tfMap)

		if item == nil {
			continue
		}

		items = append(items, *item)
	}

	return &awstypes.CustomHeaders{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenCustomHeaders(apiObject *awstypes.CustomHeaders) []interface{} {
	if apiObject.Items == nil {
		return nil
	}

	tfList := []interface{}{}

	for _, v := range apiObject.Items {
		tfList = append(tfList, flattenOriginCustomHeader(&v))
	}

	return tfList
}

func expandOriginCustomHeader(tfMap map[string]interface{}) *awstypes.OriginCustomHeader {
	if tfMap == nil {
		return nil
	}

	return &awstypes.OriginCustomHeader{
		HeaderName:  aws.String(tfMap["name"].(string)),
		HeaderValue: aws.String(tfMap["value"].(string)),
	}
}

func flattenOriginCustomHeader(apiObject *awstypes.OriginCustomHeader) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	return map[string]interface{}{
		"name":  aws.ToString(apiObject.HeaderName),
		"value": aws.ToString(apiObject.HeaderValue),
	}
}

func expandCustomOriginConfig(tfMap map[string]interface{}) *awstypes.CustomOriginConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CustomOriginConfig{
		HTTPPort:               aws.Int32(int32(tfMap["http_port"].(int))),
		HTTPSPort:              aws.Int32(int32(tfMap["https_port"].(int))),
		OriginKeepaliveTimeout: aws.Int32(int32(tfMap["origin_keepalive_timeout"].(int))),
		OriginProtocolPolicy:   awstypes.OriginProtocolPolicy(tfMap["origin_protocol_policy"].(string)),
		OriginReadTimeout:      aws.Int32(int32(tfMap["origin_read_timeout"].(int))),
		OriginSslProtocols:     expandCustomOriginConfigSSL(tfMap["origin_ssl_protocols"].(*schema.Set).List()),
	}

	return apiObject
}

func flattenCustomOriginConfig(apiObject *awstypes.CustomOriginConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"http_port":                aws.ToInt32(apiObject.HTTPPort),
		"https_port":               aws.ToInt32(apiObject.HTTPSPort),
		"origin_keepalive_timeout": aws.ToInt32(apiObject.OriginKeepaliveTimeout),
		"origin_protocol_policy":   apiObject.OriginProtocolPolicy,
		"origin_read_timeout":      aws.ToInt32(apiObject.OriginReadTimeout),
		"origin_ssl_protocols":     flattenCustomOriginConfigSSL(apiObject.OriginSslProtocols),
	}

	return tfMap
}

func expandCustomOriginConfigSSL(tfList []interface{}) *awstypes.OriginSslProtocols {
	if tfList == nil {
		return nil

	}

	return &awstypes.OriginSslProtocols{
		Items:    flex.ExpandStringyValueList[awstypes.SslProtocol](tfList),
		Quantity: aws.Int32(int32(len(tfList))),
	}
}

func flattenCustomOriginConfigSSL(apiObject *awstypes.OriginSslProtocols) []interface{} {
	if apiObject == nil {
		return nil
	}

	return flex.FlattenStringyValueList(apiObject.Items)
}

func expandS3OriginConfig(tfMap map[string]interface{}) *awstypes.S3OriginConfig {
	if tfMap == nil {
		return nil
	}

	return &awstypes.S3OriginConfig{
		OriginAccessIdentity: aws.String(tfMap["origin_access_identity"].(string)),
	}
}

func expandOriginShield(tfMap map[string]interface{}) *awstypes.OriginShield {
	if tfMap == nil {
		return nil
	}

	return &awstypes.OriginShield{
		Enabled:            aws.Bool(tfMap["enabled"].(bool)),
		OriginShieldRegion: aws.String(tfMap["origin_shield_region"].(string)),
	}
}

func flattenS3OriginConfig(apiObject *awstypes.S3OriginConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	return map[string]interface{}{
		"origin_access_identity": aws.ToString(apiObject.OriginAccessIdentity),
	}
}

func flattenOriginShield(apiObject *awstypes.OriginShield) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	return map[string]interface{}{
		"enabled":              aws.ToBool(apiObject.Enabled),
		"origin_shield_region": aws.ToString(apiObject.OriginShieldRegion),
	}
}

func expandCustomErrorResponses(tfList []interface{}) *awstypes.CustomErrorResponses {
	var items []awstypes.CustomErrorResponse

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		item := expandCustomErrorResponse(tfMap)

		if item == nil {
			continue
		}

		items = append(items, *item)
	}

	return &awstypes.CustomErrorResponses{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenCustomErrorResponses(apiObject *awstypes.CustomErrorResponses) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfList := []interface{}{}

	for _, v := range apiObject.Items {
		tfList = append(tfList, flattenCustomErrorResponse(&v))
	}

	return tfList
}

func expandCustomErrorResponse(tfMap map[string]interface{}) *awstypes.CustomErrorResponse {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CustomErrorResponse{
		ErrorCode: aws.Int32(int32(tfMap["error_code"].(int))),
	}

	if v, ok := tfMap["error_caching_min_ttl"]; ok {
		apiObject.ErrorCachingMinTTL = aws.Int64(int64(v.(int)))
	}

	if v, ok := tfMap["response_code"]; ok && v.(int) != 0 {
		apiObject.ResponseCode = flex.IntValueToString(v.(int))
	} else {
		apiObject.ResponseCode = aws.String("")
	}

	if v, ok := tfMap["response_page_path"]; ok {
		apiObject.ResponsePagePath = aws.String(v.(string))
	}

	return apiObject
}

func flattenCustomErrorResponse(apiObject *awstypes.CustomErrorResponse) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})
	tfMap["error_code"] = aws.ToInt32(apiObject.ErrorCode)

	if apiObject.ErrorCachingMinTTL != nil {
		tfMap["error_caching_min_ttl"] = aws.ToInt64(apiObject.ErrorCachingMinTTL)
	}

	if apiObject.ResponseCode != nil {
		tfMap["response_code"] = flex.StringToIntValue(apiObject.ResponseCode)
	}

	if apiObject.ResponsePagePath != nil {
		tfMap["response_page_path"] = aws.ToString(apiObject.ResponsePagePath)
	}

	return tfMap
}

func expandLoggingConfig(tfMap map[string]interface{}) *awstypes.LoggingConfig {
	apiObject := &awstypes.LoggingConfig{}

	if tfMap != nil {
		apiObject.Bucket = aws.String(tfMap["bucket"].(string))
		apiObject.Enabled = aws.Bool(true)
		apiObject.IncludeCookies = aws.Bool(tfMap["include_cookies"].(bool))
		apiObject.Prefix = aws.String(tfMap["prefix"].(string))
	} else {
		apiObject.Bucket = aws.String("")
		apiObject.Enabled = aws.Bool(false)
		apiObject.IncludeCookies = aws.Bool(false)
		apiObject.Prefix = aws.String("")
	}

	return apiObject
}

func flattenLoggingConfig(apiObject *awstypes.LoggingConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"bucket":          aws.ToString(apiObject.Bucket),
		"include_cookies": aws.ToBool(apiObject.IncludeCookies),
		"prefix":          aws.ToString(apiObject.Prefix),
	}

	return []interface{}{tfMap}
}

func expandAliases(tfList []interface{}) *awstypes.Aliases {
	apiObject := &awstypes.Aliases{
		Quantity: aws.Int32(int32(len(tfList))),
	}

	if len(tfList) > 0 {
		apiObject.Items = flex.ExpandStringValueList(tfList)
	}

	return apiObject
}

func flattenAliases(apiObject *awstypes.Aliases) []interface{} {
	if apiObject == nil {
		return nil
	}

	if apiObject.Items != nil {
		return flex.FlattenStringValueList(apiObject.Items)
	}

	return []interface{}{}
}

func expandRestrictions(tfMap map[string]interface{}) *awstypes.Restrictions {
	if tfMap == nil {
		return nil
	}

	return &awstypes.Restrictions{
		GeoRestriction: expandGeoRestriction(tfMap["geo_restriction"].([]interface{})[0].(map[string]interface{})),
	}
}

func flattenRestrictions(apiObject *awstypes.Restrictions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"geo_restriction": []interface{}{flattenGeoRestriction(apiObject.GeoRestriction)},
	}

	return []interface{}{tfMap}
}

func expandGeoRestriction(tfMap map[string]interface{}) *awstypes.GeoRestriction {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.GeoRestriction{
		Quantity:        aws.Int32(0),
		RestrictionType: awstypes.GeoRestrictionType(tfMap["restriction_type"].(string)),
	}

	if v, ok := tfMap["locations"]; ok {
		v := v.(*schema.Set)
		apiObject.Items = flex.ExpandStringValueSet(v)
		apiObject.Quantity = aws.Int32(int32(v.Len()))
	}

	return apiObject
}

func flattenGeoRestriction(apiObject *awstypes.GeoRestriction) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})
	tfMap["restriction_type"] = apiObject.RestrictionType

	if apiObject.Items != nil {
		tfMap["locations"] = flex.FlattenStringValueSet(apiObject.Items)
	}

	return tfMap
}

func expandViewerCertificate(tfMap map[string]interface{}) *awstypes.ViewerCertificate {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ViewerCertificate{}

	if v, ok := tfMap["iam_certificate_id"]; ok && v != "" {
		apiObject.IAMCertificateId = aws.String(v.(string))
		apiObject.SSLSupportMethod = awstypes.SSLSupportMethod(tfMap["ssl_support_method"].(string))
	} else if v, ok := tfMap["acm_certificate_arn"]; ok && v != "" {
		apiObject.ACMCertificateArn = aws.String(v.(string))
		apiObject.SSLSupportMethod = awstypes.SSLSupportMethod(tfMap["ssl_support_method"].(string))
	} else {
		apiObject.CloudFrontDefaultCertificate = aws.Bool(tfMap["cloudfront_default_certificate"].(bool))
	}

	if v, ok := tfMap["minimum_protocol_version"]; ok && v != "" {
		apiObject.MinimumProtocolVersion = awstypes.MinimumProtocolVersion(v.(string))
	}

	return apiObject
}

func flattenViewerCertificate(apiObject *awstypes.ViewerCertificate) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if apiObject.IAMCertificateId != nil {
		tfMap["iam_certificate_id"] = aws.ToString(apiObject.IAMCertificateId)
		tfMap["ssl_support_method"] = apiObject.SSLSupportMethod
	}

	if apiObject.ACMCertificateArn != nil {
		tfMap["acm_certificate_arn"] = aws.ToString(apiObject.ACMCertificateArn)
		tfMap["ssl_support_method"] = apiObject.SSLSupportMethod
	}

	if apiObject.CloudFrontDefaultCertificate != nil {
		tfMap["cloudfront_default_certificate"] = aws.ToBool(apiObject.CloudFrontDefaultCertificate)
	}

	tfMap["minimum_protocol_version"] = apiObject.MinimumProtocolVersion

	return []interface{}{tfMap}
}

func flattenActiveTrustedKeyGroups(apiObject *awstypes.ActiveTrustedKeyGroups) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"enabled": aws.ToBool(apiObject.Enabled),
		"items":   flattenKGKeyPairIDs(apiObject.Items),
	}

	return []interface{}{tfMap}
}

func flattenKGKeyPairIDs(apiObjects []awstypes.KGKeyPairIds) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"key_group_id": aws.ToString(apiObject.KeyGroupId),
			"key_pair_ids": apiObject.KeyPairIds.Items,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenActiveTrustedSigners(apiObject *awstypes.ActiveTrustedSigners) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"enabled": aws.ToBool(apiObject.Enabled),
		"items":   flattenSigners(apiObject.Items),
	}

	return []interface{}{tfMap}
}

func flattenSigners(apiObjects []awstypes.Signer) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"aws_account_number": aws.ToString(apiObject.AwsAccountNumber),
			"key_pair_ids":       apiObject.KeyPairIds.Items,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
