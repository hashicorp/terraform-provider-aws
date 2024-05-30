// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayV2Route_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeNone)),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, ""),
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
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigatewayv2.ResourceRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Route_authorizer(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	authorizerResourceName := "aws_apigatewayv2_authorizer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_authorizer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "authorization_scopes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeCustom)),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_id", authorizerResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, ""),
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
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "authorization_scopes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeAwsIam)),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Route_jwtAuthorization(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_jwtAuthorization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "authorization_scopes.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeJwt)),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_id", "aws_apigatewayv2_authorizer.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route_key", "GET /test"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, ""),
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
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "authorization_scopes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeJwt)),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_id", "aws_apigatewayv2_authorizer.another", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route_key", "GET /test"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Route_model(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	modelResourceName := "aws_apigatewayv2_model.test"
	// Model name must be alphanumeric.
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_model(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeNone)),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", names.AttrAction),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "request_models.test", modelResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, ""),
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
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_requestParameters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeNone)),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "request_parameter.*", map[string]string{
						"request_parameter_key": "route.request.header.authorization",
						"required":              acctest.CtTrue,
					}),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, ""),
				),
			},
			{
				Config: testAccRouteConfig_requestParametersUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeNone)),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "request_parameter.*", map[string]string{
						"request_parameter_key": "route.request.header.authorization",
						"required":              acctest.CtFalse,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "request_parameter.*", map[string]string{
						"request_parameter_key": "route.request.querystring.authToken",
						"required":              acctest.CtTrue,
					}),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, ""),
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
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeNone)),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$connect"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Route_simpleAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_simpleAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeNone)),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", "GET"),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", "$default"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, ""),
				),
			},
			{
				Config: testAccRouteConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeNone)),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, ""),
				),
			},
			{
				Config: testAccRouteConfig_simpleAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeNone)),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", "GET"),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route_key", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", "$default"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, ""),
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
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	integrationResourceName := "aws_apigatewayv2_integration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_target(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeNone)),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct0),
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
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetRouteOutput
	resourceName := "aws_apigatewayv2_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteConfig_key(rName, "GET /path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeNone)),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route_key", "GET /path"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, ""),
				),
			},
			{
				Config: testAccRouteConfig_key(rName, "POST /new/path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.AuthorizationTypeNone)),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_name", ""),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "request_parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route_key", "POST /new/path"),
					resource.TestCheckResourceAttr(resourceName, "route_response_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, ""),
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

func testAccCheckRouteDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_route" {
				continue
			}

			_, err := tfapigatewayv2.FindRouteByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 Route %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRouteExists(ctx context.Context, n string, apiID *string, v *apigatewayv2.GetRouteOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		output, err := tfapigatewayv2.FindRouteByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*apiID = rs.Primary.Attributes["api_id"]
		*v = *output

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

		return resource.TestCheckResourceAttr(resourceName, names.AttrTarget, fmt.Sprintf("integrations/%s", rs.Primary.ID))(s)
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
resource "aws_apigatewayv2_authorizer" "another" {
  api_id           = aws_apigatewayv2_api.test.id
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]
  name             = "another-authorizer"

  jwt_configuration {
    audience = ["test"]
    issuer   = "https://${aws_cognito_user_pool.test.endpoint}"
  }
}

resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "GET /test"

  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.another.id

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

func testAccRouteConfig_key(rName, routeKey string) string {
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
	return acctest.ConfigCompose(testAccRouteConfig_apiWebSocket(rName), `
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$default"

  api_key_required                    = true
  operation_name                      = "GET"
  route_response_selection_expression = "$default"
}
`)
}

func testAccRouteConfig_target(rName string) string {
	return acctest.ConfigCompose(testAccIntegrationConfig_basic(rName), `
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "$default"

  target = "integrations/${aws_apigatewayv2_integration.test.id}"
}
`)
}
