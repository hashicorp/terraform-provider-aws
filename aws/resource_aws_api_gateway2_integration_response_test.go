package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSAPIGateway2IntegrationResponse_basic(t *testing.T) {
	resourceName := "aws_api_gateway_v2_integration_response.test"
	integrationResourceName := "aws_api_gateway_v2_integration.test"
	rName := fmt.Sprintf("tf-testacc-apigwv2-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2IntegrationResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2IntegrationResponseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2IntegrationResponseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", ""),
					resource.TestCheckResourceAttrPair(resourceName, "integration_id", integrationResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_key", "/200/"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2IntegrationResponseImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGateway2IntegrationResponse_AllAttributes(t *testing.T) {
	resourceName := "aws_api_gateway_v2_integration_response.test"
	integrationResourceName := "aws_api_gateway_v2_integration.test"
	rName := fmt.Sprintf("tf-testacc-apigwv2-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2IntegrationResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2IntegrationResponseConfig_allAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2IntegrationResponseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "content_handling_strategy", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttrPair(resourceName, "integration_id", integrationResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "integration_response_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "template_selection_expression", "$request.body.name"),
				),
			},
			{
				Config: testAccAWSAPIGateway2IntegrationResponseConfig_allAttributesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2IntegrationResponseExists(resourceName),
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
				ImportStateIdFunc: testAccAWSAPIGateway2IntegrationResponseImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGateway2IntegrationResponseDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_v2_integration_response" {
			continue
		}

		_, err := conn.GetIntegrationResponse(&apigatewayv2.GetIntegrationResponseInput{
			ApiId:                 aws.String(rs.Primary.Attributes["api_id"]),
			IntegrationId:         aws.String(rs.Primary.Attributes["integration_id"]),
			IntegrationResponseId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 integration response %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGateway2IntegrationResponseExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 integration response ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.GetIntegrationResponse(&apigatewayv2.GetIntegrationResponseInput{
			ApiId:                 aws.String(rs.Primary.Attributes["api_id"]),
			IntegrationId:         aws.String(rs.Primary.Attributes["integration_id"]),
			IntegrationResponseId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSAPIGateway2IntegrationResponseImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.Attributes["integration_id"], rs.Primary.ID), nil
	}
}

func testAccAWSAPIGateway2IntegrationResponseConfig_basic(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_basic(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_integration_response" "test" {
  api_id                   = "${aws_api_gateway_v2_api.test.id}"
  integration_id           = "${aws_api_gateway_v2_integration.test.id}"
  integration_response_key = "/200/"
}
`)
}

func testAccAWSAPIGateway2IntegrationResponseConfig_allAttributes(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_basic(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_integration_response" "test" {
  api_id                   = "${aws_api_gateway_v2_api.test.id}"
  integration_id           = "${aws_api_gateway_v2_integration.test.id}"
  integration_response_key = "$default"

  content_handling_strategy     = "CONVERT_TO_TEXT"
  template_selection_expression = "$request.body.name"

  response_templates = {
    "application/json" = ""
  }
}
`)
}

func testAccAWSAPIGateway2IntegrationResponseConfig_allAttributesUpdated(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_basic(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_integration_response" "test" {
  api_id                   = "${aws_api_gateway_v2_api.test.id}"
  integration_id           = "${aws_api_gateway_v2_integration.test.id}"
  integration_response_key = "/404/"

  content_handling_strategy     = "CONVERT_TO_BINARY"
  template_selection_expression = "$request.body.id"

  response_templates = {
    "application/json" = "#set($number=42)"
    "application/xml"  = "#set($percent=$number/100)"
  }
}
`)
}
