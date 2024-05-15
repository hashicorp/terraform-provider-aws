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

func TestAccAPIGatewayIntegrationResponse_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetIntegrationResponseOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_integration_response.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationResponseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationResponseConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntegrationResponseExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "content_handling", ""),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.method.response.header.Content-Type", "integration.response.body.type"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "response_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "response_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
					resource.TestCheckResourceAttr(resourceName, "selection_pattern", ".*"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatusCode, "400"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIntegrationResponseImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccIntegrationResponseConfig_update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntegrationResponseExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "content_handling", "CONVERT_TO_BINARY"),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "response_templates.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "response_templates.application/json", "$input.path('$')"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.application/xml", ""),
					resource.TestCheckResourceAttr(resourceName, "selection_pattern", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatusCode, "400"),
				),
			},
		},
	})
}

func TestAccAPIGatewayIntegrationResponse_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetIntegrationResponseOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_integration_response.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationResponseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationResponseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationResponseExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceIntegrationResponse(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIntegrationResponseExists(ctx context.Context, n string, v *apigateway.GetIntegrationResponseOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindIntegrationResponseByFourPartKey(ctx, conn, rs.Primary.Attributes["http_method"], rs.Primary.Attributes[names.AttrResourceID], rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes[names.AttrStatusCode])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckIntegrationResponseDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_integration_response" {
				continue
			}

			_, err := tfapigateway.FindIntegrationResponseByFourPartKey(ctx, conn, rs.Primary.Attributes["http_method"], rs.Primary.Attributes[names.AttrResourceID], rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes[names.AttrStatusCode])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Integration Response %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccIntegrationResponseImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes[names.AttrResourceID], rs.Primary.Attributes["http_method"], rs.Primary.Attributes[names.AttrStatusCode]), nil
	}
}

func testAccIntegrationResponseConfig_basic(rName string) string {
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

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"

  response_models = {
    "application/json" = "Error"
  }

  response_parameters = {
    "method.response.header.Content-Type" = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  request_templates = {
    "application/json" = ""
    "application/xml"  = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  type = "MOCK"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id       = aws_api_gateway_rest_api.test.id
  resource_id       = aws_api_gateway_resource.test.id
  http_method       = aws_api_gateway_method.test.http_method
  status_code       = aws_api_gateway_method_response.error.status_code
  selection_pattern = ".*"

  response_templates = {
    "application/json" = ""
    "application/xml"  = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  response_parameters = {
    "method.response.header.Content-Type" = "integration.response.body.type"
  }
}
`, rName)
}

func testAccIntegrationResponseConfig_update(rName string) string {
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

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"

  response_models = {
    "application/json" = "Error"
  }

  response_parameters = {
    "method.response.header.Content-Type" = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  request_templates = {
    "application/json" = ""
    "application/xml"  = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  type = "MOCK"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code

  response_templates = {
    "application/json" = "$input.path('$')"
    "application/xml"  = ""
  }

  content_handling = "CONVERT_TO_BINARY"
}
`, rName)
}
