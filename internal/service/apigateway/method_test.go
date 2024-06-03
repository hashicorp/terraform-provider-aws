// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayMethod_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetMethodOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMethodExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorization", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "request_models.application/json", "Error"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.method.request.header.Content-Type", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.method.request.querystring.page", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccMethodImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"authorizer_id", "operation_name", "request_validator_id"},
			},
			{
				Config: testAccMethodConfig_update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMethodExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorization", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "request_models.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "request_models.application/json", "Error"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.method.request.querystring.page", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccAPIGatewayMethod_customAuthorizer(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetMethodOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodConfig_customAuthorizer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorization", "CUSTOM"),
					resource.TestCheckResourceAttrSet(resourceName, "authorizer_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccMethodImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"operation_name", "request_validator_id"},
			},

			{
				Config: testAccMethodConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorization", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "authorizer_id", ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayMethod_cognitoAuthorizer(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetMethodOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodConfig_cognitoAuthorizer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorization", "COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttrSet(resourceName, "authorizer_id"),
					resource.TestCheckResourceAttr(resourceName, "authorization_scopes.#", acctest.Ct2),
				),
			},

			{
				Config: testAccMethodConfig_cognitoAuthorizerUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorization", "COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttrSet(resourceName, "authorizer_id"),
					resource.TestCheckResourceAttr(resourceName, "authorization_scopes.#", acctest.Ct3),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccMethodImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayMethod_customRequestValidator(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetMethodOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodConfig_customRequestValidator(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "request_validator_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccMethodImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"authorizer_id", "operation_name"},
			},

			{
				Config: testAccMethodConfig_customRequestValidatorUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "request_validator_id", ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayMethod_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetMethodOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceMethod(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayMethod_operationName(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetMethodOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_method.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMethodDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMethodConfig_operationName(rName, "getTest"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "operation_name", "getTest"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccMethodImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"authorizer_id", "request_validator_id"},
			},
			{
				Config: testAccMethodConfig_operationName(rName, "describeTest"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMethodExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "operation_name", "describeTest"),
				),
			},
		},
	})
}

func testAccCheckMethodExists(ctx context.Context, n string, v *apigateway.GetMethodOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindMethodByThreePartKey(ctx, conn, rs.Primary.Attributes["http_method"], rs.Primary.Attributes[names.AttrResourceID], rs.Primary.Attributes["rest_api_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckMethodDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_method" {
				continue
			}

			_, err := tfapigateway.FindMethodByThreePartKey(ctx, conn, rs.Primary.Attributes["http_method"], rs.Primary.Attributes[names.AttrResourceID], rs.Primary.Attributes["rest_api_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Method %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccMethodImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes[names.AttrResourceID], rs.Primary.Attributes["http_method"]), nil
	}
}

func testAccMethodConfig_customAuthorizer(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_iam_role" "invocation_role" {
  name = "%[1]s-invocation"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "invocation_policy" {
  name = %[1]q
  role = aws_iam_role.invocation_role.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "lambda:InvokeFunction",
      "Effect": "Allow",
      "Resource": "${aws_lambda_function.authorizer.arn}"
    }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "authorizer" {
  filename         = "test-fixtures/lambdatest.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  function_name    = %[1]q
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs16.x"
}

resource "aws_api_gateway_authorizer" "test" {
  name                   = %[1]q
  rest_api_id            = aws_api_gateway_rest_api.test.id
  authorizer_uri         = aws_lambda_function.authorizer.invoke_arn
  authorizer_credentials = aws_iam_role.invocation_role.arn
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "CUSTOM"
  authorizer_id = aws_api_gateway_authorizer.test.id

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = {
    "method.request.header.Content-Type" = false
    "method.request.querystring.page"    = true
  }
}
`, rName)
}

func testAccMethodConfig_cognitoAuthorizerBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_iam_role" "invocation_role" {
  name = "%[1]s-invocation"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_cognito_user_pool" "pool" {
  name = %[1]q
}

resource "aws_api_gateway_authorizer" "test" {
  name            = %[1]q
  rest_api_id     = aws_api_gateway_rest_api.test.id
  identity_source = "method.request.header.Authorization"
  provider_arns   = [aws_cognito_user_pool.pool.arn]
  type            = "COGNITO_USER_POOLS"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}
`, rName)
}

func testAccMethodConfig_cognitoAuthorizer(rName string) string {
	return acctest.ConfigCompose(testAccMethodConfig_cognitoAuthorizerBase(rName), `
resource "aws_api_gateway_method" "test" {
  rest_api_id          = aws_api_gateway_rest_api.test.id
  resource_id          = aws_api_gateway_resource.test.id
  http_method          = "GET"
  authorization        = "COGNITO_USER_POOLS"
  authorizer_id        = aws_api_gateway_authorizer.test.id
  authorization_scopes = ["test.read", "test.write"]

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = {
    "method.request.header.Content-Type" = false
    "method.request.querystring.page"    = true
  }
}
`)
}

func testAccMethodConfig_cognitoAuthorizerUpdate(rName string) string {
	return acctest.ConfigCompose(testAccMethodConfig_cognitoAuthorizerBase(rName), `
resource "aws_api_gateway_method" "test" {
  rest_api_id          = aws_api_gateway_rest_api.test.id
  resource_id          = aws_api_gateway_resource.test.id
  http_method          = "GET"
  authorization        = "COGNITO_USER_POOLS"
  authorizer_id        = aws_api_gateway_authorizer.test.id
  authorization_scopes = ["test.read", "test.write", "test.delete"]

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = {
    "method.request.querystring.page" = false
  }
}
`)
}

func testAccMethodConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = {
    "method.request.header.Content-Type" = false
    "method.request.querystring.page"    = true
  }
}
`, rName)
}

func testAccMethodConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = {
    "method.request.querystring.page" = false
  }
}
`, rName)
}

func testAccMethodConfig_customRequestValidatorBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_request_validator" "validator" {
  rest_api_id                 = aws_api_gateway_rest_api.test.id
  name                        = "paramsValidator"
  validate_request_parameters = true
}
`, rName)
}

func testAccMethodConfig_customRequestValidator(rName string) string {
	return acctest.ConfigCompose(testAccMethodConfig_customRequestValidatorBase(rName), `
resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = {
    "method.request.header.Content-Type" = false
    "method.request.querystring.page"    = true
  }

  request_validator_id = aws_api_gateway_request_validator.validator.id
}
`)
}

func testAccMethodConfig_customRequestValidatorUpdate(rName string) string {
	return acctest.ConfigCompose(testAccMethodConfig_customRequestValidatorBase(rName), `
resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = {
    "method.request.querystring.page" = false
  }
}
`)
}

func testAccMethodConfig_operationName(rName, operationName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  authorization  = "NONE"
  http_method    = "GET"
  operation_name = %[2]q
  resource_id    = aws_api_gateway_resource.test.id
  rest_api_id    = aws_api_gateway_rest_api.test.id

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = {
    "method.request.header.Content-Type" = false
    "method.request.querystring.page"    = true
  }
}
`, rName, operationName)
}
