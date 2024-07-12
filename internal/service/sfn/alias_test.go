// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sfn"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsfn "github.com/hashicorp/terraform-provider-aws/internal/service/sfn"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSFNAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var alias sfn.DescribeStateMachineAliasOutput
	rString := sdkacctest.RandString(8)
	stateMachineName := fmt.Sprintf("tf_acc_state_machine_alias_basic_%s", rString)
	aliasName := fmt.Sprintf("tf_acc_state_machine_alias_basic_%s", rString)
	resourceName := "aws_sfn_alias.test"
	functionArnResourcePart := fmt.Sprintf("stateMachine:%s:%s", stateMachineName, aliasName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineAliasConfig_basic(stateMachineName, aliasName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &alias),
					testAccCheckAliasAttributes(&alias),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, aliasName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "states", functionArnResourcePart),
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

func TestAccSFNAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var alias sfn.DescribeStateMachineAliasOutput
	rString := sdkacctest.RandString(8)
	stateMachineName := fmt.Sprintf("tf_acc_state_machine_alias_basic_%s", rString)
	aliasName := fmt.Sprintf("tf_acc_state_machine_alias_basic_%s", rString)
	resourceName := "aws_sfn_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineAliasConfig_basic(stateMachineName, aliasName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, resourceName, &alias),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsfn.ResourceAlias(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAliasAttributes(mapping *sfn.DescribeStateMachineAliasOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := *mapping.Name
		arn := *mapping.StateMachineAliasArn
		if arn == "" {
			return fmt.Errorf("Could not read StateMachine alias ARN")
		}
		if name == "" {
			return fmt.Errorf("Could not read StateMachine alias name")
		}
		return nil
	}
}

func testAccCheckAliasDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SFNClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sfn_alias" {
				continue
			}

			_, err := tfsfn.FindAliasByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Step Functions State Machine Alias %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAliasExists(ctx context.Context, name string, v *sfn.DescribeStateMachineAliasOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SFNClient(ctx)

		output, err := tfsfn.FindAliasByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccStateMachineAliasConfig_base(rName string, rMaxAttempts int) string {
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
  runtime       = "nodejs16.x"
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

resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  publish  = true
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
`, rName, rMaxAttempts)
}

func testAccStateMachineAliasConfig_basic(statemachineName string, aliasName string, rMaxAttempts int) string {
	return acctest.ConfigCompose(testAccStateMachineAliasConfig_base(statemachineName, rMaxAttempts), fmt.Sprintf(`
resource "aws_sfn_alias" "test" {
  name = %[1]q

  routing_configuration {
    state_machine_version_arn = aws_sfn_state_machine.test.state_machine_version_arn
    weight                    = 100
  }
}
`, aliasName))
}
