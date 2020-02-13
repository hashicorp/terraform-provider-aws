package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSSfnStateMachine_createUpdate(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	roleResourceName := "aws_iam_role.iam_for_sfn"
	rName := acctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineConfig(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "states", regexp.MustCompile(`stateMachine:.+`)),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
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
					testAccCheckAWSSfnExists(resourceName, &sm),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "states", regexp.MustCompile(`stateMachine:.+`)),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSSfnStateMachine_Tags(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
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
					testAccCheckAWSSfnExists(resourceName, &sm),
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
					testAccCheckAWSSfnExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSfnStateMachineConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSfnStateMachine_disappears(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := acctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					testAccCheckAWSSfnStateMachineDisappears(&sm),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSfnExists(n string, sm *sfn.DescribeStateMachineOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Step Function ID set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sfnconn

		resp, err := conn.DescribeStateMachine(&sfn.DescribeStateMachineInput{
			StateMachineArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*sm = *resp

		return nil
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
			if isAWSErr(err, sfn.ErrCodeStateMachineDoesNotExist, "") {
				return nil
			}
			return err
		}

		if out != nil && *out.Status != sfn.StateMachineStatusDeleting {
			return fmt.Errorf("Expected AWS Step Function State Machine to be destroyed, but was still found")
		}

		return err
	}

	return nil
}

func testAccCheckAWSSfnStateMachineDisappears(sm *sfn.DescribeStateMachineOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).sfnconn

		input := &sfn.DeleteStateMachineInput{
			StateMachineArn: sm.StateMachineArn,
		}

		if _, err := conn.DeleteStateMachine(input); err != nil {
			return err
		}

		return resource.Retry(5*time.Minute, func() *resource.RetryError {
			opts := &sfn.DescribeStateMachineInput{
				StateMachineArn: sm.StateMachineArn,
			}
			resp, err := conn.DescribeStateMachine(opts)
			if err != nil {
				if isAWSErr(err, sfn.ErrCodeStateMachineDoesNotExist, "") {
					return nil
				} else {
					return resource.NonRetryableError(fmt.Errorf("Error While Deleting State Machine: %v", resp.StateMachineArn))
				}
			}
			if aws.StringValue(resp.Status) == sfn.StateMachineStatusDeleting {
				return resource.RetryableError(fmt.Errorf("Waiting for State Machine: %v Deletion", resp.StateMachineArn))
			}
			return nil
		})
	}
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

func testAccAWSSfnStateMachineConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = "${aws_iam_role.test.arn}"

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using a Pass state",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Pass",
      "Result": "Hello World!",
      "End": true
    }
  }
}
EOF
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "states.amazonaws.com"
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
  name     = %q
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
