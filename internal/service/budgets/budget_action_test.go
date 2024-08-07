// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package budgets_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/budgets/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbudgets "github.com/hashicorp/terraform-provider-aws/internal/service/budgets"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBudgetsBudgetAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_budgets_budget_action.test"
	var conf awstypes.Action

	const thresholdValue = "1000000000"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BudgetsEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BudgetsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBudgetActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBudgetActionConfig_basic(rName, string(awstypes.ApprovalModelAuto), thresholdValue),
				Check: resource.ComposeTestCheckFunc(
					testAccBudgetActionExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "budgets", regexache.MustCompile(fmt.Sprintf(`budget/%s/action/.+`, rName))),
					resource.TestCheckResourceAttrPair(resourceName, "budget_name", "aws_budgets_budget.test", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "action_type", "APPLY_IAM_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "approval_model", string(awstypes.ApprovalModelAuto)),
					resource.TestCheckResourceAttr(resourceName, "notification_type", "ACTUAL"),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.0.action_threshold_type", "ABSOLUTE_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.0.action_threshold_value", thresholdValue),
					resource.TestCheckResourceAttr(resourceName, "definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "definition.0.iam_action_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "definition.0.iam_action_definition.0.policy_arn", "aws_iam_policy.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "definition.0.iam_action_definition.0.roles.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "subscriber.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.ActionStatusStandby)),
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

func TestAccBudgetsBudgetAction_triggeredAutomatic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_budgets_budget_action.test"
	var conf awstypes.Action

	const thresholdValue = "100"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BudgetsEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BudgetsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBudgetActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBudgetActionConfig_basic(rName, string(awstypes.ApprovalModelAuto), thresholdValue),
				Check: resource.ComposeTestCheckFunc(
					testAccBudgetActionExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "budgets", regexache.MustCompile(fmt.Sprintf(`budget/%s/action/.+`, rName))),
					resource.TestCheckResourceAttrPair(resourceName, "budget_name", "aws_budgets_budget.test", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "action_type", "APPLY_IAM_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "approval_model", string(awstypes.ApprovalModelAuto)),
					resource.TestCheckResourceAttr(resourceName, "notification_type", "ACTUAL"),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.0.action_threshold_type", "ABSOLUTE_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.0.action_threshold_value", thresholdValue),
					resource.TestCheckResourceAttr(resourceName, "definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "definition.0.iam_action_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "definition.0.iam_action_definition.0.policy_arn", "aws_iam_policy.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "definition.0.iam_action_definition.0.roles.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "subscriber.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
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

func TestAccBudgetsBudgetAction_triggeredManual(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_budgets_budget_action.test"
	var conf awstypes.Action

	const thresholdValue = "100"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BudgetsEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BudgetsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBudgetActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBudgetActionConfig_basic(rName, string(awstypes.ApprovalModelManual), thresholdValue),
				Check: resource.ComposeTestCheckFunc(
					testAccBudgetActionExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "budgets", regexache.MustCompile(fmt.Sprintf(`budget/%s/action/.+`, rName))),
					resource.TestCheckResourceAttrPair(resourceName, "budget_name", "aws_budgets_budget.test", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "action_type", "APPLY_IAM_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "approval_model", string(awstypes.ApprovalModelManual)),
					resource.TestCheckResourceAttr(resourceName, "notification_type", "ACTUAL"),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.0.action_threshold_type", "ABSOLUTE_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.0.action_threshold_value", thresholdValue),
					resource.TestCheckResourceAttr(resourceName, "definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "definition.0.iam_action_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "definition.0.iam_action_definition.0.policy_arn", "aws_iam_policy.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "definition.0.iam_action_definition.0.roles.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "subscriber.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus), // Race condition between "STANDBY" and "PENDING"
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

func TestAccBudgetsBudgetAction_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_budgets_budget_action.test"
	var conf awstypes.Action

	const thresholdValue = "1000000000"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BudgetsEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BudgetsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBudgetActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBudgetActionConfig_tags1(rName, string(awstypes.ApprovalModelManual), thresholdValue, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccBudgetActionExists(ctx, resourceName, &conf),
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
				Config: testAccBudgetActionConfig_tags2(rName, string(awstypes.ApprovalModelManual), thresholdValue, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccBudgetActionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccBudgetActionConfig_tags1(rName, string(awstypes.ApprovalModelManual), thresholdValue, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccBudgetActionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccBudgetsBudgetAction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_budgets_budget_action.test"
	var conf awstypes.Action

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BudgetsEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BudgetsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBudgetActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBudgetActionConfig_basic(rName, string(awstypes.ApprovalModelAuto), "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccBudgetActionExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfbudgets.ResourceBudgetAction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccBudgetActionExists(ctx context.Context, resourceName string, config *awstypes.Action) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Budget Action ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BudgetsClient(ctx)

		accountID, actionID, budgetName, err := tfbudgets.BudgetActionParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := tfbudgets.FindBudgetWithDelay(ctx, func() (*awstypes.Action, error) {
			return tfbudgets.FindActionByThreePartKey(ctx, conn, accountID, actionID, budgetName)
		})

		if err != nil {
			return err
		}

		*config = *output

		return nil
	}
}

func testAccCheckBudgetActionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BudgetsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_budgets_budget_action" {
				continue
			}

			accountID, actionID, budgetName, err := tfbudgets.BudgetActionParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfbudgets.FindBudgetWithDelay(ctx, func() (*awstypes.Action, error) {
				return tfbudgets.FindActionByThreePartKey(ctx, conn, accountID, actionID, budgetName)
			})

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Budget Action %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBudgetActionConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_budgets_budget" "test" {
  name              = %[1]q
  budget_type       = "USAGE"
  limit_amount      = "1.0"
  limit_unit        = "dollars"
  time_period_start = "2006-01-02_15:04"
  time_unit         = "MONTHLY"
}

resource "aws_iam_policy" "test" {
  name        = %[1]q
  description = "My test policy"

  policy = <<EOF
	  {
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Action": [
			  "ec2:Describe*"
			],
			"Effect": "Allow",
			"Resource": "*"
		  }
		]
	  }
	  EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
	  {
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": [
				"budgets.${data.aws_partition.current.dns_suffix}"
			  ]
			},
			"Action": [
			  "sts:AssumeRole"
			]
		  }
		]
	  }
	  EOF
}
`, rName)
}

func testAccBudgetActionConfig_basic(rName, approvalModel, thresholdValue string) string {
	return acctest.ConfigCompose(
		testAccBudgetActionConfig_base(rName),
		fmt.Sprintf(`
resource "aws_budgets_budget_action" "test" {
  budget_name        = aws_budgets_budget.test.name
  action_type        = "APPLY_IAM_POLICY"
  approval_model     = %[2]q
  notification_type  = "ACTUAL"
  execution_role_arn = aws_iam_role.test.arn

  action_threshold {
    action_threshold_type  = "ABSOLUTE_VALUE"
    action_threshold_value = %[3]s
  }

  definition {
    iam_action_definition {
      policy_arn = aws_iam_policy.test.arn
      roles      = [aws_iam_role.test.name]
    }
  }

  subscriber {
    address           = %[4]q
    subscription_type = "EMAIL"
  }
}
`, rName, approvalModel, thresholdValue, acctest.DefaultEmailAddress))
}

func testAccBudgetActionConfig_tags1(rName, approvalModel, thresholdValue, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccBudgetActionConfig_base(rName),
		fmt.Sprintf(`
resource "aws_budgets_budget_action" "test" {
  budget_name        = aws_budgets_budget.test.name
  action_type        = "APPLY_IAM_POLICY"
  approval_model     = %[2]q
  notification_type  = "ACTUAL"
  execution_role_arn = aws_iam_role.test.arn

  action_threshold {
    action_threshold_type  = "ABSOLUTE_VALUE"
    action_threshold_value = %[3]s
  }

  definition {
    iam_action_definition {
      policy_arn = aws_iam_policy.test.arn
      roles      = [aws_iam_role.test.name]
    }
  }

  subscriber {
    address           = %[4]q
    subscription_type = "EMAIL"
  }

  tags = {
    %[5]q = %[6]q
  }
}
`, rName, approvalModel, thresholdValue, acctest.DefaultEmailAddress, tagKey1, tagValue1))
}

func testAccBudgetActionConfig_tags2(rName, approvalModel, thresholdValue, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccBudgetActionConfig_base(rName),
		fmt.Sprintf(`
resource "aws_budgets_budget_action" "test" {
  budget_name        = aws_budgets_budget.test.name
  action_type        = "APPLY_IAM_POLICY"
  approval_model     = %[2]q
  notification_type  = "ACTUAL"
  execution_role_arn = aws_iam_role.test.arn

  action_threshold {
    action_threshold_type  = "ABSOLUTE_VALUE"
    action_threshold_value = %[3]s
  }

  definition {
    iam_action_definition {
      policy_arn = aws_iam_policy.test.arn
      roles      = [aws_iam_role.test.name]
    }
  }

  subscriber {
    address           = %[4]q
    subscription_type = "EMAIL"
  }

  tags = {
    %[5]q = %[6]q
    %[7]q = %[8]q
  }
}
`, rName, approvalModel, thresholdValue, acctest.DefaultEmailAddress, tagKey1, tagValue1, tagKey2, tagValue2))
}
