package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsLambdaInvoke_basic(t *testing.T) {
	testData := "value3"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsLambdaInvoke_basic_config("tf-test-lambda-role", "tf-test-lambda-invoke", testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_lambda_invoke.invoke_test", "result.%", "3"),
					resource.TestCheckResourceAttr("data.aws_lambda_invoke.invoke_test", "result.key1", "value1"),
					resource.TestCheckResourceAttr("data.aws_lambda_invoke.invoke_test", "result.key2", "value2"),
					resource.TestCheckResourceAttr("data.aws_lambda_invoke.invoke_test", "result.key3", testData),
				),
			},
		},
	})
}

func TestAccDataSourceAwsLambdaInvoke_qualifier(t *testing.T) {
	testData := "value3"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsLambdaInvoke_qualifier_config("tf-test-lambda-role-qualifier", "tf-test-lambda-invoke-qualifier", testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_lambda_invoke.invoke_test", "result.%", "3"),
					resource.TestCheckResourceAttr("data.aws_lambda_invoke.invoke_test", "result.key1", "value1"),
					resource.TestCheckResourceAttr("data.aws_lambda_invoke.invoke_test", "result.key2", "value2"),
					resource.TestCheckResourceAttr("data.aws_lambda_invoke.invoke_test", "result.key3", testData),
				),
			},
		},
	})
}

func testAccDataSourceAwsLambdaInvoke_base_config(roleName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "lambda_assume_role_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda_role" {
  name               = "%s"
  assume_role_policy = "${data.aws_iam_policy_document.lambda_assume_role_policy.json}"
}

resource "aws_iam_role_policy_attachment" "lambda_role_policy" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = "${aws_iam_role.lambda_role.name}"
}
`, roleName)
}

func testAccDataSourceAwsLambdaInvoke_basic_config(roleName, lambdaName, testData string) string {
	return fmt.Sprintf(testAccDataSourceAwsLambdaInvoke_base_config(roleName)+`
resource "aws_lambda_function" "lambda" {
  depends_on = ["aws_iam_role_policy_attachment.lambda_role_policy"]

  filename      = "test-fixtures/lambda_invoke.zip"
  function_name = "%s"
  role          = "${aws_iam_role.lambda_role.arn}"
  handler       = "lambda_invoke.handler"
  runtime       = "nodejs8.10"

  environment {
    variables = {
      TEST_DATA = "%s"
    }
  }
}

data "aws_lambda_invoke" "invoke_test" {
  function_name = "${aws_lambda_function.lambda.function_name}"

  input {
    key1 = "value1"
    key2 = "value2"
  }
}
`, lambdaName, testData)
}

func testAccDataSourceAwsLambdaInvoke_qualifier_config(roleName, lambdaName, testData string) string {
	return fmt.Sprintf(testAccDataSourceAwsLambdaInvoke_base_config(roleName)+`
resource "aws_lambda_function" "lambda" {
  depends_on = ["aws_iam_role_policy_attachment.lambda_role_policy"]

  filename      = "test-fixtures/lambda_invoke.zip"
  function_name = "%s"
  role          = "${aws_iam_role.lambda_role.arn}"
  handler       = "lambda_invoke.handler"
  runtime       = "nodejs8.10"
  publish       = true

  environment {
    variables = {
      TEST_DATA = "%s"
    }
  }
}

data "aws_lambda_invoke" "invoke_test" {
  function_name = "${aws_lambda_function.lambda.function_name}"
  qualifier     = "${aws_lambda_function.lambda.version}"

  input {
    key1 = "value1"
    key2 = "value2"
  }
}
`, lambdaName, testData)
}
