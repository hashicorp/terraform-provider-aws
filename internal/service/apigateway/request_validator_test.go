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

func TestAccAPIGatewayRequestValidator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRequestValidatorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_request_validator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRequestValidatorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRequestValidatorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRequestValidatorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "validate_request_body", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "validate_request_parameters", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRequestValidatorImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccRequestValidatorConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRequestValidatorExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("%s-modified", rName)),
					resource.TestCheckResourceAttr(resourceName, "validate_request_body", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "validate_request_parameters", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccAPIGatewayRequestValidator_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRequestValidatorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_request_validator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRequestValidatorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRequestValidatorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRequestValidatorExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceRequestValidator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRequestValidatorExists(ctx context.Context, n string, v *apigateway.GetRequestValidatorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindRequestValidatorByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["rest_api_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRequestValidatorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_request_validator" {
				continue
			}

			_, err := tfapigateway.FindRequestValidatorByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["rest_api_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Request Validator %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccRequestValidatorImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.ID), nil
	}
}

func testAccRequestValidatorConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRequestValidatorConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccRequestValidatorConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_request_validator" "test" {
  name        = %[1]q
  rest_api_id = aws_api_gateway_rest_api.test.id
}
`, rName))
}

func testAccRequestValidatorConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccRequestValidatorConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_request_validator" "test" {
  name                        = "%[1]s-modified"
  rest_api_id                 = aws_api_gateway_rest_api.test.id
  validate_request_body       = true
  validate_request_parameters = true
}
`, rName))
}
