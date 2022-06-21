package lambda_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLambdaInvocation_basic(t *testing.T) {
	resourceName := "aws_lambda_invocation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInvocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInvocationConfig_basic(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, fmt.Sprintf(`{"key1":"value1","key2":"value2","key3":%q}`, testData)),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_qualifier(t *testing.T) {
	resourceName := "aws_lambda_invocation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInvocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInvocationConfig_qualifier(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, `{"key1":"value1","key2":"value2","key3":"`+testData+`"}`),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_complex(t *testing.T) {
	resourceName := "aws_lambda_invocation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInvocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInvocationConfig_complex(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, `{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a": "b"}},"key3":"`+testData+`"}`),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_triggers(t *testing.T) {
	resourceName := "aws_lambda_invocation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"
	testData2 := "value4"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInvocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInvocationConfig_triggers(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, `{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a": "b"}},"key3":"`+testData+`"}`),
				),
			},
			{
				Config: testAccInvocationConfig_triggers(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, `{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a": "b"}},"key3":"`+testData+`"}`),
				),
			},
			{
				Config: testAccInvocationConfig_triggers(rName, testData2),
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
`, roleName)
}

func testAccInvocationConfig_basic(rName, testData string) string {
	return acctest.ConfigCompose(
		testAccConfigInvocation_base(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs14.x"

  environment {
    variables = {
      TEST_DATA = %[2]q
    }
  }
}

resource "aws_lambda_invocation" "test" {
  function_name = aws_lambda_function.test.function_name

  input = jsonencode({
    key1 = "value1"
    key2 = "value2"
  })
}
`, rName, testData))
}

func testAccInvocationConfig_qualifier(rName, testData string) string {
	return acctest.ConfigCompose(
		testAccConfigInvocation_base(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs14.x"
  publish       = true

  environment {
    variables = {
      TEST_DATA = %[2]q
    }
  }
}

resource "aws_lambda_invocation" "test" {
  function_name = aws_lambda_function.test.function_name
  qualifier     = aws_lambda_function.test.version

  input = jsonencode({
    key1 = "value1"
    key2 = "value2"
  })
}
`, rName, testData))
}

func testAccInvocationConfig_complex(rName, testData string) string {
	return acctest.ConfigCompose(
		testAccConfigInvocation_base(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs14.x"
  publish       = true

  environment {
    variables = {
      TEST_DATA = %[2]q
    }
  }
}

resource "aws_lambda_invocation" "test" {
  function_name = aws_lambda_function.test.function_name

  input = jsonencode({
    key1 = {
      subkey1 = "subvalue1"
    }
    key2 = {
      subkey2 = "subvalue2"
      subkey3 = {
        a = "b"
      }
    }
  })
}
`, rName, testData))
}

func testAccInvocationConfig_triggers(rName, testData string) string {
	return acctest.ConfigCompose(
		testAccConfigInvocation_base(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs14.x"
  publish       = true

  environment {
    variables = {
      TEST_DATA = %[2]q
    }
  }
}

resource "aws_lambda_invocation" "test" {
  function_name = aws_lambda_function.test.function_name

  triggers = {
    redeployment = sha1(jsonencode([
      aws_lambda_function.test.environment
    ]))
  }

  input = jsonencode({
    key1 = {
      subkey1 = "subvalue1"
    }
    key2 = {
      subkey2 = "subvalue2"
      subkey3 = {
        a = "b"
      }
    }
  })
}
`, rName, testData))
}
