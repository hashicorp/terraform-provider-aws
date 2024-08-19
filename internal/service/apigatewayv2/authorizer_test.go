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

func TestAccAPIGatewayV2Authorizer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetAuthorizerOutput
	resourceName := "aws_apigatewayv2_authorizer.test"
	lambdaResourceName := "aws_lambda_function.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_payload_format_version", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "enable_simple_responses", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccAuthorizerConfig_httpNoAuthenticationSources(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAuthorizerImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Authorizer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetAuthorizerOutput
	resourceName := "aws_apigatewayv2_authorizer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &apiId, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigatewayv2.ResourceAuthorizer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Authorizer_credentials(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetAuthorizerOutput
	resourceName := "aws_apigatewayv2_authorizer.test"
	iamRoleResourceName := "aws_iam_role.test"
	lambdaResourceName := "aws_lambda_function.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_credentials(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "authorizer_payload_format_version", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "enable_simple_responses", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "identity_sources.*", "route.request.header.Auth"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAuthorizerImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAuthorizerConfig_credentialsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_payload_format_version", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "enable_simple_responses", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "identity_sources.*", "route.request.header.Auth"),
					resource.TestCheckTypeSetElemAttr(resourceName, "identity_sources.*", "route.request.querystring.Name"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("%s-updated", rName)),
				),
			},
			{
				Config: testAccAuthorizerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_payload_format_version", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "enable_simple_responses", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Authorizer_jwt(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetAuthorizerOutput
	resourceName := "aws_apigatewayv2_authorizer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_jwt(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_payload_format_version", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "JWT"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_simple_responses", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "identity_sources.*", "$request.header.Authorization"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.0.audience.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "jwt_configuration.0.audience.*", "test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAuthorizerImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAuthorizerConfig_jwtUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_payload_format_version", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "JWT"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_simple_responses", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "identity_sources.*", "$request.header.Authorization"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.0.audience.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "jwt_configuration.0.audience.*", "test"),
					resource.TestCheckTypeSetElemAttr(resourceName, "jwt_configuration.0.audience.*", "testing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Authorizer_HTTPAPILambdaRequestAuthorizer_initialMissingCacheTTL(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetAuthorizerOutput
	resourceName := "aws_apigatewayv2_authorizer.test"
	lambdaResourceName := "aws_lambda_function.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_httpAPILambdaRequest(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_payload_format_version", "2.0"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "enable_simple_responses", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "identity_sources.*", "$request.header.Auth"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAuthorizerImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAuthorizerConfig_httpAPILambdaRequestUpdated(rName, 3600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "enable_simple_responses", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "identity_sources.*", "$request.querystring.User"),
					resource.TestCheckTypeSetElemAttr(resourceName, "identity_sources.*", "$context.routeKey"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccAuthorizerConfig_httpAPILambdaRequestUpdated(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "enable_simple_responses", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "identity_sources.*", "$request.querystring.User"),
					resource.TestCheckTypeSetElemAttr(resourceName, "identity_sources.*", "$context.routeKey"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Authorizer_HTTPAPILambdaRequestAuthorizer_initialZeroCacheTTL(t *testing.T) {
	ctx := acctest.Context(t)
	var apiId string
	var v apigatewayv2.GetAuthorizerOutput
	resourceName := "aws_apigatewayv2_authorizer.test"
	lambdaResourceName := "aws_lambda_function.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_httpAPILambdaRequestUpdated(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "enable_simple_responses", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "identity_sources.*", "$request.querystring.User"),
					resource.TestCheckTypeSetElemAttr(resourceName, "identity_sources.*", "$context.routeKey"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAuthorizerImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAuthorizerConfig_httpAPILambdaRequestUpdated(rName, 600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "authorizer_credentials_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "authorizer_payload_format_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", "600"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_type", "REQUEST"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "enable_simple_responses", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "identity_sources.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "identity_sources.*", "$request.querystring.User"),
					resource.TestCheckTypeSetElemAttr(resourceName, "identity_sources.*", "$context.routeKey"),
					resource.TestCheckResourceAttr(resourceName, "jwt_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func testAccCheckAuthorizerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_authorizer" {
				continue
			}

			_, err := tfapigatewayv2.FindAuthorizerByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 Authorizer %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAuthorizerExists(ctx context.Context, n string, apiID *string, v *apigatewayv2.GetAuthorizerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		output, err := tfapigatewayv2.FindAuthorizerByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*apiID = rs.Primary.Attributes["api_id"]
		*v = *output

		return nil
	}
}

func testAccAuthorizerImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

func testAccAuthorizerConfig_apiWebSocket(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
`, rName)
}

func testAccAuthorizerConfig_apiHTTP(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
}
`, rName)
}

func testAccAuthorizerConfig_baseLambda(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLambdaBase(rName, rName, rName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"
}

resource "aws_iam_role" "test" {
  name = "%[1]s_auth_invocation_role"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {"Service": "apigateway.amazonaws.com"},
    "Effect": "Allow"
  }]
}
EOF
}
`, rName))
}

func testAccAuthorizerConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAuthorizerConfig_apiWebSocket(rName),
		testAccAuthorizerConfig_baseLambda(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_authorizer" "test" {
  api_id          = aws_apigatewayv2_api.test.id
  authorizer_type = "REQUEST"
  authorizer_uri  = aws_lambda_function.test.invoke_arn
  name            = %[1]q
}
`, rName))
}

func testAccAuthorizerConfig_credentials(rName string) string {
	return acctest.ConfigCompose(
		testAccAuthorizerConfig_apiWebSocket(rName),
		testAccAuthorizerConfig_baseLambda(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_authorizer" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  authorizer_type  = "REQUEST"
  authorizer_uri   = aws_lambda_function.test.invoke_arn
  identity_sources = ["route.request.header.Auth"]
  name             = %[1]q

  authorizer_credentials_arn = aws_iam_role.test.arn
}
`, rName))
}

func testAccAuthorizerConfig_credentialsUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccAuthorizerConfig_apiWebSocket(rName),
		testAccAuthorizerConfig_baseLambda(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_authorizer" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  authorizer_type  = "REQUEST"
  authorizer_uri   = aws_lambda_function.test.invoke_arn
  identity_sources = ["route.request.header.Auth", "route.request.querystring.Name"]
  name             = "%[1]s-updated"

  authorizer_credentials_arn = aws_iam_role.test.arn
}
`, rName))
}

func testAccAuthorizerConfig_jwt(rName string) string {
	return acctest.ConfigCompose(
		testAccAuthorizerConfig_apiHTTP(rName),
		testAccAuthorizerConfig_baseLambda(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_apigatewayv2_authorizer" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]
  name             = %[1]q

  jwt_configuration {
    audience = ["test"]
    issuer   = "https://${aws_cognito_user_pool.test.endpoint}"
  }
}
`, rName))
}

func testAccAuthorizerConfig_jwtUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccAuthorizerConfig_apiHTTP(rName),
		testAccAuthorizerConfig_baseLambda(rName),
		fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_apigatewayv2_authorizer" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]
  name             = %[1]q

  jwt_configuration {
    audience = ["test", "testing"]
    issuer   = "https://${aws_cognito_user_pool.test.endpoint}"
  }
}
`, rName))
}

func testAccAuthorizerConfig_httpAPILambdaRequest(rName string) string {
	return acctest.ConfigCompose(
		testAccAuthorizerConfig_apiHTTP(rName),
		testAccAuthorizerConfig_baseLambda(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_authorizer" "test" {
  api_id                            = aws_apigatewayv2_api.test.id
  authorizer_payload_format_version = "2.0"
  authorizer_type                   = "REQUEST"
  authorizer_uri                    = aws_lambda_function.test.invoke_arn
  enable_simple_responses           = true
  identity_sources                  = ["$request.header.Auth"]
  name                              = %[1]q
}
`, rName))
}

func testAccAuthorizerConfig_httpNoAuthenticationSources(rName string) string {
	return acctest.ConfigCompose(
		testAccAuthorizerConfig_apiHTTP(rName),
		testAccAuthorizerConfig_baseLambda(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_authorizer" "test" {
  api_id                            = aws_apigatewayv2_api.test.id
  authorizer_payload_format_version = "2.0"
  authorizer_type                   = "REQUEST"
  authorizer_uri                    = aws_lambda_function.test.invoke_arn
  enable_simple_responses           = true
  name                              = %[1]q
}
`, rName))
}

func testAccAuthorizerConfig_httpAPILambdaRequestUpdated(rName string, authorizerResultTtl int) string {
	return acctest.ConfigCompose(
		testAccAuthorizerConfig_apiHTTP(rName),
		testAccAuthorizerConfig_baseLambda(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_authorizer" "test" {
  api_id                            = aws_apigatewayv2_api.test.id
  authorizer_payload_format_version = "1.0"
  authorizer_result_ttl_in_seconds  = %[2]d
  authorizer_type                   = "REQUEST"
  authorizer_uri                    = aws_lambda_function.test.invoke_arn
  enable_simple_responses           = false
  identity_sources                  = ["$request.querystring.User", "$context.routeKey"]
  name                              = %[1]q
}
`, rName, authorizerResultTtl))
}
