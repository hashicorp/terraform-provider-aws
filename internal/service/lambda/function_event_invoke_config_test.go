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

func TestAccLambdaFunctionEventInvokeConfig_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeFunctionNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "maximum_event_age_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", "2"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", ""),
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

func TestAccLambdaFunctionEventInvokeConfig_Disappears_lambdaFunction(t *testing.T) {
	var function lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeFunctionNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(lambdaFunctionResourceName, rName, &function),
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					testAccCheckFunctionDisappears(&function),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_Disappears_lambdaFunctionEventInvoke(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeFunctionNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					testAccCheckFunctionEventInvokeDisappearsConfig(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_DestinationOnFailure_destination(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"
	sqsQueueResourceName := "aws_sqs_queue.test"
	snsTopicResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeDestinationOnFailureDestinationSQSQueueConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination", sqsQueueResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionEventInvokeDestinationOnFailureDestinationSNSTopicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination", snsTopicResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_DestinationOnSuccess_destination(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"
	sqsQueueResourceName := "aws_sqs_queue.test"
	snsTopicResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeDestinationOnSuccessDestinationSQSQueueConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_success.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_success.0.destination", sqsQueueResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionEventInvokeDestinationOnSuccessDestinationSNSTopicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_success.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_success.0.destination", snsTopicResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_Destination_remove(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"
	sqsQueueResourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeDestinationOnFailureDestinationSQSQueueConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination", sqsQueueResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionEventInvokeQualifierFunctionVersionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "0"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_Destination_swap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"
	sqsQueueResourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeDestinationOnFailureDestinationSQSQueueConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination", sqsQueueResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionEventInvokeDestinationOnSuccessDestinationSQSQueueConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_success.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_success.0.destination", sqsQueueResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_FunctionName_arn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeFunctionNameARNConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", ""),
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

func TestAccLambdaFunctionEventInvokeConfig_QualifierFunctionName_arn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeQualifierFunctionNameARNConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", tflambda.FunctionVersionLatest),
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

func TestAccLambdaFunctionEventInvokeConfig_maximumEventAgeInSeconds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeMaximumEventAgeInSecondsConfig(rName, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_event_age_in_seconds", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionEventInvokeMaximumEventAgeInSecondsConfig(rName, 200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_event_age_in_seconds", "200"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_maximumRetryAttempts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeMaximumRetryAttemptsConfig(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionEventInvokeMaximumRetryAttemptsConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", "1"),
				),
			},
			{
				Config: testAccFunctionEventInvokeMaximumRetryAttemptsConfig(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", "0"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_Qualifier_aliasName(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaAliasResourceName := "aws_lambda_alias.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeQualifierAliasNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
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

func TestAccLambdaFunctionEventInvokeConfig_Qualifier_functionVersion(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeQualifierFunctionVersionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
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

func TestAccLambdaFunctionEventInvokeConfig_Qualifier_latest(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeQualifierLatestConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeExistsConfig(resourceName),
					resource.TestCheckResourceAttr(resourceName, "qualifier", tflambda.FunctionVersionLatest),
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

func testAccCheckFunctionEventInvokeConfigDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_function_event_invoke_config" {
			continue
		}

		functionName, qualifier, err := tflambda.FunctionEventInvokeConfigParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &lambda.GetFunctionEventInvokeConfigInput{
			FunctionName: aws.String(functionName),
		}

		if qualifier != "" {
			input.Qualifier = aws.String(qualifier)
		}

		output, err := conn.GetFunctionEventInvokeConfig(input)

		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("Lambda Function Event Invoke Config (%s) still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckFunctionEventInvokeDisappearsConfig(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

		functionName, qualifier, err := tflambda.FunctionEventInvokeConfigParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &lambda.DeleteFunctionEventInvokeConfigInput{
			FunctionName: aws.String(functionName),
		}

		if qualifier != "" {
			input.Qualifier = aws.String(qualifier)
		}

		_, err = conn.DeleteFunctionEventInvokeConfig(input)

		return err
	}
}

func testAccCheckFunctionEventInvokeExistsConfig(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

		functionName, qualifier, err := tflambda.FunctionEventInvokeConfigParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &lambda.GetFunctionEventInvokeConfigInput{
			FunctionName: aws.String(functionName),
		}

		if qualifier != "" {
			input.Qualifier = aws.String(qualifier)
		}

		_, err = conn.GetFunctionEventInvokeConfig(input)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccFunctionEventInvokeBaseConfig(rName string) string {
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

  depends_on = [
    aws_iam_role_policy_attachment.test,
  ]
}
`, rName)
}

func testAccFunctionEventInvokeDestinationOnFailureDestinationSNSTopicConfig(rName string) string {
	return testAccFunctionEventInvokeBaseConfig(rName) + fmt.Sprintf(`
resource "aws_iam_role_policy_attachment" "test-AmazonSNSFullAccess" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSNSFullAccess"
  role       = aws_iam_role.test.id
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name

  destination_config {
    on_failure {
      destination = aws_sns_topic.test.arn
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonSNSFullAccess]
}
`, rName)
}

func testAccFunctionEventInvokeDestinationOnFailureDestinationSQSQueueConfig(rName string) string {
	return testAccFunctionEventInvokeBaseConfig(rName) + fmt.Sprintf(`
resource "aws_iam_role_policy_attachment" "test-AmazonSQSFullAccess" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSQSFullAccess"
  role       = aws_iam_role.test.id
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name

  destination_config {
    on_failure {
      destination = aws_sqs_queue.test.arn
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonSQSFullAccess]
}
`, rName)
}

func testAccFunctionEventInvokeDestinationOnSuccessDestinationSNSTopicConfig(rName string) string {
	return testAccFunctionEventInvokeBaseConfig(rName) + fmt.Sprintf(`
resource "aws_iam_role_policy_attachment" "test-AmazonSNSFullAccess" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSNSFullAccess"
  role       = aws_iam_role.test.id
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name

  destination_config {
    on_success {
      destination = aws_sns_topic.test.arn
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonSNSFullAccess]
}
`, rName)
}

func testAccFunctionEventInvokeDestinationOnSuccessDestinationSQSQueueConfig(rName string) string {
	return testAccFunctionEventInvokeBaseConfig(rName) + fmt.Sprintf(`
resource "aws_iam_role_policy_attachment" "test-AmazonSQSFullAccess" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSQSFullAccess"
  role       = aws_iam_role.test.id
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name

  destination_config {
    on_success {
      destination = aws_sqs_queue.test.arn
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonSQSFullAccess]
}
`, rName)
}

func testAccFunctionEventInvokeFunctionNameConfig(rName string) string {
	return testAccFunctionEventInvokeBaseConfig(rName) + `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name
}
`
}

func testAccFunctionEventInvokeFunctionNameARNConfig(rName string) string {
	return testAccFunctionEventInvokeBaseConfig(rName) + `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.arn
}
`
}

func testAccFunctionEventInvokeQualifierFunctionNameARNConfig(rName string) string {
	return testAccFunctionEventInvokeBaseConfig(rName) + `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.arn
  qualifier     = "$LATEST"
}
`
}

func testAccFunctionEventInvokeMaximumEventAgeInSecondsConfig(rName string, maximumEventAgeInSeconds int) string {
	return testAccFunctionEventInvokeBaseConfig(rName) + fmt.Sprintf(`
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name                = aws_lambda_function.test.function_name
  maximum_event_age_in_seconds = %[1]d
}
`, maximumEventAgeInSeconds)
}

func testAccFunctionEventInvokeMaximumRetryAttemptsConfig(rName string, maximumRetryAttempts int) string {
	return testAccFunctionEventInvokeBaseConfig(rName) + fmt.Sprintf(`
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name          = aws_lambda_function.test.function_name
  maximum_retry_attempts = %[1]d
}
`, maximumRetryAttempts)
}

func testAccFunctionEventInvokeQualifierAliasNameConfig(rName string) string {
	return testAccFunctionEventInvokeBaseConfig(rName) + `
resource "aws_lambda_alias" "test" {
  function_name    = aws_lambda_function.test.function_name
  function_version = aws_lambda_function.test.version
  name             = "test"
}

resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_alias.test.function_name
  qualifier     = aws_lambda_alias.test.name
}
`
}

func testAccFunctionEventInvokeQualifierFunctionVersionConfig(rName string) string {
	return testAccFunctionEventInvokeBaseConfig(rName) + `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name
  qualifier     = aws_lambda_function.test.version
}
`
}

func testAccFunctionEventInvokeQualifierLatestConfig(rName string) string {
	return testAccFunctionEventInvokeBaseConfig(rName) + `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name
  qualifier     = "$LATEST"
}
`
}
