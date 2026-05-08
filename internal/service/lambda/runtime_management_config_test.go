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
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaRuntimeManagementConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var cfg lambda.GetRuntimeManagementConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambda_runtime_management_config.test"
	functionResourceName := "aws_lambda_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuntimeManagementConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuntimeManagementConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuntimeManagementConfigExists(ctx, t, resourceName, &cfg),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "update_runtime_on", string(types.UpdateRuntimeOnFunctionUpdate)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrFunctionARN, "lambda", regexache.MustCompile(`function:+.`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccRuntimeManagementConfigImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "function_name",
				ImportStateVerifyIgnore: []string{
					"qualifier",
				},
			},
		},
	})
}

func TestAccLambdaRuntimeManagementConfig_disappears_Function(t *testing.T) {
	ctx := acctest.Context(t)

	var cfg lambda.GetRuntimeManagementConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambda_runtime_management_config.test"
	functionResourceName := "aws_lambda_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuntimeManagementConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuntimeManagementConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuntimeManagementConfigExists(ctx, t, resourceName, &cfg),
					acctest.CheckSDKResourceDisappears(ctx, t, tflambda.ResourceFunction(), functionResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaRuntimeManagementConfig_runtimeVersionARN(t *testing.T) {
	ctx := acctest.Context(t)

	var cfg lambda.GetRuntimeManagementConfigOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambda_runtime_management_config.test"
	functionResourceName := "aws_lambda_function.test"
	var runtimeVersionARN *string
	runtimeVersionConfigVars := config.Variables{}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuntimeManagementConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuntimeManagementConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuntimeManagementConfigExists(ctx, t, resourceName, &cfg),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "update_runtime_on", string(types.UpdateRuntimeOnFunctionUpdate)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrFunctionARN, "lambda", regexache.MustCompile(`function:+.`)),
				),
			},
			{
				PreConfig: func() {
					function, err := tflambda.FindFunctionByName(ctx, acctest.ProviderMeta(ctx, t).LambdaClient(ctx), rName)
					if err != nil {
						t.Fatalf("finding Lambda Function (%s): %s", rName, err)
					}

					if function.Configuration == nil || function.Configuration.RuntimeVersionConfig == nil || function.Configuration.RuntimeVersionConfig.RuntimeVersionArn == nil {
						t.Fatal("runtime version ARN not found")
					}

					runtimeVersionARN = function.Configuration.RuntimeVersionConfig.RuntimeVersionArn
					runtimeVersionConfigVars["runtime_version_arn"] = config.StringVariable(*runtimeVersionARN)
				},
				Config:          testAccRuntimeManagementConfigConfig_runtimeVersionARN(rName),
				ConfigVariables: runtimeVersionConfigVars,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuntimeManagementConfigExists(ctx, t, resourceName, &cfg),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "update_runtime_on", string(types.UpdateRuntimeOnManual)),
					resource.TestCheckResourceAttrWith(resourceName, "runtime_version_arn", func(value string) error {
						if got, want := value, *runtimeVersionARN; got != want {
							return fmt.Errorf("Attribute 'runtime_version_arn' expected %q, got: %s", want, got)
						}
						return nil
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrFunctionARN, "lambda", regexache.MustCompile(`function:+.`)),
				),
			},
		},
	})
}

func testAccCheckRuntimeManagementConfigDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_runtime_management_config" {
				continue
			}

			functionName := rs.Primary.Attributes["function_name"]
			qualifier := rs.Primary.Attributes["qualifier"]

			_, err := tflambda.FindRuntimeManagementConfigByTwoPartKey(ctx, conn, functionName, qualifier)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Lambda, create.ErrActionCheckingDestroyed, tflambda.ResNameRuntimeManagementConfig, rs.Primary.ID, err)
			}

			return create.Error(names.Lambda, create.ErrActionCheckingDestroyed, tflambda.ResNameRuntimeManagementConfig, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRuntimeManagementConfigExists(ctx context.Context, t *testing.T, name string, cfg *lambda.GetRuntimeManagementConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameRuntimeManagementConfig, name, errors.New("not found"))
		}

		functionName := rs.Primary.Attributes["function_name"]
		qualifier := rs.Primary.Attributes["qualifier"]
		if functionName == "" {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameRuntimeManagementConfig, name, errors.New("function_name not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).LambdaClient(ctx)

		out, err := tflambda.FindRuntimeManagementConfigByTwoPartKey(ctx, conn, functionName, qualifier)
		if err != nil {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameRuntimeManagementConfig, functionName, err)
		}

		*cfg = *out

		return nil
	}
}

func testAccRuntimeManagementConfigImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["function_name"], rs.Primary.Attributes["qualifier"]), nil
	}
}

func testAccRuntimeManagementConfigConfigBase(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs22.x"
}
`, rName))
}

func testAccRuntimeManagementConfigConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccRuntimeManagementConfigConfigBase(rName),
		`
resource "aws_lambda_runtime_management_config" "test" {
  function_name     = aws_lambda_function.test.function_name
  update_runtime_on = "FunctionUpdate"
}
`)
}

func testAccRuntimeManagementConfigConfig_runtimeVersionARN(rName string) string {
	return acctest.ConfigCompose(
		testAccRuntimeManagementConfigConfigBase(rName),
		`
variable "runtime_version_arn" {
  type = string
}

resource "aws_lambda_runtime_management_config" "test" {
  function_name       = aws_lambda_function.test.function_name
  update_runtime_on   = "Manual"
  runtime_version_arn = var.runtime_version_arn
}
`)
}
