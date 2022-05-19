package apigatewayv2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAPIGatewayV2IntegrationResponse_basic(t *testing.T) {
	var apiId, integrationId string
	var v apigatewayv2.GetIntegrationResponseOutput
	resourceName := "aws_apigatewayv2_integration_response.test"
	integrationResourceName := "aws_apigatewayv2_integration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIntegrationResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationResponseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationResponseExists(resourceName, &apiId, &integrationId, &v),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", ""),
					resource.TestCheckResourceAttrPair(resourceName, "integration_id", integrationResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_key", "/200/"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.%", "0"),
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
	var apiId, integrationId string
	var v apigatewayv2.GetIntegrationResponseOutput
	resourceName := "aws_apigatewayv2_integration_response.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIntegrationResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationResponseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationResponseExists(resourceName, &apiId, &integrationId, &v),
					testAccCheckIntegrationResponseDisappears(&apiId, &integrationId, &v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2IntegrationResponse_allAttributes(t *testing.T) {
	var apiId, integrationId string
	var v apigatewayv2.GetIntegrationResponseOutput
	resourceName := "aws_apigatewayv2_integration_response.test"
	integrationResourceName := "aws_apigatewayv2_integration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIntegrationResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationResponseConfig_allAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationResponseExists(resourceName, &apiId, &integrationId, &v),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttrPair(resourceName, "integration_id", integrationResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", "$request.body.name"),
				),
			},
			{
				Config: testAccIntegrationResponseConfig_allAttributesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationResponseExists(resourceName, &apiId, &integrationId, &v),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", "CONVERT_TO_BINARY"),
					resource.TestCheckResourceAttrPair(resourceName, "integration_id", integrationResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_key", "/404/"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.%", "2"),
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

func testAccCheckIntegrationResponseDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_integration_response" {
			continue
		}

		_, err := conn.GetIntegrationResponse(&apigatewayv2.GetIntegrationResponseInput{
			ApiId:                 aws.String(rs.Primary.Attributes["api_id"]),
			IntegrationId:         aws.String(rs.Primary.Attributes["integration_id"]),
			IntegrationResponseId: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 integration response %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckIntegrationResponseDisappears(apiId, integrationId *string, v *apigatewayv2.GetIntegrationResponseOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		_, err := conn.DeleteIntegrationResponse(&apigatewayv2.DeleteIntegrationResponseInput{
			ApiId:                 apiId,
			IntegrationId:         integrationId,
			IntegrationResponseId: v.IntegrationResponseId,
		})

		return err
	}
}

func testAccCheckIntegrationResponseExists(n string, vApiId, vIntegrationId *string, v *apigatewayv2.GetIntegrationResponseOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 integration response ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		apiId := aws.String(rs.Primary.Attributes["api_id"])
		integrationId := aws.String(rs.Primary.Attributes["integration_id"])
		resp, err := conn.GetIntegrationResponse(&apigatewayv2.GetIntegrationResponseInput{
			ApiId:                 apiId,
			IntegrationId:         integrationId,
			IntegrationResponseId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*vApiId = *apiId
		*vIntegrationId = *integrationId
		*v = *resp

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
