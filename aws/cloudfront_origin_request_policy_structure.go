package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandCloudFrontOriginRequestPolicyCookieNames(tfMap map[string]interface{}) *cloudfront.CookieNames {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.CookieNames{}

	var items []*string
	for _, item := range tfMap["items"].(*schema.Set).List() {
		items = append(items, aws.String(item.(string)))
	}
	apiObject.Items = items
	apiObject.Quantity = aws.Int64(int64(len(items)))

	return apiObject
}

func expandCloudFrontOriginRequestPolicyCookiesConfig(tfMap map[string]interface{}) *cloudfront.OriginRequestPolicyCookiesConfig {
	if tfMap == nil {
		return nil
	}

	var itemsAPIObject *cloudfront.CookieNames

	if itemsFlat, ok := tfMap["cookies"].([]interface{}); ok && len(itemsFlat) == 1 {
		itemsAPIObject = expandCloudFrontOriginRequestPolicyCookieNames(itemsFlat[0].(map[string]interface{}))
	} else {
		itemsAPIObject = nil
	}

	apiObject := &cloudfront.OriginRequestPolicyCookiesConfig{
		CookieBehavior: aws.String(tfMap["cookie_behavior"].(string)),
		Cookies:        itemsAPIObject,
	}

	return apiObject
}

func expandCloudFrontOriginRequestPolicyHeaders(tfMap map[string]interface{}) *cloudfront.Headers {
	if tfMap == nil {
		return nil
	}
	apiObject := &cloudfront.Headers{}

	var items []*string
	for _, item := range tfMap["items"].(*schema.Set).List() {
		items = append(items, aws.String(item.(string)))
	}
	apiObject.Items = items
	apiObject.Quantity = aws.Int64(int64(len(items)))

	return apiObject
}

func expandCloudFrontOriginRequestPolicyHeadersConfig(tfMap map[string]interface{}) *cloudfront.OriginRequestPolicyHeadersConfig {
	if tfMap == nil {
		return nil
	}
	var itemsAPIObject *cloudfront.Headers

	if itemsFlat, ok := tfMap["headers"].([]interface{}); ok && len(itemsFlat) == 1 && tfMap["header_behavior"] != "none" {
		itemsAPIObject = expandCloudFrontOriginRequestPolicyHeaders(itemsFlat[0].(map[string]interface{}))
	} else {
		itemsAPIObject = nil
	}

	apiObject := &cloudfront.OriginRequestPolicyHeadersConfig{
		HeaderBehavior: aws.String(tfMap["header_behavior"].(string)),
		Headers:        itemsAPIObject,
	}

	return apiObject
}

func expandCloudFrontOriginRequestPolicyQueryStringNames(tfMap map[string]interface{}) *cloudfront.QueryStringNames {
	if tfMap == nil {
		return nil
	}
	apiObject := &cloudfront.QueryStringNames{}

	var items []*string
	for _, item := range tfMap["items"].(*schema.Set).List() {
		items = append(items, aws.String(item.(string)))
	}
	apiObject.Items = items
	apiObject.Quantity = aws.Int64(int64(len(items)))

	return apiObject
}

func expandCloudFrontOriginRequestPolicyQueryStringsConfig(tfMap map[string]interface{}) *cloudfront.OriginRequestPolicyQueryStringsConfig {
	if tfMap == nil {
		return nil
	}
	var itemsAPIObject *cloudfront.QueryStringNames

	if itemsFlat, ok := tfMap["query_strings"].([]interface{}); ok && len(itemsFlat) == 1 {
		itemsAPIObject = expandCloudFrontOriginRequestPolicyQueryStringNames(itemsFlat[0].(map[string]interface{}))
	} else {
		itemsAPIObject = nil
	}

	apiObject := &cloudfront.OriginRequestPolicyQueryStringsConfig{
		QueryStringBehavior: aws.String(tfMap["query_string_behavior"].(string)),
		QueryStrings:        itemsAPIObject,
	}

	return apiObject
}

func expandCloudFrontOriginRequestPolicyConfig(d *schema.ResourceData) *cloudfront.OriginRequestPolicyConfig {
	apiObject := &cloudfront.OriginRequestPolicyConfig{
		Comment:            aws.String(d.Get("comment").(string)),
		Name:               aws.String(d.Get("name").(string)),
		CookiesConfig:      expandCloudFrontOriginRequestPolicyCookiesConfig(d.Get("cookies_config").([]interface{})[0].(map[string]interface{})),
		HeadersConfig:      expandCloudFrontOriginRequestPolicyHeadersConfig(d.Get("headers_config").([]interface{})[0].(map[string]interface{})),
		QueryStringsConfig: expandCloudFrontOriginRequestPolicyQueryStringsConfig(d.Get("query_strings_config").([]interface{})[0].(map[string]interface{})),
	}

	return apiObject
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
