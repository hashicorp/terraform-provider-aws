package apigateway_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAPIGatewayMethodSettings_basic(t *testing.T) {
	var stage apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsLoggingLevelConfig(rName, "INFO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.logging_level", "INFO"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodSettingsImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayMethodSettings_Settings_cacheDataEncrypted(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsCacheDataEncryptedConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.cache_data_encrypted", "true"),
				),
			},
			{
				Config: testAccMethodSettingsSettingsCacheDataEncryptedConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.cache_data_encrypted", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodSettingsImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayMethodSettings_Settings_cacheTTLInSeconds(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsCacheTTLInSecondsConfig(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.cache_ttl_in_seconds", "0"),
				),
			},
			{
				Config: testAccMethodSettingsSettingsCacheTTLInSecondsConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.cache_ttl_in_seconds", "1"),
				),
			},
			{
				Config: testAccMethodSettingsSettingsCacheTTLInSecondsConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.cache_ttl_in_seconds", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodSettingsImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayMethodSettings_Settings_cachingEnabled(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsCachingEnabledConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.caching_enabled", "true"),
				),
			},
			{
				Config: testAccMethodSettingsSettingsCachingEnabledConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.caching_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodSettingsImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayMethodSettings_Settings_dataTraceEnabled(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsDataTraceEnabledConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.data_trace_enabled", "true"),
				),
			},
			{
				Config: testAccMethodSettingsSettingsDataTraceEnabledConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.data_trace_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodSettingsImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayMethodSettings_Settings_loggingLevel(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsLoggingLevelConfig(rName, "INFO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage1),
					testAccCheckMethodSettings_loggingLevel(&stage1, "test/GET", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.logging_level", "INFO"),
				),
			},
			{
				Config: testAccMethodSettingsSettingsLoggingLevelConfig(rName, "OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage2),
					testAccCheckMethodSettings_loggingLevel(&stage2, "test/GET", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.logging_level", "OFF"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodSettingsImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayMethodSettings_Settings_metricsEnabled(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsMetricsEnabledConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage1),
					testAccCheckMethodSettings_metricsEnabled(&stage1, "test/GET", true),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.metrics_enabled", "true"),
				),
			},
			{
				Config: testAccMethodSettingsSettingsMetricsEnabledConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage2),
					testAccCheckMethodSettings_metricsEnabled(&stage2, "test/GET", false),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.metrics_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodSettingsImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayMethodSettings_Settings_multiple(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsMultipleConfig(rName, "INFO", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage1),
					testAccCheckMethodSettings_metricsEnabled(&stage1, "test/GET", true),
					testAccCheckMethodSettings_loggingLevel(&stage1, "test/GET", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.metrics_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.logging_level", "INFO"),
				),
			},
			{
				Config: testAccMethodSettingsSettingsMultipleConfig(rName, "OFF", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage2),
					testAccCheckMethodSettings_metricsEnabled(&stage2, "test/GET", false),
					testAccCheckMethodSettings_loggingLevel(&stage2, "test/GET", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.logging_level", "OFF"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodSettingsImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayMethodSettings_Settings_requireAuthorizationForCacheControl(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsRequireAuthorizationForCacheControlConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.require_authorization_for_cache_control", "true"),
				),
			},
			{
				Config: testAccMethodSettingsSettingsRequireAuthorizationForCacheControlConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.require_authorization_for_cache_control", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodSettingsImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayMethodSettings_Settings_throttlingBurstLimit(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsThrottlingBurstLimitConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_burst_limit", "1"),
				),
			},
			{
				Config: testAccMethodSettingsSettingsThrottlingBurstLimitConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_burst_limit", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodSettingsImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/5690
func TestAccAPIGatewayMethodSettings_Settings_throttlingBurstLimitDisabledByDefault(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsLoggingLevelConfig(rName, "INFO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_burst_limit", "-1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodSettingsImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccMethodSettingsSettingsThrottlingBurstLimitConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_burst_limit", "1"),
				),
			},
		},
	})
}

func TestAccAPIGatewayMethodSettings_Settings_throttlingRateLimit(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsThrottlingRateLimitConfig(rName, 1.1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_rate_limit", "1.1"),
				),
			},
			{
				Config: testAccMethodSettingsSettingsThrottlingRateLimitConfig(rName, 2.2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_rate_limit", "2.2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodSettingsImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/5690
func TestAccAPIGatewayMethodSettings_Settings_throttlingRateLimitDisabledByDefault(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsLoggingLevelConfig(rName, "INFO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_rate_limit", "-1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodSettingsImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccMethodSettingsSettingsThrottlingRateLimitConfig(rName, 1.1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_rate_limit", "1.1"),
				),
			},
		},
	})
}

func TestAccAPIGatewayMethodSettings_Settings_unauthorizedCacheControlHeaderStrategy(t *testing.T) {
	var stage1, stage2 apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsUnauthorizedCacheControlHeaderStrategyConfig(rName, "SUCCEED_WITH_RESPONSE_HEADER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage1),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.unauthorized_cache_control_header_strategy", "SUCCEED_WITH_RESPONSE_HEADER"),
				),
			},
			{
				Config: testAccMethodSettingsSettingsUnauthorizedCacheControlHeaderStrategyConfig(rName, "SUCCEED_WITHOUT_RESPONSE_HEADER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage2),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.unauthorized_cache_control_header_strategy", "SUCCEED_WITHOUT_RESPONSE_HEADER"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodSettingsImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMethodSettings_metricsEnabled(conf *apigateway.Stage, path string, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		settings, ok := conf.MethodSettings[path]
		if !ok {
			return fmt.Errorf("Expected to find method settings for %q", path)
		}

		if expected && aws.BoolValue(settings.MetricsEnabled) != expected {
			return fmt.Errorf("Expected metrics to be enabled, got %t", aws.BoolValue(settings.MetricsEnabled))
		}
		if !expected && aws.BoolValue(settings.MetricsEnabled) != expected {
			return fmt.Errorf("Expected metrics to be disabled, got %t", aws.BoolValue(settings.MetricsEnabled))
		}

		return nil
	}
}

func testAccCheckMethodSettings_loggingLevel(conf *apigateway.Stage, path string, expectedLevel string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		settings, ok := conf.MethodSettings[path]
		if !ok {
			return fmt.Errorf("Expected to find method settings for %q", path)
		}

		if aws.StringValue(settings.LoggingLevel) != expectedLevel {
			return fmt.Errorf("Expected logging level to match %q, got %q", expectedLevel, aws.StringValue(settings.LoggingLevel))
		}

		return nil
	}
}

func TestAccAPIGatewayMethodSettings_disappears(t *testing.T) {
	var stage apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMethodSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsSettingsLoggingLevelConfig(rName, "INFO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceMethodSettings(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMethodSettingsDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_method_settings" {
			continue
		}

		_, err := tfapigateway.FindStageByName(conn, rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["stage_name"])
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway Stage %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccMethodSettingsImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		restApiID := rs.Primary.Attributes["rest_api_id"]
		stageName := rs.Primary.Attributes["stage_name"]
		methodPath := rs.Primary.Attributes["method_path"]

		return fmt.Sprintf("%s/%s/%s", restApiID, stageName, methodPath), nil
	}
}

func testAccMethodSettingsBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = {
    "method.request.header.Content-Type" = false
    "method.request.querystring.page"    = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  type        = "MOCK"

  request_templates = {
    "application/xml" = <<EOF
{
   "body" : $input.json('$')
}
EOF
  }
}

resource "aws_api_gateway_deployment" "test" {
  depends_on  = [aws_api_gateway_integration.test]
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "dev"
}
`, rName)
}

func testAccMethodSettingsSettingsCacheDataEncryptedConfig(rName string, cacheDataEncrypted bool) string {
	return testAccMethodSettingsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    cache_data_encrypted = %t
  }
}
`, cacheDataEncrypted)
}

func testAccMethodSettingsSettingsCacheTTLInSecondsConfig(rName string, cacheTtlInSeconds int) string {
	return testAccMethodSettingsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    cache_ttl_in_seconds = %d
  }
}
`, cacheTtlInSeconds)
}

func testAccMethodSettingsSettingsCachingEnabledConfig(rName string, cachingEnabled bool) string {
	return testAccMethodSettingsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    caching_enabled = %t
  }
}
`, cachingEnabled)
}

func testAccMethodSettingsSettingsDataTraceEnabledConfig(rName string, dataTraceEnabled bool) string {
	return testAccMethodSettingsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    data_trace_enabled = %t
  }
}
`, dataTraceEnabled)
}

func testAccMethodSettingsSettingsLoggingLevelConfig(rName, loggingLevel string) string {
	return testAccMethodSettingsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    logging_level = %q
  }
}
`, loggingLevel)
}

func testAccMethodSettingsSettingsMetricsEnabledConfig(rName string, metricsEnabled bool) string {
	return testAccMethodSettingsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    metrics_enabled = %t
  }
}
`, metricsEnabled)
}

func testAccMethodSettingsSettingsMultipleConfig(rName, loggingLevel string, metricsEnabled bool) string {
	return testAccMethodSettingsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"

  settings {
    logging_level   = %q
    metrics_enabled = %t
  }
}
`, loggingLevel, metricsEnabled)
}

func testAccMethodSettingsSettingsRequireAuthorizationForCacheControlConfig(rName string, requireAuthorizationForCacheControl bool) string {
	return testAccMethodSettingsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    require_authorization_for_cache_control = %t
  }
}
`, requireAuthorizationForCacheControl)
}

func testAccMethodSettingsSettingsThrottlingBurstLimitConfig(rName string, throttlingBurstLimit int) string {
	return testAccMethodSettingsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    throttling_burst_limit = %d
  }
}
`, throttlingBurstLimit)
}

func testAccMethodSettingsSettingsThrottlingRateLimitConfig(rName string, throttlingRateLimit float32) string {
	return testAccMethodSettingsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    throttling_rate_limit = %f
  }
}
`, throttlingRateLimit)
}

func testAccMethodSettingsSettingsUnauthorizedCacheControlHeaderStrategyConfig(rName, unauthorizedCacheControlHeaderStrategy string) string {
	return testAccMethodSettingsBaseConfig(rName) + fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    unauthorized_cache_control_header_strategy = %q
  }
}
`, unauthorizedCacheControlHeaderStrategy)
}
