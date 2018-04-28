package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsLambdaInvocation_basic(t *testing.T) {
	testData := "value3"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsLambdaInvocation_basic_config("tf-test-lambda-role", "tf-test-lambda-invocation", testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result.%", "3"),
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result.key1", "value1"),
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result.key2", "value2"),
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result.key3", testData),
				),
			},
		},
	})
}

func TestAccDataSourceAwsLambdaInvocation_qualifier(t *testing.T) {
	testData := "value3"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsLambdaInvocation_qualifier_config("tf-test-lambda-role-qualifier", "tf-test-lambda-invocation-qualifier", testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result.%", "3"),
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result.key1", "value1"),
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result.key2", "value2"),
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result.key3", testData),
				),
			},
		},
	})
}

func testAccDataSourceAwsLambdaInvocation_base_config(roleName string) string {
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

func testAccDataSourceAwsLambdaInvocation_basic_config(roleName, lambdaName, testData string) string {
	return fmt.Sprintf(testAccDataSourceAwsLambdaInvocation_base_config(roleName)+`
resource "aws_lambda_function" "lambda" {
  depends_on = ["aws_iam_role_policy_attachment.lambda_role_policy"]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = "%s"
  role          = "${aws_iam_role.lambda_role.arn}"
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs8.10"

  environment {
    variables = {
      TEST_DATA = "%s"
    }
  }
}

data "aws_lambda_invocation" "invocation_test" {
  function_name = "${aws_lambda_function.lambda.function_name}"

  input = <<JSON
{
  "key1": "value1",
  "key2": "value2"
}
JSON
}
`, lambdaName, testData)
}

func testAccDataSourceAwsLambdaInvocation_qualifier_config(roleName, lambdaName, testData string) string {
	return fmt.Sprintf(testAccDataSourceAwsLambdaInvocation_base_config(roleName)+`
resource "aws_lambda_function" "lambda" {
  depends_on = ["aws_iam_role_policy_attachment.lambda_role_policy"]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = "%s"
  role          = "${aws_iam_role.lambda_role.arn}"
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs8.10"
  publish       = true

  environment {
    variables = {
      TEST_DATA = "%s"
    }
  }
}

data "aws_lambda_invocation" "invocation_test" {
  function_name = "${aws_lambda_function.lambda.function_name}"
  qualifier     = "${aws_lambda_function.lambda.version}"

  input = <<JSON
{
  "key1": "value1",
  "key2": "value2"
}
JSON
}
`, lambdaName, testData)
}
