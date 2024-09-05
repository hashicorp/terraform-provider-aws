// Copyright (c) HashiCorp, Inc.
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

func TestAccLambdaRuntimeManagementConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var cfg lambda.GetRuntimeManagementConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_runtime_management_config.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuntimeManagementConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuntimeManagementConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuntimeManagementConfigExists(ctx, resourceName, &cfg),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "update_runtime_on", string(types.UpdateRuntimeOnFunctionUpdate)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrFunctionARN, "lambda", regexache.MustCompile(`function:+.`)),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_runtime_management_config.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuntimeManagementConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuntimeManagementConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuntimeManagementConfigExists(ctx, resourceName, &cfg),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflambda.ResourceFunction(), functionResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaRuntimeManagementConfig_runtimeVersionARN(t *testing.T) {
	ctx := acctest.Context(t)

	var cfg lambda.GetRuntimeManagementConfigOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_runtime_management_config.test"
	functionResourceName := "aws_lambda_function.test"
	// nodejs18.x version hash in us-west-2, commercial partition
	runtimeVersion := "b475b23763329123d9e6f79f51886d0e1054f727f5b90ec945fcb2a3ec09afdd"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// The runtime version ARN contains a hash unique to a partition/region.
			// There is currently no API to retrieve this ARN, so we have to hard-code
			// the value and restrict this test to us-west-2 in the standard commercial
			// partition.
			acctest.PreCheckPartition(t, names.StandardPartitionID)
			acctest.PreCheckRegion(t, names.USWest2RegionID)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuntimeManagementConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuntimeManagementConfigConfig_runtimeVersionARN(rName, runtimeVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuntimeManagementConfigExists(ctx, resourceName, &cfg),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "update_runtime_on", string(types.UpdateRuntimeOnManual)),
					resource.TestMatchResourceAttr(resourceName, "runtime_version_arn", regexache.MustCompile(runtimeVersion)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrFunctionARN, "lambda", regexache.MustCompile(`function:+.`)),
				),
			},
		},
	})
}

func testAccCheckRuntimeManagementConfigDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

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

func testAccCheckRuntimeManagementConfigExists(ctx context.Context, name string, cfg *lambda.GetRuntimeManagementConfigOutput) resource.TestCheckFunc {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

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
  runtime       = "nodejs18.x"
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

func testAccRuntimeManagementConfigConfig_runtimeVersionARN(rName, runtimeVersion string) string {
	return acctest.ConfigCompose(
		testAccRuntimeManagementConfigConfigBase(rName),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_lambda_runtime_management_config" "test" {
  function_name       = aws_lambda_function.test.function_name
  update_runtime_on   = "Manual"
  runtime_version_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.name}::runtime:%[1]s"
}
`, runtimeVersion))
}
