// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayV2IntegrationResponse_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId, integrationId string
	var v apigatewayv2.GetIntegrationResponseOutput
	resourceName := "aws_apigatewayv2_integration_response.test"
	integrationResourceName := "aws_apigatewayv2_integration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationResponseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationResponseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationResponseExists(ctx, resourceName, &apiId, &integrationId, &v),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", ""),
					resource.TestCheckResourceAttrPair(resourceName, "integration_id", integrationResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "integration_response_key", "/200/"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccIntegrationResponseImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayV2IntegrationResponse_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId, integrationId string
	var v apigatewayv2.GetIntegrationResponseOutput
	resourceName := "aws_apigatewayv2_integration_response.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationResponseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationResponseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationResponseExists(ctx, resourceName, &apiId, &integrationId, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigatewayv2.ResourceIntegrationResponse(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2IntegrationResponse_allAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId, integrationId string
	var v apigatewayv2.GetIntegrationResponseOutput
	resourceName := "aws_apigatewayv2_integration_response.test"
	integrationResourceName := "aws_apigatewayv2_integration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationResponseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationResponseConfig_allAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationResponseExists(ctx, resourceName, &apiId, &integrationId, &v),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttrPair(resourceName, "integration_id", integrationResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "integration_response_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "response_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", "$request.body.name"),
				),
			},
			{
				Config: testAccIntegrationResponseConfig_allAttributesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationResponseExists(ctx, resourceName, &apiId, &integrationId, &v),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", "CONVERT_TO_BINARY"),
					resource.TestCheckResourceAttrPair(resourceName, "integration_id", integrationResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "integration_response_key", "/404/"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "response_templates.application/json", "#set($number=42)"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.application/xml", "#set($percent=$number/100)"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", "$request.body.id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccIntegrationResponseImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckIntegrationResponseDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_integration_response" {
				continue
			}

			_, err := tfapigatewayv2.FindIntegrationResponseByThreePartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.Attributes["integration_id"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 Integration Response %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIntegrationResponseExists(ctx context.Context, n string, apiID, integrationID *string, v *apigatewayv2.GetIntegrationResponseOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		output, err := tfapigatewayv2.FindIntegrationResponseByThreePartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.Attributes["integration_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*apiID = rs.Primary.Attributes["api_id"]
		*integrationID = rs.Primary.Attributes["integration_id"]
		*v = *output

		return nil
	}
}

func testAccIntegrationResponseImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.Attributes["integration_id"], rs.Primary.ID), nil
	}
}

func testAccIntegrationResponseConfig_basic(rName string) string {
	return testAccIntegrationConfig_basic(rName) + `
resource "aws_apigatewayv2_integration_response" "test" {
  api_id                   = aws_apigatewayv2_api.test.id
  integration_id           = aws_apigatewayv2_integration.test.id
  integration_response_key = "/200/"
}
`
}

func testAccIntegrationResponseConfig_allAttributes(rName string) string {
	return testAccIntegrationConfig_basic(rName) + `
resource "aws_apigatewayv2_integration_response" "test" {
  api_id                   = aws_apigatewayv2_api.test.id
  integration_id           = aws_apigatewayv2_integration.test.id
  integration_response_key = "$default"

  content_handling_strategy     = "CONVERT_TO_TEXT"
  template_selection_expression = "$request.body.name"

  response_templates = {
    "application/json" = ""
  }
}
`
}

func testAccIntegrationResponseConfig_allAttributesUpdated(rName string) string {
	return testAccIntegrationConfig_basic(rName) + `
resource "aws_apigatewayv2_integration_response" "test" {
  api_id                   = aws_apigatewayv2_api.test.id
  integration_id           = aws_apigatewayv2_integration.test.id
  integration_response_key = "/404/"

  content_handling_strategy     = "CONVERT_TO_BINARY"
  template_selection_expression = "$request.body.id"

  response_templates = {
    "application/json" = "#set($number=42)"
    "application/xml"  = "#set($percent=$number/100)"
  }
}
`
}
