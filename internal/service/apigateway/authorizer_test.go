// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccAPIGatewayAuthorizer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetAuthorizerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_authorizer.test"
	lambdaResourceName := "aws_lambda_function.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_lambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/restapis/.+/authorizers/.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "identity_source", "method.request.header.Authorization"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "TOKEN"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials", roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", strconv.Itoa(tfapigateway.DefaultAuthorizerTTL)),
					resource.TestCheckResourceAttr(resourceName, "identity_validation_expression", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAuthorizerImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAuthorizerConfig_lambdaUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttr(resourceName, "identity_source", "method.request.header.Authorization"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"_modified"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "TOKEN"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials", roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", strconv.Itoa(360)),
					resource.TestCheckResourceAttr(resourceName, "identity_validation_expression", ".*"),
				),
			},
		},
	})
}

func TestAccAPIGatewayAuthorizer_cognito(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_authorizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_cognito(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "provider_arns.#", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAuthorizerImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAuthorizerConfig_cognitoUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "provider_arns.#", acctest.Ct3),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/16613
func TestAccAPIGatewayAuthorizer_Cognito_authorizerCredentials(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_authorizer.test"
	iamRoleResourceName := "aws_iam_role.lambda"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_cognitoCredentials(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "provider_arns.#", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAuthorizerImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayAuthorizer_switchAuthType(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_authorizer.test"
	lambdaResourceName := "aws_lambda_function.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_lambda(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "TOKEN"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials", roleResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAuthorizerImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAuthorizerConfig_cognito(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "COGNITO_USER_POOLS"),
					resource.TestCheckResourceAttr(resourceName, "provider_arns.#", acctest.Ct2),
				),
			},
			{
				Config: testAccAuthorizerConfig_lambdaUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"_modified"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "TOKEN"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_uri", lambdaResourceName, "invoke_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "authorizer_credentials", roleResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccAPIGatewayAuthorizer_switchAuthorizerTTL(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetAuthorizerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_authorizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_lambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", strconv.Itoa(tfapigateway.DefaultAuthorizerTTL)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAuthorizerImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAuthorizerConfig_lambdaUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", strconv.Itoa(360)),
				),
			},
			{
				Config: testAccAuthorizerConfig_lambdaNoCache(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", strconv.Itoa(0)),
				),
			},
			{
				Config: testAccAuthorizerConfig_lambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", strconv.Itoa(tfapigateway.DefaultAuthorizerTTL)),
				),
			},
		},
	})
}

func TestAccAPIGatewayAuthorizer_authTypeValidation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccAuthorizerConfig_authTypeValidationDefaultToken(rName),
				ExpectError: regexache.MustCompile(`authorizer_uri must be set non-empty when authorizer type is TOKEN`),
			},
			{
				Config:      testAccAuthorizerConfig_authTypeValidationRequest(rName),
				ExpectError: regexache.MustCompile(`authorizer_uri must be set non-empty when authorizer type is REQUEST`),
			},
			{
				Config:      testAccAuthorizerConfig_authTypeValidationCognito(rName),
				ExpectError: regexache.MustCompile(`provider_arns must be set non-empty when authorizer type is COGNITO_USER_POOLS`),
			},
		},
	})
}

func TestAccAPIGatewayAuthorizer_Zero_ttl(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetAuthorizerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_authorizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_lambdaNoCache(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorizer_result_ttl_in_seconds", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAuthorizerImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayAuthorizer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetAuthorizerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_authorizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthorizerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerConfig_lambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthorizerExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceAuthorizer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAuthorizerExists(ctx context.Context, n string, v *apigateway.GetAuthorizerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindAuthorizerByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["rest_api_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAuthorizerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_authorizer" {
				continue
			}

			_, err := tfapigateway.FindAuthorizerByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["rest_api_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Authorizer %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAuthorizerImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.ID), nil
	}
}

func testAccAuthorizerConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q
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

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "lambda:InvokeFunction",
      "Effect": "Allow",
      "Resource": "${aws_lambda_function.test.arn}"
    }
  ]
}
EOF
}

resource "aws_iam_role" "lambda" {
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

resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  function_name    = %[1]q
  role             = aws_iam_role.lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs16.x"
}
`, rName)
}

func testAccAuthorizerConfig_lambda(rName string) string {
	return acctest.ConfigCompose(testAccAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_authorizer" "test" {
  name                   = %[1]q
  rest_api_id            = aws_api_gateway_rest_api.test.id
  authorizer_uri         = aws_lambda_function.test.invoke_arn
  authorizer_credentials = aws_iam_role.test.arn
}
`, rName))
}

func testAccAuthorizerConfig_lambdaUpdate(rName string) string {
	return acctest.ConfigCompose(testAccAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_authorizer" "test" {
  name                             = "%[1]s_modified"
  rest_api_id                      = aws_api_gateway_rest_api.test.id
  authorizer_uri                   = aws_lambda_function.test.invoke_arn
  authorizer_credentials           = aws_iam_role.test.arn
  authorizer_result_ttl_in_seconds = 360
  identity_validation_expression   = ".*"
}
`, rName))
}

func testAccAuthorizerConfig_lambdaNoCache(rName string) string {
	return acctest.ConfigCompose(testAccAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_authorizer" "test" {
  name                             = "%[1]s_modified"
  rest_api_id                      = aws_api_gateway_rest_api.test.id
  authorizer_uri                   = aws_lambda_function.test.invoke_arn
  authorizer_credentials           = aws_iam_role.test.arn
  authorizer_result_ttl_in_seconds = 0
  identity_validation_expression   = ".*"
}
`, rName))
}

func testAccAuthorizerConfig_cognito(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool" "test" {
  count = 2
  name  = "%[1]s-${count.index}"
}

resource "aws_api_gateway_authorizer" "test" {
  name          = %[1]q
  type          = "COGNITO_USER_POOLS"
  rest_api_id   = aws_api_gateway_rest_api.test.id
  provider_arns = aws_cognito_user_pool.test[*].arn
}
`, rName)
}

func testAccAuthorizerConfig_cognitoUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool" "test" {
  count = 3
  name  = "%[1]s-${count.index}"
}

resource "aws_api_gateway_authorizer" "test" {
  name          = %[1]q
  type          = "COGNITO_USER_POOLS"
  rest_api_id   = aws_api_gateway_rest_api.test.id
  provider_arns = aws_cognito_user_pool.test[*].arn
}
`, rName)
}

func testAccAuthorizerConfig_cognitoCredentials(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool" "test" {
  count = 2
  name  = "%[1]s-${count.index}"
}

resource "aws_api_gateway_authorizer" "test" {
  authorizer_credentials = aws_iam_role.lambda.arn
  name                   = %[1]q
  type                   = "COGNITO_USER_POOLS"
  rest_api_id            = aws_api_gateway_rest_api.test.id
  provider_arns          = aws_cognito_user_pool.test[*].arn
}
`, rName)
}

func testAccAuthorizerConfig_authTypeValidationDefaultToken(rName string) string {
	return acctest.ConfigCompose(testAccAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_authorizer" "test" {
  name                   = %[1]q
  rest_api_id            = aws_api_gateway_rest_api.test.id
  authorizer_credentials = aws_iam_role.test.arn
}
`, rName))
}

func testAccAuthorizerConfig_authTypeValidationRequest(rName string) string {
	return acctest.ConfigCompose(testAccAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_authorizer" "test" {
  name                   = %[1]q
  type                   = "REQUEST"
  rest_api_id            = aws_api_gateway_rest_api.test.id
  authorizer_credentials = aws_iam_role.test.arn
}
`, rName))
}

func testAccAuthorizerConfig_authTypeValidationCognito(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool" "test" {
  count = 2
  name  = "%[1]s-${count.index}"
}

resource "aws_api_gateway_authorizer" "test" {
  name        = %[1]q
  type        = "COGNITO_USER_POOLS"
  rest_api_id = aws_api_gateway_rest_api.test.id
}
`, rName)
}
