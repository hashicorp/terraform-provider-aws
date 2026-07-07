// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameFunctionScalingConfig = "Function Scaling Config"
)

func TestAccLambdaFunctionScalingConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var out lambda.GetFunctionScalingConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_scaling_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccCapacityProviderPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionScalingConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionScalingConfigConfig_basic(rName, 3, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionScalingConfigExists(ctx, t, resourceName, &out),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrFunctionARN),
					resource.TestCheckResourceAttrSet(resourceName, "function_state"),
					resource.TestCheckResourceAttr(resourceName, "function_scaling_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "function_scaling_config.0.min_execution_environments", "3"),
					resource.TestCheckResourceAttr(resourceName, "function_scaling_config.0.max_execution_environments", "100"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", "$LATEST.PUBLISHED"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccFunctionScalingConfigImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "function_name",
				ImportStateVerifyIgnore:              []string{"function_state"},
			},
		},
	})
}

func TestAccLambdaFunctionScalingConfig_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var out lambda.GetFunctionScalingConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_scaling_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccCapacityProviderPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionScalingConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionScalingConfigConfig_basic(rName, 3, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionScalingConfigExists(ctx, t, resourceName, &out),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrFunctionARN),
					resource.TestCheckResourceAttrSet(resourceName, "function_state"),
					resource.TestCheckResourceAttr(resourceName, "function_scaling_config.0.min_execution_environments", "3"),
					resource.TestCheckResourceAttr(resourceName, "function_scaling_config.0.max_execution_environments", "100"),
				),
			},
			{
				Config: testAccFunctionScalingConfigConfig_basic(rName, 5, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionScalingConfigExists(ctx, t, resourceName, &out),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrFunctionARN),
					resource.TestCheckResourceAttrSet(resourceName, "function_state"),
					resource.TestCheckResourceAttr(resourceName, "function_scaling_config.0.min_execution_environments", "5"),
					resource.TestCheckResourceAttr(resourceName, "function_scaling_config.0.max_execution_environments", "200"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionScalingConfig_disappears_Function(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var out lambda.GetFunctionScalingConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_scaling_config.test"
	functionResourceName := "aws_lambda_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccCapacityProviderPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionScalingConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionScalingConfigConfig_basic(rName, 3, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionScalingConfigExists(ctx, t, resourceName, &out),
					// The scaling config has no dedicated delete API and cannot be
					// deleted independently. Disappear the parent function so that
					// GetFunctionScalingConfig returns NotFound and Read removes the
					// resource from state.
					acctest.CheckSDKResourceDisappears(ctx, t, tflambda.ResourceFunction(), functionResourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

// TestAccLambdaFunctionScalingConfig_partial covers configs that set only one of
// min/max. AWS retains omitted fields (a max-only update keeps the previously-set
// min) and cannot clear an individual field, so min and max are Optional+Computed:
// omitting a field accepts whatever AWS reports rather than forcing it null.
func TestAccLambdaFunctionScalingConfig_partial(t *testing.T) {
	ctx := acctest.Context(t)

	var out lambda.GetFunctionScalingConfigOutput
	resourceName := "aws_lambda_function_scaling_config.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccCapacityProviderPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionScalingConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionScalingConfigConfig_minOnly(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionScalingConfigExists(ctx, t, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "function_scaling_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "function_scaling_config.0.min_execution_environments", "3"),
				),
			},
			{
				// max-only update: AWS retains the previously-set min, so both are reported.
				Config: testAccFunctionScalingConfigConfig_maxOnly(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionScalingConfigExists(ctx, t, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "function_scaling_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "function_scaling_config.0.min_execution_environments", "3"),
					resource.TestCheckResourceAttr(resourceName, "function_scaling_config.0.max_execution_environments", "200"),
				),
			},
		},
	})
}

// TestAccLambdaFunctionScalingConfig_emptyScalingConfig verifies that an empty
// function_scaling_config block is rejected at plan time. An empty configuration
// is a reset (no scaling config), which cannot persist as a resource, so at least
// one of min/max must be set.
func TestAccLambdaFunctionScalingConfig_emptyScalingConfig(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccFunctionScalingConfigConfig_emptyScalingConfig(rName),
				PlanOnly:    true,
				ExpectError: regexache.MustCompile(`(?i)at least one`),
			},
		},
	})
}

func testAccCheckFunctionScalingConfigDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_function_scaling_config" {
				continue
			}

			functionName := rs.Primary.Attributes["function_name"]
			qualifier := rs.Primary.Attributes["qualifier"]

			out, err := tflambda.FindFunctionScalingConfigByTwoPartKey(ctx, conn, functionName, qualifier)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.Lambda, create.ErrActionCheckingDestroyed, ResNameFunctionScalingConfig, functionName, err)
			}

			// There is no dedicated delete API; the config is "destroyed" by resetting
			// it, which leaves no execution environment values. AWS reports the values
			// under RequestedFunctionScalingConfig while a change settles and moves them
			// to AppliedFunctionScalingConfig once in effect, clearing the other, so
			// inspect both.
			requestedEmpty := out.RequestedFunctionScalingConfig == nil ||
				(out.RequestedFunctionScalingConfig.MinExecutionEnvironments == nil &&
					out.RequestedFunctionScalingConfig.MaxExecutionEnvironments == nil)
			appliedEmpty := out.AppliedFunctionScalingConfig == nil ||
				(out.AppliedFunctionScalingConfig.MinExecutionEnvironments == nil &&
					out.AppliedFunctionScalingConfig.MaxExecutionEnvironments == nil)
			if requestedEmpty && appliedEmpty {
				continue
			}

			return create.Error(names.Lambda, create.ErrActionCheckingDestroyed, ResNameFunctionScalingConfig, functionName, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckFunctionScalingConfigExists(ctx context.Context, t *testing.T, name string, out *lambda.GetFunctionScalingConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, ResNameFunctionScalingConfig, name, errors.New("not found in state"))
		}

		functionName := rs.Primary.Attributes["function_name"]
		qualifier := rs.Primary.Attributes["qualifier"]
		if functionName == "" || qualifier == "" {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, ResNameFunctionScalingConfig, name, errors.New("function_name or qualifier not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).LambdaClient(ctx)
		resp, err := tflambda.FindFunctionScalingConfigByTwoPartKey(ctx, conn, functionName, qualifier)
		if err != nil {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, ResNameFunctionScalingConfig, name, err)
		}

		*out = *resp
		return nil
	}
}

func testAccFunctionScalingConfigImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}

		functionName := rs.Primary.Attributes["function_name"]
		qualifier := rs.Primary.Attributes["qualifier"]

		return functionName + intflex.ResourceIdSeparator + qualifier, nil
	}
}

func testAccFunctionScalingConfigConfig_functionBase(rName string) string {
	return acctest.ConfigCompose(
		testAccCapacityProviderConfig_basic(rName),
		fmt.Sprintf(`
resource "aws_iam_role_policy" "lambda_operator" {
  name = "%[1]s-operator"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "lambda:PublishVersion",
          "lambda:GetFunction",
          "lambda:GetFunctionConfiguration",
          "lambda:UpdateFunctionConfiguration",
          "lambda:PutFunctionScalingConfig",
          "lambda:GetFunctionScalingConfig",
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/capacityprovider.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "index.handler"
  runtime       = "python3.14"
  memory_size   = 32768

  publish    = true
  publish_to = "LATEST_PUBLISHED"

  capacity_provider_config {
    lambda_managed_instances_capacity_provider_config {
      capacity_provider_arn = aws_lambda_capacity_provider.test.arn
    }
  }

  timeouts {
    create = "30m"
  }

  depends_on = [aws_iam_role_policy.lambda_operator]
}
`, rName))
}

func testAccFunctionScalingConfigConfig_basic(rName string, minExecEnv, maxExecEnv int) string {
	return acctest.ConfigCompose(
		testAccFunctionScalingConfigConfig_functionBase(rName),
		fmt.Sprintf(`
resource "aws_lambda_function_scaling_config" "test" {
  function_name = aws_lambda_function.test.function_name
  qualifier     = "$LATEST.PUBLISHED"

  function_scaling_config {
    min_execution_environments = %[1]d
    max_execution_environments = %[2]d
  }
}
`, minExecEnv, maxExecEnv))
}

func testAccFunctionScalingConfigConfig_minOnly(rName string, minExecEnv int) string {
	return acctest.ConfigCompose(
		testAccFunctionScalingConfigConfig_functionBase(rName),
		fmt.Sprintf(`
resource "aws_lambda_function_scaling_config" "test" {
  function_name = aws_lambda_function.test.function_name
  qualifier     = "$LATEST.PUBLISHED"

  function_scaling_config {
    min_execution_environments = %[1]d
  }
}
`, minExecEnv))
}

func testAccFunctionScalingConfigConfig_maxOnly(rName string, maxExecEnv int) string {
	return acctest.ConfigCompose(
		testAccFunctionScalingConfigConfig_functionBase(rName),
		fmt.Sprintf(`
resource "aws_lambda_function_scaling_config" "test" {
  function_name = aws_lambda_function.test.function_name
  qualifier     = "$LATEST.PUBLISHED"

  function_scaling_config {
    max_execution_environments = %[1]d
  }
}
`, maxExecEnv))
}

func testAccFunctionScalingConfigConfig_emptyScalingConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_function_scaling_config" "test" {
  function_name = %[1]q
  qualifier     = "$LATEST.PUBLISHED"

  function_scaling_config {}
}
`, rName)
}
