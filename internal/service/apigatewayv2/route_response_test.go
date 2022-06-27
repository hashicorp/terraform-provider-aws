package apigatewayv2_test

import (
	"fmt"
	"strings"
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

func TestAccAPIGatewayV2RouteResponse_basic(t *testing.T) {
	var apiId, routeId string
	var v apigatewayv2.GetRouteResponseOutput
	resourceName := "aws_apigatewayv2_route_response.test"
	routeResourceName := "aws_apigatewayv2_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteResponseConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteResponseExists(resourceName, &apiId, &routeId, &v),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "response_models.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "route_id", routeResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "route_response_key", "$default"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteResponseImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayV2RouteResponse_disappears(t *testing.T) {
	var apiId, routeId string
	var v apigatewayv2.GetRouteResponseOutput
	resourceName := "aws_apigatewayv2_route_response.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteResponseConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteResponseExists(resourceName, &apiId, &routeId, &v),
					testAccCheckRouteResponseDisappears(&apiId, &routeId, &v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2RouteResponse_model(t *testing.T) {
	var apiId, routeId string
	var v apigatewayv2.GetRouteResponseOutput
	resourceName := "aws_apigatewayv2_route_response.test"
	modelResourceName := "aws_apigatewayv2_model.test"
	routeResourceName := "aws_apigatewayv2_route.test"
	// Model name must be alphanumeric.
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteResponseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteResponseConfig_model(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteResponseExists(resourceName, &apiId, &routeId, &v),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", "action"),
					resource.TestCheckResourceAttr(resourceName, "response_models.%", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "response_models.test", modelResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "route_id", routeResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "route_response_key", "$default"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteResponseImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRouteResponseDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_route_response" {
			continue
		}

		_, err := conn.GetRouteResponse(&apigatewayv2.GetRouteResponseInput{
			ApiId:           aws.String(rs.Primary.Attributes["api_id"]),
			RouteId:         aws.String(rs.Primary.Attributes["route_id"]),
			RouteResponseId: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 route response %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckRouteResponseDisappears(apiId, routeId *string, v *apigatewayv2.GetRouteResponseOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		_, err := conn.DeleteRouteResponse(&apigatewayv2.DeleteRouteResponseInput{
			ApiId:           apiId,
			RouteId:         routeId,
			RouteResponseId: v.RouteResponseId,
		})

		return err
	}
}

func testAccCheckRouteResponseExists(n string, vApiId, vRouteId *string, v *apigatewayv2.GetRouteResponseOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 route response ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		apiId := aws.String(rs.Primary.Attributes["api_id"])
		routeId := aws.String(rs.Primary.Attributes["route_id"])
		resp, err := conn.GetRouteResponse(&apigatewayv2.GetRouteResponseInput{
			ApiId:           apiId,
			RouteId:         routeId,
			RouteResponseId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*vApiId = *apiId
		*vRouteId = *routeId
		*v = *resp

		return nil
	}
}

func testAccRouteResponseImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.Attributes["route_id"], rs.Primary.ID), nil
	}
}

func testAccRouteResponseConfig_basicWebSocket(rName string) string {
	return acctest.ConfigCompose(
		testAccRouteConfig_basicWebSocket(rName),
		`
resource "aws_apigatewayv2_route_response" "test" {
  api_id             = aws_apigatewayv2_api.test.id
  route_id           = aws_apigatewayv2_route.test.id
  route_response_key = "$default"
}
`)
}

func testAccRouteResponseConfig_model(rName string) string {
	return acctest.ConfigCompose(
		testAccRouteConfig_model(rName),
		`
resource "aws_apigatewayv2_route_response" "test" {
  api_id             = aws_apigatewayv2_api.test.id
  route_id           = aws_apigatewayv2_route.test.id
  route_response_key = "$default"

  model_selection_expression = "action"

  response_models = {
    "test" = aws_apigatewayv2_model.test.name
  }
}
`)
}
