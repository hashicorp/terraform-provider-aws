// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaFunctionRecursionConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var recursionConfig lambda.GetFunctionRecursionConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_recursion_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionRecursionConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionRecursionConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionRecursionConfigExists(ctx, resourceName, &recursionConfig),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, "recursive_loop", string(awstypes.RecursiveLoopTerminate)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccCheckFunctionRecursionConfigImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "function_name",
			},
		},
	})
}

func TestAccLambdaFunctionRecursionConfig_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var recursionConfig1, recursionConfig2 lambda.GetFunctionRecursionConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_recursion_config.test"

	updatedRecursiveLoopAllow := string(awstypes.RecursiveLoopAllow)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionRecursionConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionRecursionConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionRecursionConfigExists(ctx, resourceName, &recursionConfig1),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, "recursive_loop", string(awstypes.RecursiveLoopTerminate)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccCheckFunctionRecursionConfigImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "function_name",
			},
			{
				Config: testAccFunctionRecursionConfigConfig_updateRecursiveLoop(rName, updatedRecursiveLoopAllow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionRecursionConfigExists(ctx, resourceName, &recursionConfig2),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, "recursive_loop", updatedRecursiveLoopAllow),
				),
			},
		},
	})
}

func TestAccLambdaFunctionRecursionConfig_disappears_Function(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var recursionConfig lambda.GetFunctionRecursionConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	functionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_recursion_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionRecursionConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionRecursionConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionRecursionConfigExists(ctx, resourceName, &recursionConfig),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflambda.ResourceFunction(), functionResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFunctionRecursionConfigDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_function_recursion_config" {
				continue
			}

			functionName := rs.Primary.Attributes["function_name"]
			_, err := tflambda.FindFunctionRecursionConfigByName(ctx, conn, functionName)
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Lambda, create.ErrActionCheckingDestroyed, tflambda.ResNameFunctionRecursionConfig, functionName, err)
			}

			return create.Error(names.Lambda, create.ErrActionCheckingDestroyed, tflambda.ResNameFunctionRecursionConfig, functionName, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckFunctionRecursionConfigExists(ctx context.Context, name string, recursionconfig *lambda.GetFunctionRecursionConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameFunctionRecursionConfig, name, errors.New("not found"))
		}

		functionName := rs.Primary.Attributes["function_name"]
		if functionName == "" {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameFunctionRecursionConfig, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		resp, err := tflambda.FindFunctionRecursionConfigByName(ctx, conn, functionName)
		if err != nil {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameFunctionRecursionConfig, functionName, err)
		}

		recursionconfig = resp

		return nil
	}
}

func testAccCheckFunctionRecursionConfigImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["function_name"], nil
	}
}

func testAccFunctionRecursionConfigConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  function_name    = %[1]q
  role             = aws_iam_role.test.arn
  runtime          = "nodejs20.x"
  handler          = "index.handler"
}
`, rName)
}

func testAccFunctionRecursionConfigConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccFunctionRecursionConfigConfigBase(rName),
		`
resource "aws_lambda_function_recursion_config" "test" {
  function_name  = aws_lambda_function.test.function_name
  recursive_loop = "Terminate"
}
`)
}

func testAccFunctionRecursionConfigConfig_updateRecursiveLoop(rName, recursiveLoop string) string {
	return acctest.ConfigCompose(
		testAccFunctionRecursionConfigConfigBase(rName),
		fmt.Sprintf(`
resource "aws_lambda_function_recursion_config" "test" {
  function_name  = aws_lambda_function.test.function_name
  recursive_loop = %[1]q
}
`, recursiveLoop))
}
