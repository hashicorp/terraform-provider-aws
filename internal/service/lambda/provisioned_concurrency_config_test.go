package lambda_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
)

func TestAccLambdaProvisionedConcurrencyConfig_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProvisionedConcurrencyConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyQualifierFunctionVersionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyExistsConfig(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaFunctionResourceName, "version"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaProvisionedConcurrencyConfig_Disappears_lambdaFunction(t *testing.T) {
	var function lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProvisionedConcurrencyConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyProvisionedConcurrentExecutionsConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(lambdaFunctionResourceName, rName, &function),
					testAccCheckProvisionedConcurrencyExistsConfig(resourceName),
					testAccCheckFunctionDisappears(&function),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaProvisionedConcurrencyConfig_Disappears_lambdaProvisionedConcurrency(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProvisionedConcurrencyConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyProvisionedConcurrentExecutionsConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyExistsConfig(resourceName),
					testAccCheckProvisionedConcurrencyDisappearsConfig(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaProvisionedConcurrencyConfig_provisionedConcurrentExecutions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProvisionedConcurrencyConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyProvisionedConcurrentExecutionsConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProvisionedConcurrencyProvisionedConcurrentExecutionsConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "2"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", "1"),
				),
			},
		},
	})
}

func TestAccLambdaProvisionedConcurrencyConfig_Qualifier_aliasName(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaAliasResourceName := "aws_lambda_alias.test"
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProvisionedConcurrencyConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedConcurrencyQualifierAliasNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProvisionedConcurrencyExistsConfig(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaAliasResourceName, "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckProvisionedConcurrencyConfigDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_provisioned_concurrency_config" {
			continue
		}

		functionName, qualifier, err := tflambda.ProvisionedConcurrencyConfigParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &lambda.GetProvisionedConcurrencyConfigInput{
			FunctionName: aws.String(functionName),
			Qualifier:    aws.String(qualifier),
		}

		output, err := conn.GetProvisionedConcurrencyConfig(input)

		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeProvisionedConcurrencyConfigNotFoundException) || tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("Lambda Provisioned Concurrency Config (%s) still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckProvisionedConcurrencyDisappearsConfig(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

		functionName, qualifier, err := tflambda.ProvisionedConcurrencyConfigParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &lambda.DeleteProvisionedConcurrencyConfigInput{
			FunctionName: aws.String(functionName),
			Qualifier:    aws.String(qualifier),
		}

		_, err = conn.DeleteProvisionedConcurrencyConfig(input)

		return err
	}
}

func testAccCheckProvisionedConcurrencyExistsConfig(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

		functionName, qualifier, err := tflambda.ProvisionedConcurrencyConfigParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &lambda.GetProvisionedConcurrencyConfigInput{
			FunctionName: aws.String(functionName),
			Qualifier:    aws.String(qualifier),
		}

		output, err := conn.GetProvisionedConcurrencyConfig(input)

		if err != nil {
			return err
		}

		if got, want := aws.StringValue(output.Status), lambda.ProvisionedConcurrencyStatusEnumReady; got != want {
			return fmt.Errorf("Lambda Provisioned Concurrency Config (%s) expected status (%s), got: %s", rs.Primary.ID, want, got)
		}

		return nil
	}
}

func testAccProvisionedConcurrencyBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.test.id
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdapinpoint.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambdapinpoint.handler"
  publish       = true
  runtime       = "nodejs12.x"

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccProvisionedConcurrencyProvisionedConcurrentExecutionsConfig(rName string, provisionedConcurrentExecutions int) string {
	return testAccProvisionedConcurrencyBaseConfig(rName) + fmt.Sprintf(`
resource "aws_lambda_provisioned_concurrency_config" "test" {
  function_name                     = aws_lambda_function.test.function_name
  provisioned_concurrent_executions = %[1]d
  qualifier                         = aws_lambda_function.test.version
}
`, provisionedConcurrentExecutions)
}

func testAccProvisionedConcurrencyQualifierAliasNameConfig(rName string) string {
	return testAccProvisionedConcurrencyBaseConfig(rName) + `
resource "aws_lambda_alias" "test" {
  function_name    = aws_lambda_function.test.function_name
  function_version = aws_lambda_function.test.version
  name             = "test"
}

resource "aws_lambda_provisioned_concurrency_config" "test" {
  function_name                     = aws_lambda_alias.test.function_name
  provisioned_concurrent_executions = 1
  qualifier                         = aws_lambda_alias.test.name
}
`
}

func testAccProvisionedConcurrencyQualifierFunctionVersionConfig(rName string) string {
	return testAccProvisionedConcurrencyBaseConfig(rName) + `
resource "aws_lambda_provisioned_concurrency_config" "test" {
  function_name                     = aws_lambda_function.test.function_name
  provisioned_concurrent_executions = 1
  qualifier                         = aws_lambda_function.test.version
}
`
}
