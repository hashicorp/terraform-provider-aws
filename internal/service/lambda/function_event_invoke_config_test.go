// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaFunctionEventInvokeConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEventInvokeConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeConfigConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "maximum_event_age_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", acctest.Ct2),
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

func TestAccLambdaFunctionEventInvokeConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEventInvokeConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeConfigConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflambda.ResourceFunctionEventInvokeConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_Disappears_lambdaFunction(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEventInvokeConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeConfigConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflambda.ResourceFunction(), lambdaFunctionResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_DestinationOnFailure_destination(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"
	sqsQueueResourceName := "aws_sqs_queue.test"
	snsTopicResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEventInvokeConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeConfigConfig_destinationOnFailureDestinationSQSQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination", sqsQueueResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionEventInvokeConfigConfig_destinationOnFailureDestinationSNSTopic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination", snsTopicResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_DestinationOnSuccess_destination(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"
	sqsQueueResourceName := "aws_sqs_queue.test"
	snsTopicResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEventInvokeConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeConfigConfig_destinationOnSuccessDestinationSQSQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_success.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_success.0.destination", sqsQueueResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionEventInvokeConfigConfig_destinationOnSuccessDestinationSNSTopic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_success.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_success.0.destination", snsTopicResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_Destination_remove(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"
	sqsQueueResourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEventInvokeConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeConfigConfig_destinationOnFailureDestinationSQSQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination", sqsQueueResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionEventInvokeConfigConfig_qualifierVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_Destination_swap(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"
	sqsQueueResourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEventInvokeConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeConfigConfig_destinationOnFailureDestinationSQSQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination", sqsQueueResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionEventInvokeConfigConfig_destinationOnSuccessDestinationSQSQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_success.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_success.0.destination", sqsQueueResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_FunctionName_arn(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEventInvokeConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeConfigConfig_nameARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEventInvokeConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeConfigConfig_qualifierNameARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEventInvokeConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeConfigConfig_maximumAgeInSeconds(rName, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_event_age_in_seconds", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionEventInvokeConfigConfig_maximumAgeInSeconds(rName, 200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_event_age_in_seconds", "200"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_maximumRetryAttempts(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEventInvokeConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeConfigConfig_maximumRetryAttempts(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionEventInvokeConfigConfig_maximumRetryAttempts(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", acctest.Ct1),
				),
			},
			{
				Config: testAccFunctionEventInvokeConfigConfig_maximumRetryAttempts(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccLambdaFunctionEventInvokeConfig_Qualifier_aliasName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaAliasResourceName := "aws_lambda_alias.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEventInvokeConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeConfigConfig_qualifierAliasName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaAliasResourceName, names.AttrName),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEventInvokeConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeConfigConfig_qualifierVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaFunctionResourceName, names.AttrVersion),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionEventInvokeConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEventInvokeConfigConfig_qualifierLatest(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionEventInvokeConfigExists(ctx, resourceName),
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

func testAccCheckFunctionEventInvokeConfigDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_function_event_invoke_config" {
				continue
			}

			functionName, qualifier, err := tflambda.FunctionEventInvokeConfigParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tflambda.FindFunctionEventInvokeConfigByTwoPartKey(ctx, conn, functionName, qualifier)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lambda Function Event Invoke Config %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFunctionEventInvokeConfigExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		functionName, qualifier, err := tflambda.FunctionEventInvokeConfigParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = tflambda.FindFunctionEventInvokeConfigByTwoPartKey(ctx, conn, functionName, qualifier)

		return err
	}
}

func testAccFunctionEventInvokeConfigConfig_base(rName string) string {
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
  runtime       = "nodejs16.x"

  depends_on = [
    aws_iam_role_policy_attachment.test,
  ]
}
`, rName)
}

func testAccFunctionEventInvokeConfigConfig_destinationOnFailureDestinationSNSTopic(rName string) string {
	return acctest.ConfigCompose(testAccFunctionEventInvokeConfigConfig_base(rName), fmt.Sprintf(`
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
`, rName))
}

func testAccFunctionEventInvokeConfigConfig_destinationOnFailureDestinationSQSQueue(rName string) string {
	return acctest.ConfigCompose(testAccFunctionEventInvokeConfigConfig_base(rName), fmt.Sprintf(`
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
`, rName))
}

func testAccFunctionEventInvokeConfigConfig_destinationOnSuccessDestinationSNSTopic(rName string) string {
	return acctest.ConfigCompose(testAccFunctionEventInvokeConfigConfig_base(rName), fmt.Sprintf(`
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
`, rName))
}

func testAccFunctionEventInvokeConfigConfig_destinationOnSuccessDestinationSQSQueue(rName string) string {
	return acctest.ConfigCompose(testAccFunctionEventInvokeConfigConfig_base(rName), fmt.Sprintf(`
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
`, rName))
}

func testAccFunctionEventInvokeConfigConfig_name(rName string) string {
	return acctest.ConfigCompose(testAccFunctionEventInvokeConfigConfig_base(rName), `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name
}
`)
}

func testAccFunctionEventInvokeConfigConfig_nameARN(rName string) string {
	return acctest.ConfigCompose(testAccFunctionEventInvokeConfigConfig_base(rName), `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.arn
}
`)
}

func testAccFunctionEventInvokeConfigConfig_qualifierNameARN(rName string) string {
	return acctest.ConfigCompose(testAccFunctionEventInvokeConfigConfig_base(rName), `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.arn
  qualifier     = "$LATEST"
}
`)
}

func testAccFunctionEventInvokeConfigConfig_maximumAgeInSeconds(rName string, maximumEventAgeInSeconds int) string {
	return acctest.ConfigCompose(testAccFunctionEventInvokeConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name                = aws_lambda_function.test.function_name
  maximum_event_age_in_seconds = %[1]d
}
`, maximumEventAgeInSeconds))
}

func testAccFunctionEventInvokeConfigConfig_maximumRetryAttempts(rName string, maximumRetryAttempts int) string {
	return acctest.ConfigCompose(testAccFunctionEventInvokeConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name          = aws_lambda_function.test.function_name
  maximum_retry_attempts = %[1]d
}
`, maximumRetryAttempts))
}

func testAccFunctionEventInvokeConfigConfig_qualifierAliasName(rName string) string {
	return acctest.ConfigCompose(testAccFunctionEventInvokeConfigConfig_base(rName), `
resource "aws_lambda_alias" "test" {
  function_name    = aws_lambda_function.test.function_name
  function_version = aws_lambda_function.test.version
  name             = "test"
}

resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_alias.test.function_name
  qualifier     = aws_lambda_alias.test.name
}
`)
}

func testAccFunctionEventInvokeConfigConfig_qualifierVersion(rName string) string {
	return acctest.ConfigCompose(testAccFunctionEventInvokeConfigConfig_base(rName), `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name
  qualifier     = aws_lambda_function.test.version
}
`)
}

func testAccFunctionEventInvokeConfigConfig_qualifierLatest(rName string) string {
	return acctest.ConfigCompose(testAccFunctionEventInvokeConfigConfig_base(rName), `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name
  qualifier     = "$LATEST"
}
`)
}
