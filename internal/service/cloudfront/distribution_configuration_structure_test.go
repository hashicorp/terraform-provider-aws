package cloudfront_test

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
)

func defaultCacheBehaviorConf() map[string]interface{} {
	return map[string]interface{}{
		"viewer_protocol_policy":      "allow-all",
		"cache_policy_id":             "",
		"target_origin_id":            "myS3Origin",
		"forwarded_values":            []interface{}{forwardedValuesConf()},
		"min_ttl":                     0,
		"trusted_signers":             trustedSignersConf(),
		"lambda_function_association": lambdaFunctionAssociationsConf(),
		"function_association":        functionAssociationsConf(),
		"max_ttl":                     31536000,
		"smooth_streaming":            false,
		"default_ttl":                 86400,
		"allowed_methods":             allowedMethodsConf(),
		"origin_request_policy_id":    "ABCD1234",
		"cached_methods":              cachedMethodsConf(),
		"compress":                    true,
		"field_level_encryption_id":   "",
		"realtime_log_config_arn":     "",
		"response_headers_policy_id":  "",
	}
}

func trustedSignersConf() []interface{} {
	return []interface{}{"1234567890EX", "1234567891EX"}
}

func lambdaFunctionAssociationsConf() *schema.Set {
	x := []interface{}{
		map[string]interface{}{
			"event_type":   "viewer-request",
			"lambda_arn":   "arn:aws:lambda:us-east-1:999999999:function1:alias", //lintignore:AWSAT003,AWSAT005
			"include_body": true,
		},
		map[string]interface{}{
			"event_type":   "origin-response",
			"lambda_arn":   "arn:aws:lambda:us-east-1:999999999:function2:alias", //lintignore:AWSAT003,AWSAT005
			"include_body": true,
		},
	}

	return schema.NewSet(tfcloudfront.LambdaFunctionAssociationHash, x)
}

func functionAssociationsConf() *schema.Set {
	x := []interface{}{
		map[string]interface{}{
			"event_type":   "viewer-request",
			"function_arn": "arn:aws:cloudfront::999999999:function/function1", //lintignore:AWSAT003,AWSAT005
		},
		map[string]interface{}{
			"event_type":   "viewer-response",
			"function_arn": "arn:aws:cloudfront::999999999:function/function2", //lintignore:AWSAT003,AWSAT005
		},
	}

	return schema.NewSet(tfcloudfront.FunctionAssociationHash, x)
}

func forwardedValuesConf() map[string]interface{} {
	return map[string]interface{}{
		"query_string":            true,
		"query_string_cache_keys": queryStringCacheKeysConf(),
		"cookies":                 []interface{}{cookiePreferenceConf()},
		"headers":                 headersConf(),
	}
}

func headersConf() *schema.Set {
	return schema.NewSet(schema.HashString, []interface{}{"X-Example1", "X-Example2"})
}

func queryStringCacheKeysConf() []interface{} {
	return []interface{}{"foo", "bar"}
}

func cookiePreferenceConf() map[string]interface{} {
	return map[string]interface{}{
		"forward":           "whitelist",
		"whitelisted_names": cookieNamesConf(),
	}
}

func cookieNamesConf() *schema.Set {
	return schema.NewSet(schema.HashString, []interface{}{"Example1", "Example2"})
}

func allowedMethodsConf() *schema.Set {
	return schema.NewSet(schema.HashString, []interface{}{"DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"})
}

func cachedMethodsConf() *schema.Set {
	return schema.NewSet(schema.HashString, []interface{}{"GET", "HEAD", "OPTIONS"})
}

func originCustomHeadersConf() *schema.Set {
	return schema.NewSet(tfcloudfront.OriginCustomHeaderHash, []interface{}{originCustomHeaderConf1(), originCustomHeaderConf2()})
}

func originCustomHeaderConf1() map[string]interface{} {
	return map[string]interface{}{
		"name":  "X-Custom-Header1",
		"value": "samplevalue",
	}
}

func originCustomHeaderConf2() map[string]interface{} {
	return map[string]interface{}{
		"name":  "X-Custom-Header2",
		"value": "samplevalue",
	}
}

func customOriginConf() map[string]interface{} {
	return map[string]interface{}{
		"origin_protocol_policy":   "http-only",
		"http_port":                80,
		"https_port":               443,
		"origin_ssl_protocols":     customOriginSSLProtocolsConf(),
		"origin_read_timeout":      30,
		"origin_keepalive_timeout": 5,
	}
}

func customOriginSSLProtocolsConf() *schema.Set {
	return schema.NewSet(schema.HashString, []interface{}{"SSLv3", "TLSv1", "TLSv1.1", "TLSv1.2"})
}

func originShield() map[string]interface{} {
	return map[string]interface{}{
		"enabled":              true,
		"origin_shield_region": "testRegion",
	}
}

func s3OriginConf() map[string]interface{} {
	return map[string]interface{}{
		"origin_access_identity": "origin-access-identity/cloudfront/E127EXAMPLE51Z",
	}
}

func originWithCustomConf() map[string]interface{} {
	return map[string]interface{}{
		"origin_id":            "CustomOrigin",
		"domain_name":          "www.example.com",
		"origin_path":          "/",
		"custom_origin_config": []interface{}{customOriginConf()},
		"custom_header":        originCustomHeadersConf(),
	}
}

func originWithS3Conf() map[string]interface{} {
	return map[string]interface{}{
		"origin_id":        "S3Origin",
		"domain_name":      "s3.example.com",
		"origin_path":      "/",
		"s3_origin_config": []interface{}{s3OriginConf()},
		"custom_header":    originCustomHeadersConf(),
	}
}

func multiOriginConf() *schema.Set {
	return schema.NewSet(tfcloudfront.OriginHash, []interface{}{originWithCustomConf(), originWithS3Conf()})
}

func originGroupMembers() []interface{} {
	return []interface{}{map[string]interface{}{
		"origin_id": "S3origin",
	}, map[string]interface{}{
		"origin_id": "S3failover",
	}}
}

func failoverStatusCodes() map[string]interface{} {
	return map[string]interface{}{
		"status_codes": schema.NewSet(schema.HashInt, []interface{}{503, 504}),
	}
}

func originGroupConf() map[string]interface{} {
	return map[string]interface{}{
		"origin_id":         "groupS3",
		"failover_criteria": []interface{}{failoverStatusCodes()},
		"member":            originGroupMembers(),
	}
}

func originGroupsConf() *schema.Set {
	return schema.NewSet(tfcloudfront.OriginGroupHash, []interface{}{originGroupConf()})
}

func geoRestrictionWhitelistConf() map[string]interface{} {
	return map[string]interface{}{
		"restriction_type": "whitelist",
		"locations":        schema.NewSet(schema.HashString, []interface{}{"CA", "GB", "US"}),
	}
}

func geoRestrictionsConf() map[string]interface{} {
	return map[string]interface{}{
		"geo_restriction": []interface{}{geoRestrictionWhitelistConf()},
	}
}

func geoRestrictionConfNoItems() map[string]interface{} {
	return map[string]interface{}{
		"restriction_type": "none",
	}
}

func customErrorResponsesConf() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"error_code":            404,
			"error_caching_min_ttl": 30,
			"response_code":         200,
			"response_page_path":    "/error-pages/404.html",
		},
		map[string]interface{}{
			"error_code":            403,
			"error_caching_min_ttl": 15,
			"response_code":         404,
			"response_page_path":    "/error-pages/404.html",
		},
	}
}

func aliasesConf() *schema.Set {
	return schema.NewSet(tfcloudfront.AliasesHash, []interface{}{"example.com", "www.example.com"})
}

func loggingConfigConf() map[string]interface{} {
	return map[string]interface{}{
		"include_cookies": false,
		"bucket":          "mylogs.s3.amazonaws.com",
		"prefix":          "myprefix",
	}
}

func customErrorResponsesConfSet() *schema.Set {
	return schema.NewSet(tfcloudfront.CustomErrorResponseHash, customErrorResponsesConf())
}

func customErrorResponsesConfFirst() map[string]interface{} {
	return customErrorResponsesConf()[0].(map[string]interface{})
}

func customErrorResponseConfNoResponseCode() map[string]interface{} {
	er := customErrorResponsesConf()[0].(map[string]interface{})
	er["response_code"] = 0
	er["response_page_path"] = ""
	return er
}

func viewerCertificateConfSetDefault() map[string]interface{} {
	return map[string]interface{}{
		"acm_certificate_arn":            "",
		"cloudfront_default_certificate": true,
		"iam_certificate_id":             "",
		"minimum_protocol_version":       "",
		"ssl_support_method":             "",
	}
}

func viewerCertificateConfSetIAM() map[string]interface{} {
	return map[string]interface{}{
		"acm_certificate_arn":            "",
		"cloudfront_default_certificate": false,
		"iam_certificate_id":             "iamcert-01234567",
		"ssl_support_method":             "vip",
		"minimum_protocol_version":       "TLSv1",
	}
}

func viewerCertificateConfSetACM() map[string]interface{} {
	return map[string]interface{}{
		"acm_certificate_arn":            "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012", //lintignore:AWSAT003,AWSAT005
		"cloudfront_default_certificate": false,
		"iam_certificate_id":             "",
		"ssl_support_method":             "sni-only",
		"minimum_protocol_version":       "TLSv1",
	}
}

func TestStructure_expandDefaultCacheBehavior(t *testing.T) {
	data := defaultCacheBehaviorConf()
	dcb := tfcloudfront.ExpandDefaultCacheBehavior(data)
	if dcb == nil {
		t.Fatalf("ExpandDefaultCacheBehavior returned nil")
	}
	if !*dcb.Compress {
		t.Fatalf("Expected Compress to be true, got %v", *dcb.Compress)
	}
	if *dcb.ViewerProtocolPolicy != "allow-all" {
		t.Fatalf("Expected ViewerProtocolPolicy to be allow-all, got %v", *dcb.ViewerProtocolPolicy)
	}
	if *dcb.TargetOriginId != "myS3Origin" {
		t.Fatalf("Expected TargetOriginId to be allow-all, got %v", *dcb.TargetOriginId)
	}
	if !reflect.DeepEqual(dcb.ForwardedValues.Headers.Items, flex.ExpandStringSet(headersConf())) {
		t.Fatalf("Expected Items to be %v, got %v", headersConf(), dcb.ForwardedValues.Headers.Items)
	}
	if *dcb.MinTTL != 0 {
		t.Fatalf("Expected MinTTL to be 0, got %v", *dcb.MinTTL)
	}
	if !reflect.DeepEqual(dcb.TrustedSigners.Items, flex.ExpandStringList(trustedSignersConf())) {
		t.Fatalf("Expected TrustedSigners.Items to be %v, got %v", trustedSignersConf(), dcb.TrustedSigners.Items)
	}
	if *dcb.MaxTTL != 31536000 {
		t.Fatalf("Expected MaxTTL to be 31536000, got %v", *dcb.MaxTTL)
	}
	if *dcb.SmoothStreaming {
		t.Fatalf("Expected SmoothStreaming to be false, got %v", *dcb.SmoothStreaming)
	}
	if *dcb.DefaultTTL != 86400 {
		t.Fatalf("Expected DefaultTTL to be 86400, got %v", *dcb.DefaultTTL)
	}
	if *dcb.LambdaFunctionAssociations.Quantity != 2 {
		t.Fatalf("Expected LambdaFunctionAssociations to be 2, got %v", *dcb.LambdaFunctionAssociations.Quantity)
	}
	if *dcb.FunctionAssociations.Quantity != 2 {
		t.Fatalf("Expected FunctionAssociations to be 2, got %v", *dcb.FunctionAssociations.Quantity)
	}
	if !reflect.DeepEqual(dcb.AllowedMethods.Items, flex.ExpandStringSet(allowedMethodsConf())) {
		t.Fatalf("Expected AllowedMethods.Items to be %v, got %v", allowedMethodsConf().List(), dcb.AllowedMethods.Items)
	}
	if !reflect.DeepEqual(dcb.AllowedMethods.CachedMethods.Items, flex.ExpandStringSet(cachedMethodsConf())) {
		t.Fatalf("Expected AllowedMethods.CachedMethods.Items to be %v, got %v", cachedMethodsConf().List(), dcb.AllowedMethods.CachedMethods.Items)
	}
}

func TestStructure_expandTrustedSigners(t *testing.T) {
	data := trustedSignersConf()
	ts := tfcloudfront.ExpandTrustedSigners(data)
	if *ts.Quantity != 2 {
		t.Fatalf("Expected Quantity to be 2, got %v", *ts.Quantity)
	}
	if !*ts.Enabled {
		t.Fatalf("Expected Enabled to be true, got %v", *ts.Enabled)
	}
	if !reflect.DeepEqual(ts.Items, flex.ExpandStringList(data)) {
		t.Fatalf("Expected Items to be %v, got %v", data, ts.Items)
	}
}

func TestStructure_flattenTrustedSigners(t *testing.T) {
	in := trustedSignersConf()
	ts := tfcloudfront.ExpandTrustedSigners(in)
	out := tfcloudfront.FlattenTrustedSigners(ts)

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestStructure_expandTrustedSigners_empty(t *testing.T) {
	data := []interface{}{}
	ts := tfcloudfront.ExpandTrustedSigners(data)
	if *ts.Quantity != 0 {
		t.Fatalf("Expected Quantity to be 0, got %v", *ts.Quantity)
	}
	if *ts.Enabled {
		t.Fatalf("Expected Enabled to be true, got %v", *ts.Enabled)
	}
	if ts.Items != nil {
		t.Fatalf("Expected Items to be nil, got %v", ts.Items)
	}
}

func TestStructure_expandLambdaFunctionAssociations(t *testing.T) {
	data := lambdaFunctionAssociationsConf()
	lfa := tfcloudfront.ExpandLambdaFunctionAssociations(data.List())
	if *lfa.Quantity != 2 {
		t.Fatalf("Expected Quantity to be 2, got %v", *lfa.Quantity)
	}
	if len(lfa.Items) != 2 {
		t.Fatalf("Expected Items to be len 2, got %v", len(lfa.Items))
	}
	if et := "viewer-request"; *lfa.Items[0].EventType != et {
		t.Fatalf("Expected first Item's EventType to be %q, got %q", et, *lfa.Items[0].EventType)
	}
	if et := "origin-response"; *lfa.Items[1].EventType != et {
		t.Fatalf("Expected second Item's EventType to be %q, got %q", et, *lfa.Items[1].EventType)
	}
}

func TestStructure_flattenlambdaFunctionAssociations(t *testing.T) {
	in := lambdaFunctionAssociationsConf()
	lfa := tfcloudfront.ExpandLambdaFunctionAssociations(in.List())
	out := tfcloudfront.FlattenLambdaFunctionAssociations(lfa)

	if !reflect.DeepEqual(in.List(), out.List()) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestStructure_expandlambdaFunctionAssociations_empty(t *testing.T) {
	data := new(schema.Set)
	lfa := tfcloudfront.ExpandLambdaFunctionAssociations(data.List())
	if *lfa.Quantity != 0 {
		t.Fatalf("Expected Quantity to be 0, got %v", *lfa.Quantity)
	}
	if len(lfa.Items) != 0 {
		t.Fatalf("Expected Items to be len 0, got %v", len(lfa.Items))
	}
	if !reflect.DeepEqual(lfa.Items, []*cloudfront.LambdaFunctionAssociation{}) {
		t.Fatalf("Expected Items to be empty, got %v", lfa.Items)
	}
}

func TestStructure_expandFunctionAssociations(t *testing.T) {
	data := functionAssociationsConf()
	lfa := tfcloudfront.ExpandFunctionAssociations(data.List())
	if *lfa.Quantity != 2 {
		t.Fatalf("Expected Quantity to be 2, got %v", *lfa.Quantity)
	}
	if len(lfa.Items) != 2 {
		t.Fatalf("Expected Items to be len 2, got %v", len(lfa.Items))
	}
	if et := "viewer-response"; *lfa.Items[0].EventType != et {
		t.Fatalf("Expected first Item's EventType to be %q, got %q", et, *lfa.Items[0].EventType)
	}
	if et := "viewer-request"; *lfa.Items[1].EventType != et {
		t.Fatalf("Expected second Item's EventType to be %q, got %q", et, *lfa.Items[1].EventType)
	}
}

func TestStructure_flattenFunctionAssociations(t *testing.T) {
	in := functionAssociationsConf()
	lfa := tfcloudfront.ExpandFunctionAssociations(in.List())
	out := tfcloudfront.FlattenFunctionAssociations(lfa)

	if !reflect.DeepEqual(in.List(), out.List()) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestStructure_expandFunctionAssociations_empty(t *testing.T) {
	data := new(schema.Set)
	lfa := tfcloudfront.ExpandFunctionAssociations(data.List())
	if *lfa.Quantity != 0 {
		t.Fatalf("Expected Quantity to be 0, got %v", *lfa.Quantity)
	}
	if len(lfa.Items) != 0 {
		t.Fatalf("Expected Items to be len 0, got %v", len(lfa.Items))
	}
	if !reflect.DeepEqual(lfa.Items, []*cloudfront.FunctionAssociation{}) {
		t.Fatalf("Expected Items to be empty, got %v", lfa.Items)
	}
}

func TestStructure_expandForwardedValues(t *testing.T) {
	data := forwardedValuesConf()
	fv := tfcloudfront.ExpandForwardedValues(data)
	if !*fv.QueryString {
		t.Fatalf("Expected QueryString to be true, got %v", *fv.QueryString)
	}
	if !reflect.DeepEqual(fv.Cookies.WhitelistedNames.Items, flex.ExpandStringSet(cookieNamesConf())) {
		t.Fatalf("Expected Cookies.WhitelistedNames.Items to be %v, got %v", cookieNamesConf(), fv.Cookies.WhitelistedNames.Items)
	}
	if !reflect.DeepEqual(fv.Headers.Items, flex.ExpandStringSet(headersConf())) {
		t.Fatalf("Expected Headers.Items to be %v, got %v", headersConf(), fv.Headers.Items)
	}
}

func TestStructure_flattenForwardedValues(t *testing.T) {
	in := forwardedValuesConf()
	fv := tfcloudfront.ExpandForwardedValues(in)
	out := tfcloudfront.FlattenForwardedValues(fv)

	if !out["query_string"].(bool) {
		t.Fatalf("Expected out[query_string] to be true, got %v", out["query_string"])
	}
}

func TestStructure_expandHeaders(t *testing.T) {
	data := headersConf()
	h := tfcloudfront.ExpandHeaders(data.List())
	if *h.Quantity != 2 {
		t.Fatalf("Expected Quantity to be 2, got %v", *h.Quantity)
	}
	if !reflect.DeepEqual(h.Items, flex.ExpandStringSet(data)) {
		t.Fatalf("Expected Items to be %v, got %v", data, h.Items)
	}
}

func TestStructure_flattenHeaders(t *testing.T) {
	in := headersConf()
	h := tfcloudfront.ExpandHeaders(in.List())
	out := schema.NewSet(schema.HashString, tfcloudfront.FlattenHeaders(h))

	if !in.Equal(out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestStructure_expandQueryStringCacheKeys(t *testing.T) {
	data := queryStringCacheKeysConf()
	k := tfcloudfront.ExpandQueryStringCacheKeys(data)
	if *k.Quantity != 2 {
		t.Fatalf("Expected Quantity to be 2, got %v", *k.Quantity)
	}
	if !reflect.DeepEqual(k.Items, flex.ExpandStringList(data)) {
		t.Fatalf("Expected Items to be %v, got %v", data, k.Items)
	}
}

func TestStructure_flattenQueryStringCacheKeys(t *testing.T) {
	in := queryStringCacheKeysConf()
	k := tfcloudfront.ExpandQueryStringCacheKeys(in)
	out := tfcloudfront.FlattenQueryStringCacheKeys(k)

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestStructure_expandCookiePreference(t *testing.T) {
	data := cookiePreferenceConf()
	cp := tfcloudfront.ExpandCookiePreference(data)
	if *cp.Forward != "whitelist" {
		t.Fatalf("Expected Forward to be whitelist, got %v", *cp.Forward)
	}
	if !reflect.DeepEqual(cp.WhitelistedNames.Items, flex.ExpandStringSet(cookieNamesConf())) {
		t.Fatalf("Expected WhitelistedNames.Items to be %v, got %v", cookieNamesConf(), cp.WhitelistedNames.Items)
	}
}

func TestStructure_flattenCookiePreference(t *testing.T) {
	in := cookiePreferenceConf()
	cp := tfcloudfront.ExpandCookiePreference(in)
	out := tfcloudfront.FlattenCookiePreference(cp)

	if e, a := in["forward"], out["forward"]; e != a {
		t.Fatalf("Expected forward to be %v, got %v", e, a)
	}
}

func TestStructure_expandCookieNames(t *testing.T) {
	data := cookieNamesConf()
	cn := tfcloudfront.ExpandCookieNames(data.List())
	if *cn.Quantity != 2 {
		t.Fatalf("Expected Quantity to be 2, got %v", *cn.Quantity)
	}
	if !reflect.DeepEqual(cn.Items, flex.ExpandStringSet(data)) {
		t.Fatalf("Expected Items to be %v, got %v", data, cn.Items)
	}
}

func TestStructure_flattenCookieNames(t *testing.T) {
	in := cookieNamesConf()
	cn := tfcloudfront.ExpandCookieNames(in.List())
	out := schema.NewSet(schema.HashString, tfcloudfront.FlattenCookieNames(cn))

	if !in.Equal(out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestStructure_expandAllowedMethods(t *testing.T) {
	data := allowedMethodsConf()
	am := tfcloudfront.ExpandAllowedMethods(data)
	if *am.Quantity != 7 {
		t.Fatalf("Expected Quantity to be 7, got %v", *am.Quantity)
	}
	if !reflect.DeepEqual(am.Items, flex.ExpandStringSet(data)) {
		t.Fatalf("Expected Items to be %v, got %v", data, am.Items)
	}
}

func TestStructure_flattenAllowedMethods(t *testing.T) {
	in := allowedMethodsConf()
	am := tfcloudfront.ExpandAllowedMethods(in)
	out := tfcloudfront.FlattenAllowedMethods(am)

	if !in.Equal(out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestStructure_expandCachedMethods(t *testing.T) {
	data := cachedMethodsConf()
	cm := tfcloudfront.ExpandCachedMethods(data)
	if *cm.Quantity != 3 {
		t.Fatalf("Expected Quantity to be 3, got %v", *cm.Quantity)
	}
	if !reflect.DeepEqual(cm.Items, flex.ExpandStringSet(data)) {
		t.Fatalf("Expected Items to be %v, got %v", data, cm.Items)
	}
}

func TestStructure_flattenCachedMethods(t *testing.T) {
	in := cachedMethodsConf()
	cm := tfcloudfront.ExpandCachedMethods(in)
	out := tfcloudfront.FlattenCachedMethods(cm)

	if !in.Equal(out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestStructure_expandOrigins(t *testing.T) {
	data := multiOriginConf()
	origins := tfcloudfront.ExpandOrigins(data)
	if *origins.Quantity != 2 {
		t.Fatalf("Expected Quantity to be 2, got %v", *origins.Quantity)
	}
	if *origins.Items[0].OriginPath != "/" {
		t.Fatalf("Expected first Item's OriginPath to be /, got %v", *origins.Items[0].OriginPath)
	}
}

func TestStructure_flattenOrigins(t *testing.T) {
	in := multiOriginConf()
	origins := tfcloudfront.ExpandOrigins(in)
	out := tfcloudfront.FlattenOrigins(origins)
	diff := in.Difference(out)

	if len(diff.List()) > 0 {
		t.Fatalf("Expected out to be %v, got %v, diff: %v", in, out, diff)
	}
}

func TestStructure_expandOriginGroups(t *testing.T) {
	in := originGroupsConf()
	groups := tfcloudfront.ExpandOriginGroups(in)

	if *groups.Quantity != 1 {
		t.Fatalf("Expected origin group quantity to be %v, got %v", 1, *groups.Quantity)
	}
	originGroup := groups.Items[0]
	if *originGroup.Id != "groupS3" {
		t.Fatalf("Expected origin group id to be %v, got %v", "groupS3", *originGroup.Id)
	}
	if *originGroup.FailoverCriteria.StatusCodes.Quantity != 2 {
		t.Fatalf("Expected 2 origin group status codes, got %v", *originGroup.FailoverCriteria.StatusCodes.Quantity)
	}
	statusCodes := originGroup.FailoverCriteria.StatusCodes.Items
	for _, code := range statusCodes {
		if *code != 503 && *code != 504 {
			t.Fatalf("Expected origin group failover status code to either 503 or 504 got %v", *code)
		}
	}

	if *originGroup.Members.Quantity > 2 {
		t.Fatalf("Expected origin group member quantity to be 2, got %v", *originGroup.Members.Quantity)
	}

	members := originGroup.Members.Items
	if len(members) > 2 {
		t.Fatalf("Expected 2 origin group members, got %v", len(members))
	}
	for _, member := range members {
		if *member.OriginId != "S3failover" && *member.OriginId != "S3origin" {
			t.Fatalf("Expected origin group member to either S3failover or s3origin got %v", *member.OriginId)
		}
	}
}

func TestStructure_flattenOriginGroups(t *testing.T) {
	in := originGroupsConf()
	groups := tfcloudfront.ExpandOriginGroups(in)
	out := tfcloudfront.FlattenOriginGroups(groups)
	diff := in.Difference(out)

	if len(diff.List()) > 0 {
		t.Fatalf("Expected out to be %v, got %v, diff: %v", in, out, diff)
	}
}

func TestStructure_expandOrigin(t *testing.T) {
	data := originWithCustomConf()
	or := tfcloudfront.ExpandOrigin(data)
	if *or.Id != "CustomOrigin" {
		t.Fatalf("Expected Id to be CustomOrigin, got %v", *or.Id)
	}
	if *or.DomainName != "www.example.com" {
		t.Fatalf("Expected DomainName to be www.example.com, got %v", *or.DomainName)
	}
	if *or.OriginPath != "/" {
		t.Fatalf("Expected OriginPath to be /, got %v", *or.OriginPath)
	}
	if *or.CustomOriginConfig.OriginProtocolPolicy != "http-only" {
		t.Fatalf("Expected CustomOriginConfig.OriginProtocolPolicy to be http-only, got %v", *or.CustomOriginConfig.OriginProtocolPolicy)
	}
	if *or.CustomHeaders.Items[0].HeaderValue != "samplevalue" {
		t.Fatalf("Expected CustomHeaders.Items[0].HeaderValue to be samplevalue, got %v", *or.CustomHeaders.Items[0].HeaderValue)
	}
}

func TestStructure_flattenOrigin(t *testing.T) {
	in := originWithCustomConf()
	or := tfcloudfront.ExpandOrigin(in)
	out := tfcloudfront.FlattenOrigin(or)

	if out["origin_id"] != "CustomOrigin" {
		t.Fatalf("Expected out[origin_id] to be CustomOrigin, got %v", out["origin_id"])
	}
	if out["domain_name"] != "www.example.com" {
		t.Fatalf("Expected out[domain_name] to be www.example.com, got %v", out["domain_name"])
	}
	if out["origin_path"] != "/" {
		t.Fatalf("Expected out[origin_path] to be /, got %v", out["origin_path"])
	}
}

func TestStructure_expandCustomHeaders(t *testing.T) {
	in := originCustomHeadersConf()
	chs := tfcloudfront.ExpandCustomHeaders(in)
	if *chs.Quantity != 2 {
		t.Fatalf("Expected Quantity to be 2, got %v", *chs.Quantity)
	}
	if *chs.Items[0].HeaderValue != "samplevalue" {
		t.Fatalf("Expected first Item's HeaderValue to be samplevalue, got %v", *chs.Items[0].HeaderValue)
	}
}

func TestStructure_flattenCustomHeaders(t *testing.T) {
	in := originCustomHeadersConf()
	chs := tfcloudfront.ExpandCustomHeaders(in)
	out := tfcloudfront.FlattenCustomHeaders(chs)
	diff := in.Difference(out)

	if len(diff.List()) > 0 {
		t.Fatalf("Expected out to be %v, got %v, diff: %v", in, out, diff)
	}
}

func TestStructure_flattenOriginCustomHeader(t *testing.T) {
	in := originCustomHeaderConf1()
	och := tfcloudfront.ExpandOriginCustomHeader(in)
	out := tfcloudfront.FlattenOriginCustomHeader(och)

	if out["name"] != "X-Custom-Header1" {
		t.Fatalf("Expected out[name] to be X-Custom-Header1, got %v", out["name"])
	}
	if out["value"] != "samplevalue" {
		t.Fatalf("Expected out[value] to be samplevalue, got %v", out["value"])
	}
}

func TestStructure_expandOriginCustomHeader(t *testing.T) {
	in := originCustomHeaderConf1()
	och := tfcloudfront.ExpandOriginCustomHeader(in)

	if *och.HeaderName != "X-Custom-Header1" {
		t.Fatalf("Expected HeaderName to be X-Custom-Header1, got %v", *och.HeaderName)
	}
	if *och.HeaderValue != "samplevalue" {
		t.Fatalf("Expected HeaderValue to be samplevalue, got %v", *och.HeaderValue)
	}
}

func TestStructure_expandCustomOriginConfig(t *testing.T) {
	data := customOriginConf()
	co := tfcloudfront.ExpandCustomOriginConfig(data)
	if *co.OriginProtocolPolicy != "http-only" {
		t.Fatalf("Expected OriginProtocolPolicy to be http-only, got %v", *co.OriginProtocolPolicy)
	}
	if *co.HTTPPort != 80 {
		t.Fatalf("Expected HTTPPort to be 80, got %v", *co.HTTPPort)
	}
	if *co.HTTPSPort != 443 {
		t.Fatalf("Expected HTTPSPort to be 443, got %v", *co.HTTPSPort)
	}
	if *co.OriginReadTimeout != 30 {
		t.Fatalf("Expected Origin Read Timeout to be 30, got %v", *co.OriginReadTimeout)
	}
	if *co.OriginKeepaliveTimeout != 5 {
		t.Fatalf("Expected Origin Keepalive Timeout to be 5, got %v", *co.OriginKeepaliveTimeout)
	}
}

func TestStructure_flattenCustomOriginConfig(t *testing.T) {
	in := customOriginConf()
	co := tfcloudfront.ExpandCustomOriginConfig(in)
	out := tfcloudfront.FlattenCustomOriginConfig(co)

	if e, a := in["http_port"], out["http_port"]; e != a {
		t.Fatalf("Expected http_port to be %v, got %v", e, a)
	}
	if e, a := in["https_port"], out["https_port"]; e != a {
		t.Fatalf("Expected https_port to be %v, got %v", e, a)
	}
	if e, a := in["origin_keepalive_timeout"], out["origin_keepalive_timeout"]; e != a {
		t.Fatalf("Expected origin_keepalive_timeout to be %v, got %v", e, a)
	}
	if e, a := in["origin_protocol_policy"], out["origin_protocol_policy"]; e != a {
		t.Fatalf("Expected origin_protocol_policy to be %v, got %v", e, a)
	}
	if e, a := in["origin_read_timeout"], out["origin_read_timeout"]; e != a {
		t.Fatalf("Expected origin_read_timeout to be %v, got %v", e, a)
	}
	if e, a := in["origin_ssl_protocols"].(*schema.Set), out["origin_ssl_protocols"].(*schema.Set); !e.Equal(a) {
		t.Fatalf("Expected origin_ssl_protocols to be %v, got %v", e, a)
	}
}

func TestStructure_expandCustomOriginConfigSSL(t *testing.T) {
	in := customOriginSSLProtocolsConf()
	ocs := tfcloudfront.ExpandCustomOriginConfigSSL(in.List())
	if *ocs.Quantity != 4 {
		t.Fatalf("Expected Quantity to be 4, got %v", *ocs.Quantity)
	}
}

func TestStructure_flattenCustomOriginConfigSSL(t *testing.T) {
	in := customOriginSSLProtocolsConf()
	ocs := tfcloudfront.ExpandCustomOriginConfigSSL(in.List())
	out := tfcloudfront.FlattenCustomOriginConfigSSL(ocs)

	if !in.Equal(out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestStructure_expandOriginShield(t *testing.T) {
	data := originShield()
	o := tfcloudfront.ExpandOriginShield(data)
	if *o.Enabled != true {
		t.Fatalf("Expected Enabled to be true, got %v", *o.Enabled)
	}
	if *o.OriginShieldRegion != "testRegion" {
		t.Fatalf("Expected OriginShieldRegion to be testRegion, got %v", *o.OriginShieldRegion)
	}
}

func TestStructure_flattenOriginShield(t *testing.T) {
	in := originShield()
	o := tfcloudfront.ExpandOriginShield(in)
	out := tfcloudfront.FlattenOriginShield(o)

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestStructure_expandS3OriginConfig(t *testing.T) {
	data := s3OriginConf()
	s3o := tfcloudfront.ExpandS3OriginConfig(data)
	if *s3o.OriginAccessIdentity != "origin-access-identity/cloudfront/E127EXAMPLE51Z" {
		t.Fatalf("Expected OriginAccessIdentity to be origin-access-identity/cloudfront/E127EXAMPLE51Z, got %v", *s3o.OriginAccessIdentity)
	}
}

func TestStructure_flattenS3OriginConfig(t *testing.T) {
	in := s3OriginConf()
	s3o := tfcloudfront.ExpandS3OriginConfig(in)
	out := tfcloudfront.FlattenS3OriginConfig(s3o)

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestStructure_expandCustomErrorResponses(t *testing.T) {
	data := customErrorResponsesConfSet()
	ers := tfcloudfront.ExpandCustomErrorResponses(data)
	if *ers.Quantity != 2 {
		t.Fatalf("Expected Quantity to be 2, got %v", *ers.Quantity)
	}
	if *ers.Items[0].ResponsePagePath != "/error-pages/404.html" {
		t.Fatalf("Expected ResponsePagePath in first Item to be /error-pages/404.html, got %v", *ers.Items[0].ResponsePagePath)
	}
}

func TestStructure_flattenCustomErrorResponses(t *testing.T) {
	in := customErrorResponsesConfSet()
	ers := tfcloudfront.ExpandCustomErrorResponses(in)
	out := tfcloudfront.FlattenCustomErrorResponses(ers)

	if !in.Equal(out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestStructure_expandCustomErrorResponse(t *testing.T) {
	data := customErrorResponsesConfFirst()
	er := tfcloudfront.ExpandCustomErrorResponse(data)
	if *er.ErrorCode != 404 {
		t.Fatalf("Expected ErrorCode to be 404, got %v", *er.ErrorCode)
	}
	if *er.ErrorCachingMinTTL != 30 {
		t.Fatalf("Expected ErrorCachingMinTTL to be 30, got %v", *er.ErrorCachingMinTTL)
	}
	if *er.ResponseCode != "200" {
		t.Fatalf("Expected ResponseCode to be 200 (as string), got %v", *er.ResponseCode)
	}
	if *er.ResponsePagePath != "/error-pages/404.html" {
		t.Fatalf("Expected ResponsePagePath to be /error-pages/404.html, got %v", *er.ResponsePagePath)
	}
}

func TestStructure_expandCustomErrorResponse_emptyResponseCode(t *testing.T) {
	data := customErrorResponseConfNoResponseCode()
	er := tfcloudfront.ExpandCustomErrorResponse(data)
	if *er.ResponseCode != "" {
		t.Fatalf("Expected ResponseCode to be empty string, got %v", *er.ResponseCode)
	}
	if *er.ResponsePagePath != "" {
		t.Fatalf("Expected ResponsePagePath to be empty string, got %v", *er.ResponsePagePath)
	}
}

func TestStructure_flattenCustomErrorResponse(t *testing.T) {
	in := customErrorResponsesConfFirst()
	er := tfcloudfront.ExpandCustomErrorResponse(in)
	out := tfcloudfront.FlattenCustomErrorResponse(er)

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("Expected out to be %v, got %v", in, out)
	}
}

func TestStructure_expandLoggingConfig(t *testing.T) {
	data := loggingConfigConf()

	lc := tfcloudfront.ExpandLoggingConfig(data)
	if !*lc.Enabled {
		t.Fatalf("Expected Enabled to be true, got %v", *lc.Enabled)
	}
	if *lc.Prefix != "myprefix" {
		t.Fatalf("Expected Prefix to be myprefix, got %v", *lc.Prefix)
	}
	if *lc.Bucket != "mylogs.s3.amazonaws.com" {
		t.Fatalf("Expected Bucket to be mylogs.s3.amazonaws.com, got %v", *lc.Bucket)
	}
	if *lc.IncludeCookies {
		t.Fatalf("Expected IncludeCookies to be false, got %v", *lc.IncludeCookies)
	}
}

func TestStructure_expandLoggingConfig_nilValue(t *testing.T) {
	lc := tfcloudfront.ExpandLoggingConfig(nil)
	if *lc.Enabled {
		t.Fatalf("Expected Enabled to be false, got %v", *lc.Enabled)
	}
	if *lc.Prefix != "" {
		t.Fatalf("Expected Prefix to be blank, got %v", *lc.Prefix)
	}
	if *lc.Bucket != "" {
		t.Fatalf("Expected Bucket to be blank, got %v", *lc.Bucket)
	}
	if *lc.IncludeCookies {
		t.Fatalf("Expected IncludeCookies to be false, got %v", *lc.IncludeCookies)
	}
}

func TestStructure_expandAliases(t *testing.T) {
	data := aliasesConf()
	a := tfcloudfront.ExpandAliases(data)
	if *a.Quantity != 2 {
		t.Fatalf("Expected Quantity to be 2, got %v", *a.Quantity)
	}
	if !reflect.DeepEqual(a.Items, flex.ExpandStringSet(data)) {
		t.Fatalf("Expected Items to be [example.com www.example.com], got %v", a.Items)
	}
}

func TestStructure_flattenAliases(t *testing.T) {
	in := aliasesConf()
	a := tfcloudfront.ExpandAliases(in)
	out := tfcloudfront.FlattenAliases(a)
	diff := in.Difference(out)

	if len(diff.List()) > 0 {
		t.Fatalf("Expected out to be %v, got %v, diff: %v", in, out, diff)
	}
}

func TestStructure_expandRestrictions(t *testing.T) {
	data := geoRestrictionsConf()
	r := tfcloudfront.ExpandRestrictions(data)
	if *r.GeoRestriction.RestrictionType != "whitelist" {
		t.Fatalf("Expected GeoRestriction.RestrictionType to be whitelist, got %v", *r.GeoRestriction.RestrictionType)
	}
}

func TestStructure_expandGeoRestriction_whitelist(t *testing.T) {
	data := geoRestrictionWhitelistConf()
	gr := tfcloudfront.ExpandGeoRestriction(data)
	if *gr.RestrictionType != "whitelist" {
		t.Fatalf("Expected RestrictionType to be whitelist, got %v", *gr.RestrictionType)
	}
	if *gr.Quantity != 3 {
		t.Fatalf("Expected Quantity to be 3, got %v", *gr.Quantity)
	}
	if !reflect.DeepEqual(aws.StringValueSlice(gr.Items), []string{"GB", "US", "CA"}) {
		t.Fatalf("Expected Items be [CA, GB, US], got %v", aws.StringValueSlice(gr.Items))
	}
}

func TestStructure_flattenGeoRestriction_whitelist(t *testing.T) {
	in := geoRestrictionWhitelistConf()
	gr := tfcloudfront.ExpandGeoRestriction(in)
	out := tfcloudfront.FlattenGeoRestriction(gr)

	if e, a := in["restriction_type"], out["restriction_type"]; e != a {
		t.Fatalf("Expected restriction_type to be %s, got %s", e, a)
	}
	if e, a := in["locations"].(*schema.Set), out["locations"].(*schema.Set); !e.Equal(a) {
		t.Fatalf("Expected out to be %v, got %v", e, a)
	}
}

func TestStructure_expandGeoRestriction_no_items(t *testing.T) {
	data := geoRestrictionConfNoItems()
	gr := tfcloudfront.ExpandGeoRestriction(data)
	if *gr.RestrictionType != "none" {
		t.Fatalf("Expected RestrictionType to be none, got %v", *gr.RestrictionType)
	}
	if *gr.Quantity != 0 {
		t.Fatalf("Expected Quantity to be 0, got %v", *gr.Quantity)
	}
	if gr.Items != nil {
		t.Fatalf("Expected Items to not be set, got %v", gr.Items)
	}
}

func TestStructure_flattenGeoRestriction_no_items(t *testing.T) {
	in := geoRestrictionConfNoItems()
	gr := tfcloudfront.ExpandGeoRestriction(in)
	out := tfcloudfront.FlattenGeoRestriction(gr)

	if e, a := in["restriction_type"], out["restriction_type"]; e != a {
		t.Fatalf("Expected restriction_type to be %s, got %s", e, a)
	}
	if out["locations"] != nil {
		t.Fatalf("Expected locations to be nil, got %v", out["locations"])
	}
}

func TestStructure_expandViewerCertificateDefaultCertificate(t *testing.T) {
	data := viewerCertificateConfSetDefault()
	vc := tfcloudfront.ExpandViewerCertificate(data)
	if vc.ACMCertificateArn != nil {
		t.Fatalf("Expected ACMCertificateArn to be unset, got %v", *vc.ACMCertificateArn)
	}
	if !*vc.CloudFrontDefaultCertificate {
		t.Fatalf("Expected CloudFrontDefaultCertificate to be true, got %v", *vc.CloudFrontDefaultCertificate)
	}
	if vc.IAMCertificateId != nil {
		t.Fatalf("Expected IAMCertificateId to not be set, got %v", *vc.IAMCertificateId)
	}
	if vc.SSLSupportMethod != nil {
		t.Fatalf("Expected IAMCertificateId to not be set, got %v", *vc.SSLSupportMethod)
	}
	if vc.MinimumProtocolVersion != nil {
		t.Fatalf("Expected IAMCertificateId to not be set, got %v", *vc.MinimumProtocolVersion)
	}
}

func TestStructure_expandViewerCertificate_iam_certificate_id(t *testing.T) {
	data := viewerCertificateConfSetIAM()
	vc := tfcloudfront.ExpandViewerCertificate(data)
	if vc.ACMCertificateArn != nil {
		t.Fatalf("Expected ACMCertificateArn to be unset, got %v", *vc.ACMCertificateArn)
	}
	if vc.CloudFrontDefaultCertificate != nil {
		t.Fatalf("Expected CloudFrontDefaultCertificate to be unset, got %v", *vc.CloudFrontDefaultCertificate)
	}
	if *vc.IAMCertificateId != "iamcert-01234567" {
		t.Fatalf("Expected IAMCertificateId to be iamcert-01234567, got %v", *vc.IAMCertificateId)
	}
	if *vc.SSLSupportMethod != "vip" {
		t.Fatalf("Expected IAMCertificateId to be vip, got %v", *vc.SSLSupportMethod)
	}
	if *vc.MinimumProtocolVersion != "TLSv1" {
		t.Fatalf("Expected IAMCertificateId to be TLSv1, got %v", *vc.MinimumProtocolVersion)
	}
}

func TestStructure_expandViewerCertificate_acm_certificate_arn(t *testing.T) {
	data := viewerCertificateConfSetACM()
	vc := tfcloudfront.ExpandViewerCertificate(data)

	// lintignore:AWSAT003,AWSAT005
	if *vc.ACMCertificateArn != "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012" {
		t.Fatalf("Expected ACMCertificateArn to be arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012, got %v", *vc.ACMCertificateArn) // lintignore:AWSAT003,AWSAT005
	}
	if vc.CloudFrontDefaultCertificate != nil {
		t.Fatalf("Expected CloudFrontDefaultCertificate to be unset, got %v", *vc.CloudFrontDefaultCertificate)
	}
	if vc.IAMCertificateId != nil {
		t.Fatalf("Expected IAMCertificateId to be unset, got %v", *vc.IAMCertificateId)
	}
	if *vc.SSLSupportMethod != "sni-only" {
		t.Fatalf("Expected IAMCertificateId to be sni-only, got %v", *vc.SSLSupportMethod)
	}
	if *vc.MinimumProtocolVersion != "TLSv1" {
		t.Fatalf("Expected IAMCertificateId to be TLSv1, got %v", *vc.MinimumProtocolVersion)
	}
}
