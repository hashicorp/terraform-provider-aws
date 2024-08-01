// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccFunctionURLPreCheck(t *testing.T) {
	acctest.PreCheckPartition(t, names.StandardPartitionID)
}

func TestAccLambdaFunctionURL_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionUrlConfigOutput
	resourceName := "aws_lambda_function_url.test"
	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccFunctionURLPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionURLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionURLConfig_basic(funcName, policyName, roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionURLExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.FunctionUrlAuthTypeNone)),
					resource.TestCheckResourceAttr(resourceName, "cors.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrFunctionARN),
					resource.TestCheckResourceAttr(resourceName, "function_name", funcName),
					resource.TestCheckResourceAttrSet(resourceName, "function_url"),
					resource.TestCheckResourceAttr(resourceName, "invoke_mode", "BUFFERED"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", ""),
					resource.TestCheckResourceAttrSet(resourceName, "url_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaFunctionURL_Cors(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionUrlConfigOutput
	resourceName := "aws_lambda_function_url.test"

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccFunctionURLPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionURLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionURLConfig_cors(funcName, policyName, roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionURLExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.FunctionUrlAuthTypeAwsIam)),
					resource.TestCheckResourceAttr(resourceName, "cors.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_credentials", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_headers.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors.0.allow_headers.*", "date"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors.0.allow_headers.*", "keep-alive"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_methods.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors.0.allow_methods.*", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_origins.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors.0.allow_origins.*", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.expose_headers.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors.0.expose_headers.*", "date"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors.0.expose_headers.*", "keep-alive"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.max_age", "86400"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionURLConfig_corsUpdated(funcName, policyName, roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionURLExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.FunctionUrlAuthTypeAwsIam)),
					resource.TestCheckResourceAttr(resourceName, "cors.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_credentials", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_headers.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors.0.allow_headers.*", "x-custom-header"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_methods.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors.0.allow_methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors.0.allow_methods.*", "POST"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_origins.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors.0.allow_origins.*", "https://www.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors.0.allow_origins.*", "http://localhost:60905"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.expose_headers.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors.0.expose_headers.*", "date"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.max_age", "72000"),
				),
			},
			{
				Config: testAccFunctionURLConfig_basic(funcName, policyName, roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionURLExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", string(awstypes.FunctionUrlAuthTypeNone)),
					resource.TestCheckResourceAttr(resourceName, "cors.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccLambdaFunctionURL_Alias(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionUrlConfigOutput
	resourceName := "aws_lambda_function_url.test"

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	aliasName := "live"
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccFunctionURLPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionURLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionURLConfig_alias(funcName, aliasName, policyName, roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionURLExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "function_name", funcName),
					resource.TestCheckResourceAttr(resourceName, "qualifier", aliasName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaFunctionURL_TwoURLs(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionUrlConfigOutput
	latestResourceName := "aws_lambda_function_url.latest"
	liveResourceName := "aws_lambda_function_url.live"
	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	aliasName := "live"
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccFunctionURLPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionURLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionURLConfig_two(funcName, aliasName, policyName, roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionURLExists(ctx, latestResourceName, &conf),
					resource.TestCheckResourceAttr(latestResourceName, "authorization_type", string(awstypes.FunctionUrlAuthTypeNone)),
					resource.TestCheckResourceAttr(latestResourceName, "cors.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(latestResourceName, names.AttrFunctionARN),
					resource.TestCheckResourceAttr(latestResourceName, "function_name", funcName),
					resource.TestCheckResourceAttrSet(latestResourceName, "function_url"),
					resource.TestCheckResourceAttr(latestResourceName, "qualifier", ""),
					resource.TestCheckResourceAttrSet(latestResourceName, "url_id"),

					testAccCheckFunctionURLExists(ctx, liveResourceName, &conf),
					resource.TestCheckResourceAttr(liveResourceName, "authorization_type", string(awstypes.FunctionUrlAuthTypeNone)),
					resource.TestCheckResourceAttr(liveResourceName, "cors.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(liveResourceName, names.AttrFunctionARN),
					resource.TestCheckResourceAttr(liveResourceName, "function_name", funcName),
					resource.TestCheckResourceAttrSet(liveResourceName, "function_url"),
					resource.TestCheckResourceAttr(liveResourceName, "qualifier", "live"),
					resource.TestCheckResourceAttrSet(liveResourceName, "url_id"),
				),
			},
			{
				ResourceName:      latestResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      liveResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaFunctionURL_invokeMode(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionUrlConfigOutput
	resourceName := "aws_lambda_function_url.test"
	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccFunctionURLPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionURLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionURLConfig_invokeMode(funcName, policyName, roleName, "BUFFERED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionURLExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "invoke_mode", "BUFFERED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionURLConfig_invokeMode(funcName, policyName, roleName, "RESPONSE_STREAM"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionURLExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "invoke_mode", "RESPONSE_STREAM"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionURLConfig_basic(funcName, policyName, roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionURLExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "invoke_mode", "BUFFERED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckFunctionURLExists(ctx context.Context, n string, v *lambda.GetFunctionUrlConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		output, err := tflambda.FindFunctionURLByTwoPartKey(ctx, conn, rs.Primary.Attributes["function_name"], rs.Primary.Attributes["qualifier"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckFunctionURLDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_function_url" {
				continue
			}

			_, err := tflambda.FindFunctionURLByTwoPartKey(ctx, conn, rs.Primary.Attributes["function_name"], rs.Primary.Attributes["qualifier"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lambda Function URL %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccFunctionURLConfig_base(policyName, roleName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = %[1]q
  role = aws_iam_role.iam_for_lambda.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "SNS:Publish"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "xray:PutTraceSegments"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = %[2]q

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

`, policyName, roleName)
}

func testAccFunctionURLConfig_basic(funcName, policyName, roleName string) string {
	return acctest.ConfigCompose(testAccFunctionURLConfig_base(policyName, roleName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
}

resource "aws_lambda_function_url" "test" {
  function_name      = aws_lambda_function.test.function_name
  authorization_type = "NONE"
}
`, funcName))
}

func testAccFunctionURLConfig_cors(funcName, policyName, roleName string) string {
	return acctest.ConfigCompose(testAccFunctionURLConfig_base(policyName, roleName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
}

resource "aws_lambda_function_url" "test" {
  function_name      = aws_lambda_function.test.function_name
  authorization_type = "AWS_IAM"

  cors {
    allow_credentials = true
    allow_origins     = ["*"]
    allow_methods     = ["*"]
    allow_headers     = ["date", "keep-alive"]
    expose_headers    = ["keep-alive", "date"]
    max_age           = 86400
  }
}
`, funcName))
}

func testAccFunctionURLConfig_corsUpdated(funcName, policyName, roleName string) string {
	return acctest.ConfigCompose(testAccFunctionURLConfig_base(policyName, roleName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
}

resource "aws_lambda_function_url" "test" {
  function_name      = aws_lambda_function.test.function_name
  authorization_type = "AWS_IAM"

  cors {
    allow_credentials = false
    allow_origins     = ["https://www.example.com", "http://localhost:60905"]
    allow_methods     = ["GET", "POST"]
    allow_headers     = ["x-custom-header"]
    expose_headers    = ["date"]
    max_age           = 72000
  }
}
`, funcName))
}

func testAccFunctionURLConfig_alias(funcName, aliasName, policyName, roleName string) string {
	return acctest.ConfigCompose(testAccFunctionURLConfig_base(policyName, roleName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
  publish       = true
}

resource "aws_lambda_alias" "live" {
  name             = %[2]q
  description      = "a sample description"
  function_name    = aws_lambda_function.test.function_name
  function_version = "1"
}

resource "aws_lambda_function_url" "test" {
  function_name      = aws_lambda_function.test.function_name
  qualifier          = aws_lambda_alias.live.name
  authorization_type = "AWS_IAM"

  cors {
    allow_credentials = true
    allow_origins     = ["*"]
    allow_methods     = ["*"]
    allow_headers     = ["date", "keep-alive"]
    expose_headers    = ["keep-alive", "date"]
    max_age           = 86400
  }
}
`, funcName, aliasName))
}

func testAccFunctionURLConfig_invokeMode(funcName, policyName, roleName, invokeMode string) string {
	return acctest.ConfigCompose(testAccFunctionURLConfig_base(policyName, roleName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
}

resource "aws_lambda_function_url" "test" {
  function_name      = aws_lambda_function.test.function_name
  authorization_type = "NONE"
  invoke_mode        = %[2]q
}
`, funcName, invokeMode))
}

func testAccFunctionURLConfig_two(funcName, aliasName, policyName, roleName string) string {
	return acctest.ConfigCompose(testAccFunctionURLConfig_base(policyName, roleName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
  publish       = true
}

resource "aws_lambda_function_url" "latest" {
  function_name      = aws_lambda_function.test.function_name
  authorization_type = "NONE"
}

resource "aws_lambda_alias" "live" {
  name             = %[2]q
  description      = "a sample description"
  function_name    = aws_lambda_function.test.function_name
  function_version = aws_lambda_function.test.version
}

resource "aws_lambda_function_url" "live" {
  function_name      = aws_lambda_function.test.function_name
  qualifier          = aws_lambda_alias.live.name
  authorization_type = "NONE"
}
`, funcName, aliasName))
}
