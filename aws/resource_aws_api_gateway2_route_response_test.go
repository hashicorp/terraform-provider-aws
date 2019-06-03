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

func TestAccAWSAPIGateway2RouteResponse_basic(t *testing.T) {
	resourceName := "aws_api_gateway_v2_route_response.test"
	routeResourceName := "aws_api_gateway_v2_route.test"
	rName := fmt.Sprintf("tf-testacc-apigwv2-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2RouteResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2RouteResponseConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2RouteResponseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "response_models.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "route_id", routeResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "route_response_key", "$default"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2RouteResponseImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAPIGateway2RouteResponse_Model(t *testing.T) {
	resourceName := "aws_api_gateway_v2_route_response.test"
	modelResourceName := "aws_api_gateway_v2_model.test"
	routeResourceName := "aws_api_gateway_v2_route.test"
	rName := fmt.Sprintf("tftestaccapigwv2%s", acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGateway2RouteResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGateway2RouteResponseConfig_model(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGateway2RouteResponseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", "action"),
					resource.TestCheckResourceAttr(resourceName, "response_models.%", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "response_models.test", modelResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "route_id", routeResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "route_response_key", "$default"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAPIGateway2RouteResponseImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAPIGateway2RouteResponseDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_v2_route_response" {
			continue
		}

		_, err := conn.GetRouteResponse(&apigatewayv2.GetRouteResponseInput{
			ApiId:           aws.String(rs.Primary.Attributes["api_id"]),
			RouteId:         aws.String(rs.Primary.Attributes["route_id"]),
			RouteResponseId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 route response %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGateway2RouteResponseExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 route response ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		_, err := conn.GetRouteResponse(&apigatewayv2.GetRouteResponseInput{
			ApiId:           aws.String(rs.Primary.Attributes["api_id"]),
			RouteId:         aws.String(rs.Primary.Attributes["route_id"]),
			RouteResponseId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSAPIGateway2RouteResponseImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.Attributes["route_id"], rs.Primary.ID), nil
	}
}

func testAccAWSAPIGateway2RouteResponseConfig_basic(rName string) string {
	return testAccAWSAPIGatewayV2RouteConfig_basic(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_route_response" "test" {
  api_id             = "${aws_api_gateway_v2_api.test.id}"
  route_id           = "${aws_api_gateway_v2_route.test.id}"
  route_response_key = "$default"
}
`)
}

func testAccAWSAPIGateway2RouteResponseConfig_model(rName string) string {
	return testAccAWSAPIGatewayV2RouteConfig_model(rName) + fmt.Sprintf(`
resource "aws_api_gateway_v2_route_response" "test" {
  api_id             = "${aws_api_gateway_v2_api.test.id}"
  route_id           = "${aws_api_gateway_v2_route.test.id}"
  route_response_key = "$default"

  model_selection_expression = "action"

  response_models = {
    "test" = "${aws_api_gateway_v2_model.test.name}"
  }
}
`)
}
