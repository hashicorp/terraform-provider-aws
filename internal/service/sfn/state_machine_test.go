package sfn_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/sfn"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsfn "github.com/hashicorp/terraform-provider-aws/internal/service/sfn"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSFNStateMachine_createUpdate(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	roleResourceName := "aws_iam_role.for_sfn"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sfn.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_basic(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName, &sm),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "states", fmt.Sprintf("stateMachine:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.include_execution_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "type", "STANDARD"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStateMachineConfig_basic(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName, &sm),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "states", fmt.Sprintf("stateMachine:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.include_execution_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "type", "STANDARD"),
				),
			},
		},
	})
}

func TestAccSFNStateMachine_expressUpdate(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sfn.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_typed(rName, "EXPRESS", 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "EXPRESS"),
				),
			},
			{
				Config: testAccStateMachineConfig_typed(rName, "EXPRESS", 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "EXPRESS"),
				),
			},
		},
	})
}

func TestAccSFNStateMachine_standardUpdate(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sfn.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_typed(rName, "STANDARD", 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "STANDARD"),
				),
			},
			{
				Config: testAccStateMachineConfig_typed(rName, "STANDARD", 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "STANDARD"),
				),
			},
		},
	})
}

func TestAccSFNStateMachine_tags(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sfn.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStateMachineConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccStateMachineConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSFNStateMachine_tracing(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sfn.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_tracingDisable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.0.enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStateMachineConfig_tracingEnable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccSFNStateMachine_disappears(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sfn.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_basic(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName, &sm),
					acctest.CheckResourceDisappears(acctest.Provider, tfsfn.ResourceStateMachine(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSFNStateMachine_expressLogging(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sfn.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_expressLogConfiguration(rName, sfn.LogLevelError),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.level", sfn.LogLevelError),
				),
			},
			{
				Config: testAccStateMachineConfig_expressLogConfiguration(rName, sfn.LogLevelAll),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.level", sfn.LogLevelAll),
				),
			},
		},
	})
}

func testAccCheckExists(n string, v *sfn.DescribeStateMachineOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Step Function State Machine ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SFNConn

		output, err := tfsfn.FindStateMachineByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStateMachineDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SFNConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sfn_state_machine" {
			continue
		}

		_, err := tfsfn.FindStateMachineByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Step Function State Machine %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccStateMachineBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role_policy" "for_lambda" {
  name = "%[1]s-lambda"
  role = aws_iam_role.for_lambda.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ],
    "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
  }]
}
EOF
}

resource "aws_iam_role" "for_lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "for_sfn" {
  name = "%[1]s-sfn"
  role = aws_iam_role.for_sfn.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "lambda:InvokeFunction",
      "logs:CreateLogDelivery",
      "logs:GetLogDelivery",
      "logs:UpdateLogDelivery",
      "logs:DeleteLogDelivery",
      "logs:ListLogDeliveries",
      "logs:PutResourcePolicy",
      "logs:DescribeResourcePolicies",
      "logs:DescribeLogGroups",
      "xray:PutTraceSegments",
      "xray:PutTelemetryRecords",
      "xray:GetSamplingRules",
      "xray:GetSamplingTargets"
    ],
    "Resource": "*"
  }]
}
EOF
}

resource "aws_iam_role" "for_sfn" {
  name = "%[1]s-sfn"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "states.${data.aws_region.current.name}.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}
`, rName)
}

func testAccStateMachineConfig_basic(rName string, rMaxAttempts int) string {
	return acctest.ConfigCompose(testAccStateMachineBaseConfig(rName), fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.test.arn}",
      "Retry": [
        {
          "ErrorEquals": [
            "States.ALL"
          ],
          "IntervalSeconds": 5,
          "MaxAttempts": %[2]d,
          "BackoffRate": 8
        }
      ],
      "End": true
    }
  }
}
EOF
}
`, rName, rMaxAttempts))
}

func testAccStateMachineConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccStateMachineBaseConfig(rName), fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.test.arn}",
      "Retry": [
        {
          "ErrorEquals": [
            "States.ALL"
          ],
          "IntervalSeconds": 5,
          "MaxAttempts": 5,
          "BackoffRate": 8
        }
      ],
      "End": true
    }
  }
}
EOF

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value))
}

func testAccStateMachineConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccStateMachineBaseConfig(rName), fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.test.arn}",
      "Retry": [
        {
          "ErrorEquals": [
            "States.ALL"
          ],
          "IntervalSeconds": 5,
          "MaxAttempts": 5,
          "BackoffRate": 8
        }
      ],
      "End": true
    }
  }
}
EOF

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccStateMachineConfig_typed(rName, rType string, rMaxAttempts int) string {
	return acctest.ConfigCompose(testAccStateMachineBaseConfig(rName), fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn
  type     = %[2]q

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.test.arn}",
      "Retry": [
        {
          "ErrorEquals": ["States.ALL"],
          "IntervalSeconds": 5,
          "MaxAttempts": %[3]d,
          "BackoffRate": 8.0
        }
      ],
      "End": true
    }
  }
}
EOF
}
`, rName, rType, rMaxAttempts))
}

func testAccStateMachineConfig_expressLogConfiguration(rName string, rLevel string) string {
	return acctest.ConfigCompose(testAccStateMachineBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn
  type     = "EXPRESS"

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.test.arn}",
      "Retry": [
        {
          "ErrorEquals": [
            "States.ALL"
          ],
          "IntervalSeconds": 5,
          "MaxAttempts": 5,
          "BackoffRate": 8
        }
      ],
      "End": true
    }
  }
}
EOF

  logging_configuration {
    log_destination        = "${aws_cloudwatch_log_group.test.arn}:*"
    include_execution_data = false
    level                  = %[2]q
  }
}
`, rName, rLevel))
}

func testAccStateMachineConfig_tracingEnable(rName string) string {
	return acctest.ConfigCompose(testAccStateMachineBaseConfig(rName), fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn

  tracing_configuration {
    enabled = true
  }

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.test.arn}",
      "Retry": [
        {
          "ErrorEquals": [
            "States.ALL"
          ],
          "IntervalSeconds": 5,
          "MaxAttempts": 5,
          "BackoffRate": 8
        }
      ],
      "End": true
    }
  }
}
EOF
}
`, rName))
}

func testAccStateMachineConfig_tracingDisable(rName string) string {
	return acctest.ConfigCompose(testAccStateMachineBaseConfig(rName), fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn

  tracing_configuration {
    enabled = false
  }

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.test.arn}",
      "Retry": [
        {
          "ErrorEquals": [
            "States.ALL"
          ],
          "IntervalSeconds": 5,
          "MaxAttempts": 5,
          "BackoffRate": 8
        }
      ],
      "End": true
    }
  }
}
EOF
}
`, rName))
}
