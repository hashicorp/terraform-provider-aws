// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecretRotationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret_rotation.test"
	datasourceName := "data.aws_secretsmanager_secret_rotation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSecretRotationDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`couldn't find resource`),
			},
			{
				Config: testAccSecretRotationDataSourceConfig_default(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "rotation_enabled", resourceName, "rotation_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "rotation_lambda_arn", resourceName, "rotation_lambda_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "rotation_rules.#", resourceName, "rotation_rules.#"),
				),
			},
		},
	})
}

const testAccSecretRotationDataSourceConfig_nonExistent = `
data "aws_secretsmanager_secret_rotation" "test" {
  secret_id = "tf-acc-test-does-not-exist"
}
`

func testAccSecretRotationDataSourceConfig_default(rName string, automaticallyAfterDays int) string {
	return acctest.ConfigCompose(acctest.ConfigLambdaBase(rName, rName, rName), fmt.Sprintf(`
# Not a real rotation function
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-1"
  handler       = "exports.example"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs20.x"
}

resource "aws_lambda_permission" "test" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "secretsmanager.amazonaws.com"
  statement_id  = "AllowExecutionFromSecretsManager"
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_rotation" "test" {
  secret_id           = aws_secretsmanager_secret.test.id
  rotation_lambda_arn = aws_lambda_function.test.arn

  rotation_rules {
    automatically_after_days = %[2]d
  }
}

data "aws_secretsmanager_secret_rotation" "test" {
  secret_id = aws_secretsmanager_secret_rotation.test.secret_id
}
`, rName, automaticallyAfterDays))
}
