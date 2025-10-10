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

func TestAccAPIGatewayGatewayResponse_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetGatewayResponseOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_gateway_response.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayResponseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayResponseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayResponseExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatusCode, "401"),
					resource.TestCheckResourceAttr(resourceName, "response_parameters.gatewayresponse.header.Authorization", "'Basic'"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
					resource.TestCheckNoResourceAttr(resourceName, "response_templates.application/json"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccGatewayResponseImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccGatewayResponseConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayResponseExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatusCode, "477"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.application/json", "{'message':$context.error.messageString}"),
					resource.TestCheckNoResourceAttr(resourceName, "response_templates.application/xml"),
					resource.TestCheckNoResourceAttr(resourceName, "response_parameters.gatewayresponse.header.Authorization"),
				),
			},
		},
	})
}

func TestAccAPIGatewayGatewayResponse_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetGatewayResponseOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_gateway_response.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayResponseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayResponseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayResponseExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceGatewayResponse(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGatewayResponseExists(ctx context.Context, n string, v *apigateway.GetGatewayResponseOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindGatewayResponseByTwoPartKey(ctx, conn, rs.Primary.Attributes["response_type"], rs.Primary.Attributes["rest_api_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGatewayResponseDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_gateway_response" {
				continue
			}

			_, err := tfapigateway.FindGatewayResponseByTwoPartKey(ctx, conn, rs.Primary.Attributes["response_type"], rs.Primary.Attributes["rest_api_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Gateway Response %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccGatewayResponseImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["response_type"]), nil
	}
}

func testAccGatewayResponseConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_gateway_response" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  status_code   = "401"
  response_type = "UNAUTHORIZED"

  response_templates = {
    "application/xml" = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  response_parameters = {
    "gatewayresponse.header.Authorization" = "'Basic'"
  }
}
`, rName)
}

func testAccGatewayResponseConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_gateway_response" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  status_code   = "477"
  response_type = "UNAUTHORIZED"

  response_templates = {
    "application/json" = "{'message':$context.error.messageString}"
  }
}
`, rName)
}
