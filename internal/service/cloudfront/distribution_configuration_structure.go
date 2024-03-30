// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// CloudFront DistributionConfig structure helpers.
//
// These functions assist in pulling in data from Terraform resource
// configuration for the aws_cloudfront_distribution resource, as there are
// several sub-fields that require their own data type, and do not necessarily
// 1-1 translate to resource configuration.

package cloudfront

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// Assemble the *awstypes.DistributionConfig variable. Calls out to various
// expander functions to convert attributes and sub-attributes to the various
// complex structures which are necessary to properly build the
// DistributionConfig structure.
//
// Used by the aws_cloudfront_distribution Create and Update functions.
func expandDistributionConfig(d *schema.ResourceData) *awstypes.DistributionConfig {
	distributionConfig := &awstypes.DistributionConfig{
		CacheBehaviors:               expandCacheBehaviors(d.Get("ordered_cache_behavior").([]interface{})),
		CallerReference:              aws.String(id.UniqueId()),
		Comment:                      aws.String(d.Get("comment").(string)),
		ContinuousDeploymentPolicyId: aws.String(d.Get("continuous_deployment_policy_id").(string)),
		CustomErrorResponses:         ExpandCustomErrorResponses(d.Get("custom_error_response").(*schema.Set)),
		DefaultCacheBehavior:         ExpandDefaultCacheBehavior(d.Get("default_cache_behavior").([]interface{})[0].(map[string]interface{})),
		DefaultRootObject:            aws.String(d.Get("default_root_object").(string)),
		Enabled:                      aws.Bool(d.Get("enabled").(bool)),
		IsIPV6Enabled:                aws.Bool(d.Get("is_ipv6_enabled").(bool)),
		HttpVersion:                  awstypes.HttpVersion(d.Get("http_version").(string)),
		Origins:                      ExpandOrigins(d.Get("origin").(*schema.Set)),
		PriceClass:                   awstypes.PriceClass(d.Get("price_class").(string)),
		Staging:                      aws.Bool(d.Get("staging").(bool)),
		WebACLId:                     aws.String(d.Get("web_acl_id").(string)),
	}

	// This sets CallerReference if it's still pending computation (ie: new resource)
	if v, ok := d.GetOk("caller_reference"); ok {
		distributionConfig.CallerReference = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging_config"); ok {
		distributionConfig.Logging = ExpandLoggingConfig(v.([]interface{})[0].(map[string]interface{}))
	} else {
		distributionConfig.Logging = ExpandLoggingConfig(nil)
	}
	if v, ok := d.GetOk("aliases"); ok {
		distributionConfig.Aliases = ExpandAliases(v.(*schema.Set))
	} else {
		distributionConfig.Aliases = ExpandAliases(schema.NewSet(AliasesHash, []interface{}{}))
	}
	if v, ok := d.GetOk("restrictions"); ok {
		distributionConfig.Restrictions = ExpandRestrictions(v.([]interface{})[0].(map[string]interface{}))
	}
	if v, ok := d.GetOk("viewer_certificate"); ok {
		distributionConfig.ViewerCertificate = ExpandViewerCertificate(v.([]interface{})[0].(map[string]interface{}))
	}
	if v, ok := d.GetOk("origin_group"); ok {
		distributionConfig.OriginGroups = ExpandOriginGroups(v.(*schema.Set))
	}
	return distributionConfig
}

// Unpack the *awstypes.DistributionConfig variable and set resource data.
// Calls out to flatten functions to convert the DistributionConfig
// sub-structures to their respective attributes in the
// aws_cloudfront_distribution resource.
//
// Used by the aws_cloudfront_distribution Read function.
func flattenDistributionConfig(d *schema.ResourceData, distributionConfig *awstypes.DistributionConfig) error {
	var err error

	d.Set("enabled", distributionConfig.Enabled)
	d.Set("is_ipv6_enabled", distributionConfig.IsIPV6Enabled)
	d.Set("price_class", distributionConfig.PriceClass)

	err = d.Set("default_cache_behavior", []interface{}{flattenDefaultCacheBehavior(distributionConfig.DefaultCacheBehavior)})
	if err != nil {
		return err // nosemgrep:ci.bare-error-returns
	}
	err = d.Set("viewer_certificate", flattenViewerCertificate(distributionConfig.ViewerCertificate))
	if err != nil {
		return err // nosemgrep:ci.bare-error-returns
	}

	d.Set("caller_reference", distributionConfig.CallerReference)
	if distributionConfig.Comment != nil {
		if aws.ToString(distributionConfig.Comment) != "" {
			d.Set("comment", distributionConfig.Comment)
		}
	}
	d.Set("default_root_object", distributionConfig.DefaultRootObject)
	d.Set("http_version", distributionConfig.HttpVersion)
	d.Set("staging", distributionConfig.Staging)
	d.Set("web_acl_id", distributionConfig.WebACLId)

	// Not having this set for staging distributions causes IllegalUpdate errors when making updates of any kind.
	// If this absolutely must not be optional/computed, the policy ID will need to be retrieved and set for each
	// API call for staging distributions.
	d.Set("continuous_deployment_policy_id", distributionConfig.ContinuousDeploymentPolicyId)

	if distributionConfig.CustomErrorResponses != nil {
		err = d.Set("custom_error_response", FlattenCustomErrorResponses(distributionConfig.CustomErrorResponses))
		if err != nil {
			return err // nosemgrep:ci.bare-error-returns
		}
	}
	if distributionConfig.CacheBehaviors != nil {
		if err := d.Set("ordered_cache_behavior", flattenCacheBehaviors(distributionConfig.CacheBehaviors)); err != nil {
			return err // nosemgrep:ci.bare-error-returns
		}
	}

	if distributionConfig.Logging != nil && *distributionConfig.Logging.Enabled {
		err = d.Set("logging_config", flattenLoggingConfig(distributionConfig.Logging))
	} else {
		err = d.Set("logging_config", []interface{}{})
	}
	if err != nil {
		return err // nosemgrep:ci.bare-error-returns
	}

	if distributionConfig.Aliases != nil {
		err = d.Set("aliases", FlattenAliases(distributionConfig.Aliases))
		if err != nil {
			return err // nosemgrep:ci.bare-error-returns
		}
	}
	if distributionConfig.Restrictions != nil {
		err = d.Set("restrictions", flattenRestrictions(distributionConfig.Restrictions))
		if err != nil {
			return err // nosemgrep:ci.bare-error-returns
		}
	}
	if *aws.Int32(*distributionConfig.Origins.Quantity) > 0 {
		err = d.Set("origin", FlattenOrigins(distributionConfig.Origins))
		if err != nil {
			return err // nosemgrep:ci.bare-error-returns
		}
	}
	if *aws.Int32(*distributionConfig.OriginGroups.Quantity) > 0 {
		err = d.Set("origin_group", FlattenOriginGroups(distributionConfig.OriginGroups))
		if err != nil {
			return err // nosemgrep:ci.bare-error-returns
		}
	}

	return nil
}

func expandCacheBehaviors(lst []interface{}) *awstypes.CacheBehaviors {
	var qty int32
	var items []awstypes.CacheBehavior
	for _, v := range lst {
		items = append(items, *expandCacheBehavior(v.(map[string]interface{})))
		qty++
	}
	return &awstypes.CacheBehaviors{
		Quantity: aws.Int32(qty),
		Items:    items,
	}
}

func flattenCacheBehaviors(cbs *awstypes.CacheBehaviors) []interface{} {
	lst := []interface{}{}
	for _, v := range cbs.Items {
		lst = append(lst, flattenCacheBehavior(&v))
	}
	return lst
}

func ExpandDefaultCacheBehavior(m map[string]interface{}) *awstypes.DefaultCacheBehavior {
	dcb := &awstypes.DefaultCacheBehavior{
		CachePolicyId:           aws.String(m["cache_policy_id"].(string)),
		Compress:                aws.Bool(m["compress"].(bool)),
		FieldLevelEncryptionId:  aws.String(m["field_level_encryption_id"].(string)),
		OriginRequestPolicyId:   aws.String(m["origin_request_policy_id"].(string)),
		ResponseHeadersPolicyId: aws.String(m["response_headers_policy_id"].(string)),
		TargetOriginId:          aws.String(m["target_origin_id"].(string)),
		ViewerProtocolPolicy:    awstypes.ViewerProtocolPolicy(m["viewer_protocol_policy"].(string)),
	}

	if forwardedValuesFlat, ok := m["forwarded_values"].([]interface{}); ok && len(forwardedValuesFlat) == 1 {
		dcb.ForwardedValues = ExpandForwardedValues(m["forwarded_values"].([]interface{})[0].(map[string]interface{}))
	}

	if m["cache_policy_id"].(string) == "" {
		dcb.MinTTL = aws.Int64(int64(m["min_ttl"].(int)))
		dcb.MaxTTL = aws.Int64(int64(m["max_ttl"].(int)))
		dcb.DefaultTTL = aws.Int64(int64(m["default_ttl"].(int)))
	}

	if v, ok := m["trusted_key_groups"]; ok {
		dcb.TrustedKeyGroups = expandTrustedKeyGroups(v.([]interface{}))
	} else {
		dcb.TrustedKeyGroups = expandTrustedKeyGroups([]interface{}{})
	}

	if v, ok := m["trusted_signers"]; ok {
		dcb.TrustedSigners = ExpandTrustedSigners(v.([]interface{}))
	} else {
		dcb.TrustedSigners = ExpandTrustedSigners([]interface{}{})
	}

	if v, ok := m["lambda_function_association"]; ok {
		dcb.LambdaFunctionAssociations = ExpandLambdaFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := m["function_association"]; ok {
		dcb.FunctionAssociations = ExpandFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := m["smooth_streaming"]; ok {
		dcb.SmoothStreaming = aws.Bool(v.(bool))
	}
	if v, ok := m["allowed_methods"]; ok {
		dcb.AllowedMethods = ExpandAllowedMethods(v.(*schema.Set))
	}
	if v, ok := m["cached_methods"]; ok {
		dcb.AllowedMethods.CachedMethods = ExpandCachedMethods(v.(*schema.Set))
	}
	if v, ok := m["realtime_log_config_arn"]; ok && v.(string) != "" {
		dcb.RealtimeLogConfigArn = aws.String(v.(string))
	}

	return dcb
}

func expandCacheBehavior(m map[string]interface{}) *awstypes.CacheBehavior {
	var forwardedValues *awstypes.ForwardedValues
	if forwardedValuesFlat, ok := m["forwarded_values"].([]interface{}); ok && len(forwardedValuesFlat) == 1 {
		forwardedValues = ExpandForwardedValues(m["forwarded_values"].([]interface{})[0].(map[string]interface{}))
	}

	cb := &awstypes.CacheBehavior{
		CachePolicyId:           aws.String(m["cache_policy_id"].(string)),
		Compress:                aws.Bool(m["compress"].(bool)),
		FieldLevelEncryptionId:  aws.String(m["field_level_encryption_id"].(string)),
		ForwardedValues:         forwardedValues,
		OriginRequestPolicyId:   aws.String(m["origin_request_policy_id"].(string)),
		ResponseHeadersPolicyId: aws.String(m["response_headers_policy_id"].(string)),
		TargetOriginId:          aws.String(m["target_origin_id"].(string)),
		ViewerProtocolPolicy:    awstypes.ViewerProtocolPolicy(m["viewer_protocol_policy"].(string)),
	}

	if m["cache_policy_id"].(string) == "" {
		cb.MinTTL = aws.Int64(int64(m["min_ttl"].(int)))
		cb.MaxTTL = aws.Int64(int64(m["max_ttl"].(int)))
		cb.DefaultTTL = aws.Int64(int64(m["default_ttl"].(int)))
	}

	if v, ok := m["trusted_key_groups"]; ok {
		cb.TrustedKeyGroups = expandTrustedKeyGroups(v.([]interface{}))
	} else {
		cb.TrustedKeyGroups = expandTrustedKeyGroups([]interface{}{})
	}

	if v, ok := m["trusted_signers"]; ok {
		cb.TrustedSigners = ExpandTrustedSigners(v.([]interface{}))
	} else {
		cb.TrustedSigners = ExpandTrustedSigners([]interface{}{})
	}

	if v, ok := m["lambda_function_association"]; ok {
		cb.LambdaFunctionAssociations = ExpandLambdaFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := m["function_association"]; ok {
		cb.FunctionAssociations = ExpandFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := m["smooth_streaming"]; ok {
		cb.SmoothStreaming = aws.Bool(v.(bool))
	}
	if v, ok := m["allowed_methods"]; ok {
		cb.AllowedMethods = ExpandAllowedMethods(v.(*schema.Set))
	}
	if v, ok := m["cached_methods"]; ok {
		cb.AllowedMethods.CachedMethods = ExpandCachedMethods(v.(*schema.Set))
	}
	if v, ok := m["path_pattern"]; ok {
		cb.PathPattern = aws.String(v.(string))
	}
	if v, ok := m["realtime_log_config_arn"]; ok && v.(string) != "" {
		cb.RealtimeLogConfigArn = aws.String(v.(string))
	}

	return cb
}

func flattenDefaultCacheBehavior(dcb *awstypes.DefaultCacheBehavior) map[string]interface{} {
	m := map[string]interface{}{
		"cache_policy_id":            aws.ToString(dcb.CachePolicyId),
		"compress":                   aws.Bool(*dcb.Compress),
		"field_level_encryption_id":  aws.ToString(dcb.FieldLevelEncryptionId),
		"viewer_protocol_policy":     aws.ToString((*string)(&dcb.ViewerProtocolPolicy)),
		"target_origin_id":           aws.ToString(dcb.TargetOriginId),
		"min_ttl":                    aws.Int64(*dcb.MinTTL),
		"origin_request_policy_id":   aws.ToString(dcb.OriginRequestPolicyId),
		"realtime_log_config_arn":    aws.ToString(dcb.RealtimeLogConfigArn),
		"response_headers_policy_id": aws.ToString(dcb.ResponseHeadersPolicyId),
	}

	if dcb.ForwardedValues != nil {
		m["forwarded_values"] = []interface{}{FlattenForwardedValues(dcb.ForwardedValues)}
	}
	if len(dcb.TrustedKeyGroups.Items) > 0 {
		m["trusted_key_groups"] = flattenTrustedKeyGroups(dcb.TrustedKeyGroups)
	}
	if len(dcb.TrustedSigners.Items) > 0 {
		m["trusted_signers"] = FlattenTrustedSigners(dcb.TrustedSigners)
	}
	if len(dcb.LambdaFunctionAssociations.Items) > 0 {
		m["lambda_function_association"] = FlattenLambdaFunctionAssociations(dcb.LambdaFunctionAssociations)
	}
	if len(dcb.FunctionAssociations.Items) > 0 {
		m["function_association"] = FlattenFunctionAssociations(dcb.FunctionAssociations)
	}
	if dcb.MaxTTL != nil {
		m["max_ttl"] = aws.Int64(*dcb.MaxTTL)
	}
	if dcb.SmoothStreaming != nil {
		m["smooth_streaming"] = aws.Bool(*dcb.SmoothStreaming)
	}
	if dcb.DefaultTTL != nil {
		m["default_ttl"] = int(*dcb.DefaultTTL)
	}
	if dcb.AllowedMethods != nil {
		m["allowed_methods"] = FlattenAllowedMethods(dcb.AllowedMethods)
	}
	if dcb.AllowedMethods.CachedMethods != nil {
		m["cached_methods"] = FlattenCachedMethods(dcb.AllowedMethods.CachedMethods)
	}

	return m
}

func flattenCacheBehavior(cb *awstypes.CacheBehavior) map[string]interface{} {
	m := make(map[string]interface{})

	m["cache_policy_id"] = aws.ToString(cb.CachePolicyId)
	m["compress"] = aws.Bool(*cb.Compress)
	m["field_level_encryption_id"] = aws.ToString(cb.FieldLevelEncryptionId)
	m["viewer_protocol_policy"] = cb.ViewerProtocolPolicy
	m["target_origin_id"] = aws.ToString(cb.TargetOriginId)
	m["min_ttl"] = cb.MinTTL
	m["origin_request_policy_id"] = aws.ToString(cb.OriginRequestPolicyId)
	m["realtime_log_config_arn"] = aws.ToString(cb.RealtimeLogConfigArn)
	m["response_headers_policy_id"] = aws.ToString(cb.ResponseHeadersPolicyId)

	if cb.ForwardedValues != nil {
		m["forwarded_values"] = []interface{}{FlattenForwardedValues(cb.ForwardedValues)}
	}
	if len(cb.TrustedKeyGroups.Items) > 0 {
		m["trusted_key_groups"] = flattenTrustedKeyGroups(cb.TrustedKeyGroups)
	}
	if len(cb.TrustedSigners.Items) > 0 {
		m["trusted_signers"] = FlattenTrustedSigners(cb.TrustedSigners)
	}
	if len(cb.LambdaFunctionAssociations.Items) > 0 {
		m["lambda_function_association"] = FlattenLambdaFunctionAssociations(cb.LambdaFunctionAssociations)
	}
	if len(cb.FunctionAssociations.Items) > 0 {
		m["function_association"] = FlattenFunctionAssociations(cb.FunctionAssociations)
	}
	if cb.MaxTTL != nil {
		m["max_ttl"] = int(*aws.Int64(*cb.MaxTTL))
	}
	if cb.SmoothStreaming != nil {
		m["smooth_streaming"] = cb.SmoothStreaming
	}
	if cb.DefaultTTL != nil {
		m["default_ttl"] = int(*aws.Int64(*cb.DefaultTTL))
	}
	if cb.AllowedMethods != nil {
		m["allowed_methods"] = FlattenAllowedMethods(cb.AllowedMethods)
	}
	if cb.AllowedMethods.CachedMethods != nil {
		m["cached_methods"] = FlattenCachedMethods(cb.AllowedMethods.CachedMethods)
	}
	if cb.PathPattern != nil {
		m["path_pattern"] = aws.ToString(cb.PathPattern)
	}
	return m
}

func expandTrustedKeyGroups(s []interface{}) *awstypes.TrustedKeyGroups {
	var tkg awstypes.TrustedKeyGroups
	if len(s) > 0 {
		tkg.Quantity = aws.Int32(int32(len(s)))
		tkg.Items = flex.ExpandStringValueList(s)
		tkg.Enabled = aws.Bool(true)
	} else {
		tkg.Quantity = aws.Int32(0)
		tkg.Enabled = aws.Bool(false)
	}
	return &tkg
}

func flattenTrustedKeyGroups(tkg *awstypes.TrustedKeyGroups) []interface{} {
	if tkg.Items != nil {
		return flex.FlattenStringValueList(tkg.Items)
	}
	return []interface{}{}
}

func ExpandTrustedSigners(s []interface{}) *awstypes.TrustedSigners {
	var ts awstypes.TrustedSigners
	if len(s) > 0 {
		ts.Quantity = aws.Int32(int32(len(s)))
		ts.Items = flex.ExpandStringValueList(s)
		ts.Enabled = aws.Bool(true)
	} else {
		ts.Quantity = aws.Int32(0)
		ts.Enabled = aws.Bool(false)
	}
	return &ts
}

func FlattenTrustedSigners(ts *awstypes.TrustedSigners) []interface{} {
	if ts.Items != nil {
		return flex.FlattenStringValueList(ts.Items)
	}
	return []interface{}{}
}

func LambdaFunctionAssociationHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["event_type"].(string)))
	buf.WriteString(m["lambda_arn"].(string))
	buf.WriteString(fmt.Sprintf("%t", m["include_body"].(bool)))
	return create.StringHashcode(buf.String())
}

func FunctionAssociationHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["event_type"].(string)))
	buf.WriteString(m["function_arn"].(string))
	return create.StringHashcode(buf.String())
}

func ExpandLambdaFunctionAssociations(v interface{}) *awstypes.LambdaFunctionAssociations {
	if v == nil {
		return &awstypes.LambdaFunctionAssociations{
			Quantity: aws.Int32(0),
		}
	}

	s := v.([]interface{})
	var lfa awstypes.LambdaFunctionAssociations
	lfa.Quantity = aws.Int32(int32(len(s)))
	lfa.Items = make([]awstypes.LambdaFunctionAssociation, len(s))
	for i, lf := range s {
		lfa.Items[i] = *expandLambdaFunctionAssociation(lf.(map[string]interface{}))
	}
	return &lfa
}

func expandLambdaFunctionAssociation(lf map[string]interface{}) *awstypes.LambdaFunctionAssociation {
	var lfa awstypes.LambdaFunctionAssociation
	if v, ok := lf["event_type"]; ok {
		lfa.EventType = awstypes.EventType(v.(string))
	}
	if v, ok := lf["lambda_arn"]; ok {
		lfa.LambdaFunctionARN = aws.String(v.(string))
	}
	if v, ok := lf["include_body"]; ok {
		lfa.IncludeBody = aws.Bool(v.(bool))
	}
	return &lfa
}

func ExpandFunctionAssociations(v interface{}) *awstypes.FunctionAssociations {
	if v == nil {
		return &awstypes.FunctionAssociations{
			Quantity: aws.Int32(0),
		}
	}

	s := v.([]interface{})
	var fa awstypes.FunctionAssociations
	fa.Quantity = aws.Int32(int32(len(s)))
	fa.Items = make([]awstypes.FunctionAssociation, len(s))

	for i, f := range s {
		fa.Items[i] = *expandFunctionAssociation(f.(map[string]interface{}))
	}
	return &fa
}

func expandFunctionAssociation(f map[string]interface{}) *awstypes.FunctionAssociation {
	var fa awstypes.FunctionAssociation
	if v, ok := f["event_type"]; ok {
		fa.EventType = awstypes.EventType(v.(string))
	}
	if v, ok := f["function_arn"]; ok {
		fa.FunctionARN = aws.String(v.(string))
	}
	return &fa
}

func FlattenLambdaFunctionAssociations(lfa *awstypes.LambdaFunctionAssociations) *schema.Set {
	s := schema.NewSet(LambdaFunctionAssociationHash, []interface{}{})
	for _, v := range lfa.Items {
		s.Add(flattenLambdaFunctionAssociation(&v))
	}
	return s
}

func flattenLambdaFunctionAssociation(lfa *awstypes.LambdaFunctionAssociation) map[string]interface{} {
	m := map[string]interface{}{}
	if lfa != nil {
		m["event_type"] = awstypes.EventType(lfa.EventType)
		m["lambda_arn"] = aws.ToString(lfa.LambdaFunctionARN)
		m["include_body"] = aws.Bool(*lfa.IncludeBody)
	}
	return m
}

func FlattenFunctionAssociations(fa *awstypes.FunctionAssociations) *schema.Set {
	s := schema.NewSet(FunctionAssociationHash, []interface{}{})
	for _, v := range fa.Items {
		s.Add(flattenFunctionAssociation(&v))
	}
	return s
}

func flattenFunctionAssociation(fa *awstypes.FunctionAssociation) map[string]interface{} {
	m := map[string]interface{}{}
	eventType := string(fa.EventType)
	if fa != nil {
		m["event_type"] = aws.ToString(&eventType)
		m["function_arn"] = aws.ToString(fa.FunctionARN)
	}
	return m
}

func ExpandForwardedValues(m map[string]interface{}) *awstypes.ForwardedValues {
	if len(m) < 1 {
		return nil
	}

	fv := &awstypes.ForwardedValues{
		QueryString: aws.Bool(m["query_string"].(bool)),
	}
	if v, ok := m["cookies"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		fv.Cookies = ExpandCookiePreference(v.([]interface{})[0].(map[string]interface{}))
	}
	if v, ok := m["headers"]; ok {
		fv.Headers = ExpandHeaders(v.(*schema.Set).List())
	}
	if v, ok := m["query_string_cache_keys"]; ok {
		fv.QueryStringCacheKeys = ExpandQueryStringCacheKeys(v.([]interface{}))
	}
	return fv
}

func FlattenForwardedValues(fv *awstypes.ForwardedValues) map[string]interface{} {
	m := make(map[string]interface{})
	m["query_string"] = aws.ToBool(fv.QueryString)
	if fv.Cookies != nil {
		m["cookies"] = []interface{}{FlattenCookiePreference(fv.Cookies)}
	}
	if fv.Headers != nil {
		m["headers"] = schema.NewSet(schema.HashString, FlattenHeaders(fv.Headers))
	}
	if fv.QueryStringCacheKeys != nil {
		m["query_string_cache_keys"] = FlattenQueryStringCacheKeys(fv.QueryStringCacheKeys)
	}
	return m
}

func ExpandHeaders(d []interface{}) *awstypes.Headers {
	return &awstypes.Headers{
		Quantity: aws.Int32(int32(len(d))),
		Items:    flex.ExpandStringValueList(d),
	}
}

func FlattenHeaders(h *awstypes.Headers) []interface{} {
	if h.Items != nil {
		return flex.FlattenStringValueList(h.Items)
	}
	return []interface{}{}
}

func ExpandQueryStringCacheKeys(d []interface{}) *awstypes.QueryStringCacheKeys {
	return &awstypes.QueryStringCacheKeys{
		Quantity: aws.Int32(int32(len(d))),
		Items:    flex.ExpandStringValueList(d),
	}
}

func FlattenQueryStringCacheKeys(k *awstypes.QueryStringCacheKeys) []interface{} {
	if k.Items != nil {
		return flex.FlattenStringValueList(k.Items)
	}
	return []interface{}{}
}

func ExpandCookiePreference(m map[string]interface{}) *awstypes.CookiePreference {
	cp := &awstypes.CookiePreference{
		Forward: awstypes.ItemSelection(m["forward"].(string)),
	}
	if v, ok := m["whitelisted_names"]; ok {
		cp.WhitelistedNames = ExpandCookieNames(v.(*schema.Set).List())
	}
	return cp
}

func FlattenCookiePreference(cp *awstypes.CookiePreference) map[string]interface{} {
	m := make(map[string]interface{})
	m["forward"] = awstypes.ItemSelection(cp.Forward)
	if cp.WhitelistedNames != nil {
		m["whitelisted_names"] = schema.NewSet(schema.HashString, FlattenCookieNames(cp.WhitelistedNames))
	}
	return m
}

func ExpandCookieNames(d []interface{}) *awstypes.CookieNames {
	return &awstypes.CookieNames{
		Quantity: aws.Int32(int32(len(d))),
		Items:    flex.ExpandStringValueList(d),
	}
}

func FlattenCookieNames(cn *awstypes.CookieNames) []interface{} {
	if cn.Items != nil {
		return flex.FlattenStringValueList(cn.Items)
	}
	return []interface{}{}
}

func ExpandAllowedMethods(s *schema.Set) *awstypes.AllowedMethods {
	items := make([]awstypes.Method, 0)
	for _, v := range flex.ExpandStringSet(s) {
		items = append(items, awstypes.Method(*v))
	}
	return &awstypes.AllowedMethods{
		Quantity: aws.Int32(int32(s.Len())),
		Items:    items,
	}
}

func FlattenAllowedMethods(am *awstypes.AllowedMethods) *schema.Set {
	items := make([]*string, 0)
	return flex.FlattenStringSet(items)
}

func ExpandCachedMethods(s *schema.Set) *awstypes.CachedMethods {
	items := make([]awstypes.Method, 0)
	for _, v := range flex.ExpandStringSet(s) {
		items = append(items, awstypes.Method(*v))
	}
	return &awstypes.CachedMethods{
		Quantity: aws.Int32(int32(s.Len())),
		Items:    items,
	}
}

func FlattenCachedMethods(cm *awstypes.CachedMethods) *schema.Set {
	if cm.Items != nil {
		cmItems := make([]string, 0)
		for _, v := range cm.Items {
			temp := string(v)
			cmItems = append(cmItems, temp)
		}
		return flex.FlattenStringValueSet(cmItems)
	}
	return nil
}

func ExpandOrigins(s *schema.Set) *awstypes.Origins {
	qty := 0
	items := []awstypes.Origin{}
	for _, v := range s.List() {
		items = append(items, ExpandOrigin(v.(map[string]interface{})))
		qty++
	}
	return &awstypes.Origins{
		Quantity: aws.Int32(int32(qty)),
		Items:    items,
	}
}

func FlattenOrigins(ors *awstypes.Origins) *schema.Set {
	s := []interface{}{}
	for _, v := range ors.Items {
		s = append(s, FlattenOrigin(&v))
	}
	return schema.NewSet(OriginHash, s)
}

func ExpandOrigin(m map[string]interface{}) awstypes.Origin {
	origin := &awstypes.Origin{
		Id:         aws.String(m["origin_id"].(string)),
		DomainName: aws.String(m["domain_name"].(string)),
	}

	if v, ok := m["connection_attempts"]; ok {
		origin.ConnectionAttempts = aws.Int32(int32(v.(int)))
	}
	if v, ok := m["connection_timeout"]; ok {
		origin.ConnectionTimeout = aws.Int32(int32(v.(int)))
	}
	if v, ok := m["custom_header"]; ok {
		origin.CustomHeaders = ExpandCustomHeaders(v.(*schema.Set))
	}
	if v, ok := m["custom_origin_config"]; ok {
		if s := v.([]interface{}); len(s) > 0 {
			origin.CustomOriginConfig = ExpandCustomOriginConfig(s[0].(map[string]interface{}))
		}
	}
	if v, ok := m["origin_access_control_id"]; ok {
		origin.OriginAccessControlId = aws.String(v.(string))
	}
	if v, ok := m["origin_path"]; ok {
		origin.OriginPath = aws.String(v.(string))
	}

	if v, ok := m["origin_shield"]; ok {
		if s := v.([]interface{}); len(s) > 0 {
			origin.OriginShield = ExpandOriginShield(s[0].(map[string]interface{}))
		}
	}

	if v, ok := m["s3_origin_config"]; ok {
		if s := v.([]interface{}); len(s) > 0 {
			origin.S3OriginConfig = ExpandS3OriginConfig(s[0].(map[string]interface{}))
		}
	}

	// if both custom and s3 origin are missing, add an empty s3 origin
	// One or the other must be specified, but the S3 origin can be "empty"
	if origin.S3OriginConfig == nil && origin.CustomOriginConfig == nil {
		origin.S3OriginConfig = &awstypes.S3OriginConfig{
			OriginAccessIdentity: aws.String(""),
		}
	}

	return *origin
}

func FlattenOrigin(or *awstypes.Origin) map[string]interface{} {
	m := make(map[string]interface{})
	m["origin_id"] = aws.ToString(or.Id)
	m["domain_name"] = aws.ToString(or.DomainName)
	if or.ConnectionAttempts != nil {
		m["connection_attempts"] = int(*aws.Int32(*or.ConnectionAttempts))
	}
	if or.ConnectionTimeout != nil {
		m["connection_timeout"] = int(*aws.Int32(*or.ConnectionTimeout))
	}
	if or.CustomHeaders != nil {
		m["custom_header"] = FlattenCustomHeaders(or.CustomHeaders)
	}
	if or.CustomOriginConfig != nil {
		m["custom_origin_config"] = []interface{}{FlattenCustomOriginConfig(or.CustomOriginConfig)}
	}
	if or.OriginAccessControlId != nil {
		m["origin_access_control_id"] = aws.ToString(or.OriginAccessControlId)
	}
	if or.OriginPath != nil {
		m["origin_path"] = aws.ToString(or.OriginPath)
	}
	if or.OriginShield != nil && aws.ToBool(or.OriginShield.Enabled) {
		m["origin_shield"] = []interface{}{FlattenOriginShield(or.OriginShield)}
	}
	if or.S3OriginConfig != nil && aws.ToString(or.S3OriginConfig.OriginAccessIdentity) != "" {
		m["s3_origin_config"] = []interface{}{FlattenS3OriginConfig(or.S3OriginConfig)}
	}
	return m
}

func ExpandOriginGroups(s *schema.Set) *awstypes.OriginGroups {
	qty := 0
	items := []awstypes.OriginGroup{}
	for _, v := range s.List() {
		items = append(items, *expandOriginGroup(v.(map[string]interface{})))
		qty++
	}
	return &awstypes.OriginGroups{
		Quantity: aws.Int32(int32(qty)),
		Items:    items,
	}
}

func FlattenOriginGroups(ogs *awstypes.OriginGroups) *schema.Set {
	s := []interface{}{}
	for _, v := range ogs.Items {
		s = append(s, flattenOriginGroup(&v))
	}
	return schema.NewSet(OriginGroupHash, s)
}

func expandOriginGroup(m map[string]interface{}) *awstypes.OriginGroup {
	failoverCriteria := m["failover_criteria"].([]interface{})[0].(map[string]interface{})
	members := m["member"].([]interface{})
	originGroup := &awstypes.OriginGroup{
		Id:               aws.String(m["origin_id"].(string)),
		FailoverCriteria: expandOriginGroupFailoverCriteria(failoverCriteria),
		Members:          expandMembers(members),
	}
	return originGroup
}

func flattenOriginGroup(og *awstypes.OriginGroup) map[string]interface{} {
	m := make(map[string]interface{})
	m["origin_id"] = aws.ToString(og.Id)
	if og.FailoverCriteria != nil {
		m["failover_criteria"] = flattenOriginGroupFailoverCriteria(og.FailoverCriteria)
	}
	if og.Members != nil {
		m["member"] = flattenOriginGroupMembers(og.Members)
	}
	return m
}

func expandOriginGroupFailoverCriteria(m map[string]interface{}) *awstypes.OriginGroupFailoverCriteria {
	failoverCriteria := &awstypes.OriginGroupFailoverCriteria{}
	if v, ok := m["status_codes"]; ok {
		codes := []int32{}
		for _, code := range v.(*schema.Set).List() {
			codes = append(codes, *aws.Int32(int32(code.(int))))
		}
		failoverCriteria.StatusCodes = &awstypes.StatusCodes{
			Items:    codes,
			Quantity: aws.Int32(int32(len(codes))),
		}
	}
	return failoverCriteria
}

func flattenOriginGroupFailoverCriteria(ogfc *awstypes.OriginGroupFailoverCriteria) []interface{} {
	m := make(map[string]interface{})
	if ogfc.StatusCodes.Items != nil {
		l := []interface{}{}
		for _, i := range ogfc.StatusCodes.Items {
			l = append(l, int(*aws.Int32(i)))
		}
		m["status_codes"] = schema.NewSet(schema.HashInt, l)
	}
	return []interface{}{m}
}

func expandMembers(l []interface{}) *awstypes.OriginGroupMembers {
	qty := 0
	items := []awstypes.OriginGroupMember{}
	for _, m := range l {
		ogm := &awstypes.OriginGroupMember{
			OriginId: aws.String(m.(map[string]interface{})["origin_id"].(string)),
		}
		items = append(items, *ogm)
		qty++
	}
	return &awstypes.OriginGroupMembers{
		Quantity: aws.Int32(int32(qty)),
		Items:    items,
	}
}

func flattenOriginGroupMembers(ogm *awstypes.OriginGroupMembers) []interface{} {
	s := []interface{}{}
	for _, i := range ogm.Items {
		m := map[string]interface{}{
			"origin_id": aws.ToString(i.OriginId),
		}
		s = append(s, m)
	}
	return s
}

// Assemble the hash for the aws_awstypes_distribution origin
// TypeSet attribute.
func OriginHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["origin_id"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["domain_name"].(string)))
	if v, ok := m["connection_attempts"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["connection_timeout"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["custom_header"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", customHeadersHash(v.(*schema.Set))))
	}
	if v, ok := m["custom_origin_config"]; ok {
		if s := v.([]interface{}); len(s) > 0 && s[0] != nil {
			buf.WriteString(fmt.Sprintf("%d-", customOriginConfigHash((s[0].(map[string]interface{})))))
		}
	}

	if v, ok := m["origin_access_control_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["origin_path"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["origin_shield"]; ok {
		if s := v.([]interface{}); len(s) > 0 && s[0] != nil {
			buf.WriteString(fmt.Sprintf("%d-", originShieldHash((s[0].(map[string]interface{})))))
		}
	}

	if v, ok := m["s3_origin_config"]; ok {
		if s := v.([]interface{}); len(s) > 0 && s[0] != nil {
			buf.WriteString(fmt.Sprintf("%d-", s3OriginConfigHash((s[0].(map[string]interface{})))))
		}
	}
	return create.StringHashcode(buf.String())
}

// Assemble the hash for the aws_awstypes_distribution origin group
// TypeSet attribute.
func OriginGroupHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["origin_id"].(string)))
	if v, ok := m["failover_criteria"]; ok {
		if l := v.([]interface{}); len(l) > 0 {
			buf.WriteString(fmt.Sprintf("%d-", failoverCriteriaHash(l[0])))
		}
	}
	if v, ok := m["member"]; ok {
		if members := v.([]interface{}); len(members) > 0 {
			for _, member := range members {
				buf.WriteString(fmt.Sprintf("%d-", memberHash(member)))
			}
		}
	}
	return create.StringHashcode(buf.String())
}

func memberHash(v interface{}) int {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", v.(map[string]interface{})["origin_id"]))
	return create.StringHashcode(buf.String())
}

func failoverCriteriaHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["status_codes"]; ok {
		for _, w := range v.(*schema.Set).List() {
			buf.WriteString(fmt.Sprintf("%d-", w))
		}
	}
	return create.StringHashcode(buf.String())
}

func ExpandCustomHeaders(s *schema.Set) *awstypes.CustomHeaders {
	qty := 0
	items := []awstypes.OriginCustomHeader{}
	for _, v := range s.List() {
		items = append(items, *ExpandOriginCustomHeader(v.(map[string]interface{})))
		qty++
	}
	return &awstypes.CustomHeaders{
		Quantity: aws.Int32(int32(qty)),
		Items:    items,
	}
}

func FlattenCustomHeaders(chs *awstypes.CustomHeaders) *schema.Set {
	s := []interface{}{}
	for _, v := range chs.Items {
		s = append(s, FlattenOriginCustomHeader(&v))
	}
	return schema.NewSet(OriginCustomHeaderHash, s)
}

func ExpandOriginCustomHeader(m map[string]interface{}) *awstypes.OriginCustomHeader {
	return &awstypes.OriginCustomHeader{
		HeaderName:  aws.String(m["name"].(string)),
		HeaderValue: aws.String(m["value"].(string)),
	}
}

func FlattenOriginCustomHeader(och *awstypes.OriginCustomHeader) map[string]interface{} {
	return map[string]interface{}{
		"name":  aws.ToString(och.HeaderName),
		"value": aws.ToString(och.HeaderValue),
	}
}

// Helper function used by OriginHash to get a composite hash for all
// aws_awstypes_distribution custom_header attributes.
func customHeadersHash(s *schema.Set) int {
	var buf bytes.Buffer
	for _, v := range s.List() {
		buf.WriteString(fmt.Sprintf("%d-", OriginCustomHeaderHash(v)))
	}
	return create.StringHashcode(buf.String())
}

// Assemble the hash for the aws_awstypes_distribution custom_header
// TypeSet attribute.
func OriginCustomHeaderHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["value"].(string)))
	return create.StringHashcode(buf.String())
}

func ExpandCustomOriginConfig(m map[string]interface{}) *awstypes.CustomOriginConfig {
	customOrigin := &awstypes.CustomOriginConfig{
		OriginProtocolPolicy:   awstypes.OriginProtocolPolicy(m["origin_protocol_policy"].(string)),
		HTTPPort:               aws.Int32(int32(m["http_port"].(int))),
		HTTPSPort:              aws.Int32(int32(m["https_port"].(int))),
		OriginSslProtocols:     ExpandCustomOriginConfigSSL(m["origin_ssl_protocols"].(*schema.Set).List()),
		OriginReadTimeout:      aws.Int32(int32(m["origin_read_timeout"].(int))),
		OriginKeepaliveTimeout: aws.Int32(int32(m["origin_keepalive_timeout"].(int))),
	}

	return customOrigin
}

func FlattenCustomOriginConfig(cor *awstypes.CustomOriginConfig) map[string]interface{} {
	customOrigin := map[string]interface{}{
		"origin_protocol_policy":   awstypes.OriginProtocolPolicy(cor.OriginProtocolPolicy),
		"http_port":                int(*aws.Int32(*cor.HTTPPort)),
		"https_port":               int(*aws.Int32(*cor.HTTPSPort)),
		"origin_ssl_protocols":     FlattenCustomOriginConfigSSL(cor.OriginSslProtocols),
		"origin_read_timeout":      int(*aws.Int32(*cor.OriginReadTimeout)),
		"origin_keepalive_timeout": int(*aws.Int32(*cor.OriginKeepaliveTimeout)),
	}

	return customOrigin
}

// Assemble the hash for the aws_awstypes_distribution custom_origin_config
// TypeSet attribute.
func customOriginConfigHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["origin_protocol_policy"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["http_port"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["https_port"].(int)))
	for _, v := range sortInterfaceSlice(m["origin_ssl_protocols"].(*schema.Set).List()) {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	buf.WriteString(fmt.Sprintf("%d-", m["origin_keepalive_timeout"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["origin_read_timeout"].(int)))

	return create.StringHashcode(buf.String())
}

func ExpandCustomOriginConfigSSL(s []interface{}) *awstypes.OriginSslProtocols {
	items := flex.ExpandStringList(s)
	ospItems := make([]awstypes.SslProtocol, len(items))
	for _, v := range items {
		ospItems = append(ospItems, awstypes.SslProtocol(*v))
	}
	return &awstypes.OriginSslProtocols{
		Quantity: aws.Int32(int32(len(items))),
		Items:    ospItems,
	}
}

func FlattenCustomOriginConfigSSL(osp *awstypes.OriginSslProtocols) *schema.Set {
	items := []*string{}
	for _, v := range osp.Items {
		items = append(items, aws.String(string(v)))
	}
	return flex.FlattenStringSet(items)
}

func ExpandS3OriginConfig(m map[string]interface{}) *awstypes.S3OriginConfig {
	return &awstypes.S3OriginConfig{
		OriginAccessIdentity: aws.String(m["origin_access_identity"].(string)),
	}
}

func ExpandOriginShield(m map[string]interface{}) *awstypes.OriginShield {
	return &awstypes.OriginShield{
		Enabled:            aws.Bool(m["enabled"].(bool)),
		OriginShieldRegion: aws.String(m["origin_shield_region"].(string)),
	}
}

func FlattenS3OriginConfig(s3o *awstypes.S3OriginConfig) map[string]interface{} {
	return map[string]interface{}{
		"origin_access_identity": aws.ToString(s3o.OriginAccessIdentity),
	}
}

func FlattenOriginShield(o *awstypes.OriginShield) map[string]interface{} {
	return map[string]interface{}{
		"origin_shield_region": aws.ToString(o.OriginShieldRegion),
		"enabled":              aws.Bool(*o.Enabled),
	}
}

// Assemble the hash for the aws_awstypes_distribution s3_origin_config
// TypeSet attribute.
func s3OriginConfigHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["origin_access_identity"].(string)))
	return create.StringHashcode(buf.String())
}

func originShieldHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["enabled"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["origin_shield_region"].(string)))
	return create.StringHashcode(buf.String())
}

func ExpandCustomErrorResponses(s *schema.Set) *awstypes.CustomErrorResponses {
	qty := 0
	items := []awstypes.CustomErrorResponse{}
	for _, v := range s.List() {
		items = append(items, *ExpandCustomErrorResponse(v.(map[string]interface{})))
		qty++
	}
	return &awstypes.CustomErrorResponses{
		Quantity: aws.Int32(int32(qty)),
		Items:    items,
	}
}

func FlattenCustomErrorResponses(ers *awstypes.CustomErrorResponses) *schema.Set {
	s := []interface{}{}
	for _, v := range ers.Items {
		s = append(s, FlattenCustomErrorResponse(&v))
	}
	return schema.NewSet(CustomErrorResponseHash, s)
}

func ExpandCustomErrorResponse(m map[string]interface{}) *awstypes.CustomErrorResponse {
	er := awstypes.CustomErrorResponse{
		ErrorCode: aws.Int32(int32(m["error_code"].(int))),
	}
	if v, ok := m["error_caching_min_ttl"]; ok {
		er.ErrorCachingMinTTL = aws.Int64(int64(v.(int)))
	}
	if v, ok := m["response_code"]; ok && v.(int) != 0 {
		er.ResponseCode = aws.String(strconv.Itoa(v.(int)))
	} else {
		er.ResponseCode = aws.String("")
	}
	if v, ok := m["response_page_path"]; ok {
		er.ResponsePagePath = aws.String(v.(string))
	}

	return &er
}

func FlattenCustomErrorResponse(er *awstypes.CustomErrorResponse) map[string]interface{} {
	m := make(map[string]interface{})
	m["error_code"] = int(*aws.Int32(*er.ErrorCode))
	if er.ErrorCachingMinTTL != nil {
		m["error_caching_min_ttl"] = int64(*aws.Int64(*er.ErrorCachingMinTTL))
	}
	if er.ResponseCode != nil {
		m["response_code"], _ = strconv.Atoi(aws.ToString(er.ResponseCode))
	}
	if er.ResponsePagePath != nil {
		m["response_page_path"] = aws.ToString(er.ResponsePagePath)
	}
	return m
}

// Assemble the hash for the aws_awstypes_distribution custom_error_response
// TypeSet attribute.
func CustomErrorResponseHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["error_code"].(int)))
	if v, ok := m["error_caching_min_ttl"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["response_code"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["response_page_path"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	return create.StringHashcode(buf.String())
}

func ExpandLoggingConfig(m map[string]interface{}) *awstypes.LoggingConfig {
	var lc awstypes.LoggingConfig
	if m != nil {
		lc.Prefix = aws.String(m["prefix"].(string))
		lc.Bucket = aws.String(m["bucket"].(string))
		lc.IncludeCookies = aws.Bool(m["include_cookies"].(bool))
		lc.Enabled = aws.Bool(true)
	} else {
		lc.Prefix = aws.String("")
		lc.Bucket = aws.String("")
		lc.IncludeCookies = aws.Bool(false)
		lc.Enabled = aws.Bool(false)
	}
	return &lc
}

func flattenLoggingConfig(lc *awstypes.LoggingConfig) []interface{} {
	m := map[string]interface{}{
		"bucket":          aws.ToString(lc.Bucket),
		"include_cookies": aws.ToBool(lc.IncludeCookies),
		"prefix":          aws.ToString(lc.Prefix),
	}

	return []interface{}{m}
}

func ExpandAliases(s *schema.Set) *awstypes.Aliases {
	aliases := awstypes.Aliases{
		Quantity: aws.Int32(int32(s.Len())),
	}
	if s.Len() > 0 {
		aliases.Items = flex.ExpandStringValueSet(s)
	}
	return &aliases
}

func FlattenAliases(aliases *awstypes.Aliases) *schema.Set {
	if aliases.Items != nil {
		return flex.FlattenStringValueSet(aliases.Items)
	}
	return schema.NewSet(AliasesHash, []interface{}{})
}

// Assemble the hash for the aws_cloudfront_distribution aliases
// TypeSet attribute.
func AliasesHash(v interface{}) int {
	return create.StringHashcode(v.(string))
}

func ExpandRestrictions(m map[string]interface{}) *awstypes.Restrictions {
	return &awstypes.Restrictions{
		GeoRestriction: ExpandGeoRestriction(m["geo_restriction"].([]interface{})[0].(map[string]interface{})),
	}
}

func flattenRestrictions(r *awstypes.Restrictions) []interface{} {
	m := map[string]interface{}{
		"geo_restriction": []interface{}{FlattenGeoRestriction(r.GeoRestriction)},
	}

	return []interface{}{m}
}

func ExpandGeoRestriction(m map[string]interface{}) *awstypes.GeoRestriction {
	gr := &awstypes.GeoRestriction{
		Quantity:        aws.Int32(0),
		RestrictionType: awstypes.GeoRestrictionType(m["restriction_type"].(string)),
	}

	if v, ok := m["locations"]; ok {
		gr.Items = flex.ExpandStringValueSet(v.(*schema.Set))
		gr.Quantity = aws.Int32(int32(v.(*schema.Set).Len()))
	}

	return gr
}

func FlattenGeoRestriction(gr *awstypes.GeoRestriction) map[string]interface{} {
	m := make(map[string]interface{})

	m["restriction_type"] = gr.RestrictionType
	if gr.Items != nil {
		m["locations"] = flex.FlattenStringValueSet(gr.Items)
	}
	return m
}

func ExpandViewerCertificate(m map[string]interface{}) *awstypes.ViewerCertificate {
	var vc awstypes.ViewerCertificate
	if v, ok := m["iam_certificate_id"]; ok && v != "" {
		vc.IAMCertificateId = aws.String(v.(string))
		vc.SSLSupportMethod = awstypes.SSLSupportMethod(m["ssl_support_method"].(string))
	} else if v, ok := m["acm_certificate_arn"]; ok && v != "" {
		vc.ACMCertificateArn = aws.String(v.(string))
		vc.SSLSupportMethod = awstypes.SSLSupportMethod(m["ssl_support_method"].(string))
	} else {
		vc.CloudFrontDefaultCertificate = aws.Bool(m["cloudfront_default_certificate"].(bool))
	}
	if v, ok := m["minimum_protocol_version"]; ok && v != "" {
		vc.MinimumProtocolVersion = awstypes.MinimumProtocolVersion(v.(string))
	}
	return &vc
}

func flattenViewerCertificate(vc *awstypes.ViewerCertificate) []interface{} {
	m := make(map[string]interface{})

	if vc.IAMCertificateId != nil {
		m["iam_certificate_id"] = aws.ToString(vc.IAMCertificateId)
		m["ssl_support_method"] = awstypes.SSLSupportMethod(vc.SSLSupportMethod)
	}
	if vc.ACMCertificateArn != nil {
		m["acm_certificate_arn"] = aws.ToString(vc.ACMCertificateArn)
		m["ssl_support_method"] = awstypes.SSLSupportMethod(vc.SSLSupportMethod)
	}
	if vc.CloudFrontDefaultCertificate != nil {
		m["cloudfront_default_certificate"] = aws.ToBool(vc.CloudFrontDefaultCertificate)
	}
	if vc.MinimumProtocolVersion != awstypes.MinimumProtocolVersion("") {
		m["minimum_protocol_version"] = awstypes.MinimumProtocolVersion(vc.MinimumProtocolVersion)
	}
	return []interface{}{m}
}

func flattenActiveTrustedKeyGroups(atkg *awstypes.ActiveTrustedKeyGroups) []interface{} {
	if atkg == nil {
		return []interface{}{}
	}

	items := make([]*awstypes.KGKeyPairIds, 0, len(atkg.Items))
	for _, v := range atkg.Items {
		items = append(items, &v)
	}
	m := map[string]interface{}{
		"enabled": aws.ToBool(atkg.Enabled),
		"items":   flattenKGKeyPairIds(items),
	}

	return []interface{}{m}
}

func flattenKGKeyPairIds(keyPairIds []*awstypes.KGKeyPairIds) []interface{} {
	result := make([]interface{}, 0, len(keyPairIds))

	for _, keyPairId := range keyPairIds {
		m := map[string]interface{}{
			"key_group_id": aws.ToString(keyPairId.KeyGroupId),
			"key_pair_ids": aws.StringSlice(keyPairId.KeyPairIds.Items),
		}

		result = append(result, m)
	}

	return result
}

func flattenActiveTrustedSigners(ats *awstypes.ActiveTrustedSigners) []interface{} {
	if ats == nil {
		return []interface{}{}
	}
	items := make([]*awstypes.Signer, 0, len(ats.Items))
	for _, v := range ats.Items {
		items = append(items, &v)
	}
	m := map[string]interface{}{
		"enabled": aws.ToBool(ats.Enabled),
		"items":   flattenSigners(items),
	}

	return []interface{}{m}
}

func flattenSigners(signers []*awstypes.Signer) []interface{} {
	result := make([]interface{}, 0, len(signers))

	for _, signer := range signers {
		m := map[string]interface{}{
			"aws_account_number": aws.ToString(signer.AwsAccountNumber),
			"key_pair_ids":       aws.StringSlice(signer.KeyPairIds.Items),
		}

		result = append(result, m)
	}

	return result
}
