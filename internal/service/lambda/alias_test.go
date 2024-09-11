// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetAliasOutput
	resourceName := "aws_lambda_alias.test"
	rString := sdkacctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_alias_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_alias_basic_%s", rString)
	attachmentName := fmt.Sprintf("tf_acc_attachment_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_alias_basic_%s", rString)
	aliasName := fmt.Sprintf("tf_acc_lambda_alias_basic_%s", rString)
	functionArnResourcePart := fmt.Sprintf("function:%s:%s", funcName, aliasName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_basic(roleName, policyName, attachmentName, funcName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &conf),
					testAccCheckAliasAttributes(&conf),
					testAccCheckAliasRoutingDoesNotExistConfig(&conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", functionArnResourcePart),
					testAccCheckAliasInvokeARN(resourceName, &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAliasImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config:   testAccAliasConfig_usingFunctionName(roleName, policyName, attachmentName, funcName, aliasName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccLambdaAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetAliasOutput
	resourceName := "aws_lambda_alias.test"
	rString := sdkacctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_alias_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_alias_basic_%s", rString)
	attachmentName := fmt.Sprintf("tf_acc_attachment_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_alias_basic_%s", rString)
	aliasName := fmt.Sprintf("tf_acc_lambda_alias_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_basic(roleName, policyName, attachmentName, funcName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflambda.ResourceAlias(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaAlias_FunctionName_name(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetAliasOutput
	resourceName := "aws_lambda_alias.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_usingFunctionName(rName, rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &conf),
					testAccCheckAliasAttributes(&conf),
					testAccCheckAliasRoutingDoesNotExistConfig(&conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", fmt.Sprintf("function:%s:%s", rName, rName)),
					testAccCheckAliasInvokeARN(resourceName, &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAliasImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaAlias_nameUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetAliasOutput
	resourceName := "aws_lambda_alias.test"
	rString := sdkacctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_alias_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_alias_basic_%s", rString)
	attachmentName := fmt.Sprintf("tf_acc_attachment_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_alias_basic_%s", rString)
	aliasName := fmt.Sprintf("tf_acc_lambda_alias_basic_%s", rString)
	aliasNameUpdate := fmt.Sprintf("tf_acc_lambda_alias_basic_%s", sdkacctest.RandString(8))
	functionArnResourcePart := fmt.Sprintf("function:%s:%s", funcName, aliasName)
	functionArnResourcePartUpdate := fmt.Sprintf("function:%s:%s", funcName, aliasNameUpdate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_basic(roleName, policyName, attachmentName, funcName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &conf),
					testAccCheckAliasAttributes(&conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", functionArnResourcePart),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAliasImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAliasConfig_basic(roleName, policyName, attachmentName, funcName, aliasNameUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &conf),
					testAccCheckAliasAttributes(&conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", functionArnResourcePartUpdate),
				),
			},
		},
	})
}

func TestAccLambdaAlias_routing(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetAliasOutput
	resourceName := "aws_lambda_alias.test"
	rString := sdkacctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_alias_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_alias_basic_%s", rString)
	attachmentName := fmt.Sprintf("tf_acc_attachment_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_alias_basic_%s", rString)
	aliasName := fmt.Sprintf("tf_acc_lambda_alias_basic_%s", rString)
	functionArnResourcePart := fmt.Sprintf("function:%s:%s", funcName, aliasName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_basic(roleName, policyName, attachmentName, funcName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &conf),
					testAccCheckAliasAttributes(&conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", functionArnResourcePart),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAliasImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAliasConfig_routing(roleName, policyName, attachmentName, funcName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &conf),
					testAccCheckAliasAttributes(&conf),
					testAccCheckAliasRoutingExistsConfig(&conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", functionArnResourcePart),
				),
			},
			{
				Config: testAccAliasConfig_basic(roleName, policyName, attachmentName, funcName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &conf),
					testAccCheckAliasAttributes(&conf),
					testAccCheckAliasRoutingDoesNotExistConfig(&conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", functionArnResourcePart),
				),
			},
		},
	})
}

func testAccCheckAliasDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_alias" {
				continue
			}

			_, err := tflambda.FindAliasByTwoPartKey(ctx, conn, rs.Primary.Attributes["function_name"], rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lambda Alias %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAliasExists(ctx context.Context, n string, v *lambda.GetAliasOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		output, err := tflambda.FindAliasByTwoPartKey(ctx, conn, rs.Primary.Attributes["function_name"], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAliasAttributes(v *lambda.GetAliasOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := aws.ToString(v.Name)
		arn := aws.ToString(v.AliasArn)
		if arn == "" {
			return fmt.Errorf("Could not read Lambda alias ARN")
		}
		if name == "" {
			return fmt.Errorf("Could not read Lambda alias name")
		}
		return nil
	}
}

func testAccCheckAliasInvokeARN(n string, v *lambda.GetAliasOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arn := aws.ToString(v.AliasArn)
		return acctest.CheckResourceAttrRegionalARNAccountID(n, "invoke_arn", "apigateway", "lambda", fmt.Sprintf("path/2015-03-31/functions/%s/invocations", arn))(s)
	}
}

func testAccCheckAliasRoutingExistsConfig(v *lambda.GetAliasOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routingConfig := v.RoutingConfig

		if routingConfig == nil {
			return fmt.Errorf("Could not read Lambda alias routing config")
		}
		if len(routingConfig.AdditionalVersionWeights) != 1 {
			return fmt.Errorf("Could not read Lambda alias additional version weights")
		}
		return nil
	}
}

func testAccCheckAliasRoutingDoesNotExistConfig(v *lambda.GetAliasOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routingConfig := v.RoutingConfig

		if routingConfig != nil {
			return fmt.Errorf("Lambda alias routing config still exists after removal")
		}
		return nil
	}
}

func testAccAliasImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["function_name"], rs.Primary.Attributes[names.AttrName]), nil
	}
}

func testAccAliasConfig_base(roleName, policyName, attachmentName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
  name = %[1]q

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

resource "aws_iam_policy" "policy_for_role" {
  name        = %[2]q
  path        = "/"
  description = "IAM policy for Lamda alias testing"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "policy_attachment_for_role" {
  name       = %[3]q
  roles      = [aws_iam_role.iam_for_lambda.name]
  policy_arn = aws_iam_policy.policy_for_role.arn
}
`, roleName, policyName, attachmentName)
}

func testAccAliasConfig_basic(roleName, policyName, attachmentName, funcName, aliasName string) string {
	return acctest.ConfigCompose(
		testAccAliasConfig_base(roleName, policyName, attachmentName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  function_name    = %[1]q
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs16.x"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  publish          = "true"
}

resource "aws_lambda_alias" "test" {
  name             = %[2]q
  description      = "a sample description"
  function_name    = aws_lambda_function.test.arn
  function_version = "1"
}
`, funcName, aliasName))
}

func testAccAliasConfig_usingFunctionName(roleName, policyName, attachmentName, funcName, aliasName string) string {
	return acctest.ConfigCompose(
		testAccAliasConfig_base(roleName, policyName, attachmentName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  function_name    = %[1]q
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs16.x"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  publish          = "true"
}

resource "aws_lambda_alias" "test" {
  name             = %[2]q
  description      = "a sample description"
  function_name    = aws_lambda_function.test.function_name
  function_version = "1"
}
`, funcName, aliasName))
}

func testAccAliasConfig_routing(roleName, policyName, attachmentName, funcName, aliasName string) string {
	return acctest.ConfigCompose(
		testAccAliasConfig_base(roleName, policyName, attachmentName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest_modified.zip"
  function_name    = %[1]q
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs16.x"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest_modified.zip")
  publish          = "true"
}

resource "aws_lambda_alias" "test" {
  name             = %[2]q
  description      = "a sample description"
  function_name    = aws_lambda_function.test.arn
  function_version = "1"

  routing_config {
    additional_version_weights = {
      "2" = 0.5
    }
  }
}
`, funcName, aliasName))
}
