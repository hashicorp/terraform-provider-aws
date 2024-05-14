// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"testing"

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

func TestAccAPIGatewayUsagePlanKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetUsagePlanKeyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	apiGatewayApiKeyResourceName := "aws_api_gateway_api_key.test"
	apiGatewayUsagePlanResourceName := "aws_api_gateway_usage_plan.test"
	resourceName := "aws_api_gateway_usage_plan_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsagePlanKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanKeyConfig_typeAPI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanKeyExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKeyID, apiGatewayApiKeyResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "key_type", "API_KEY"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "usage_plan_id", apiGatewayUsagePlanResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrValue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccCheckUsagePlanKeyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayUsagePlanKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetUsagePlanKeyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_usage_plan_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsagePlanKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanKeyConfig_typeAPI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanKeyExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceUsagePlanKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayUsagePlanKey_KeyID_concurrency(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetUsagePlanKeyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUsagePlanKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUsagePlanKeyConfig_idConcurrency(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUsagePlanKeyExists(ctx, "aws_api_gateway_usage_plan_key.test.0", &conf),
					testAccCheckUsagePlanKeyExists(ctx, "aws_api_gateway_usage_plan_key.test.1", &conf),
					testAccCheckUsagePlanKeyExists(ctx, "aws_api_gateway_usage_plan_key.test.2", &conf),
					testAccCheckUsagePlanKeyExists(ctx, "aws_api_gateway_usage_plan_key.test.3", &conf),
					testAccCheckUsagePlanKeyExists(ctx, "aws_api_gateway_usage_plan_key.test.4", &conf),
					testAccCheckUsagePlanKeyExists(ctx, "aws_api_gateway_usage_plan_key.test.5", &conf),
					testAccCheckUsagePlanKeyExists(ctx, "aws_api_gateway_usage_plan_key.test.6", &conf),
					testAccCheckUsagePlanKeyExists(ctx, "aws_api_gateway_usage_plan_key.test.7", &conf),
					testAccCheckUsagePlanKeyExists(ctx, "aws_api_gateway_usage_plan_key.test.8", &conf),
					testAccCheckUsagePlanKeyExists(ctx, "aws_api_gateway_usage_plan_key.test.9", &conf),
				),
			},
		},
	})
}

func testAccCheckUsagePlanKeyExists(ctx context.Context, n string, v *apigateway.GetUsagePlanKeyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindUsagePlanKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes["usage_plan_id"], rs.Primary.Attributes[names.AttrKeyID])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckUsagePlanKeyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_usage_plan_key" {
				continue
			}

			_, err := tfapigateway.FindUsagePlanKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes["usage_plan_id"], rs.Primary.Attributes[names.AttrKeyID])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Usage Plan Key %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUsagePlanKeyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["usage_plan_id"], rs.Primary.ID), nil
	}
}

func testAccUsagePlanKeyBaseConfig(rName string) string {
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
  depends_on = [aws_api_gateway_integration_response.test]

  description = "This is a test"
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "test"
}
`, rName)
}

func testAccUsagePlanKeyConfig_typeAPI(rName string) string {
	return acctest.ConfigCompose(
		testAccUsagePlanKeyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  name = %[1]q
}

resource "aws_api_gateway_usage_plan" "test" {
  name = %[1]q

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.test.stage_name
  }
}

resource "aws_api_gateway_usage_plan_key" "test" {
  key_id        = aws_api_gateway_api_key.test.id
  key_type      = "API_KEY"
  usage_plan_id = aws_api_gateway_usage_plan.test.id
}
`, rName))
}

func testAccUsagePlanKeyConfig_idConcurrency(rName string) string {
	return acctest.ConfigCompose(
		testAccUsagePlanKeyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  count = 10

  name = "%[1]s-${count.index}"
}

resource "aws_api_gateway_usage_plan" "test" {
  name = %[1]q

  api_stages {
    api_id = aws_api_gateway_rest_api.test.id
    stage  = aws_api_gateway_deployment.test.stage_name
  }
}

resource "aws_api_gateway_usage_plan_key" "test" {
  count = 10

  key_id        = aws_api_gateway_api_key.test[count.index].id
  key_type      = "API_KEY"
  usage_plan_id = aws_api_gateway_usage_plan.test.id
}
`, rName))
}
