package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func testAccCheckLambdaInvocationResult(name, expectedResult string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		result, ok := rs.Primary.Attributes["result"]

		if !ok {
			return fmt.Errorf("No result is set")
		}

		if !suppressEquivalentJsonDiffs("", result, expectedResult, nil) {
			return fmt.Errorf("%s: Attribute 'result' expected %s, got %s", name, expectedResult, result)
		}

		return nil
	}
}

func TestAccDataSourceAwsLambdaInvocation_basic(t *testing.T) {
	testData := "value3"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsLambdaInvocation_basic_config("tf-test-lambda-role", "tf-test-lambda-invocation", testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result_map.%", "3"),
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result_map.key1", "value1"),
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result_map.key2", "value2"),
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result_map.key3", testData),
					testAccCheckLambdaInvocationResult("data.aws_lambda_invocation.invocation_test", `{"key1":"value1","key2":"value2","key3":"`+testData+`"}`),
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
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result_map.%", "3"),
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result_map.key1", "value1"),
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result_map.key2", "value2"),
					resource.TestCheckResourceAttr("data.aws_lambda_invocation.invocation_test", "result_map.key3", testData),
				),
			},
		},
	})
}

func TestAccDataSourceAwsLambdaInvocation_complex(t *testing.T) {
	testData := "value3"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsLambdaInvocation_complex_config("tf-test-lambda-role-complex", "tf-test-lambda-invocation-complex", testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("data.aws_lambda_invocation.invocation_test", "result_map"),
					testAccCheckLambdaInvocationResult("data.aws_lambda_invocation.invocation_test", `{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a": "b"}},"key3":"`+testData+`"}`),
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

func testAccDataSourceAwsLambdaInvocation_complex_config(roleName, lambdaName, testData string) string {
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

  input = <<JSON
{
  "key1": {"subkey1": "subvalue1"},
  "key2": {"subkey2": "subvalue2", "subkey3": {"a": "b"}}
}
JSON
}
`, lambdaName, testData)
}
