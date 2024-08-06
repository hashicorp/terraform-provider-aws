// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMethodSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_loggingLevel(rName, "INFO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
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

func testAccMethodSettings_Settings_cacheDataEncrypted(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_cacheDataEncrypted(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.cache_data_encrypted", acctest.CtTrue),
				),
			},
			{
				Config: testAccMethodSettingsConfig_cacheDataEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.cache_data_encrypted", acctest.CtFalse),
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

func testAccMethodSettings_Settings_cacheTTLInSeconds(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_cacheTTLInSeconds(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.cache_ttl_in_seconds", acctest.Ct0),
				),
			},
			{
				Config: testAccMethodSettingsConfig_cacheTTLInSeconds(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.cache_ttl_in_seconds", acctest.Ct1),
				),
			},
			{
				Config: testAccMethodSettingsConfig_cacheTTLInSeconds(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.cache_ttl_in_seconds", acctest.Ct2),
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

func testAccMethodSettings_Settings_cachingEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_cachingEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.caching_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccMethodSettingsConfig_cachingEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.caching_enabled", acctest.CtFalse),
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

func testAccMethodSettings_Settings_dataTraceEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_dataTraceEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.data_trace_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccMethodSettingsConfig_dataTraceEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.data_trace_enabled", acctest.CtFalse),
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

func testAccMethodSettings_Settings_loggingLevel(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_loggingLevel(rName, "INFO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.logging_level", "INFO"),
				),
			},
			{
				Config: testAccMethodSettingsConfig_loggingLevel(rName, "OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
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

func testAccMethodSettings_Settings_metricsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_metricsEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.metrics_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccMethodSettingsConfig_metricsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.metrics_enabled", acctest.CtFalse),
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

func testAccMethodSettings_Settings_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_multiple(rName, "INFO", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.metrics_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "settings.0.logging_level", "INFO"),
				),
			},
			{
				Config: testAccMethodSettingsConfig_multiple(rName, "OFF", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.metrics_enabled", acctest.CtFalse),
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

func testAccMethodSettings_Settings_requireAuthorizationForCacheControl(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_requireAuthorizationForCacheControl(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.require_authorization_for_cache_control", acctest.CtTrue),
				),
			},
			{
				Config: testAccMethodSettingsConfig_requireAuthorizationForCacheControl(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.require_authorization_for_cache_control", acctest.CtFalse),
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

func testAccMethodSettings_Settings_throttlingBurstLimit(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_throttlingBurstLimit(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_burst_limit", acctest.Ct1),
				),
			},
			{
				Config: testAccMethodSettingsConfig_throttlingBurstLimit(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_burst_limit", acctest.Ct2),
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
func testAccMethodSettings_Settings_throttlingBurstLimitDisabledByDefault(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_loggingLevel(rName, "INFO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
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
				Config: testAccMethodSettingsConfig_throttlingBurstLimit(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_burst_limit", acctest.Ct1),
				),
			},
		},
	})
}

func testAccMethodSettings_Settings_throttlingRateLimit(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_throttlingRateLimit(rName, 1.1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_rate_limit", "1.1"),
				),
			},
			{
				Config: testAccMethodSettingsConfig_throttlingRateLimit(rName, 2.2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
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
func testAccMethodSettings_Settings_throttlingRateLimitDisabledByDefault(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_loggingLevel(rName, "INFO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
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
				Config: testAccMethodSettingsConfig_throttlingRateLimit(rName, 1.1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.throttling_rate_limit", "1.1"),
				),
			},
		},
	})
}

func testAccMethodSettings_Settings_unauthorizedCacheControlHeaderStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_unauthorizedCacheControlHeaderStrategy(rName, "SUCCEED_WITH_RESPONSE_HEADER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.unauthorized_cache_control_header_strategy", "SUCCEED_WITH_RESPONSE_HEADER"),
				),
			},
			{
				Config: testAccMethodSettingsConfig_unauthorizedCacheControlHeaderStrategy(rName, "SUCCEED_WITHOUT_RESPONSE_HEADER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
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

func testAccMethodSettings_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodSettingsConfig_loggingLevel(rName, "INFO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodSettingsExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceMethodSettings(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMethodSettingsExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		_, err := tfapigateway.FindMethodSettingsByThreePartKey(ctx, conn, rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["stage_name"], rs.Primary.Attributes["method_path"])

		return err
	}
}

func testAccCheckMethodSettingsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_method_settings" {
				continue
			}

			_, err := tfapigateway.FindMethodSettingsByThreePartKey(ctx, conn, rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["stage_name"], rs.Primary.Attributes["method_path"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Method Settings %s still exists", rs.Primary.ID)
		}

		return nil
	}
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

func testAccMethodSettingsConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccAccountConfig_role0(rName), fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
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
`, rName))
}

func testAccMethodSettingsConfig_cacheDataEncrypted(rName string, cacheDataEncrypted bool) string {
	return acctest.ConfigCompose(testAccMethodSettingsConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    cache_data_encrypted = %[1]t
  }
}
`, cacheDataEncrypted))
}

func testAccMethodSettingsConfig_cacheTTLInSeconds(rName string, cacheTtlInSeconds int) string {
	return acctest.ConfigCompose(testAccMethodSettingsConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    cache_ttl_in_seconds = %[1]d
  }
}
`, cacheTtlInSeconds))
}

func testAccMethodSettingsConfig_cachingEnabled(rName string, cachingEnabled bool) string {
	return acctest.ConfigCompose(testAccMethodSettingsConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    caching_enabled = %[1]t
  }
}
`, cachingEnabled))
}

func testAccMethodSettingsConfig_dataTraceEnabled(rName string, dataTraceEnabled bool) string {
	return acctest.ConfigCompose(testAccMethodSettingsConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    data_trace_enabled = %[1]t
  }
}
`, dataTraceEnabled))
}

func testAccMethodSettingsConfig_loggingLevel(rName, loggingLevel string) string {
	return acctest.ConfigCompose(testAccMethodSettingsConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    logging_level = %[1]q
  }

  depends_on = [aws_api_gateway_account.test]
}
`, loggingLevel))
}

func testAccMethodSettingsConfig_metricsEnabled(rName string, metricsEnabled bool) string {
	return acctest.ConfigCompose(testAccMethodSettingsConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    metrics_enabled = %[1]t
  }
}
`, metricsEnabled))
}

func testAccMethodSettingsConfig_multiple(rName, loggingLevel string, metricsEnabled bool) string {
	return acctest.ConfigCompose(testAccMethodSettingsConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"

  settings {
    logging_level   = %[1]q
    metrics_enabled = %[2]t
  }

  depends_on = [aws_api_gateway_account.test]
}
`, loggingLevel, metricsEnabled))
}

func testAccMethodSettingsConfig_requireAuthorizationForCacheControl(rName string, requireAuthorizationForCacheControl bool) string {
	return acctest.ConfigCompose(testAccMethodSettingsConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    require_authorization_for_cache_control = %[1]t
  }
}
`, requireAuthorizationForCacheControl))
}

func testAccMethodSettingsConfig_throttlingBurstLimit(rName string, throttlingBurstLimit int) string {
	return acctest.ConfigCompose(testAccMethodSettingsConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    throttling_burst_limit = %[1]d
  }
}
`, throttlingBurstLimit))
}

func testAccMethodSettingsConfig_throttlingRateLimit(rName string, throttlingRateLimit float32) string {
	return acctest.ConfigCompose(testAccMethodSettingsConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    throttling_rate_limit = %[1]f
  }
}
`, throttlingRateLimit))
}

func testAccMethodSettingsConfig_unauthorizedCacheControlHeaderStrategy(rName, unauthorizedCacheControlHeaderStrategy string) string {
	return acctest.ConfigCompose(testAccMethodSettingsConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_method_settings" "test" {
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_deployment.test.stage_name

  settings {
    unauthorized_cache_control_header_strategy = %[1]q
  }
}
`, unauthorizedCacheControlHeaderStrategy))
}
