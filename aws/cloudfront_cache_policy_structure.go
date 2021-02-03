package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandCloudFrontCachePolicyCookieNames(cookieNamesFlat map[string]interface{}) *cloudfront.CookieNames {
	cookieNames := &cloudfront.CookieNames{}

	var newCookieItems []*string
	for _, cookie := range cookieNamesFlat["items"].(*schema.Set).List() {
		newCookieItems = append(newCookieItems, aws.String(cookie.(string)))
	}
	cookieNames.Items = newCookieItems
	cookieNames.Quantity = aws.Int64(int64(len(newCookieItems)))

	return cookieNames
}

func expandCloudFrontCachePolicyCookiesConfig(cookiesConfigFlat map[string]interface{}) *cloudfront.CachePolicyCookiesConfig {
	cookiesConfig := &cloudfront.CachePolicyCookiesConfig{
		CookieBehavior: aws.String(cookiesConfigFlat["cookie_behavior"].(string)),
	}

	if cookiesFlat, ok := cookiesConfigFlat["cookies"].([]interface{}); ok && len(cookiesFlat) == 1 {
		cookiesConfig.Cookies = expandCloudFrontCachePolicyCookieNames(cookiesFlat[0].(map[string]interface{}))
	}

	return cookiesConfig
}

func expandCloudFrontCachePolicyHeaders(headerNamesFlat map[string]interface{}) *cloudfront.Headers {
	headers := &cloudfront.Headers{}

	var newHeaderItems []*string
	for _, header := range headerNamesFlat["items"].(*schema.Set).List() {
		newHeaderItems = append(newHeaderItems, aws.String(header.(string)))
	}
	headers.Items = newHeaderItems
	headers.Quantity = aws.Int64(int64(len(newHeaderItems)))

	return headers
}

func expandCloudFrontCachePolicyHeadersConfig(headersConfigFlat map[string]interface{}) *cloudfront.CachePolicyHeadersConfig {
	headersConfig := &cloudfront.CachePolicyHeadersConfig{
		HeaderBehavior: aws.String(headersConfigFlat["header_behavior"].(string)),
	}

	if headersFlat, ok := headersConfigFlat["headers"].([]interface{}); ok && len(headersFlat) == 1 && headersConfigFlat["header_behavior"] != "none" {
		headersConfig.Headers = expandCloudFrontCachePolicyHeaders(headersFlat[0].(map[string]interface{}))
	}

	return headersConfig
}

func expandCloudFrontCachePolicyQueryStringNames(queryStringNamesFlat map[string]interface{}) *cloudfront.QueryStringNames {
	queryStringNames := &cloudfront.QueryStringNames{}

	var newQueryStringItems []*string
	for _, queryString := range queryStringNamesFlat["items"].(*schema.Set).List() {
		newQueryStringItems = append(newQueryStringItems, aws.String(queryString.(string)))
	}
	queryStringNames.Items = newQueryStringItems
	queryStringNames.Quantity = aws.Int64(int64(len(newQueryStringItems)))

	return queryStringNames
}

func expandCloudFrontCachePolicyQueryStringConfig(queryStringConfigFlat map[string]interface{}) *cloudfront.CachePolicyQueryStringsConfig {
	queryStringConfig := &cloudfront.CachePolicyQueryStringsConfig{
		QueryStringBehavior: aws.String(queryStringConfigFlat["query_string_behavior"].(string)),
	}

	if queryStringFlat, ok := queryStringConfigFlat["query_strings"].([]interface{}); ok && len(queryStringFlat) == 1 {
		queryStringConfig.QueryStrings = expandCloudFrontCachePolicyQueryStringNames(queryStringFlat[0].(map[string]interface{}))
	}

	return queryStringConfig
}

func expandCloudFrontCachePolicyParametersConfig(parameters map[string]interface{}) *cloudfront.ParametersInCacheKeyAndForwardedToOrigin {
	var cookiesConfig *cloudfront.CachePolicyCookiesConfig
	var headersConfig *cloudfront.CachePolicyHeadersConfig
	var queryStringsConfig *cloudfront.CachePolicyQueryStringsConfig

	if cookiesFlat, ok := parameters["cookies_config"].([]interface{}); ok && len(cookiesFlat) == 1 {
		cookiesConfig = expandCloudFrontCachePolicyCookiesConfig(cookiesFlat[0].(map[string]interface{}))
	}

	if headersFlat, ok := parameters["headers_config"].([]interface{}); ok && len(headersFlat) == 1 {
		headersConfig = expandCloudFrontCachePolicyHeadersConfig(headersFlat[0].(map[string]interface{}))
	}

	if queryStringsFlat, ok := parameters["query_strings_config"].([]interface{}); ok && len(queryStringsFlat) == 1 {
		queryStringsConfig = expandCloudFrontCachePolicyQueryStringConfig(queryStringsFlat[0].(map[string]interface{}))
	}

	parametersConfig := &cloudfront.ParametersInCacheKeyAndForwardedToOrigin{
		CookiesConfig:              cookiesConfig,
		EnableAcceptEncodingBrotli: aws.Bool(parameters["enable_accept_encoding_brotli"].(bool)),
		EnableAcceptEncodingGzip:   aws.Bool(parameters["enable_accept_encoding_gzip"].(bool)),
		HeadersConfig:              headersConfig,
		QueryStringsConfig:         queryStringsConfig,
	}

	return parametersConfig
}

func expandCloudFrontCachePolicyConfig(d *schema.ResourceData) *cloudfront.CachePolicyConfig {
	parametersConfig := &cloudfront.ParametersInCacheKeyAndForwardedToOrigin{}

	if parametersFlat, ok := d.GetOk("parameters_in_cache_key_and_forwarded_to_origin"); ok {
		parametersConfig = expandCloudFrontCachePolicyParametersConfig(parametersFlat.([]interface{})[0].(map[string]interface{}))
	}
	cachePolicy := &cloudfront.CachePolicyConfig{
		Comment:                                  aws.String(d.Get("comment").(string)),
		DefaultTTL:                               aws.Int64(int64(d.Get("default_ttl").(int))),
		MaxTTL:                                   aws.Int64(int64(d.Get("max_ttl").(int))),
		MinTTL:                                   aws.Int64(int64(d.Get("min_ttl").(int))),
		Name:                                     aws.String(d.Get("name").(string)),
		ParametersInCacheKeyAndForwardedToOrigin: parametersConfig,
	}

	return cachePolicy
}

func flattenCloudFrontCachePolicyCookiesConfig(cookiesConfig *cloudfront.CachePolicyCookiesConfig) []map[string]interface{} {
	cookiesConfigFlat := map[string]interface{}{}

	cookies := []map[string]interface{}{}
	if cookiesConfig.Cookies != nil {
		cookies = []map[string]interface{}{
			{
				"items": cookiesConfig.Cookies.Items,
			},
		}
	}

	cookiesConfigFlat["cookie_behavior"] = aws.StringValue(cookiesConfig.CookieBehavior)
	cookiesConfigFlat["cookies"] = cookies

	return []map[string]interface{}{
		cookiesConfigFlat,
	}
}

func flattenCloudFrontCachePolicyHeadersConfig(headersConfig *cloudfront.CachePolicyHeadersConfig) []map[string]interface{} {
	headersConfigFlat := map[string]interface{}{}

	headers := []map[string]interface{}{}
	if headersConfig.Headers != nil {
		headers = []map[string]interface{}{
			{
				"items": headersConfig.Headers.Items,
			},
		}
	}

	headersConfigFlat["header_behavior"] = aws.StringValue(headersConfig.HeaderBehavior)
	headersConfigFlat["headers"] = headers

	return []map[string]interface{}{
		headersConfigFlat,
	}
}

func flattenCloudFrontCachePolicyQueryStringsConfig(queryStringsConfig *cloudfront.CachePolicyQueryStringsConfig) []map[string]interface{} {
	queryStringsConfigFlat := map[string]interface{}{}

	queryStrings := []map[string]interface{}{}
	if queryStringsConfig.QueryStrings != nil {
		queryStrings = []map[string]interface{}{
			{
				"items": queryStringsConfig.QueryStrings.Items,
			},
		}
	}

	queryStringsConfigFlat["query_string_behavior"] = aws.StringValue(queryStringsConfig.QueryStringBehavior)
	queryStringsConfigFlat["query_strings"] = queryStrings

	return []map[string]interface{}{
		queryStringsConfigFlat,
	}
}

func setParametersConfig(parametersConfig *cloudfront.ParametersInCacheKeyAndForwardedToOrigin) []map[string]interface{} {
	parametersConfigFlat := map[string]interface{}{
		"enable_accept_encoding_brotli": aws.BoolValue(parametersConfig.EnableAcceptEncodingBrotli),
		"enable_accept_encoding_gzip":   aws.BoolValue(parametersConfig.EnableAcceptEncodingGzip),
		"cookies_config":                flattenCloudFrontCachePolicyCookiesConfig(parametersConfig.CookiesConfig),
		"headers_config":                flattenCloudFrontCachePolicyHeadersConfig(parametersConfig.HeadersConfig),
		"query_strings_config":          flattenCloudFrontCachePolicyQueryStringsConfig(parametersConfig.QueryStringsConfig),
	}

	return []map[string]interface{}{
		parametersConfigFlat,
	}
}

func setCloudFrontCachePolicy(d *schema.ResourceData, cachePolicy *cloudfront.CachePolicyConfig) {
	d.Set("comment", aws.StringValue(cachePolicy.Comment))
	d.Set("default_ttl", aws.Int64Value(cachePolicy.DefaultTTL))
	d.Set("max_ttl", aws.Int64Value(cachePolicy.MaxTTL))
	d.Set("min_ttl", aws.Int64Value(cachePolicy.MinTTL))
	d.Set("name", aws.StringValue(cachePolicy.Name))
	d.Set("parameters_in_cache_key_and_forwarded_to_origin", setParametersConfig(cachePolicy.ParametersInCacheKeyAndForwardedToOrigin))
}
