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
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
)

func TestAccAPIGatewayV2Route_basic(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Route_disappears(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigatewayv2.ResourceRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Route_authorizer(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	authorizerResourceName := "aws_apigatewayv2_authorizer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_authorizer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_scopes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeCustom),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_id", authorizerResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRouteConfig_authorizerUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_scopes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeAwsIam),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Route_jwtAuthorization(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	authorizerResourceName := "aws_apigatewayv2_authorizer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_jwtAuthorization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_scopes.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeJwt),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_id", authorizerResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "GET /test"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRouteConfig_jwtAuthorizationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_scopes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeJwt),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_id", authorizerResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "GET /test"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Route_model(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	modelResourceName := "aws_apigatewayv2_model.test"
	// Model name must be alphanumeric.
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_model(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", "action"),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "request_models.test", modelResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Route_requestParameters(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_requestParameters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "request_parameter.*", map[string]string{
						"request_parameter_key": "route.request.header.authorization",
						"required":              "true",
					}),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				Config: testAccRouteConfig_requestParametersUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "request_parameter.*", map[string]string{
						"request_parameter_key": "route.request.header.authorization",
						"required":              "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "request_parameter.*", map[string]string{
						"request_parameter_key": "route.request.querystring.authToken",
						"required":              "true",
					}),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRouteConfig_noRequestParameters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Route_simpleAttributes(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_simpleAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "true"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", "GET"),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", "$default"),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				Config: testAccRouteConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				Config: testAccRouteConfig_simpleAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "true"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", "GET"),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", "$default"),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Route_target(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	integrationResourceName := "aws_apigatewayv2_integration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_target(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					testAccCheckRouteTarget(resourceName, integrationResourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Route_updateRouteKey(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_routeKey(rName, "GET /path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "GET /path"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				Config: testAccRouteConfig_routeKey(rName, "POST /new/path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", apigatewayv2.AuthorizationTypeNone),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route_key", "POST /new/path"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "target", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRouteDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_route" {
			continue
		}

		_, err := conn.GetRoute(&apigatewayv2.GetRouteInput{
			ApiId:   aws.String(rs.Primary.Attributes["api_id"]),
			RouteId: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 route %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckRouteExists(n string, vApiId *string, v *apigatewayv2.GetRouteOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 route ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		apiId := aws.String(rs.Primary.Attributes["api_id"])
		resp, err := conn.GetRoute(&apigatewayv2.GetRouteInput{
			ApiId:   apiId,
			RouteId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*vApiId = *apiId
		*v = *resp

		return nil
	}
}

func testAccRouteImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

func testAccCheckRouteTarget(resourceName, integrationResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[integrationResourceName]
		if !ok {
			return fmt.Errorf("Not Found: %s", integrationResourceName)
		}

		return resource.TestCheckResourceAttr(resourceName, "target", fmt.Sprintf("integrations/%s", rs.Primary.ID))(s)
	}
}

func testAccRouteConfig_apiWebSocket(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
`, rName)
}

func testAccRouteConfig_apiHTTP(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
}
`, rName)
}

func testAccRouteConfig_basicWebSocket(rName string) string {
	return acctest.ConfigCompose(
		testAccRouteConfig_apiWebSocket(rName),
		`
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$default"
}
`)
}

func testAccRouteConfig_authorizer(rName string) string {
	return acctest.ConfigCompose(
		testAccAuthorizerConfig_basic(rName),
		`
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$connect"

  authorization_type = "CUSTOM"
  authorizer_id      = aws_apigatewayv2_authorizer.test.id
}
`)
}

func testAccRouteConfig_authorizerUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccAuthorizerConfig_basic(rName),
		`
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$connect"

  authorization_type = "AWS_IAM"
}
`)
}

func testAccRouteConfig_jwtAuthorization(rName string) string {
	return acctest.ConfigCompose(
		testAccAuthorizerConfig_jwt(rName),
		`
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "GET /test"

  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.test.id

  authorization_scopes = ["user.id", "user.email"]
}
`)
}

func testAccRouteConfig_jwtAuthorizationUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccAuthorizerConfig_jwt(rName),
		`
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "GET /test"

  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.test.id

  authorization_scopes = ["user.email"]
}
`)
}

func testAccRouteConfig_model(rName string) string {
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

	return acctest.ConfigCompose(
		testAccModelConfig_basic(rName, schema),
		`
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$default"

  model_selection_expression = "action"

  request_models = {
    "test" = aws_apigatewayv2_model.test.name
  }
}
`)
}

func testAccRouteConfig_noRequestParameters(rName string) string {
	return acctest.ConfigCompose(
		testAccRouteConfig_apiWebSocket(rName),
		`
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$connect"
}
`)
}

func testAccRouteConfig_requestParameters(rName string) string {
	return acctest.ConfigCompose(
		testAccRouteConfig_apiWebSocket(rName),
		`
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$connect"

  request_parameter {
    request_parameter_key = "route.request.header.authorization"
    required              = true
  }
}
`)
}

func testAccRouteConfig_requestParametersUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccRouteConfig_apiWebSocket(rName),
		`
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$connect"

  request_parameter {
    request_parameter_key = "route.request.header.authorization"
    required              = false
  }

  request_parameter {
    request_parameter_key = "route.request.querystring.authToken"
    required              = true
  }
}
`)
}

func testAccRouteConfig_routeKey(rName, routeKey string) string {
	return acctest.ConfigCompose(
		testAccRouteConfig_apiHTTP(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = %[1]q
}
`, routeKey))
}

// Simple attributes - No authorization, models or targets.
func testAccRouteConfig_simpleAttributes(rName string) string {
	return testAccRouteConfig_apiWebSocket(rName) + `
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$default"

  api_key_required                    = true
  operation_name                      = "GET"
  route_response_selection_expression = "$default"
}
`
}

func testAccRouteConfig_target(rName string) string {
	return testAccIntegrationConfig_basic(rName) + `
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$default"

  target = "integrations/${aws_apigatewayv2_integration.test.id}"
}
`
}
