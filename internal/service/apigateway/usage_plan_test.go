// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayUsagePlan_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetUsagePlanOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedName := sdkacctest.RandomWithPrefix("tf-acc-test-2")
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsagePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/usageplans/+.`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "api_stages.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.%", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUsagePlanConfig_basic(updatedName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, updatedName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayUsagePlan_description(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetUsagePlanOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsagePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUsagePlanConfig_description(rName, "This is a description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This is a description"),
				),
			},
			{
				Config: testAccUsagePlanConfig_description(rName, "This is a new description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This is a new description"),
				),
			},
			{
				Config: testAccUsagePlanConfig_description(rName, "This is a description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "This is a description"),
				),
			},
			{
				Config: testAccUsagePlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayUsagePlan_productCode(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetUsagePlanOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsagePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanConfig_productCode(rName, "MYCODE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "product_code", "MYCODE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUsagePlanConfig_productCode(rName, "MYCODE2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "product_code", "MYCODE2"),
				),
			},
			{
				Config: testAccUsagePlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "product_code", ""),
				),
			},
			{
				Config: testAccUsagePlanConfig_productCode(rName, "MYCODE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "product_code", "MYCODE"),
				),
			},
		},
	})
}

func TestAccAPIGatewayUsagePlan_throttling(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetUsagePlanOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsagePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUsagePlanConfig_throttling(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.burst_limit", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.rate_limit", "5"),
				),
			},
			{
				Config: testAccUsagePlanConfig_throttlingModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.burst_limit", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.rate_limit", "6"),
				),
			},
			{
				Config: testAccUsagePlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.#", acctest.Ct0),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/2057
func TestAccAPIGatewayUsagePlan_throttlingInitialRateLimit(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetUsagePlanOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsagePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanConfig_throttling(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.0.rate_limit", "5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayUsagePlan_quota(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetUsagePlanOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsagePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUsagePlanConfig_quota(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.limit", "100"),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.offset", "6"),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.period", "WEEK"),
				),
			},
			{
				Config: testAccUsagePlanConfig_quotaModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.limit", "200"),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.offset", "20"),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.0.period", "MONTH"),
				),
			},
			{
				Config: testAccUsagePlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "quota_settings.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAPIGatewayUsagePlan_apiStages(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetUsagePlanOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsagePlanDestroy(ctx),
		Steps: []resource.TestStep{
			// Create UsagePlan WITH Stages as the API calls are different
			// when creating or updating.
			{
				Config: testAccUsagePlanConfig_apiStages(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						names.AttrStage: "test",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Handle api stages removal
			{
				Config: testAccUsagePlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "api_stages.#", acctest.Ct0),
				),
			},
			// Handle api stages additions
			{
				Config: testAccUsagePlanConfig_apiStages(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						names.AttrStage: "test",
					}),
				),
			},
			// Handle api stages updates
			{
				Config: testAccUsagePlanConfig_apiStagesMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						names.AttrStage: "foo",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						names.AttrStage: "test",
					}),
				),
			},
			{
				Config: testAccUsagePlanConfig_apiStagesModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						names.AttrStage: "foo",
					}),
				),
			},
			{
				Config: testAccUsagePlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "api_stages.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAPIGatewayUsagePlan_APIStages_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetUsagePlanOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsagePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanConfig_apiStagesMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						names.AttrStage: "foo",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						names.AttrStage: "test",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayUsagePlan_APIStages_throttle(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetUsagePlanOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsagePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanConfig_apiStagesThrottle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "api_stages.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						names.AttrStage: "test",
						"throttle.#":    acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.0.throttle.*", map[string]string{
						names.AttrPath: "/test/GET",
						"burst_limit":  acctest.Ct3,
						"rate_limit":   "6",
					}),
				),
			},
			{
				Config: testAccUsagePlanConfig_apiStagesThrottleMulti(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "api_stages.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						names.AttrStage: "foo",
						"throttle.#":    acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.*", map[string]string{
						names.AttrStage: "test",
						"throttle.#":    acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "api_stages.0.throttle.*", map[string]string{
						names.AttrPath: "/test/GET",
						"burst_limit":  acctest.Ct3,
						"rate_limit":   "6",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayUsagePlan_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetUsagePlanOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsagePlanDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceUsagePlan(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUsagePlanExists(ctx context.Context, n string, v *apigateway.GetUsagePlanOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindUsagePlanByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckUsagePlanDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_usage_plan" {
				continue
			}

			_, err := tfapigateway.FindUsagePlanByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Usage Plan %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccUsagePlanConfig_base(rName string) string {
	return fmt.Sprintf(`
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
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "test"
  description = "This is a test"

  variables = {
    "a" = "2"
  }
}

resource "aws_api_gateway_deployment" "foo" {
  depends_on = [
    aws_api_gateway_deployment.test,
    aws_api_gateway_integration.test,
  ]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "foo"
  description = "This is a prod stage"
}
`, rName)
}

func testAccUsagePlanConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccUsagePlanConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = %[1]q
}
`, rName))
}

func testAccUsagePlanConfig_description(rName, desc string) string {
	return acctest.ConfigCompose(testAccUsagePlanConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, desc))
}

func testAccUsagePlanConfig_productCode(rName, code string) string {
	return acctest.ConfigCompose(testAccUsagePlanConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name         = %[1]q
  product_code = %[2]q
}
`, rName, code))
}

func testAccUsagePlanConfig_throttling(rName string) string {
	return acctest.ConfigCompose(testAccUsagePlanConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = %[1]q

  throttle_settings {
    burst_limit = 2
    rate_limit  = 5
  }
}
`, rName))
}

func testAccUsagePlanConfig_throttlingModified(rName string) string {
	return acctest.ConfigCompose(testAccUsagePlanConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = %[1]q

  throttle_settings {
    burst_limit = 3
    rate_limit  = 6
  }
}
`, rName))
}

func testAccUsagePlanConfig_quota(rName string) string {
	return acctest.ConfigCompose(testAccUsagePlanConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = %[1]q

  quota_settings {
    limit  = 100
    offset = 6
    period = "WEEK"
  }
}
`, rName))
}

func testAccUsagePlanConfig_quotaModified(rName string) string {
	return acctest.ConfigCompose(testAccUsagePlanConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = %[1]q

  quota_settings {
    limit  = 200
    offset = 20
    period = "MONTH"
  }
}
`, rName))
}

func testAccUsagePlanConfig_apiStages(rName string) string {
	return acctest.ConfigCompose(testAccUsagePlanConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = %[1]q

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.test.stage_name
  }
}
`, rName))
}

func testAccUsagePlanConfig_apiStagesThrottle(rName string) string {
	return acctest.ConfigCompose(testAccUsagePlanConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = %[1]q

  throttle_settings {
    burst_limit = 3
    rate_limit  = 6
  }

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.test.stage_name

    throttle {
      path        = "${aws_api_gateway_resource.test.path}/${aws_api_gateway_method.test.http_method}"
      burst_limit = 3
      rate_limit  = 6
    }
  }
}
`, rName))
}

func testAccUsagePlanConfig_apiStagesThrottleMulti(rName string) string {
	return acctest.ConfigCompose(testAccUsagePlanConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = %[1]q

  throttle_settings {
    burst_limit = 3
    rate_limit  = 6
  }

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.test.stage_name

    throttle {
      path        = "${aws_api_gateway_resource.test.path}/${aws_api_gateway_method.test.http_method}"
      burst_limit = 3
      rate_limit  = 6
    }
  }

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.foo.stage_name

    throttle {
      path        = "${aws_api_gateway_resource.test.path}/${aws_api_gateway_method.test.http_method}"
      burst_limit = 3
      rate_limit  = 6
    }
  }
}
`, rName))
}

func testAccUsagePlanConfig_apiStagesModified(rName string) string {
	return acctest.ConfigCompose(testAccUsagePlanConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = %[1]q

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.foo.stage_name
  }
}
`, rName))
}

func testAccUsagePlanConfig_apiStagesMultiple(rName string) string {
	return acctest.ConfigCompose(testAccUsagePlanConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_usage_plan" "test" {
  name = %[1]q

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.foo.stage_name
  }

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.test.stage_name
  }
}
`, rName))
}
