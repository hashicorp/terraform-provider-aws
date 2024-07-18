// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsfn "github.com/hashicorp/terraform-provider-aws/internal/service/sfn"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSFNStateMachine_createUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	roleResourceName := "aws_iam_role.for_sfn"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStateMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_basic(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "states", fmt.Sprintf("stateMachine:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StateMachineStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexache.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.include_execution_data", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination", ""),
					resource.TestCheckResourceAttr(resourceName, "publish", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "revision_id", ""),
					resource.TestCheckResourceAttr(resourceName, "state_machine_version_arn", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "STANDARD"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStateMachineConfig_basic(rName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "states", fmt.Sprintf("stateMachine:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StateMachineStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestMatchResourceAttr(resourceName, "definition", regexache.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.include_execution_data", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination", ""),
					resource.TestCheckResourceAttr(resourceName, "publish", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "state_machine_version_arn", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "STANDARD"),
				),
			},
		},
	})
}

func TestAccSFNStateMachine_expressUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStateMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_typed(rName, "EXPRESS", 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StateMachineStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexache.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttr(resourceName, "revision_id", ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttrWith(resourceName, "state_machine_version_arn", func(value string) error {
						if !strings.HasSuffix(value, ":1") {
							return fmt.Errorf("incorrect version number: %s", value)
						}

						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "EXPRESS"),
				),
			},
			{
				Config: testAccStateMachineConfig_typed(rName, "EXPRESS", 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StateMachineStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestMatchResourceAttr(resourceName, "definition", regexache.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttrWith(resourceName, "state_machine_version_arn", func(value string) error {
						if !strings.HasSuffix(value, ":2") {
							return fmt.Errorf("incorrect version number: %s", value)
						}

						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "EXPRESS"),
				),
			},
		},
	})
}

func TestAccSFNStateMachine_standardUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStateMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_typed(rName, "STANDARD", 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StateMachineStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexache.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttr(resourceName, "publish", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revision_id", ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttrWith(resourceName, "state_machine_version_arn", func(value string) error {
						if !strings.HasSuffix(value, ":1") {
							return fmt.Errorf("incorrect version number: %s", value)
						}

						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "STANDARD"),
				),
			},
			{
				Config: testAccStateMachineConfig_typed(rName, "STANDARD", 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StateMachineStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestMatchResourceAttr(resourceName, "definition", regexache.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					resource.TestCheckResourceAttr(resourceName, "publish", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttrWith(resourceName, "state_machine_version_arn", func(value string) error {
						if !strings.HasSuffix(value, ":2") {
							return fmt.Errorf("incorrect version number: %s", value)
						}

						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "STANDARD"),
				),
			},
		},
	})
}

func TestAccSFNStateMachine_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStateMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, id.UniqueIdPrefix),
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

func TestAccSFNStateMachine_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStateMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_namePrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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

func TestAccSFNStateMachine_publish(t *testing.T) {
	ctx := acctest.Context(t)
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStateMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_publish(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "publish", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "state_machine_version_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"publish"},
			},
		},
	})
}

func TestAccSFNStateMachine_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStateMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStateMachineConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccStateMachineConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSFNStateMachine_tracing(t *testing.T) {
	ctx := acctest.Context(t)
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStateMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_tracingDisable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.0.enabled", acctest.CtFalse),
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
					testAccCheckExists(ctx, resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.0.enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccSFNStateMachine_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStateMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_basic(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsfn.ResourceStateMachine(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSFNStateMachine_expressLogging(t *testing.T) {
	ctx := acctest.Context(t)
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStateMachineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineConfig_expressLogConfiguration(rName, string(awstypes.LogLevelError)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StateMachineStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexache.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.level", string(awstypes.LogLevelError)),
				),
			},
			{
				Config: testAccStateMachineConfig_expressLogConfiguration(rName, string(awstypes.LogLevelAll)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StateMachineStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestMatchResourceAttr(resourceName, "definition", regexache.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.level", string(awstypes.LogLevelAll)),
				),
			},
		},
	})
}

func testAccCheckExists(ctx context.Context, n string, v *sfn.DescribeStateMachineOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SFNClient(ctx)

		output, err := tfsfn.FindStateMachineByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStateMachineDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SFNClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sfn_state_machine" {
				continue
			}

			_, err := tfsfn.FindStateMachineByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Step Functions State Machine %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccStateMachineConfig_base(rName string) string {
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
`, rName)
}

func testAccStateMachineConfig_basic(rName string, rMaxAttempts int) string {
	return acctest.ConfigCompose(testAccStateMachineConfig_base(rName), fmt.Sprintf(`
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

func testAccStateMachineConfig_nameGenerated(rName string) string {
	return acctest.ConfigCompose(testAccStateMachineConfig_base(rName), `
resource "aws_sfn_state_machine" "test" {
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
}
`)
}

func testAccStateMachineConfig_namePrefix(rName, namePrefix string) string {
	return acctest.ConfigCompose(testAccStateMachineConfig_base(rName), fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name_prefix = %[1]q
  role_arn    = aws_iam_role.for_sfn.arn

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
`, namePrefix))
}

func testAccStateMachineConfig_publish(rName string) string {
	return acctest.ConfigCompose(testAccStateMachineConfig_base(rName), `
resource "aws_sfn_state_machine" "test" {
  role_arn = aws_iam_role.for_sfn.arn
  publish  = true

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
`)
}

func testAccStateMachineConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccStateMachineConfig_base(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccStateMachineConfig_base(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccStateMachineConfig_base(rName), fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn
  type     = %[2]q
  publish  = true

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
	return acctest.ConfigCompose(testAccStateMachineConfig_base(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccStateMachineConfig_base(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccStateMachineConfig_base(rName), fmt.Sprintf(`
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
