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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// cloudFrontRoute53ZoneID defines the route 53 zone ID for CloudFront. This
// is used to set the zone_id attribute.
const cloudFrontRoute53ZoneID = "Z2FDTNDATAQYW2"

// cloudFrontCNRoute53ZoneID defines the route 53 zone ID for CloudFront in AWS CN.
// This is used to set the zone_id attribute.
// ref: https://docs.amazonaws.cn/en_us/aws/latest/userguide/route53.html
const cloudFrontCNRoute53ZoneID = "Z3RFFRIM2A3IF5"

// Assemble the *cloudfront.DistributionConfig variable. Calls out to various
// expander functions to convert attributes and sub-attributes to the various
// complex structures which are necessary to properly build the
// DistributionConfig structure.
//
// Used by the aws_cloudfront_distribution Create and Update functions.
func expandDistributionConfig(d *schema.ResourceData) *cloudfront.DistributionConfig {
	distributionConfig := &cloudfront.DistributionConfig{
		CacheBehaviors:       expandCacheBehaviors(d.Get("ordered_cache_behavior").([]interface{})),
		CallerReference:      aws.String(resource.UniqueId()),
		Comment:              aws.String(d.Get("comment").(string)),
		CustomErrorResponses: ExpandCustomErrorResponses(d.Get("custom_error_response").(*schema.Set)),
		DefaultCacheBehavior: ExpandDefaultCacheBehavior(d.Get("default_cache_behavior").([]interface{})[0].(map[string]interface{})),
		DefaultRootObject:    aws.String(d.Get("default_root_object").(string)),
		Enabled:              aws.Bool(d.Get("enabled").(bool)),
		IsIPV6Enabled:        aws.Bool(d.Get("is_ipv6_enabled").(bool)),
		HttpVersion:          aws.String(d.Get("http_version").(string)),
		Origins:              ExpandOrigins(d.Get("origin").(*schema.Set)),
		PriceClass:           aws.String(d.Get("price_class").(string)),
		WebACLId:             aws.String(d.Get("web_acl_id").(string)),
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

// Unpack the *cloudfront.DistributionConfig variable and set resource data.
// Calls out to flatten functions to convert the DistributionConfig
// sub-structures to their respective attributes in the
// aws_cloudfront_distribution resource.
//
// Used by the aws_cloudfront_distribution Read function.
func flattenDistributionConfig(d *schema.ResourceData, distributionConfig *cloudfront.DistributionConfig) error {
	var err error

	d.Set("enabled", distributionConfig.Enabled)
	d.Set("is_ipv6_enabled", distributionConfig.IsIPV6Enabled)
	d.Set("price_class", distributionConfig.PriceClass)
	d.Set("hosted_zone_id", cloudFrontRoute53ZoneID)

	err = d.Set("default_cache_behavior", []interface{}{flattenDefaultCacheBehavior(distributionConfig.DefaultCacheBehavior)})
	if err != nil {
		return err
	}
	err = d.Set("viewer_certificate", flattenViewerCertificate(distributionConfig.ViewerCertificate))
	if err != nil {
		return err
	}

	if distributionConfig.CallerReference != nil {
		d.Set("caller_reference", distributionConfig.CallerReference)
	}
	if distributionConfig.Comment != nil {
		if aws.StringValue(distributionConfig.Comment) != "" {
			d.Set("comment", distributionConfig.Comment)
		}
	}
	if distributionConfig.DefaultRootObject != nil {
		d.Set("default_root_object", distributionConfig.DefaultRootObject)
	}
	if distributionConfig.HttpVersion != nil {
		d.Set("http_version", distributionConfig.HttpVersion)
	}
	if distributionConfig.WebACLId != nil {
		d.Set("web_acl_id", distributionConfig.WebACLId)
	}

	if distributionConfig.CustomErrorResponses != nil {
		err = d.Set("custom_error_response", FlattenCustomErrorResponses(distributionConfig.CustomErrorResponses))
		if err != nil {
			return err
		}
	}
	if distributionConfig.CacheBehaviors != nil {
		if err := d.Set("ordered_cache_behavior", flattenCacheBehaviors(distributionConfig.CacheBehaviors)); err != nil {
			return err
		}
	}

	if distributionConfig.Logging != nil && *distributionConfig.Logging.Enabled {
		err = d.Set("logging_config", flattenLoggingConfig(distributionConfig.Logging))
	} else {
		err = d.Set("logging_config", []interface{}{})
	}
	if err != nil {
		return err
	}

	if distributionConfig.Aliases != nil {
		err = d.Set("aliases", FlattenAliases(distributionConfig.Aliases))
		if err != nil {
			return err
		}
	}
	if distributionConfig.Restrictions != nil {
		err = d.Set("restrictions", flattenRestrictions(distributionConfig.Restrictions))
		if err != nil {
			return err
		}
	}
	if aws.Int64Value(distributionConfig.Origins.Quantity) > 0 {
		err = d.Set("origin", FlattenOrigins(distributionConfig.Origins))
		if err != nil {
			return err
		}
	}
	if aws.Int64Value(distributionConfig.OriginGroups.Quantity) > 0 {
		err = d.Set("origin_group", FlattenOriginGroups(distributionConfig.OriginGroups))
		if err != nil {
			return err
		}
	}

	return nil
}

func expandCacheBehaviors(lst []interface{}) *cloudfront.CacheBehaviors {
	var qty int64
	var items []*cloudfront.CacheBehavior
	for _, v := range lst {
		items = append(items, expandCacheBehavior(v.(map[string]interface{})))
		qty++
	}
	return &cloudfront.CacheBehaviors{
		Quantity: aws.Int64(qty),
		Items:    items,
	}
}

func flattenCacheBehaviors(cbs *cloudfront.CacheBehaviors) []interface{} {
	lst := []interface{}{}
	for _, v := range cbs.Items {
		lst = append(lst, flattenCacheBehavior(v))
	}
	return lst
}

func ExpandDefaultCacheBehavior(m map[string]interface{}) *cloudfront.DefaultCacheBehavior {
	dcb := &cloudfront.DefaultCacheBehavior{
		CachePolicyId:           aws.String(m["cache_policy_id"].(string)),
		Compress:                aws.Bool(m["compress"].(bool)),
		FieldLevelEncryptionId:  aws.String(m["field_level_encryption_id"].(string)),
		OriginRequestPolicyId:   aws.String(m["origin_request_policy_id"].(string)),
		ResponseHeadersPolicyId: aws.String(m["response_headers_policy_id"].(string)),
		TargetOriginId:          aws.String(m["target_origin_id"].(string)),
		ViewerProtocolPolicy:    aws.String(m["viewer_protocol_policy"].(string)),
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

func expandCacheBehavior(m map[string]interface{}) *cloudfront.CacheBehavior {
	var forwardedValues *cloudfront.ForwardedValues
	if forwardedValuesFlat, ok := m["forwarded_values"].([]interface{}); ok && len(forwardedValuesFlat) == 1 {
		forwardedValues = ExpandForwardedValues(m["forwarded_values"].([]interface{})[0].(map[string]interface{}))
	}

	cb := &cloudfront.CacheBehavior{
		CachePolicyId:           aws.String(m["cache_policy_id"].(string)),
		Compress:                aws.Bool(m["compress"].(bool)),
		FieldLevelEncryptionId:  aws.String(m["field_level_encryption_id"].(string)),
		ForwardedValues:         forwardedValues,
		OriginRequestPolicyId:   aws.String(m["origin_request_policy_id"].(string)),
		ResponseHeadersPolicyId: aws.String(m["response_headers_policy_id"].(string)),
		TargetOriginId:          aws.String(m["target_origin_id"].(string)),
		ViewerProtocolPolicy:    aws.String(m["viewer_protocol_policy"].(string)),
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

func flattenDefaultCacheBehavior(dcb *cloudfront.DefaultCacheBehavior) map[string]interface{} {
	m := map[string]interface{}{
		"cache_policy_id":            aws.StringValue(dcb.CachePolicyId),
		"compress":                   aws.BoolValue(dcb.Compress),
		"field_level_encryption_id":  aws.StringValue(dcb.FieldLevelEncryptionId),
		"viewer_protocol_policy":     aws.StringValue(dcb.ViewerProtocolPolicy),
		"target_origin_id":           aws.StringValue(dcb.TargetOriginId),
		"min_ttl":                    aws.Int64Value(dcb.MinTTL),
		"origin_request_policy_id":   aws.StringValue(dcb.OriginRequestPolicyId),
		"realtime_log_config_arn":    aws.StringValue(dcb.RealtimeLogConfigArn),
		"response_headers_policy_id": aws.StringValue(dcb.ResponseHeadersPolicyId),
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
		m["max_ttl"] = aws.Int64Value(dcb.MaxTTL)
	}
	if dcb.SmoothStreaming != nil {
		m["smooth_streaming"] = aws.BoolValue(dcb.SmoothStreaming)
	}
	if dcb.DefaultTTL != nil {
		m["default_ttl"] = int(aws.Int64Value(dcb.DefaultTTL))
	}
	if dcb.AllowedMethods != nil {
		m["allowed_methods"] = FlattenAllowedMethods(dcb.AllowedMethods)
	}
	if dcb.AllowedMethods.CachedMethods != nil {
		m["cached_methods"] = FlattenCachedMethods(dcb.AllowedMethods.CachedMethods)
	}

	return m
}

func flattenCacheBehavior(cb *cloudfront.CacheBehavior) map[string]interface{} {
	m := make(map[string]interface{})

	m["cache_policy_id"] = aws.StringValue(cb.CachePolicyId)
	m["compress"] = aws.BoolValue(cb.Compress)
	m["field_level_encryption_id"] = aws.StringValue(cb.FieldLevelEncryptionId)
	m["viewer_protocol_policy"] = aws.StringValue(cb.ViewerProtocolPolicy)
	m["target_origin_id"] = aws.StringValue(cb.TargetOriginId)
	m["min_ttl"] = int(aws.Int64Value(cb.MinTTL))
	m["origin_request_policy_id"] = aws.StringValue(cb.OriginRequestPolicyId)
	m["realtime_log_config_arn"] = aws.StringValue(cb.RealtimeLogConfigArn)
	m["response_headers_policy_id"] = aws.StringValue(cb.ResponseHeadersPolicyId)

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
		m["max_ttl"] = int(aws.Int64Value(cb.MaxTTL))
	}
	if cb.SmoothStreaming != nil {
		m["smooth_streaming"] = aws.BoolValue(cb.SmoothStreaming)
	}
	if cb.DefaultTTL != nil {
		m["default_ttl"] = int(aws.Int64Value(cb.DefaultTTL))
	}
	if cb.AllowedMethods != nil {
		m["allowed_methods"] = FlattenAllowedMethods(cb.AllowedMethods)
	}
	if cb.AllowedMethods.CachedMethods != nil {
		m["cached_methods"] = FlattenCachedMethods(cb.AllowedMethods.CachedMethods)
	}
	if cb.PathPattern != nil {
		m["path_pattern"] = aws.StringValue(cb.PathPattern)
	}
	return m
}

func expandTrustedKeyGroups(s []interface{}) *cloudfront.TrustedKeyGroups {
	var tkg cloudfront.TrustedKeyGroups
	if len(s) > 0 {
		tkg.Quantity = aws.Int64(int64(len(s)))
		tkg.Items = flex.ExpandStringList(s)
		tkg.Enabled = aws.Bool(true)
	} else {
		tkg.Quantity = aws.Int64(0)
		tkg.Enabled = aws.Bool(false)
	}
	return &tkg
}

func flattenTrustedKeyGroups(tkg *cloudfront.TrustedKeyGroups) []interface{} {
	if tkg.Items != nil {
		return flex.FlattenStringList(tkg.Items)
	}
	return []interface{}{}
}

func ExpandTrustedSigners(s []interface{}) *cloudfront.TrustedSigners {
	var ts cloudfront.TrustedSigners
	if len(s) > 0 {
		ts.Quantity = aws.Int64(int64(len(s)))
		ts.Items = flex.ExpandStringList(s)
		ts.Enabled = aws.Bool(true)
	} else {
		ts.Quantity = aws.Int64(0)
		ts.Enabled = aws.Bool(false)
	}
	return &ts
}

func FlattenTrustedSigners(ts *cloudfront.TrustedSigners) []interface{} {
	if ts.Items != nil {
		return flex.FlattenStringList(ts.Items)
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

func ExpandLambdaFunctionAssociations(v interface{}) *cloudfront.LambdaFunctionAssociations {
	if v == nil {
		return &cloudfront.LambdaFunctionAssociations{
			Quantity: aws.Int64(0),
		}
	}

	s := v.([]interface{})
	var lfa cloudfront.LambdaFunctionAssociations
	lfa.Quantity = aws.Int64(int64(len(s)))
	lfa.Items = make([]*cloudfront.LambdaFunctionAssociation, len(s))
	for i, lf := range s {
		lfa.Items[i] = expandLambdaFunctionAssociation(lf.(map[string]interface{}))
	}
	return &lfa
}

func expandLambdaFunctionAssociation(lf map[string]interface{}) *cloudfront.LambdaFunctionAssociation {
	var lfa cloudfront.LambdaFunctionAssociation
	if v, ok := lf["event_type"]; ok {
		lfa.EventType = aws.String(v.(string))
	}
	if v, ok := lf["lambda_arn"]; ok {
		lfa.LambdaFunctionARN = aws.String(v.(string))
	}
	if v, ok := lf["include_body"]; ok {
		lfa.IncludeBody = aws.Bool(v.(bool))
	}
	return &lfa
}

func ExpandFunctionAssociations(v interface{}) *cloudfront.FunctionAssociations {
	if v == nil {
		return &cloudfront.FunctionAssociations{
			Quantity: aws.Int64(0),
		}
	}

	s := v.([]interface{})
	var fa cloudfront.FunctionAssociations
	fa.Quantity = aws.Int64(int64(len(s)))
	fa.Items = make([]*cloudfront.FunctionAssociation, len(s))
	for i, f := range s {
		fa.Items[i] = expandFunctionAssociation(f.(map[string]interface{}))
	}
	return &fa
}

func expandFunctionAssociation(f map[string]interface{}) *cloudfront.FunctionAssociation {
	var fa cloudfront.FunctionAssociation
	if v, ok := f["event_type"]; ok {
		fa.EventType = aws.String(v.(string))
	}
	if v, ok := f["function_arn"]; ok {
		fa.FunctionARN = aws.String(v.(string))
	}
	return &fa
}

func FlattenLambdaFunctionAssociations(lfa *cloudfront.LambdaFunctionAssociations) *schema.Set {
	s := schema.NewSet(LambdaFunctionAssociationHash, []interface{}{})
	for _, v := range lfa.Items {
		s.Add(flattenLambdaFunctionAssociation(v))
	}
	return s
}

func flattenLambdaFunctionAssociation(lfa *cloudfront.LambdaFunctionAssociation) map[string]interface{} {
	m := map[string]interface{}{}
	if lfa != nil {
		m["event_type"] = aws.StringValue(lfa.EventType)
		m["lambda_arn"] = aws.StringValue(lfa.LambdaFunctionARN)
		m["include_body"] = aws.BoolValue(lfa.IncludeBody)
	}
	return m
}

func FlattenFunctionAssociations(fa *cloudfront.FunctionAssociations) *schema.Set {
	s := schema.NewSet(FunctionAssociationHash, []interface{}{})
	for _, v := range fa.Items {
		s.Add(flattenFunctionAssociation(v))
	}
	return s
}

func flattenFunctionAssociation(fa *cloudfront.FunctionAssociation) map[string]interface{} {
	m := map[string]interface{}{}
	if fa != nil {
		m["event_type"] = aws.StringValue(fa.EventType)
		m["function_arn"] = aws.StringValue(fa.FunctionARN)
	}
	return m
}

func ExpandForwardedValues(m map[string]interface{}) *cloudfront.ForwardedValues {
	if len(m) < 1 {
		return nil
	}

	fv := &cloudfront.ForwardedValues{
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

func FlattenForwardedValues(fv *cloudfront.ForwardedValues) map[string]interface{} {
	m := make(map[string]interface{})
	m["query_string"] = aws.BoolValue(fv.QueryString)
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

func ExpandHeaders(d []interface{}) *cloudfront.Headers {
	return &cloudfront.Headers{
		Quantity: aws.Int64(int64(len(d))),
		Items:    flex.ExpandStringList(d),
	}
}

func FlattenHeaders(h *cloudfront.Headers) []interface{} {
	if h.Items != nil {
		return flex.FlattenStringList(h.Items)
	}
	return []interface{}{}
}

func ExpandQueryStringCacheKeys(d []interface{}) *cloudfront.QueryStringCacheKeys {
	return &cloudfront.QueryStringCacheKeys{
		Quantity: aws.Int64(int64(len(d))),
		Items:    flex.ExpandStringList(d),
	}
}

func FlattenQueryStringCacheKeys(k *cloudfront.QueryStringCacheKeys) []interface{} {
	if k.Items != nil {
		return flex.FlattenStringList(k.Items)
	}
	return []interface{}{}
}

func ExpandCookiePreference(m map[string]interface{}) *cloudfront.CookiePreference {
	cp := &cloudfront.CookiePreference{
		Forward: aws.String(m["forward"].(string)),
	}
	if v, ok := m["whitelisted_names"]; ok {
		cp.WhitelistedNames = ExpandCookieNames(v.(*schema.Set).List())
	}
	return cp
}

func FlattenCookiePreference(cp *cloudfront.CookiePreference) map[string]interface{} {
	m := make(map[string]interface{})
	m["forward"] = aws.StringValue(cp.Forward)
	if cp.WhitelistedNames != nil {
		m["whitelisted_names"] = schema.NewSet(schema.HashString, FlattenCookieNames(cp.WhitelistedNames))
	}
	return m
}

func ExpandCookieNames(d []interface{}) *cloudfront.CookieNames {
	return &cloudfront.CookieNames{
		Quantity: aws.Int64(int64(len(d))),
		Items:    flex.ExpandStringList(d),
	}
}

func FlattenCookieNames(cn *cloudfront.CookieNames) []interface{} {
	if cn.Items != nil {
		return flex.FlattenStringList(cn.Items)
	}
	return []interface{}{}
}

func ExpandAllowedMethods(s *schema.Set) *cloudfront.AllowedMethods {
	return &cloudfront.AllowedMethods{
		Quantity: aws.Int64(int64(s.Len())),
		Items:    flex.ExpandStringSet(s),
	}
}

func FlattenAllowedMethods(am *cloudfront.AllowedMethods) *schema.Set {
	if am.Items != nil {
		return flex.FlattenStringSet(am.Items)
	}
	return nil
}

func ExpandCachedMethods(s *schema.Set) *cloudfront.CachedMethods {
	return &cloudfront.CachedMethods{
		Quantity: aws.Int64(int64(s.Len())),
		Items:    flex.ExpandStringSet(s),
	}
}

func FlattenCachedMethods(cm *cloudfront.CachedMethods) *schema.Set {
	if cm.Items != nil {
		return flex.FlattenStringSet(cm.Items)
	}
	return nil
}

func ExpandOrigins(s *schema.Set) *cloudfront.Origins {
	qty := 0
	items := []*cloudfront.Origin{}
	for _, v := range s.List() {
		items = append(items, ExpandOrigin(v.(map[string]interface{})))
		qty++
	}
	return &cloudfront.Origins{
		Quantity: aws.Int64(int64(qty)),
		Items:    items,
	}
}

func FlattenOrigins(ors *cloudfront.Origins) *schema.Set {
	s := []interface{}{}
	for _, v := range ors.Items {
		s = append(s, FlattenOrigin(v))
	}
	return schema.NewSet(OriginHash, s)
}

func ExpandOrigin(m map[string]interface{}) *cloudfront.Origin {
	origin := &cloudfront.Origin{
		Id:         aws.String(m["origin_id"].(string)),
		DomainName: aws.String(m["domain_name"].(string)),
	}

	if v, ok := m["connection_attempts"]; ok {
		origin.ConnectionAttempts = aws.Int64(int64(v.(int)))
	}
	if v, ok := m["connection_timeout"]; ok {
		origin.ConnectionTimeout = aws.Int64(int64(v.(int)))
	}
	if v, ok := m["custom_header"]; ok {
		origin.CustomHeaders = ExpandCustomHeaders(v.(*schema.Set))
	}
	if v, ok := m["custom_origin_config"]; ok {
		if s := v.([]interface{}); len(s) > 0 {
			origin.CustomOriginConfig = ExpandCustomOriginConfig(s[0].(map[string]interface{}))
		}
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
		origin.S3OriginConfig = &cloudfront.S3OriginConfig{
			OriginAccessIdentity: aws.String(""),
		}
	}

	return origin
}

func FlattenOrigin(or *cloudfront.Origin) map[string]interface{} {
	m := make(map[string]interface{})
	m["origin_id"] = aws.StringValue(or.Id)
	m["domain_name"] = aws.StringValue(or.DomainName)
	if or.ConnectionAttempts != nil {
		m["connection_attempts"] = int(aws.Int64Value(or.ConnectionAttempts))
	}
	if or.ConnectionTimeout != nil {
		m["connection_timeout"] = int(aws.Int64Value(or.ConnectionTimeout))
	}
	if or.CustomHeaders != nil {
		m["custom_header"] = FlattenCustomHeaders(or.CustomHeaders)
	}
	if or.CustomOriginConfig != nil {
		m["custom_origin_config"] = []interface{}{FlattenCustomOriginConfig(or.CustomOriginConfig)}
	}
	if or.OriginPath != nil {
		m["origin_path"] = aws.StringValue(or.OriginPath)
	}
	if or.OriginShield != nil && aws.BoolValue(or.OriginShield.Enabled) {
		m["origin_shield"] = []interface{}{FlattenOriginShield(or.OriginShield)}
	}
	if or.S3OriginConfig != nil && aws.StringValue(or.S3OriginConfig.OriginAccessIdentity) != "" {
		m["s3_origin_config"] = []interface{}{FlattenS3OriginConfig(or.S3OriginConfig)}
	}
	return m
}

func ExpandOriginGroups(s *schema.Set) *cloudfront.OriginGroups {
	qty := 0
	items := []*cloudfront.OriginGroup{}
	for _, v := range s.List() {
		items = append(items, expandOriginGroup(v.(map[string]interface{})))
		qty++
	}
	return &cloudfront.OriginGroups{
		Quantity: aws.Int64(int64(qty)),
		Items:    items,
	}
}

func FlattenOriginGroups(ogs *cloudfront.OriginGroups) *schema.Set {
	s := []interface{}{}
	for _, v := range ogs.Items {
		s = append(s, flattenOriginGroup(v))
	}
	return schema.NewSet(OriginGroupHash, s)
}

func expandOriginGroup(m map[string]interface{}) *cloudfront.OriginGroup {
	failoverCriteria := m["failover_criteria"].([]interface{})[0].(map[string]interface{})
	members := m["member"].([]interface{})
	originGroup := &cloudfront.OriginGroup{
		Id:               aws.String(m["origin_id"].(string)),
		FailoverCriteria: expandOriginGroupFailoverCriteria(failoverCriteria),
		Members:          expandMembers(members),
	}
	return originGroup
}

func flattenOriginGroup(og *cloudfront.OriginGroup) map[string]interface{} {
	m := make(map[string]interface{})
	m["origin_id"] = aws.StringValue(og.Id)
	if og.FailoverCriteria != nil {
		m["failover_criteria"] = flattenOriginGroupFailoverCriteria(og.FailoverCriteria)
	}
	if og.Members != nil {
		m["member"] = flattenOriginGroupMembers(og.Members)
	}
	return m
}

func expandOriginGroupFailoverCriteria(m map[string]interface{}) *cloudfront.OriginGroupFailoverCriteria {
	failoverCriteria := &cloudfront.OriginGroupFailoverCriteria{}
	if v, ok := m["status_codes"]; ok {
		codes := []*int64{}
		for _, code := range v.(*schema.Set).List() {
			codes = append(codes, aws.Int64(int64(code.(int))))
		}
		failoverCriteria.StatusCodes = &cloudfront.StatusCodes{
			Items:    codes,
			Quantity: aws.Int64(int64(len(codes))),
		}
	}
	return failoverCriteria
}

func flattenOriginGroupFailoverCriteria(ogfc *cloudfront.OriginGroupFailoverCriteria) []interface{} {
	m := make(map[string]interface{})
	if ogfc.StatusCodes.Items != nil {
		l := []interface{}{}
		for _, i := range ogfc.StatusCodes.Items {
			l = append(l, int(aws.Int64Value(i)))
		}
		m["status_codes"] = schema.NewSet(schema.HashInt, l)
	}
	return []interface{}{m}
}

func expandMembers(l []interface{}) *cloudfront.OriginGroupMembers {
	qty := 0
	items := []*cloudfront.OriginGroupMember{}
	for _, m := range l {
		ogm := &cloudfront.OriginGroupMember{
			OriginId: aws.String(m.(map[string]interface{})["origin_id"].(string)),
		}
		items = append(items, ogm)
		qty++
	}
	return &cloudfront.OriginGroupMembers{
		Quantity: aws.Int64(int64(qty)),
		Items:    items,
	}
}

func flattenOriginGroupMembers(ogm *cloudfront.OriginGroupMembers) []interface{} {
	s := []interface{}{}
	for _, i := range ogm.Items {
		m := map[string]interface{}{
			"origin_id": aws.StringValue(i.OriginId),
		}
		s = append(s, m)
	}
	return s
}

// Assemble the hash for the aws_cloudfront_distribution origin
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

// Assemble the hash for the aws_cloudfront_distribution origin group
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

func ExpandCustomHeaders(s *schema.Set) *cloudfront.CustomHeaders {
	qty := 0
	items := []*cloudfront.OriginCustomHeader{}
	for _, v := range s.List() {
		items = append(items, ExpandOriginCustomHeader(v.(map[string]interface{})))
		qty++
	}
	return &cloudfront.CustomHeaders{
		Quantity: aws.Int64(int64(qty)),
		Items:    items,
	}
}

func FlattenCustomHeaders(chs *cloudfront.CustomHeaders) *schema.Set {
	s := []interface{}{}
	for _, v := range chs.Items {
		s = append(s, FlattenOriginCustomHeader(v))
	}
	return schema.NewSet(OriginCustomHeaderHash, s)
}

func ExpandOriginCustomHeader(m map[string]interface{}) *cloudfront.OriginCustomHeader {
	return &cloudfront.OriginCustomHeader{
		HeaderName:  aws.String(m["name"].(string)),
		HeaderValue: aws.String(m["value"].(string)),
	}
}

func FlattenOriginCustomHeader(och *cloudfront.OriginCustomHeader) map[string]interface{} {
	return map[string]interface{}{
		"name":  aws.StringValue(och.HeaderName),
		"value": aws.StringValue(och.HeaderValue),
	}
}

// Helper function used by OriginHash to get a composite hash for all
// aws_cloudfront_distribution custom_header attributes.
func customHeadersHash(s *schema.Set) int {
	var buf bytes.Buffer
	for _, v := range s.List() {
		buf.WriteString(fmt.Sprintf("%d-", OriginCustomHeaderHash(v)))
	}
	return create.StringHashcode(buf.String())
}

// Assemble the hash for the aws_cloudfront_distribution custom_header
// TypeSet attribute.
func OriginCustomHeaderHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["value"].(string)))
	return create.StringHashcode(buf.String())
}

func ExpandCustomOriginConfig(m map[string]interface{}) *cloudfront.CustomOriginConfig {

	customOrigin := &cloudfront.CustomOriginConfig{
		OriginProtocolPolicy:   aws.String(m["origin_protocol_policy"].(string)),
		HTTPPort:               aws.Int64(int64(m["http_port"].(int))),
		HTTPSPort:              aws.Int64(int64(m["https_port"].(int))),
		OriginSslProtocols:     ExpandCustomOriginConfigSSL(m["origin_ssl_protocols"].(*schema.Set).List()),
		OriginReadTimeout:      aws.Int64(int64(m["origin_read_timeout"].(int))),
		OriginKeepaliveTimeout: aws.Int64(int64(m["origin_keepalive_timeout"].(int))),
	}

	return customOrigin
}

func FlattenCustomOriginConfig(cor *cloudfront.CustomOriginConfig) map[string]interface{} {

	customOrigin := map[string]interface{}{
		"origin_protocol_policy":   aws.StringValue(cor.OriginProtocolPolicy),
		"http_port":                int(aws.Int64Value(cor.HTTPPort)),
		"https_port":               int(aws.Int64Value(cor.HTTPSPort)),
		"origin_ssl_protocols":     FlattenCustomOriginConfigSSL(cor.OriginSslProtocols),
		"origin_read_timeout":      int(aws.Int64Value(cor.OriginReadTimeout)),
		"origin_keepalive_timeout": int(aws.Int64Value(cor.OriginKeepaliveTimeout)),
	}

	return customOrigin
}

// Assemble the hash for the aws_cloudfront_distribution custom_origin_config
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

func ExpandCustomOriginConfigSSL(s []interface{}) *cloudfront.OriginSslProtocols {
	items := flex.ExpandStringList(s)
	return &cloudfront.OriginSslProtocols{
		Quantity: aws.Int64(int64(len(items))),
		Items:    items,
	}
}

func FlattenCustomOriginConfigSSL(osp *cloudfront.OriginSslProtocols) *schema.Set {
	return flex.FlattenStringSet(osp.Items)
}

func ExpandS3OriginConfig(m map[string]interface{}) *cloudfront.S3OriginConfig {
	return &cloudfront.S3OriginConfig{
		OriginAccessIdentity: aws.String(m["origin_access_identity"].(string)),
	}
}

func ExpandOriginShield(m map[string]interface{}) *cloudfront.OriginShield {
	return &cloudfront.OriginShield{
		Enabled:            aws.Bool(m["enabled"].(bool)),
		OriginShieldRegion: aws.String(m["origin_shield_region"].(string)),
	}
}

func FlattenS3OriginConfig(s3o *cloudfront.S3OriginConfig) map[string]interface{} {
	return map[string]interface{}{
		"origin_access_identity": aws.StringValue(s3o.OriginAccessIdentity),
	}
}

func FlattenOriginShield(o *cloudfront.OriginShield) map[string]interface{} {
	return map[string]interface{}{
		"origin_shield_region": aws.StringValue(o.OriginShieldRegion),
		"enabled":              aws.BoolValue(o.Enabled),
	}
}

// Assemble the hash for the aws_cloudfront_distribution s3_origin_config
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

func ExpandCustomErrorResponses(s *schema.Set) *cloudfront.CustomErrorResponses {
	qty := 0
	items := []*cloudfront.CustomErrorResponse{}
	for _, v := range s.List() {
		items = append(items, ExpandCustomErrorResponse(v.(map[string]interface{})))
		qty++
	}
	return &cloudfront.CustomErrorResponses{
		Quantity: aws.Int64(int64(qty)),
		Items:    items,
	}
}

func FlattenCustomErrorResponses(ers *cloudfront.CustomErrorResponses) *schema.Set {
	s := []interface{}{}
	for _, v := range ers.Items {
		s = append(s, FlattenCustomErrorResponse(v))
	}
	return schema.NewSet(CustomErrorResponseHash, s)
}

func ExpandCustomErrorResponse(m map[string]interface{}) *cloudfront.CustomErrorResponse {
	er := cloudfront.CustomErrorResponse{
		ErrorCode: aws.Int64(int64(m["error_code"].(int))),
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

func FlattenCustomErrorResponse(er *cloudfront.CustomErrorResponse) map[string]interface{} {
	m := make(map[string]interface{})
	m["error_code"] = int(aws.Int64Value(er.ErrorCode))
	if er.ErrorCachingMinTTL != nil {
		m["error_caching_min_ttl"] = int(aws.Int64Value(er.ErrorCachingMinTTL))
	}
	if er.ResponseCode != nil {
		m["response_code"], _ = strconv.Atoi(aws.StringValue(er.ResponseCode))
	}
	if er.ResponsePagePath != nil {
		m["response_page_path"] = aws.StringValue(er.ResponsePagePath)
	}
	return m
}

// Assemble the hash for the aws_cloudfront_distribution custom_error_response
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

func ExpandLoggingConfig(m map[string]interface{}) *cloudfront.LoggingConfig {
	var lc cloudfront.LoggingConfig
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

func flattenLoggingConfig(lc *cloudfront.LoggingConfig) []interface{} {
	m := map[string]interface{}{
		"bucket":          aws.StringValue(lc.Bucket),
		"include_cookies": aws.BoolValue(lc.IncludeCookies),
		"prefix":          aws.StringValue(lc.Prefix),
	}

	return []interface{}{m}
}

func ExpandAliases(s *schema.Set) *cloudfront.Aliases {
	aliases := cloudfront.Aliases{
		Quantity: aws.Int64(int64(s.Len())),
	}
	if s.Len() > 0 {
		aliases.Items = flex.ExpandStringSet(s)
	}
	return &aliases
}

func FlattenAliases(aliases *cloudfront.Aliases) *schema.Set {
	if aliases.Items != nil {
		return flex.FlattenStringSet(aliases.Items)
	}
	return schema.NewSet(AliasesHash, []interface{}{})
}

// Assemble the hash for the aws_cloudfront_distribution aliases
// TypeSet attribute.
func AliasesHash(v interface{}) int {
	return create.StringHashcode(v.(string))
}

func ExpandRestrictions(m map[string]interface{}) *cloudfront.Restrictions {
	return &cloudfront.Restrictions{
		GeoRestriction: ExpandGeoRestriction(m["geo_restriction"].([]interface{})[0].(map[string]interface{})),
	}
}

func flattenRestrictions(r *cloudfront.Restrictions) []interface{} {
	m := map[string]interface{}{
		"geo_restriction": []interface{}{FlattenGeoRestriction(r.GeoRestriction)},
	}

	return []interface{}{m}
}

func ExpandGeoRestriction(m map[string]interface{}) *cloudfront.GeoRestriction {
	gr := &cloudfront.GeoRestriction{
		Quantity:        aws.Int64(0),
		RestrictionType: aws.String(m["restriction_type"].(string)),
	}

	if v, ok := m["locations"]; ok {
		gr.Items = flex.ExpandStringSet(v.(*schema.Set))
		gr.Quantity = aws.Int64(int64(v.(*schema.Set).Len()))
	}

	return gr
}

func FlattenGeoRestriction(gr *cloudfront.GeoRestriction) map[string]interface{} {
	m := make(map[string]interface{})

	m["restriction_type"] = aws.StringValue(gr.RestrictionType)
	if gr.Items != nil {
		m["locations"] = flex.FlattenStringSet(gr.Items)
	}
	return m
}

func ExpandViewerCertificate(m map[string]interface{}) *cloudfront.ViewerCertificate {
	var vc cloudfront.ViewerCertificate
	if v, ok := m["iam_certificate_id"]; ok && v != "" {
		vc.IAMCertificateId = aws.String(v.(string))
		vc.SSLSupportMethod = aws.String(m["ssl_support_method"].(string))
	} else if v, ok := m["acm_certificate_arn"]; ok && v != "" {
		vc.ACMCertificateArn = aws.String(v.(string))
		vc.SSLSupportMethod = aws.String(m["ssl_support_method"].(string))
	} else {
		vc.CloudFrontDefaultCertificate = aws.Bool(m["cloudfront_default_certificate"].(bool))
	}
	if v, ok := m["minimum_protocol_version"]; ok && v != "" {
		vc.MinimumProtocolVersion = aws.String(v.(string))
	}
	return &vc
}

func flattenViewerCertificate(vc *cloudfront.ViewerCertificate) []interface{} {
	m := make(map[string]interface{})

	if vc.IAMCertificateId != nil {
		m["iam_certificate_id"] = aws.StringValue(vc.IAMCertificateId)
		m["ssl_support_method"] = aws.StringValue(vc.SSLSupportMethod)
	}
	if vc.ACMCertificateArn != nil {
		m["acm_certificate_arn"] = aws.StringValue(vc.ACMCertificateArn)
		m["ssl_support_method"] = aws.StringValue(vc.SSLSupportMethod)
	}
	if vc.CloudFrontDefaultCertificate != nil {
		m["cloudfront_default_certificate"] = aws.BoolValue(vc.CloudFrontDefaultCertificate)
	}
	if vc.MinimumProtocolVersion != nil {
		m["minimum_protocol_version"] = aws.StringValue(vc.MinimumProtocolVersion)
	}
	return []interface{}{m}
}

func flattenCloudfrontActiveTrustedKeyGroups(atkg *cloudfront.ActiveTrustedKeyGroups) []interface{} {
	if atkg == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"enabled": aws.BoolValue(atkg.Enabled),
		"items":   flattenCloudfrontKGKeyPairIds(atkg.Items),
	}

	return []interface{}{m}
}

func flattenCloudfrontKGKeyPairIds(keyPairIds []*cloudfront.KGKeyPairIds) []interface{} {
	result := make([]interface{}, 0, len(keyPairIds))

	for _, keyPairId := range keyPairIds {
		m := map[string]interface{}{
			"key_group_id": aws.StringValue(keyPairId.KeyGroupId),
			"key_pair_ids": aws.StringValueSlice(keyPairId.KeyPairIds.Items),
		}

		result = append(result, m)
	}

	return result
}

func flattenCloudfrontActiveTrustedSigners(ats *cloudfront.ActiveTrustedSigners) []interface{} {
	if ats == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"enabled": aws.BoolValue(ats.Enabled),
		"items":   flattenCloudfrontSigners(ats.Items),
	}

	return []interface{}{m}
}

func flattenCloudfrontSigners(signers []*cloudfront.Signer) []interface{} {
	result := make([]interface{}, 0, len(signers))

	for _, signer := range signers {
		m := map[string]interface{}{
			"aws_account_number": aws.StringValue(signer.AwsAccountNumber),
			"key_pair_ids":       aws.StringValueSlice(signer.KeyPairIds.Items),
		}

		result = append(result, m)
	}

	return result
}
