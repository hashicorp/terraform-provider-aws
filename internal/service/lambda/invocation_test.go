package lambda_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLambdaInvocation_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckInvocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigInvocation_basic(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult("aws_lambda_invocation.invocation_test", `{"key1":"value1","key2":"value2","key3":"`+testData+`"}`),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_qualifier(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckInvocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigInvocation_qualifier(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult("aws_lambda_invocation.invocation_test", `{"key1":"value1","key2":"value2","key3":"`+testData+`"}`),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_complex(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckInvocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigInvocation_complex(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult("aws_lambda_invocation.invocation_test", `{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a": "b"}},"key3":"`+testData+`"}`),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_triggers(t *testing.T) {
	resourceName := "aws_lambda_invocation.invocation_test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"
	testData2 := "value4"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckInvocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigInvocation_triggers(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, `{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a": "b"}},"key3":"`+testData+`"}`),
				),
			},
			{
				Config: testAccConfigInvocation_triggers(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, `{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a": "b"}},"key3":"`+testData+`"}`),
				),
			},
			{
				Config: testAccConfigInvocation_triggers(rName, testData2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, `{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a": "b"}},"key3":"`+testData2+`"}`),
				),
			},
		},
	})
}

func testAccCheckInvocationDestroy(s *terraform.State) error {
	// Nothing to check on destroy
	return nil
}

func testAccConfigInvocation_base(roleName string) string {
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

func testAccConfigInvocation_basic(rName, testData string) string {
	return fmt.Sprintf(testAccConfigInvocation_base(rName)+`
resource "aws_lambda_function" "lambda" {
  depends_on = ["aws_iam_role_policy_attachment.lambda_role_policy"]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = "%s"
  role          = "${aws_iam_role.lambda_role.arn}"
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs14.x"

  environment {
    variables = {
      TEST_DATA = "%s"
    }
  }
}

resource "aws_lambda_invocation" "invocation_test" {
  function_name = "${aws_lambda_function.lambda.function_name}"

  input = <<JSON
{
  "key1": "value1",
  "key2": "value2"
}
JSON
}
`, rName, testData)
}

func testAccConfigInvocation_qualifier(rName, testData string) string {
	return fmt.Sprintf(testAccConfigInvocation_base(rName)+`
resource "aws_lambda_function" "lambda" {
  depends_on = ["aws_iam_role_policy_attachment.lambda_role_policy"]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = "%s"
  role          = "${aws_iam_role.lambda_role.arn}"
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs14.x"
  publish       = true

  environment {
    variables = {
      TEST_DATA = "%s"
    }
  }
}

resource "aws_lambda_invocation" "invocation_test" {
  function_name = "${aws_lambda_function.lambda.function_name}"
  qualifier     = "${aws_lambda_function.lambda.version}"

  input = <<JSON
{
  "key1": "value1",
  "key2": "value2"
}
JSON
}
`, rName, testData)
}

func testAccConfigInvocation_complex(rName, testData string) string {
	return fmt.Sprintf(testAccConfigInvocation_base(rName)+`
resource "aws_lambda_function" "lambda" {
  depends_on = ["aws_iam_role_policy_attachment.lambda_role_policy"]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = "%s"
  role          = "${aws_iam_role.lambda_role.arn}"
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs14.x"
  publish       = true

  environment {
    variables = {
      TEST_DATA = "%s"
    }
  }
}

resource "aws_lambda_invocation" "invocation_test" {
  function_name = "${aws_lambda_function.lambda.function_name}"

  input = <<JSON
{
  "key1": {"subkey1": "subvalue1"},
  "key2": {"subkey2": "subvalue2", "subkey3": {"a": "b"}}
}
JSON
}
`, rName, testData)
}

func testAccConfigInvocation_triggers(rName, testData string) string {
	return testAccConfigInvocation_base(rName) + fmt.Sprintf(`
resource "aws_lambda_function" "lambda" {
  depends_on = ["aws_iam_role_policy_attachment.lambda_role_policy"]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = %[1]q
  role          = "${aws_iam_role.lambda_role.arn}"
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs14.x"
  publish       = true

  environment {
    variables = {
      TEST_DATA   = %[2]q
    }
  }
}

resource "aws_lambda_invocation" "invocation_test" {
  function_name = "${aws_lambda_function.lambda.function_name}"

  triggers = {
    redeployment = sha1(jsonencode([
      aws_lambda_function.lambda.environment
    ]))
  }

  input = <<JSON
{
  "key1": {"subkey1": "subvalue1"},
  "key2": {"subkey2": "subvalue2", "subkey3": {"a": "b"}}
}
JSON
}
`, rName, testData)
}
