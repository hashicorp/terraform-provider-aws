// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecretsmanager "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSecretsManagerSecretRotation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_rotation.test"
	lambdaFunctionResourceName := "aws_lambda_function.test1"
	days01 := 7
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretRotationDestroy(ctx),
		Steps: []resource.TestStep{
			// Test creating secret rotation resource
			{
				Config: testAccSecretRotationConfig_basic(rName, days01),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.automatically_after_days", strconv.Itoa(days01)),
				),
			},
			// Test updating rotation
			// We need a valid rotation function for this testing
			// InvalidRequestException: A previous rotation isn’t complete. That rotation will be reattempted.
			/*
				{
					Config: testAccSecretRotationConfig_managerUpdated(rName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSecretRotationExists(resourceName, &secret),
						resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
						resource.TestMatchResourceAttr(resourceName, "rotation_lambda_arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:lambda:[^:]+:[^:]+:function:%s-2$", rName))),
					),
				},
			*/
			// Test importing secret rotation
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSecretsManagerSecretRotation_scheduleExpression(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_rotation.test"
	lambdaFunctionResourceName := "aws_lambda_function.test1"
	scheduleExpression := "rate(10 days)"
	scheduleExpression02 := "rate(10 days)"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretRotationDestroy(ctx),
		Steps: []resource.TestStep{
			// Test creating secret rotation resource
			{
				Config: testAccSecretRotationConfig_scheduleExpression(rName, scheduleExpression),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.schedule_expression", scheduleExpression),
				),
			},
			{
				Config: testAccSecretRotationConfig_scheduleExpression(rName, scheduleExpression02),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.schedule_expression", scheduleExpression02),
				),
			},
			// Test importing secret rotation
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSecretsManagerSecretRotation_scheduleExpressionHours(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_rotation.test"
	lambdaFunctionResourceName := "aws_lambda_function.test1"
	scheduleExpression := "rate(6 hours)"
	scheduleExpression02 := "rate(10 hours)"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretRotationDestroy(ctx),
		Steps: []resource.TestStep{
			// Test creating secret rotation resource
			{
				Config: testAccSecretRotationConfig_scheduleExpression(rName, scheduleExpression),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.schedule_expression", scheduleExpression),
					testSecretValueIsCurrent(ctx, rName),
				),
			},
			{
				Config: testAccSecretRotationConfig_scheduleExpression(rName, scheduleExpression02),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.schedule_expression", scheduleExpression02),
				),
			},
			// Test importing secret rotation
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSecretsManagerSecretRotation_duration(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_rotation.test"
	lambdaFunctionResourceName := "aws_lambda_function.test1"
	days01 := 7
	duration := "3h"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretRotationDestroy(ctx),
		Steps: []resource.TestStep{
			// Test creating secret rotation resource
			{
				Config: testAccSecretRotationConfig_duration(rName, days01, duration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.automatically_after_days", strconv.Itoa(days01)),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.duration", duration),
				),
			},
			// Test importing secret rotation
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckSecretRotationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_secretsmanager_secret_rotation" {
				continue
			}

			output, err := tfsecretsmanager.FindSecretByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if !aws.BoolValue(output.RotationEnabled) {
				continue
			}

			return fmt.Errorf("Secrets Manager Secret %s rotation still enabled", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSecretRotationExists(ctx context.Context, n string, v *secretsmanager.DescribeSecretOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Secrets Manager Secret Rotation ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerConn(ctx)

		output, err := tfsecretsmanager.FindSecretByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if !aws.BoolValue(output.RotationEnabled) {
			return fmt.Errorf("Secrets Manager Secret %s rotation not enabled", rs.Primary.ID)
		}

		*v = *output

		return nil
	}
}

func testAccSecretRotationConfig_basic(rName string, automaticallyAfterDays int) string {
	return acctest.ConfigCompose(acctest.ConfigLambdaBase(rName, rName, rName), fmt.Sprintf(`
# Not a real rotation function
resource "aws_lambda_function" "test1" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-1"
  handler       = "exports.example"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs16.x"
}

resource "aws_lambda_permission" "test1" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test1.function_name
  principal     = "secretsmanager.amazonaws.com"
  statement_id  = "AllowExecutionFromSecretsManager1"
}

# Not a real rotation function
resource "aws_lambda_function" "test2" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-2"
  handler       = "exports.example"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs16.x"
}

resource "aws_lambda_permission" "test2" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test2.function_name
  principal     = "secretsmanager.amazonaws.com"
  statement_id  = "AllowExecutionFromSecretsManager2"
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_rotation" "test" {
  secret_id           = aws_secretsmanager_secret.test.id
  rotation_lambda_arn = aws_lambda_function.test1.arn

  rotation_rules {
    automatically_after_days = %[2]d
  }

  depends_on = [aws_lambda_permission.test1]
}
`, rName, automaticallyAfterDays))
}

func testAccSecretRotationConfig_scheduleExpression(rName string, scheduleExpression string) string {
	return acctest.ConfigCompose(acctest.ConfigLambdaBase(rName, rName, rName), fmt.Sprintf(`
# Not a real rotation function
resource "aws_lambda_function" "test1" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-1"
  handler       = "exports.example"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs16.x"
}

resource "aws_lambda_permission" "test1" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test1.function_name
  principal     = "secretsmanager.amazonaws.com"
  statement_id  = "AllowExecutionFromSecretsManager1"
}

# Not a real rotation function
resource "aws_lambda_function" "test2" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-2"
  handler       = "exports.example"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs16.x"
}

resource "aws_lambda_permission" "test2" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test2.function_name
  principal     = "secretsmanager.amazonaws.com"
  statement_id  = "AllowExecutionFromSecretsManager2"
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_rotation" "test" {
  secret_id           = aws_secretsmanager_secret.test.id
  rotation_lambda_arn = aws_lambda_function.test1.arn

  rotation_rules {
    schedule_expression = "%[2]s"
  }

  depends_on = [aws_lambda_permission.test1]
}
`, rName, scheduleExpression))
}

func testAccSecretRotationConfig_duration(rName string, automaticallyAfterDays int, duration string) string {
	return acctest.ConfigCompose(acctest.ConfigLambdaBase(rName, rName, rName), fmt.Sprintf(`
# Not a real rotation function
resource "aws_lambda_function" "test1" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-1"
  handler       = "exports.example"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs16.x"
}

resource "aws_lambda_permission" "test1" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test1.function_name
  principal     = "secretsmanager.amazonaws.com"
  statement_id  = "AllowExecutionFromSecretsManager1"
}

# Not a real rotation function
resource "aws_lambda_function" "test2" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-2"
  handler       = "exports.example"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs16.x"
}

resource "aws_lambda_permission" "test2" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test2.function_name
  principal     = "secretsmanager.amazonaws.com"
  statement_id  = "AllowExecutionFromSecretsManager2"
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_rotation" "test" {
  secret_id           = aws_secretsmanager_secret.test.id
  rotation_lambda_arn = aws_lambda_function.test1.arn

  rotation_rules {
    automatically_after_days = %[2]d
    duration                 = "%[3]s"
  }

  depends_on = [aws_lambda_permission.test1]
}
`, rName, automaticallyAfterDays, duration))
}

func testSecretValueIsCurrent(ctx context.Context, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerConn(ctx)
		// Write secret value to clear in-rotation state, otherwise updating the secret rotation
		// will fail with "A previous rotation isn't complete. That rotation will be reattempted."
		put_secret_input := &secretsmanager.PutSecretValueInput{
			SecretId:     aws.String(rName),
			SecretString: aws.String("secret-value"),
		}
		_, err := conn.PutSecretValueWithContext(ctx, put_secret_input)
		if err != nil {
			return err
		}
		input := &secretsmanager.DescribeSecretInput{
			SecretId: aws.String(rName),
		}
		output, err := conn.DescribeSecretWithContext(ctx, input)
		if err != nil {
			return err
		} else {
			// Ensure that the current version of the secret is in the AWSCURRENT stage
			for _, stage := range output.VersionIdsToStages {
				if *stage[0] == "AWSCURRENT" {
					return nil
				} else {
					return fmt.Errorf("Secret version is not in AWSCURRENT stage: %s", *stage[0])
				}
			}
			return nil
		}
	}
}
