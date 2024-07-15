// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecretsmanager "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecretRotation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const (
		resourceName               = "aws_secretsmanager_secret_rotation.test"
		lambdaFunctionResourceName = "aws_lambda_function.test"
		days                       = 7
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretRotationConfig_basic(rName, days),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "rotate_immediately", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.automatically_after_days", strconv.Itoa(days)),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.duration", ""),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.schedule_expression", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rotate_immediately"},
			},
		},
	})
}

func TestAccSecretsManagerSecretRotation_upgradePreRotateImmediately(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const (
		resourceName = "aws_secretsmanager_secret_rotation.test"
		days         = 7
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		CheckDestroy: testAccCheckSecretRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.32.0",
					},
				},
				Config: testAccSecretRotationConfig_basic(rName, days),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckNoResourceAttr(resourceName, "rotate_immediately"),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccSecretRotationConfig_basic(rName, days),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotate_immediately", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecretRotation_rotateImmediately(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const (
		resourceName               = "aws_secretsmanager_secret_rotation.test"
		lambdaFunctionResourceName = "aws_lambda_function.test"
		days                       = 7
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretRotationConfig_rotateImmediately(rName, days),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "rotate_immediately", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.automatically_after_days", strconv.Itoa(days)),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.duration", ""),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.schedule_expression", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rotate_immediately"},
			},
		},
	})
}

func TestAccSecretsManagerSecretRotation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const (
		resourceName = "aws_secretsmanager_secret_rotation.test"
		days         = 7
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretRotationConfig_basic(rName, days),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecretsmanager.ResourceSecretRotation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSecretsManagerSecretRotation_Disappears_secret(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const (
		resourceName       = "aws_secretsmanager_secret_rotation.test"
		secretResourceName = "aws_secretsmanager_secret.test"
		days               = 7
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretRotationConfig_basic(rName, days),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecretsmanager.ResourceSecret(), secretResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSecretsManagerSecretRotation_scheduleExpression(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const (
		resourceName               = "aws_secretsmanager_secret_rotation.test"
		lambdaFunctionResourceName = "aws_lambda_function.test"
		scheduleExpression         = "rate(10 days)"
		scheduleExpression02       = "rate(10 days)"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretRotationConfig_scheduleExpression(rName, scheduleExpression),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.schedule_expression", scheduleExpression),
				),
			},
			{
				Config: testAccSecretRotationConfig_scheduleExpression(rName, scheduleExpression02),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.schedule_expression", scheduleExpression02),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rotate_immediately"},
			},
		},
	})
}

func TestAccSecretsManagerSecretRotation_scheduleExpressionToDays(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const (
		resourceName               = "aws_secretsmanager_secret_rotation.test"
		lambdaFunctionResourceName = "aws_lambda_function.test"
		scheduleExpression         = "rate(10 days)"
		days                       = 7
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretRotationConfig_scheduleExpression(rName, scheduleExpression),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.schedule_expression", scheduleExpression),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.automatically_after_days", acctest.Ct0),
				),
			},
			{
				Config: testAccSecretRotationConfig_basic(rName, days),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.automatically_after_days", strconv.Itoa(days)),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.schedule_expression", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rotate_immediately"},
			},
		},
	})
}

func TestAccSecretsManagerSecretRotation_scheduleExpressionHours(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const (
		resourceName               = "aws_secretsmanager_secret_rotation.test"
		lambdaFunctionResourceName = "aws_lambda_function.test"
		scheduleExpression         = "rate(6 hours)"
		scheduleExpression02       = "rate(10 hours)"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretRotationConfig_scheduleExpression(rName, scheduleExpression),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.schedule_expression", scheduleExpression),
					testSecretValueIsCurrent(ctx, rName),
				),
			},
			{
				Config: testAccSecretRotationConfig_scheduleExpression(rName, scheduleExpression02),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.schedule_expression", scheduleExpression02),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rotate_immediately"},
			},
		},
	})
}

func TestAccSecretsManagerSecretRotation_duration(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	const (
		resourceName               = "aws_secretsmanager_secret_rotation.test"
		lambdaFunctionResourceName = "aws_lambda_function.test"
		days                       = 7
		duration                   = "3h"
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretRotationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretRotationConfig_duration(rName, days, duration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretRotationExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "rotation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "rotation_lambda_arn", lambdaFunctionResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.automatically_after_days", strconv.Itoa(days)),
					resource.TestCheckResourceAttr(resourceName, "rotation_rules.0.duration", duration),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rotate_immediately"},
			},
		},
	})
}

func testAccCheckSecretRotationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerClient(ctx)

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

			if !aws.ToBool(output.RotationEnabled) {
				continue
			}

			return fmt.Errorf("Secrets Manager Secret Rotation %s still enabled", rs.Primary.ID)
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerClient(ctx)

		output, err := tfsecretsmanager.FindSecretByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if !aws.ToBool(output.RotationEnabled) {
			return fmt.Errorf("Secrets Manager Secret Rotation %s not enabled", rs.Primary.ID)
		}

		*v = *output

		return nil
	}
}

func testSecretValueIsCurrent(ctx context.Context, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerClient(ctx)
		// Write secret value to clear in-rotation state, otherwise updating the secret rotation
		// will fail with "A previous rotation isn't complete. That rotation will be reattempted."
		put_secret_input := &secretsmanager.PutSecretValueInput{
			SecretId:     aws.String(rName),
			SecretString: aws.String("secret-value"),
		}
		_, err := conn.PutSecretValue(ctx, put_secret_input)
		if err != nil {
			return err
		}
		input := &secretsmanager.DescribeSecretInput{
			SecretId: aws.String(rName),
		}
		output, err := conn.DescribeSecret(ctx, input)
		if err != nil {
			return err
		} else {
			// Ensure that the current version of the secret is in the AWSCURRENT stage
			for _, stage := range output.VersionIdsToStages {
				if stage[0] == "AWSCURRENT" {
					return nil
				} else {
					return fmt.Errorf("Secret version is not in AWSCURRENT stage: %s", stage[0])
				}
			}
			return nil
		}
	}
}

func testAccSecretRotationConfig_base(rName string) string {
	return fmt.Sprintf(`
# Not a real rotation function
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-1"
  handler       = "exports.example"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs16.x"
}

resource "aws_lambda_permission" "test" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "secretsmanager.amazonaws.com"
  statement_id  = "AllowExecutionFromSecretsManager1"
}
`, rName)
}

func testAccSecretRotationConfig_basic(rName string, automaticallyAfterDays int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		testAccSecretRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"
}

resource "aws_secretsmanager_secret_rotation" "test" {
  secret_id           = aws_secretsmanager_secret.test.id
  rotation_lambda_arn = aws_lambda_function.test.arn

  rotation_rules {
    automatically_after_days = %[2]d
  }

  depends_on = [aws_lambda_permission.test]
}
`, rName, automaticallyAfterDays))
}

func testAccSecretRotationConfig_rotateImmediately(rName string, automaticallyAfterDays int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		testAccSecretRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "test-string"
}

resource "aws_secretsmanager_secret_rotation" "test" {
  secret_id           = aws_secretsmanager_secret.test.id
  rotation_lambda_arn = aws_lambda_function.test.arn
  rotate_immediately  = false

  rotation_rules {
    automatically_after_days = %[2]d
  }

  depends_on = [aws_lambda_permission.test]
}
`, rName, automaticallyAfterDays))
}

func testAccSecretRotationConfig_scheduleExpression(rName string, scheduleExpression string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		testAccSecretRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_rotation" "test" {
  secret_id           = aws_secretsmanager_secret.test.id
  rotation_lambda_arn = aws_lambda_function.test.arn

  rotation_rules {
    schedule_expression = "%[2]s"
  }

  depends_on = [aws_lambda_permission.test]
}
`, rName, scheduleExpression))
}

func testAccSecretRotationConfig_duration(rName string, automaticallyAfterDays int, duration string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		testAccSecretRotationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_rotation" "test" {
  secret_id           = aws_secretsmanager_secret.test.id
  rotation_lambda_arn = aws_lambda_function.test.arn

  rotation_rules {
    automatically_after_days = %[2]d
    duration                 = "%[3]s"
  }

  depends_on = [aws_lambda_permission.test]
}
`, rName, automaticallyAfterDays, duration))
}
