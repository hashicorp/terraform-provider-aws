package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandCloudFrontOriginRequestPolicyCookieNames(cookieNamesFlat map[string]interface{}) *cloudfront.CookieNames {
	cookieNames := &cloudfront.CookieNames{}

	var newCookieItems []*string
	for _, cookie := range cookieNamesFlat["items"].(*schema.Set).List() {
		newCookieItems = append(newCookieItems, aws.String(cookie.(string)))
	}
	cookieNames.Items = newCookieItems
	cookieNames.Quantity = aws.Int64(int64(len(newCookieItems)))

	return cookieNames
}

func expandCloudFrontOriginRequestPolicyCookiesConfig(cookiesConfigFlat map[string]interface{}) *cloudfront.OriginRequestPolicyCookiesConfig {
	var cookies *cloudfront.CookieNames

	if cookiesFlat, ok := cookiesConfigFlat["cookies"].([]interface{}); ok && len(cookiesFlat) == 1 {
		cookies = expandCloudFrontOriginRequestPolicyCookieNames(cookiesFlat[0].(map[string]interface{}))
	} else {
		cookies = nil
	}

	cookiesConfig := &cloudfront.OriginRequestPolicyCookiesConfig{
		CookieBehavior: aws.String(cookiesConfigFlat["cookie_behavior"].(string)),
		Cookies:        cookies,
	}

	return cookiesConfig
}

func expandCloudFrontOriginRequestPolicyHeaders(headerNamesFlat map[string]interface{}) *cloudfront.Headers {
	headers := &cloudfront.Headers{}

	var newHeaderItems []*string
	for _, header := range headerNamesFlat["items"].(*schema.Set).List() {
		newHeaderItems = append(newHeaderItems, aws.String(header.(string)))
	}
	headers.Items = newHeaderItems
	headers.Quantity = aws.Int64(int64(len(newHeaderItems)))

	return headers
}

func expandCloudFrontOriginRequestPolicyHeadersConfig(headersConfigFlat map[string]interface{}) *cloudfront.OriginRequestPolicyHeadersConfig {
	var headers *cloudfront.Headers

	if headersFlat, ok := headersConfigFlat["headers"].([]interface{}); ok && len(headersFlat) == 1 && headersConfigFlat["header_behavior"] != "none" {
		headers = expandCloudFrontOriginRequestPolicyHeaders(headersFlat[0].(map[string]interface{}))
	} else {
		headers = nil
	}

	headersConfig := &cloudfront.OriginRequestPolicyHeadersConfig{
		HeaderBehavior: aws.String(headersConfigFlat["header_behavior"].(string)),
		Headers:        headers,
	}

	return headersConfig
}

func expandCloudFrontOriginRequestPolicyQueryStringNames(queryStringNamesFlat map[string]interface{}) *cloudfront.QueryStringNames {
	queryStringNames := &cloudfront.QueryStringNames{}

	var newQueryStringItems []*string
	for _, queryString := range queryStringNamesFlat["items"].(*schema.Set).List() {
		newQueryStringItems = append(newQueryStringItems, aws.String(queryString.(string)))
	}
	queryStringNames.Items = newQueryStringItems
	queryStringNames.Quantity = aws.Int64(int64(len(newQueryStringItems)))

	return queryStringNames
}

func expandCloudFrontOriginRequestPolicyQueryStringsConfig(queryStringConfigFlat map[string]interface{}) *cloudfront.OriginRequestPolicyQueryStringsConfig {
	var queryStrings *cloudfront.QueryStringNames

	if queryStringFlat, ok := queryStringConfigFlat["query_strings"].([]interface{}); ok && len(queryStringFlat) == 1 {
		queryStrings = expandCloudFrontOriginRequestPolicyQueryStringNames(queryStringFlat[0].(map[string]interface{}))
	} else {
		queryStrings = nil
	}

	queryStringConfig := &cloudfront.OriginRequestPolicyQueryStringsConfig{
		QueryStringBehavior: aws.String(queryStringConfigFlat["query_string_behavior"].(string)),
		QueryStrings:        queryStrings,
	}

	return queryStringConfig
}

func expandCloudFrontOriginRequestPolicyConfig(d *schema.ResourceData) *cloudfront.OriginRequestPolicyConfig {

	originRequestPolicy := &cloudfront.OriginRequestPolicyConfig{
		Comment:            aws.String(d.Get("comment").(string)),
		Name:               aws.String(d.Get("name").(string)),
		CookiesConfig:      expandCloudFrontOriginRequestPolicyCookiesConfig(d.Get("cookies_config").([]interface{})[0].(map[string]interface{})),
		HeadersConfig:      expandCloudFrontOriginRequestPolicyHeadersConfig(d.Get("headers_config").([]interface{})[0].(map[string]interface{})),
		QueryStringsConfig: expandCloudFrontOriginRequestPolicyQueryStringsConfig(d.Get("query_strings_config").([]interface{})[0].(map[string]interface{})),
	}

	return originRequestPolicy
}

func flattenCloudFrontOriginRequestPolicyCookiesConfig(cookiesConfig *cloudfront.OriginRequestPolicyCookiesConfig) []map[string]interface{} {
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

func flattenCloudFrontOriginRequestPolicyHeadersConfig(headersConfig *cloudfront.OriginRequestPolicyHeadersConfig) []map[string]interface{} {
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

func flattenCloudFrontOriginRequestPolicyQueryStringsConfig(queryStringsConfig *cloudfront.OriginRequestPolicyQueryStringsConfig) []map[string]interface{} {
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

func flattenCloudFrontOriginRequestPolicy(d *schema.ResourceData, originRequestPolicy *cloudfront.OriginRequestPolicyConfig) {
	d.Set("comment", aws.StringValue(originRequestPolicy.Comment))
	d.Set("name", aws.StringValue(originRequestPolicy.Name))
	d.Set("cookies_config", flattenCloudFrontOriginRequestPolicyCookiesConfig(originRequestPolicy.CookiesConfig))
	d.Set("headers_config", flattenCloudFrontOriginRequestPolicyHeadersConfig(originRequestPolicy.HeadersConfig))
	d.Set("query_strings_config", flattenCloudFrontOriginRequestPolicyQueryStringsConfig(originRequestPolicy.QueryStringsConfig))
}
