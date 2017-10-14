package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAwsBudget_basic(t *testing.T) {
	rInt := acctest.RandInt()
	name := fmt.Sprintf("test-budget-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return testCheckBudgetDestroy(name, testAccProvider)
		},
		Steps: []resource.TestStep{
			{
				Config: testBudgetConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_budget.foo", "budget_name", name),
					resource.TestCheckResourceAttr("aws_budget.foo", "budget_type", "COST"),
					resource.TestCheckResourceAttr("aws_budget.foo", "limit_amount", "100"),
					resource.TestCheckResourceAttr("aws_budget.foo", "limit_unit", "USD"),
					resource.TestCheckResourceAttr("aws_budget.foo", "include_tax", "true"),
					resource.TestCheckResourceAttr("aws_budget.foo", "include_subscriptions", "false"),
					resource.TestCheckResourceAttr("aws_budget.foo", "include_blended", "false"),
					resource.TestCheckResourceAttr("aws_budget.foo", "time_period_start", "2017-01-01_12:00"),
					resource.TestCheckResourceAttr("aws_budget.foo", "time_period_end", "2018-01-01_12:00"),
					resource.TestCheckResourceAttr("aws_budget.foo", "time_unit", "MONTHLY"),
				),
			},
		},
	})
}

func testCheckBudgetDestroy(budgetName string, provider *schema.Provider) error {
	client := provider.Meta().(*AWSClient).budgetconn
	accountID := provider.Meta().(*AWSClient).accountid
	describeBudgetInput := new(budgets.DescribeBudgetInput)
	describeBudgetInput.SetBudgetName(budgetName)
	describeBudgetInput.SetAccountId(accountID)
	b, err := client.DescribeBudget(describeBudgetInput)
	if b.Budget != nil {
		return fmt.Errorf("Budget not deleted properly %v %v", b.Budget, err)
	}
	return nil
}

func testBudgetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_budget" "foo" {
	budget_name = "%s"
	budget_type = "COST"
 	limit_amount = "100"
 	limit_unit = "USD"
	include_tax = "true"
	include_subscriptions = "false"
	include_blended = "false"
	time_period_start = "2017-01-01_12:00" 
	time_period_end = "2018-01-01_12:00"
 	time_unit = "MONTHLY"
	cost_filters {
		AZ = "us-east-1"
	}
}`, name)
}
