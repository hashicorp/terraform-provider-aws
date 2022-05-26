package budgets_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/budgets"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbudgets "github.com/hashicorp/terraform-provider-aws/internal/service/budgets"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccBudgetsBudgetAction_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_budgets_budget_action.test"
	var conf budgets.Action

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(budgets.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, budgets.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccBudgetActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBudgetActionBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccBudgetActionExists(resourceName, &conf),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "budgets", regexp.MustCompile(fmt.Sprintf(`budget/%s/action/.+`, rName))),
					resource.TestCheckResourceAttrPair(resourceName, "budget_name", "aws_budgets_budget.test", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "execution_role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "action_type", "APPLY_IAM_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "approval_model", "AUTOMATIC"),
					resource.TestCheckResourceAttr(resourceName, "notification_type", "ACTUAL"),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.0.action_threshold_type", "ABSOLUTE_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "action_threshold.0.action_threshold_value", "100"),
					resource.TestCheckResourceAttr(resourceName, "definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.iam_action_definition.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "definition.0.iam_action_definition.0.policy_arn", "aws_iam_policy.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.iam_action_definition.0.roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subscriber.#", "1"),
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

func TestAccBudgetsBudgetAction_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_budgets_budget_action.test"
	var conf budgets.Action

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(budgets.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, budgets.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccBudgetActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBudgetActionBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccBudgetActionExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfbudgets.ResourceBudgetAction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccBudgetActionExists(resourceName string, config *budgets.Action) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Budget Action ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BudgetsConn

		accountID, actionID, budgetName, err := tfbudgets.BudgetActionParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := tfbudgets.FindActionByAccountIDActionIDAndBudgetName(conn, accountID, actionID, budgetName)

		if err != nil {
			return err
		}

		*config = *output

		return nil
	}
}

func testAccBudgetActionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BudgetsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_budgets_budget_action" {
			continue
		}

		accountID, actionID, budgetName, err := tfbudgets.BudgetActionParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfbudgets.FindActionByAccountIDActionIDAndBudgetName(conn, accountID, actionID, budgetName)

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

func testAccBudgetActionBasicConfig(rName string) string {
	return fmt.Sprintf(`
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

resource "aws_budgets_budget" "test" {
  name              = %[1]q
  budget_type       = "USAGE"
  limit_amount      = "10.0"
  limit_unit        = "dollars"
  time_period_start = "2006-01-02_15:04"
  time_unit         = "MONTHLY"
}

resource "aws_budgets_budget_action" "test" {
  budget_name        = aws_budgets_budget.test.name
  action_type        = "APPLY_IAM_POLICY"
  approval_model     = "AUTOMATIC"
  notification_type  = "ACTUAL"
  execution_role_arn = aws_iam_role.test.arn

  action_threshold {
    action_threshold_type  = "ABSOLUTE_VALUE"
    action_threshold_value = 100
  }

  definition {
    iam_action_definition {
      policy_arn = aws_iam_policy.test.arn
      roles      = [aws_iam_role.test.name]
    }
  }

  subscriber {
    address           = "test@test.test"
    subscription_type = "EMAIL"
  }
}
`, rName)
}
