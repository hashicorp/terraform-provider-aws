// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"context"
	"fmt"
	"strings"
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

func TestAccAPIGatewayV2RouteResponse_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId, routeId string
	var v apigatewayv2.GetRouteResponseOutput
	resourceName := "aws_apigatewayv2_route_response.test"
	routeResourceName := "aws_apigatewayv2_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteResponseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteResponseConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteResponseExists(ctx, resourceName, &apiId, &routeId, &v),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", ""),
					resource.TestCheckResourceAttr(resourceName, "response_models.%", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "route_id", routeResourceName, names.AttrID),
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
	ctx := acctest.Context(t)
	var apiId, routeId string
	var v apigatewayv2.GetRouteResponseOutput
	resourceName := "aws_apigatewayv2_route_response.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteResponseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteResponseConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteResponseExists(ctx, resourceName, &apiId, &routeId, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigatewayv2.ResourceRouteResponse(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2RouteResponse_model(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId, routeId string
	var v apigatewayv2.GetRouteResponseOutput
	resourceName := "aws_apigatewayv2_route_response.test"
	modelResourceName := "aws_apigatewayv2_model.test"
	routeResourceName := "aws_apigatewayv2_route.test"
	// Model name must be alphanumeric.
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteResponseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteResponseConfig_model(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteResponseExists(ctx, resourceName, &apiId, &routeId, &v),
					resource.TestCheckResourceAttr(resourceName, "model_selection_expression", names.AttrAction),
					resource.TestCheckResourceAttr(resourceName, "response_models.%", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "response_models.test", modelResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "route_id", routeResourceName, names.AttrID),
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

func testAccCheckRouteResponseDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_route_response" {
				continue
			}

			_, err := tfapigatewayv2.FindRouteResponseByThreePartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.Attributes["route_id"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 Route Response %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRouteResponseExists(ctx context.Context, n string, apiID, routeID *string, v *apigatewayv2.GetRouteResponseOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		output, err := tfapigatewayv2.FindRouteResponseByThreePartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.Attributes["route_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*apiID = rs.Primary.Attributes["api_id"]
		*routeID = rs.Primary.Attributes["route_id"]
		*v = *output

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
