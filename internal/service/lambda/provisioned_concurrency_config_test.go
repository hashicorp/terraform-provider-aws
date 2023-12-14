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
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaProvisionedConcurrencyConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyConfigConfig_concurrentExecutions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaFunctionResourceName, "version"),
					resource.TestCheckResourceAttr(resourceName, "skip_destroy", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccLambdaProvisionedConcurrencyConfig_Disappears_lambdaFunction(t *testing.T) {
	ctx := acctest.Context(t)
	var function lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyConfigConfig_concurrentExecutions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, lambdaFunctionResourceName, &function),
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					testAccCheckFunctionDisappears(ctx, &function),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaProvisionedConcurrencyConfig_Disappears_lambdaProvisionedConcurrency(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyConfigConfig_concurrentExecutions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					testAccCheckProvisionedConcurrencyDisappearsConfig(ctx, resourceName),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyConfigConfig_concurrentExecutions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
			{
				Config: testAccProvisionedConcurrencyConfigConfig_concurrentExecutions(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "2"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", "1"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyConfigConfig_FunctionName_arn(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
			{
				Config: testAccProvisionedConcurrencyConfigConfig_FunctionName_arn(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "2"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", "1"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyConfigConfig_qualifierAliasName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaAliasResourceName, "name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccLambdaProvisionedConcurrencyConfig_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	filename1 := "test-fixtures/lambdapinpoint.zip"
	filename2 := "test-fixtures/lambdapinpoint_modified.zip"
	version1 := "1"
	version2 := "2"
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedConcurrencyConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyConfigConfig_skipDestroy(rName, filename1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaFunctionResourceName, "version"),
					resource.TestCheckResourceAttr(resourceName, "skip_destroy", "true"),
				),
			},
			{
				Config: testAccProvisionedConcurrencyConfigConfig_skipDestroy(rName, filename2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					testAccCheckProvisionedConcurrencyConfigExistsByName(ctx, rName, version1), // verify config on previous version still exists
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaFunctionResourceName, "version"),
					resource.TestCheckResourceAttr(resourceName, "skip_destroy", "true"),
				),
			},
			{
				Config: testAccProvisionedConcurrencyConfigConfigBase_withFilename(rName, filename2), // remove the provisioned concurrency config completely
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExistsByName(ctx, rName, version1),
					testAccCheckProvisionedConcurrencyConfigExistsByName(ctx, rName, version2),
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
		ErrorCheck:   acctest.ErrorCheck(t, names.LambdaEndpointID),
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
					testAccCheckProvisionedConcurrencyConfigExists_v0Schema(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaFunctionResourceName, "version"),
					resource.TestCheckResourceAttr(resourceName, "skip_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s:1", rName)),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccProvisionedConcurrencyConfigConfig_concurrentExecutions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaFunctionResourceName, "version"),
					resource.TestCheckResourceAttr(resourceName, "skip_destroy", "false"),
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s,1", rName)),
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

			parts, err := flex.ExpandResourceId(rs.Primary.ID, tflambda.ProvisionedConcurrencyIDPartCount, false)
			if err != nil {
				return err
			}

			input := &lambda.GetProvisionedConcurrencyConfigInput{
				FunctionName: aws.String(parts[0]),
				Qualifier:    aws.String(parts[1]),
			}

			output, err := conn.GetProvisionedConcurrencyConfig(ctx, input)
			if err != nil {
				var pccnfe *types.ProvisionedConcurrencyConfigNotFoundException
				if errors.As(err, &pccnfe) {
					continue
				}
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					continue
				}
				return err
			}

			if output != nil {
				return fmt.Errorf("Lambda Provisioned Concurrency Config (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckProvisionedConcurrencyDisappearsConfig(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		parts, err := flex.ExpandResourceId(rs.Primary.ID, tflambda.ProvisionedConcurrencyIDPartCount, false)
		if err != nil {
			return err
		}

		input := &lambda.DeleteProvisionedConcurrencyConfigInput{
			FunctionName: aws.String(parts[0]),
			Qualifier:    aws.String(parts[1]),
		}

		_, err = conn.DeleteProvisionedConcurrencyConfig(ctx, input)

		return err
	}
}

// testAccCheckProvisionedConcurrencyConfigExists_v0Schema is a variant of the check
// exists functions for v0 schemas.
func testAccCheckProvisionedConcurrencyConfigExists_v0Schema(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		// flex.ExpandResourceId will fail for unmigrated (v0) schemas. For checking existence
		// in the migration test, read the required attributes directly instead.
		functionName, ok := rs.Primary.Attributes["function_name"]
		if !ok {
			return fmt.Errorf("Resource (%s) function_name attribute not set", resourceName)
		}
		qualifier, ok := rs.Primary.Attributes["qualifier"]
		if !ok {
			return fmt.Errorf("Resource (%s) qualifier attribute not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		input := &lambda.GetProvisionedConcurrencyConfigInput{
			FunctionName: aws.String(functionName),
			Qualifier:    aws.String(qualifier),
		}

		output, err := conn.GetProvisionedConcurrencyConfig(ctx, input)

		if err != nil {
			return err
		}

		if got, want := output.Status, types.ProvisionedConcurrencyStatusEnumReady; got != want {
			return fmt.Errorf("Lambda Provisioned Concurrency Config (%s) expected status (%s), got: %s", rs.Primary.ID, want, got)
		}

		return nil
	}
}

func testAccCheckProvisionedConcurrencyConfigExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		parts, err := flex.ExpandResourceId(rs.Primary.ID, tflambda.ProvisionedConcurrencyIDPartCount, false)
		if err != nil {
			return err
		}

		input := &lambda.GetProvisionedConcurrencyConfigInput{
			FunctionName: aws.String(parts[0]),
			Qualifier:    aws.String(parts[1]),
		}

		output, err := conn.GetProvisionedConcurrencyConfig(ctx, input)

		if err != nil {
			return err
		}

		if got, want := output.Status, types.ProvisionedConcurrencyStatusEnumReady; got != want {
			return fmt.Errorf("Lambda Provisioned Concurrency Config (%s) expected status (%s), got: %s", rs.Primary.ID, want, got)
		}

		return nil
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
		input := &lambda.GetProvisionedConcurrencyConfigInput{
			FunctionName: aws.String(functionName),
			Qualifier:    aws.String(qualifier),
		}

		output, err := conn.GetProvisionedConcurrencyConfig(ctx, input)

		if err != nil {
			return err
		}

		if got, want := output.Status, types.ProvisionedConcurrencyStatusEnumReady; got != want {
			return fmt.Errorf("Lambda Provisioned Concurrency Config (%s) expected status (%s), got: %s", functionName, want, got)
		}

		return nil
	}
}

func testAccProvisionedConcurrencyConfigConfigBase(rName string) string {
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
		testAccProvisionedConcurrencyConfigConfigBase(rName),
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
		testAccProvisionedConcurrencyConfigConfigBase(rName),
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
		testAccProvisionedConcurrencyConfigConfigBase(rName),
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
