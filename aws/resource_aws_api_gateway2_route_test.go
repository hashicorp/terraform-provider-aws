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

func TestAccAWSAPIGateway2Route_basic(t *testing.T) {
	resourceName := "aws_api_gateway_v2_route.test"
	rName := fmt.Sprintf("tf-testacc-apigwv2-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2RouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2RouteConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2RouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2RouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGateway2Route_Authorizer(t *testing.T) {
	resourceName := "aws_api_gateway_v2_route.test"
	authorizerResourceName := "aws_api_gateway_v2_authorizer.test"
	rName := fmt.Sprintf("tf-testacc-apigwv2-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2RouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2RouteConfig_authorizer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2RouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeCustom),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_id", authorizerResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2RouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGateway2RouteConfig_authorizerUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2RouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeAwsIam),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
		},
	})
}

func TestAccAWSAPIGateway2Route_Model(t *testing.T) {
	resourceName := "aws_api_gateway_v2_route.test"
	modelResourceName := "aws_api_gateway_v2_model.test"
	rName := fmt.Sprintf("tftestaccapigwv2%s", acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2RouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2RouteConfig_model(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2RouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", "action"),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "request_models.test", modelResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2RouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGateway2Route_SimpleAttributes(t *testing.T) {
	resourceName := "aws_api_gateway_v2_route.test"
	rName := fmt.Sprintf("tf-testacc-apigwv2-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2RouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2RouteConfig_simpleAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2RouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "true"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", "GET"),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", "$default"),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				Config: testAccAWSAPIGateway2RouteConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2RouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				Config: testAccAWSAPIGateway2RouteConfig_simpleAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2RouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "true"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", "GET"),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", "$default"),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2RouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGateway2Route_Target(t *testing.T) {
	resourceName := "aws_api_gateway_v2_route.test"
	integrationResourceName := "aws_api_gateway_v2_integration.test"
	rName := fmt.Sprintf("tf-testacc-apigwv2-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2RouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2RouteConfig_target(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2RouteExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					testAccCheckAWSAPIGateway2RouteTarget(resourceName, integrationResourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2RouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGateway2RouteDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_v2_route" {
			continue
		}

		_, err := conn.GetRoute(&apigatewayv2.GetRouteInput{
			ApiId:   aws.String(rs.Primary.Attributes["api_id"]),
			RouteId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 route %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGateway2RouteExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 route ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.GetRoute(&apigatewayv2.GetRouteInput{
			ApiId:   aws.String(rs.Primary.Attributes["api_id"]),
			RouteId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSAPIGateway2RouteImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

func testAccCheckAWSAPIGateway2RouteTarget(resourceName, integrationResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[integrationResourceName]
		if !ok {
			return fmt.Errorf("Not Found: %s", integrationResourceName)
		}

		return resource.TestCheckResourceAttr(resourceName, "target", fmt.Sprintf("integrations/%s", rs.Primary.ID))(s)
	}
}

func testAccAWSAPIGateway2RouteConfig_api(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_v2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
`, rName)
}

func testAccAWSAPIGateway2RouteConfig_basic(rName string) string {
	return testAccAWSAPIGateway2RouteConfig_api(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_route" "test" {
  api_id    = "${aws_api_gateway_v2_api.test.id}"
  route_key = "$default"
}
`)
}

func testAccAWSAPIGateway2RouteConfig_authorizer(rName string) string {
	return testAccAWSAPIGatewayV2AuthorizerConfig_basic(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_route" "test" {
  api_id    = "${aws_api_gateway_v2_api.test.id}"
  route_key = "$connect"

  authorization_type = "CUSTOM"
  authorizer_id      = "${aws_api_gateway_v2_authorizer.test.id}"
}
`)
}

func testAccAWSAPIGateway2RouteConfig_authorizerUpdated(rName string) string {
	return testAccAWSAPIGatewayV2AuthorizerConfig_basic(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_route" "test" {
  api_id    = "${aws_api_gateway_v2_api.test.id}"
  route_key = "$connect"

  authorization_type = "AWS_IAM"
}
`)
}

func testAccAWSAPIGateway2RouteConfig_model(rName string) string {
	schema := `
	{
	  "$schema": "http://json-schema.org/draft-04/schema#",
	  "title": "ExampleModel",
	  "type": "object",
	  "properties": {
		"id": {
		  "type": "string"
		}
	  }
	}
	`

	return testAccAWSAPIGatewayV2ModelConfig_basic(rName, schema) + fmt.Sprintf(`
resource "aws_api_gateway_v2_route" "test" {
  api_id    = "${aws_api_gateway_v2_api.test.id}"
  route_key = "$default"

  model_selection_expression = "action"

  request_models = {
    "test" = "${aws_api_gateway_v2_model.test.name}"
  }
}
`)
}

// Simple attributes - No authorization, models or targets.
func testAccAWSAPIGateway2RouteConfig_simpleAttributes(rName string) string {
	return testAccAWSAPIGateway2RouteConfig_api(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_route" "test" {
  api_id    = "${aws_api_gateway_v2_api.test.id}"
  route_key = "$default"

  api_key_required                    = true
  operation_name                      = "GET"
  route_response_selection_expression = "$default"
}
`)
}

func testAccAWSAPIGateway2RouteConfig_target(rName string) string {
	return testAccAWSAPIGatewayV2IntegrationConfig_basic(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_route" "test" {
  api_id    = "${aws_api_gateway_v2_api.test.id}"
  route_key = "$default"

  target = "integrations/${aws_api_gateway_v2_integration.test.id}"
}
`)
}
