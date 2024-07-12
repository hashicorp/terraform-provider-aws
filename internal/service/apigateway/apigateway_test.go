// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.APIGatewayServiceID, testAccErrorCheckSkip)
}

// skips tests that have error messages indicating unsupported features
func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"no matching Route53Zone found",
	)
}

func TestAccAPIGateway_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Account": {
			acctest.CtBasic: testAccAccount_basic,
		},
		// Some aws_api_gateway_method_settings tests require the account-level CloudWatch Logs role ARN to be set.
		// Serialize all this resource's acceptance tests.
		"MethodSettings": {
			acctest.CtBasic:                          testAccMethodSettings_basic,
			acctest.CtDisappears:                     testAccMethodSettings_disappears,
			"CacheDataEncrypted":                     testAccMethodSettings_Settings_cacheDataEncrypted,
			"CacheTTLInSeconds":                      testAccMethodSettings_Settings_cacheTTLInSeconds,
			"CachingEnabled":                         testAccMethodSettings_Settings_cachingEnabled,
			"DataTraceEnabled":                       testAccMethodSettings_Settings_dataTraceEnabled,
			"LoggingLevel":                           testAccMethodSettings_Settings_loggingLevel,
			"MetricsEnabled":                         testAccMethodSettings_Settings_metricsEnabled,
			"Multiple":                               testAccMethodSettings_Settings_multiple,
			"RequireAuthorizationForCacheControl":    testAccMethodSettings_Settings_requireAuthorizationForCacheControl,
			"ThrottlingBurstLimit":                   testAccMethodSettings_Settings_throttlingBurstLimit,
			"ThrottlingBurstLimitDisabledByDefault":  testAccMethodSettings_Settings_throttlingBurstLimitDisabledByDefault,
			"ThrottlingRateLimit":                    testAccMethodSettings_Settings_throttlingRateLimit,
			"ThrottlingRateLimitDisabledByDefault":   testAccMethodSettings_Settings_throttlingRateLimitDisabledByDefault,
			"UnauthorizedCacheControlHeaderStrategy": testAccMethodSettings_Settings_unauthorizedCacheControlHeaderStrategy,
		},
		// Some aws_api_gateway_stage tests require the account-level CloudWatch Logs role ARN to be set.
		// Serialize all this resource's acceptance tests.
		"Stage": {
			acctest.CtBasic:             testAccStage_basic,
			"tags":                      testAccAPIGatewayStage_tagsSerial,
			acctest.CtDisappears:        testAccStage_disappears,
			"disappears_restAPI":        testAccStage_Disappears_restAPI,
			"Cache":                     testAccStage_cache,
			"CacheSizeCacheDisabled":    testAccStage_cacheSizeCacheDisabled,
			"AccessLogSettings":         testAccStage_accessLogSettings,
			"AccessLogSettings_kinesis": testAccStage_AccessLogSettings_kinesis,
			"WAF":                       testAccStage_waf,
			"CanarySettings":            testAccStage_canarySettings,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
