package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSSfnStateMachine_createUpdate(t *testing.T) {
	resourceName := "aws_sfn_state_machine.test"
	rName := acctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineConfig(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", sfn.StateMachineTypeStandard),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					testAccMatchResourceAttrGlobalARN(resourceName, "role_arn", "iam", regexp.MustCompile(`role/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSfnStateMachineConfig(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", sfn.StateMachineTypeStandard),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					testAccMatchResourceAttrGlobalARN(resourceName, "role_arn", "iam", regexp.MustCompile(`role/.+`)),
				),
			},
		},
	})
}

func TestAccAWSSfnStateMachine_type_express(t *testing.T) {
	resourceName := "aws_sfn_state_machine.test"
	rName := acctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineConfigType(rName, "EXPRESS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "EXPRESS"),
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

func TestAccAWSSfnStateMachine_logging_configuration(t *testing.T) {
	resourceName := "aws_sfn_state_machine.test"
	rName := acctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineConfigLogging(rName, "ALL", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.include_execution_data", "true"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.level", "ALL"),
					//testAccMatchResourceAttrRegionalARN(resourceName, "logging_configuration.0.destinations.0.cloudwatch_log_group_arn", "cloudwatch", regexp.MustCompile(`role/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSfnStateMachineConfigLogging(rName, "ERROR", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.include_execution_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.level", "ERROR"),
					//testAccMatchResourceAttrRegionalARN(resourceName, "logging_configuration.0.destinations.0.cloudwatch_log_group_arn", "cloudwatch", regexp.MustCompile(`role/.+`)),
				),
			},
		},
	})
}

func TestAccAWSSfnStateMachine_Tags(t *testing.T) {
	resourceName := "aws_sfn_state_machine.test"
	rName := acctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName),
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
				Config: testAccAWSSfnStateMachineConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSfnStateMachineConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSSfnExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Step Function ID set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sfnconn

		_, err := conn.DescribeStateMachine(&sfn.DescribeStateMachineInput{
			StateMachineArn: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccCheckAWSSfnStateMachineDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sfnconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sfn_state_machine" {
			continue
		}

		out, err := conn.DescribeStateMachine(&sfn.DescribeStateMachineInput{
			StateMachineArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if wserr, ok := err.(awserr.Error); ok && wserr.Code() == sfn.ErrCodeStateMachineDoesNotExist {
				return nil
			}
			return err
		}

		if out != nil && *out.Status != sfn.StateMachineStatusDeleting {
			return fmt.Errorf("Expected AWS Step Function State Machine to be destroyed, but was still found")
		}

		return nil
	}

	return fmt.Errorf("Default error in Step Function Test")
}

func testAccAWSSfnStateMachineConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda_%s"

  assume_role_policy = <<EOF
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
EOF
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

data "aws_region" "current" {}

resource "aws_iam_role_policy" "iam_policy_for_sfn" {
  name = %[1]q
  role = "${aws_iam_role.iam_for_sfn.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:InvokeFunction"
      ],
        "Resource": "${aws_lambda_function.test.arn}"
	},
	{
    "Effect": "Allow",
    "Action": [
		"logs:CreateLogDelivery",
		"logs:GetLogDelivery",
		"logs:UpdateLogDelivery",
		"logs:DeleteLogDelivery",
		"logs:ListLogDeliveries",
		"logs:PutResourcePolicy",
		"logs:DescribeResourcePolicies",
		"logs:DescribeLogGroups"
    ],
    "Resource": "*"
  }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_sfn" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "states.${data.aws_region.current.name}.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Sid": ""
    }
  ]
}
EOF
}
`, rName)
}

func testAccAWSSfnStateMachineConfig(rName string, rMaxAttempts int) string {
	return testAccAWSSfnStateMachineConfigBase(rName) + fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = "%s"
  role_arn = "${aws_iam_role.iam_for_sfn.arn}"

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
          "MaxAttempts": %d,
          "BackoffRate": 8.0
        }
      ],
      "End": true
    }
  }
}
EOF
}
`, rName, rMaxAttempts)
}

func testAccAWSSfnStateMachineConfigLogging(rName, logLevel string, includeData bool) string {
	return testAccAWSSfnStateMachineConfigBase(rName) + fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  type     = "EXPRESS"
  role_arn = "${aws_iam_role.iam_for_sfn.arn}"

  logging_configuration {
	destinations {
		cloudwatch_log_group_arn = "${aws_cloudwatch_log_group.test.arn}"
    }
	include_execution_data = %[3]t
	level                  = %[2]q
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
          "ErrorEquals": ["States.ALL"],
          "IntervalSeconds": 5,
          "MaxAttempts": 5,
          "BackoffRate": 8.0
        }
      ],
      "End": true
    }
  }
}
EOF
}
`, rName, logLevel, includeData)
}

func testAccAWSSfnStateMachineConfigType(rName, typ string) string {
	return testAccAWSSfnStateMachineConfigBase(rName) + fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  type     = %[2]q
  role_arn = "${aws_iam_role.iam_for_sfn.arn}"

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
          "MaxAttempts": 5,
          "BackoffRate": 8.0
        }
      ],
      "End": true
    }
  }
}
EOF
}
`, rName, typ)
}

func testAccAWSSfnStateMachineConfigTags1(rName, tag1Key, tag1Value string) string {
	return testAccAWSSfnStateMachineConfigBase(rName) + fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = "${aws_iam_role.iam_for_sfn.arn}"

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
          "MaxAttempts": 5,
          "BackoffRate": 8.0
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
`, rName, tag1Key, tag1Value)
}

func testAccAWSSfnStateMachineConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return testAccAWSSfnStateMachineConfigBase(rName) + fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = "${aws_iam_role.iam_for_sfn.arn}"

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
          "MaxAttempts": 5,
          "BackoffRate": 8.0
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
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
