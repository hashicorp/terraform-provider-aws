// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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

func TestAccLambdaRecursionConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var recursionConfig lambda.GetFunctionRecursionConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_recursion_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecursionConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecursionConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecursionConfigExists(ctx, resourceName, &recursionConfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, rName),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, "recursive_loop", string(awstypes.RecursiveLoopTerminate)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccCheckRecursionConfigImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "function_name",
			},
		},
	})
}

func TestAccLambdaRecursionConfig_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var recursionConfig1, recursionConfig2 lambda.GetFunctionRecursionConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_recursion_config.test"

	updatedRecursiveLoopAllow := string(awstypes.RecursiveLoopAllow)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecursionConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecursionConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecursionConfigExists(ctx, resourceName, &recursionConfig1),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, rName),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, "recursive_loop", string(awstypes.RecursiveLoopTerminate)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccCheckRecursionConfigImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "function_name",
			},
			{
				Config: testAccRecursionConfigConfig_updateRecursiveLoop(rName, updatedRecursiveLoopAllow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecursionConfigExists(ctx, resourceName, &recursionConfig2),
					testAccCheckRecursionConfigNotRecreated(&recursionConfig1, &recursionConfig2),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, rName),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, "recursive_loop", updatedRecursiveLoopAllow),
				),
			},
		},
	})
}

func TestAccLambdaRecursionConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var recursionConfig lambda.GetFunctionRecursionConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	functionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_recursion_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecursionConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecursionConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRecursionConfigExists(ctx, resourceName, &recursionConfig),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflambda.ResourceFunction(), functionResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRecursionConfigDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_recursion_config" {
				continue
			}

			input := &lambda.GetFunctionRecursionConfigInput{
				FunctionName: aws.String(rs.Primary.ID),
			}
			_, err := conn.GetFunctionRecursionConfig(ctx, input)
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Lambda, create.ErrActionCheckingDestroyed, tflambda.ResNameRecursionConfig, rs.Primary.ID, err)
			}

			return create.Error(names.Lambda, create.ErrActionCheckingDestroyed, tflambda.ResNameRecursionConfig, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRecursionConfigExists(ctx context.Context, name string, recursionconfig *lambda.GetFunctionRecursionConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameRecursionConfig, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameRecursionConfig, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)
		resp, err := conn.GetFunctionRecursionConfig(ctx, &lambda.GetFunctionRecursionConfigInput{
			FunctionName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameRecursionConfig, rs.Primary.ID, err)
		}

		recursionconfig = resp

		return nil
	}
}

func testAccCheckRecursionConfigImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["function_name"], nil
	}
}

func testAccCheckRecursionConfigNotRecreated(before, after *lambda.GetFunctionRecursionConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := string(before.RecursiveLoop), string(after.RecursiveLoop); before != after {
			return create.Error(names.Lambda, create.ErrActionCheckingNotRecreated, tflambda.ResNameRecursionConfig, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccRecursionConfigConfigBase(rName string) string {
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
  runtime          = "nodejs16.x"
  handler          = "index.handler"
}
`, rName)
}

func testAccRecursionConfigConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccRecursionConfigConfigBase(rName),
		`
resource "aws_lambda_recursion_config" "test" {
  function_name  = aws_lambda_function.test.function_name
  recursive_loop = "Terminate"
}
`)
}

func testAccRecursionConfigConfig_updateRecursiveLoop(rName, resRecursiveLoopValue string) string {
	return acctest.ConfigCompose(
		testAccRecursionConfigConfigBase(rName),
		fmt.Sprintf(`
resource "aws_lambda_recursion_config" "test" {
  function_name  = aws_lambda_function.test.function_name
  recursive_loop = %[1]q
}
`, resRecursiveLoopValue))
}
