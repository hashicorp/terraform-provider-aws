// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaProvisionedConcurrencyConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyConfigConfig_concurrentExecutions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaFunctionResourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccLambdaProvisionedConcurrencyConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyConfigConfig_concurrentExecutions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflambda.ResourceProvisionedConcurrencyConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaProvisionedConcurrencyConfig_Disappears_lambdaFunction(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyConfigConfig_concurrentExecutions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflambda.ResourceFunction(), lambdaFunctionResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaProvisionedConcurrencyConfig_provisionedConcurrentExecutions(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyConfigConfig_concurrentExecutions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "qualifier", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
			{
				Config: testAccProvisionedConcurrencyConfigConfig_concurrentExecutions(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "qualifier", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccLambdaProvisionedConcurrencyConfig_FunctionName_arn(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_provisioned_concurrency_config.test"
	lambdaFunctionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyConfigConfig_FunctionName_arn(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "qualifier", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
			{
				Config: testAccProvisionedConcurrencyConfigConfig_FunctionName_arn(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "qualifier", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccLambdaProvisionedConcurrencyConfig_Qualifier_aliasName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaAliasResourceName := "aws_lambda_alias.test"
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyConfigConfig_qualifierAliasName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaAliasResourceName, names.AttrName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccLambdaProvisionedConcurrencyConfig_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	filename1 := "test-fixtures/lambdapinpoint.zip"
	filename2 := "test-fixtures/lambdapinpoint_modified.zip"
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyConfigConfig_skipDestroy(rName, filename1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaFunctionResourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
			{
				Config: testAccProvisionedConcurrencyConfigConfig_skipDestroy(rName, filename2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					testAccCheckProvisionedConcurrencyConfigExistsByName(ctx, rName, acctest.Ct1), // verify config on previous version still exists
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaFunctionResourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
			{
				Config: testAccProvisionedConcurrencyConfigConfigBase_withFilename(rName, filename2), // remove the provisioned concurrency config completely
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExistsByName(ctx, rName, acctest.Ct1),
					testAccCheckProvisionedConcurrencyConfigExistsByName(ctx, rName, acctest.Ct2),
				),
			},
		},
	})
}

func TestAccLambdaProvisionedConcurrencyConfig_idMigration530(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.LambdaServiceID),
		CheckDestroy: testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// At v5.3.0 the resource's schema is v0 and id is colon-delimited
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.3.0",
					},
				},
				Config: testAccProvisionedConcurrencyConfigConfig_concurrentExecutions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaFunctionResourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, fmt.Sprintf("%s:1", rName)),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccProvisionedConcurrencyConfigConfig_concurrentExecutions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaFunctionResourceName, names.AttrVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, fmt.Sprintf("%s,1", rName)),
				),
			},
		},
	})
}

func testAccCheckProvisionedConcurrencyConfigDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_provisioned_concurrency_config" {
				continue
			}

			_, err := tflambda.FindProvisionedConcurrencyConfigByTwoPartKey(ctx, conn, rs.Primary.Attributes["function_name"], rs.Primary.Attributes["qualifier"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lambda Provisioned Concurrency Config %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProvisionedConcurrencyConfigExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		_, err := tflambda.FindProvisionedConcurrencyConfigByTwoPartKey(ctx, conn, rs.Primary.Attributes["function_name"], rs.Primary.Attributes["qualifier"])

		return err
	}
}

// testAccCheckProvisionedConcurrencyConfigExistsByName is a helper to verify a
// provisioned concurrency setting is in place on a specific function version.
// This variant of the test check function accepts function name and qualifer arguments
// directly to support skip_destroy checks where the provisioned concurrency configuration
// resource is removed from state, but should still exist remotely.
func testAccCheckProvisionedConcurrencyConfigExistsByName(ctx context.Context, functionName, qualifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		_, err := tflambda.FindProvisionedConcurrencyConfigByTwoPartKey(ctx, conn, functionName, qualifier)

		return err
	}
}

func testAccProvisionedConcurrencyConfigConfig_base(rName string) string {
	return testAccProvisionedConcurrencyConfigConfigBase_withFilename(rName, "test-fixtures/lambdapinpoint.zip")
}

func testAccProvisionedConcurrencyConfigConfigBase_withFilename(rName, filename string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
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
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.test.id
}

resource "aws_lambda_function" "test" {
  function_name = %[1]q
  filename      = %[2]q
  role          = aws_iam_role.test.arn
  handler       = "lambdapinpoint.handler"
  publish       = true
  runtime       = "nodejs16.x"

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, filename)
}

func testAccProvisionedConcurrencyConfigConfig_concurrentExecutions(rName string, provisionedConcurrentExecutions int) string {
	return acctest.ConfigCompose(
		testAccProvisionedConcurrencyConfigConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lambda_provisioned_concurrency_config" "test" {
  function_name                     = aws_lambda_function.test.function_name
  provisioned_concurrent_executions = %[1]d
  qualifier                         = aws_lambda_function.test.version
}
`, provisionedConcurrentExecutions),
	)
}

func testAccProvisionedConcurrencyConfigConfig_FunctionName_arn(rName string, provisionedConcurrentExecutions int) string {
	return acctest.ConfigCompose(
		testAccProvisionedConcurrencyConfigConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lambda_provisioned_concurrency_config" "test" {
  function_name                     = aws_lambda_function.test.arn
  provisioned_concurrent_executions = %[1]d
  qualifier                         = aws_lambda_function.test.version
}
`, provisionedConcurrentExecutions),
	)
}

func testAccProvisionedConcurrencyConfigConfig_qualifierAliasName(rName string) string {
	return acctest.ConfigCompose(
		testAccProvisionedConcurrencyConfigConfig_base(rName),
		`
resource "aws_lambda_alias" "test" {
  function_name    = aws_lambda_function.test.function_name
  function_version = aws_lambda_function.test.version
  name             = "test"
}

resource "aws_lambda_provisioned_concurrency_config" "test" {
  function_name                     = aws_lambda_alias.test.function_name
  provisioned_concurrent_executions = 1
  qualifier                         = aws_lambda_alias.test.name
}
`,
	)
}

func testAccProvisionedConcurrencyConfigConfig_skipDestroy(rName, filename string, skipDestroy bool) string {
	return acctest.ConfigCompose(
		testAccProvisionedConcurrencyConfigConfigBase_withFilename(rName, filename),
		fmt.Sprintf(`
resource "aws_lambda_provisioned_concurrency_config" "test" {
  function_name                     = aws_lambda_function.test.function_name
  provisioned_concurrent_executions = 1
  qualifier                         = aws_lambda_function.test.version

  skip_destroy = %[1]t
}
`, skipDestroy))
}
