// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sfn_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sfn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSFNStartExecutionAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	inputJSON := `{"key1":"value1","key2":"value2"}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccStartExecutionActionConfig_basic(rName, inputJSON),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStartExecutionAction(ctx, rName, inputJSON),
				),
			},
		},
	})
}

func TestAccSFNStartExecutionAction_withName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	executionName := sdkacctest.RandomWithPrefix("execution")
	inputJSON := `{"test":"data"}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccStartExecutionActionConfig_withName(rName, executionName, inputJSON),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStartExecutionActionWithName(ctx, rName, executionName, inputJSON),
				),
			},
		},
	})
}

func TestAccSFNStartExecutionAction_emptyInput(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccStartExecutionActionConfig_emptyInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStartExecutionAction(ctx, rName, "{}"),
				),
			},
		},
	})
}

// Test helper functions

func testAccCheckStartExecutionAction(ctx context.Context, stateMachineName, expectedInput string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SFNClient(ctx)

		// Get the state machine ARN
		stateMachines, err := conn.ListStateMachines(ctx, &sfn.ListStateMachinesInput{})
		if err != nil {
			return fmt.Errorf("failed to list state machines: %w", err)
		}

		var stateMachineArn string
		for _, sm := range stateMachines.StateMachines {
			if *sm.Name == stateMachineName {
				stateMachineArn = *sm.StateMachineArn
				break
			}
		}

		if stateMachineArn == "" {
			return fmt.Errorf("state machine %s not found", stateMachineName)
		}

		// List executions to verify one was created (check all statuses)
		executions, err := conn.ListExecutions(ctx, &sfn.ListExecutionsInput{
			StateMachineArn: &stateMachineArn,
		})
		if err != nil {
			return fmt.Errorf("failed to list executions for state machine %s: %w", stateMachineName, err)
		}

		if len(executions.Executions) == 0 {
			return fmt.Errorf("no executions found for state machine %s", stateMachineName)
		}

		// Verify the execution input matches expected
		execution := executions.Executions[0]
		executionDetails, err := conn.DescribeExecution(ctx, &sfn.DescribeExecutionInput{
			ExecutionArn: execution.ExecutionArn,
		})
		if err != nil {
			return fmt.Errorf("failed to describe execution %s: %w", *execution.ExecutionArn, err)
		}

		if *executionDetails.Input != expectedInput {
			return fmt.Errorf("execution input mismatch. Expected: %s, Got: %s", expectedInput, *executionDetails.Input)
		}

		return nil
	}
}

func testAccCheckStartExecutionActionWithName(ctx context.Context, stateMachineName, executionName, expectedInput string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SFNClient(ctx)

		// Get the state machine ARN
		stateMachines, err := conn.ListStateMachines(ctx, &sfn.ListStateMachinesInput{})
		if err != nil {
			return fmt.Errorf("failed to list state machines: %w", err)
		}

		var stateMachineArn string
		for _, sm := range stateMachines.StateMachines {
			if *sm.Name == stateMachineName {
				stateMachineArn = *sm.StateMachineArn
				break
			}
		}

		if stateMachineArn == "" {
			return fmt.Errorf("state machine %s not found", stateMachineName)
		}

		// Find execution by name (check all statuses)
		executions, err := conn.ListExecutions(ctx, &sfn.ListExecutionsInput{
			StateMachineArn: &stateMachineArn,
		})
		if err != nil {
			return fmt.Errorf("failed to list executions for state machine %s: %w", stateMachineName, err)
		}

		var foundExecution *awstypes.ExecutionListItem
		for _, execution := range executions.Executions {
			if *execution.Name == executionName {
				foundExecution = &execution
				break
			}
		}

		if foundExecution == nil {
			return fmt.Errorf("execution with name %s not found for state machine %s", executionName, stateMachineName)
		}

		// Verify the execution input
		executionDetails, err := conn.DescribeExecution(ctx, &sfn.DescribeExecutionInput{
			ExecutionArn: foundExecution.ExecutionArn,
		})
		if err != nil {
			return fmt.Errorf("failed to describe execution %s: %w", *foundExecution.ExecutionArn, err)
		}

		if *executionDetails.Input != expectedInput {
			return fmt.Errorf("execution input mismatch. Expected: %s, Got: %s", expectedInput, *executionDetails.Input)
		}

		return nil
	}
}

// Configuration functions

func testAccStartExecutionActionConfig_basic(rName, inputJSON string) string {
	return acctest.ConfigCompose(
		testAccStartExecutionActionConfig_base(rName),
		fmt.Sprintf(`
action "aws_sfn_start_execution" "test" {
  config {
    state_machine_arn = aws_sfn_state_machine.test.arn
    input             = %[1]q
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_sfn_start_execution.test]
    }
  }
}
`, inputJSON))
}

func testAccStartExecutionActionConfig_withName(rName, executionName, inputJSON string) string {
	return acctest.ConfigCompose(
		testAccStartExecutionActionConfig_base(rName),
		fmt.Sprintf(`
action "aws_sfn_start_execution" "test" {
  config {
    state_machine_arn = aws_sfn_state_machine.test.arn
    name              = %[1]q
    input             = %[2]q
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_sfn_start_execution.test]
    }
  }
}
`, executionName, inputJSON))
}

func testAccStartExecutionActionConfig_emptyInput(rName string) string {
	return acctest.ConfigCompose(
		testAccStartExecutionActionConfig_base(rName),
		`
action "aws_sfn_start_execution" "test" {
  config {
    state_machine_arn = aws_sfn_state_machine.test.arn
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_sfn_start_execution.test]
    }
  }
}
`)
}

func testAccStartExecutionActionConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

data "aws_service_principal" "lambda" {
  service_name = "lambda"
}

data "aws_service_principal" "states" {
  service_name = "states"
  region       = data.aws_region.current.name
}

resource "aws_iam_role" "for_lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Principal = {
        Service = data.aws_service_principal.lambda.name
      }
      Effect = "Allow"
    }]
  })
}

resource "aws_iam_role_policy" "for_lambda" {
  name = "%[1]s-lambda"
  role = aws_iam_role.for_lambda.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ]
      Resource = "arn:${data.aws_partition.current.partition}:logs:*:*:*"
    }]
  })
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
}

resource "aws_iam_role" "for_sfn" {
  name = "%[1]s-sfn"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = data.aws_service_principal.states.name
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "for_sfn" {
  name = "%[1]s-sfn"
  role = aws_iam_role.for_sfn.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "lambda:InvokeFunction"
      ]
      Resource = "*"
    }]
  })
}

resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn

  definition = jsonencode({
    Comment = "A simple minimal example"
    StartAt = "Hello"
    States = {
      Hello = {
        Type     = "Task"
        Resource = aws_lambda_function.test.arn
        End      = true
      }
    }
  })
}
`, rName)
}
