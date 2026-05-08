// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaInvocationEphemeral_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	echoResourceName := "echo.test"
	dp := tfjsonpath.New("data")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(ctx, t) },
		ErrorCheck: acctest.ErrorCheck(t, names.LambdaServiceID),
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories(ctx, acctest.ProviderNameEcho),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInvocationEphemeralConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(echoResourceName, dp.AtMapKey("executed_version"), knownvalue.StringExact("$LATEST")),
					statecheck.ExpectKnownValue(echoResourceName, dp.AtMapKey("function_name"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(echoResourceName, dp.AtMapKey("log_result"), knownvalue.Null()),
					statecheck.ExpectKnownValue(echoResourceName, dp.AtMapKey("result"), knownvalue.StringExact(`{"output":{"key1":"value1","key2":"value2"}}`)),
					statecheck.ExpectKnownValue(echoResourceName, dp.AtMapKey(names.AttrStatusCode), knownvalue.NumberExact(big.NewFloat(200))),
				},
			},
		},
	})
}

func testAccInvocationEphemeralConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigWithEchoProvider("ephemeral.aws_lambda_invocation.test"),
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.test.name
}

resource "aws_lambda_function" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  filename      = "test-fixtures/lambda_invocation_ephemeral.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambda_invocation_ephemeral.handler"
  runtime       = "nodejs22.x"
}

ephemeral "aws_lambda_invocation" "test" {
  function_name = aws_lambda_function.test.arn

  payload = jsonencode({
    key1 = "value1"
    key2 = "value2"
  })
}
`, rName))
}
