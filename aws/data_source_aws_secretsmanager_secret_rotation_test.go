package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsSecretsManagerSecretRotation_Basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret_rotation.test"
	datasourceName := "data.aws_secretsmanager_secret_rotation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSSecretsManager(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsSecretsManagerSecretRotationConfig_NonExistent,
				ExpectError: regexp.MustCompile(`not found`),
			},
			{
				Config: testAccDataSourceAwsSecretsManagerSecretRotationConfig_Default(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsSecretsManagerSecretRotationCheck(datasourceName, resourceName),
				),
			},
		},
	})
}

func testAccDataSourceAwsSecretsManagerSecretRotationCheck(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		dataSource, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no datasource called %s", datasourceName)
		}

		attrNames := []string{
			"rotation_enabled",
			"rotation_lambda_arn",
			"rotation_rules.#",
		}

		for _, attrName := range attrNames {
			if resource.Primary.Attributes[attrName] != dataSource.Primary.Attributes[attrName] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrName,
					resource.Primary.Attributes[attrName],
					dataSource.Primary.Attributes[attrName],
				)
			}
		}

		return nil
	}
}

const testAccDataSourceAwsSecretsManagerSecretRotationConfig_NonExistent = `
data "aws_secretsmanager_secret_rotation" "test" {
  secret_id = "tf-acc-test-does-not-exist"
}
`

func testAccDataSourceAwsSecretsManagerSecretRotationConfig_Default(rName string, automaticallyAfterDays int) string {
	return baseAccAWSLambdaConfig(rName, rName, rName) + fmt.Sprintf(`
# Not a real rotation function
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-1"
  handler       = "exports.example"
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  runtime       = "nodejs8.10"
}

resource "aws_lambda_permission" "test" {
  action         = "lambda:InvokeFunction"
  function_name  = "${aws_lambda_function.test.function_name}"
  principal      = "secretsmanager.amazonaws.com"
  statement_id   = "AllowExecutionFromSecretsManager"
}

resource "aws_secretsmanager_secret" "test" {
  name = "%[1]s"
}

resource "aws_secretsmanager_secret_rotation" "test" {
	secret_id 					= "${aws_secretsmanager_secret.test.id}"
	rotation_lambda_arn = "${aws_lambda_function.test.arn}"

	rotation_rules {
    automatically_after_days = %[2]d
	}
}

data "aws_secretsmanager_secret_rotation" "test" {
  secret_id = "${aws_secretsmanager_secret_rotation.test.secret_id}"
}
`, rName, automaticallyAfterDays)
}
