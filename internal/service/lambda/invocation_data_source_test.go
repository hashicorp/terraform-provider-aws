package lambda_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func testAccCheckInvocationResult(name, expectedResult string) resource.TestCheckFunc {
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

		if !verify.SuppressEquivalentJSONDiffs("", result, expectedResult, nil) {
			return fmt.Errorf("%s: Attribute 'result' expected %s, got %s", name, expectedResult, result)
		}

		return nil
	}
}

func TestAccLambdaInvocationDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInvocationDataSourceConfig_basic(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult("data.aws_lambda_invocation.invocation_test", `{"key1":"value1","key2":"value2","key3":"`+testData+`"}`),
				),
			},
		},
	})
}

func TestAccLambdaInvocationDataSource_qualifier(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInvocationDataSourceConfig_qualifier(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult("data.aws_lambda_invocation.invocation_test", `{"key1":"value1","key2":"value2","key3":"`+testData+`"}`),
				),
			},
		},
	})
}

func TestAccLambdaInvocationDataSource_complex(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInvocationDataSourceConfig_complex(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult("data.aws_lambda_invocation.invocation_test", `{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a": "b"}},"key3":"`+testData+`"}`),
				),
			},
		},
	})
}

func testAccInvocationDataSource_base_config(roleName string) string {
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
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role_policy.json
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "lambda_role_policy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.lambda_role.name
}
`, roleName)
}

func testAccInvocationDataSourceConfig_basic(rName, testData string) string {
	return fmt.Sprintf(testAccInvocationDataSource_base_config(rName)+`
resource "aws_lambda_function" "lambda" {
  depends_on = [aws_iam_role_policy_attachment.lambda_role_policy]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = "%s"
  role          = aws_iam_role.lambda_role.arn
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs12.x"

  environment {
    variables = {
      TEST_DATA = "%s"
    }
  }
}

data "aws_lambda_invocation" "invocation_test" {
  function_name = aws_lambda_function.lambda.function_name

  input = <<JSON
{
  "key1": "value1",
  "key2": "value2"
}
JSON
}
`, rName, testData)
}

func testAccInvocationDataSourceConfig_qualifier(rName, testData string) string {
	return fmt.Sprintf(testAccInvocationDataSource_base_config(rName)+`
resource "aws_lambda_function" "lambda" {
  depends_on = [aws_iam_role_policy_attachment.lambda_role_policy]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = "%s"
  role          = aws_iam_role.lambda_role.arn
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs12.x"
  publish       = true

  environment {
    variables = {
      TEST_DATA = "%s"
    }
  }
}

data "aws_lambda_invocation" "invocation_test" {
  function_name = aws_lambda_function.lambda.function_name
  qualifier     = aws_lambda_function.lambda.version

  input = <<JSON
{
  "key1": "value1",
  "key2": "value2"
}
JSON
}
`, rName, testData)
}

func testAccInvocationDataSourceConfig_complex(rName, testData string) string {
	return fmt.Sprintf(testAccInvocationDataSource_base_config(rName)+`
resource "aws_lambda_function" "lambda" {
  depends_on = [aws_iam_role_policy_attachment.lambda_role_policy]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = "%s"
  role          = aws_iam_role.lambda_role.arn
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs12.x"
  publish       = true

  environment {
    variables = {
      TEST_DATA = "%s"
    }
  }
}

data "aws_lambda_invocation" "invocation_test" {
  function_name = aws_lambda_function.lambda.function_name

  input = <<JSON
{
  "key1": {
    "subkey1": "subvalue1"
  },
  "key2": {
    "subkey2": "subvalue2",
    "subkey3": {
      "a": "b"
    }
  }
}
JSON
}
`, rName, testData)
}
