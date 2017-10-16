package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAwsBudget_basic(t *testing.T) {
	name := fmt.Sprintf("test-budget-%d", acctest.RandInt())
	controlLimitA := "100"
	controlLimitB := "500"
	filterKey := "AZ"
	filterValueA := "us-east-1"
	filterValueB := "us-east-2"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return testCheckBudgetDestroy(name, testAccProvider)
		},
		Steps: []resource.TestStep{
			{
				Config: testBudgetConfig_basic(name, controlLimitA, filterKey, filterValueA),
				Check:  newComposedBudgetTestCheck(name, controlLimitA, filterKey, filterValueA, testAccProvider),
			},

			{
				Config: testBudgetConfig_basic(name, controlLimitB, filterKey, filterValueB),
				Check:  newComposedBudgetTestCheck(name, controlLimitB, filterKey, filterValueB, testAccProvider),
			},
		},
	})
}

func newComposedBudgetTestCheck(name, limit, filterKey, filterValue string, provider *schema.Provider) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("aws_budget.foo", "budget_name", name),
		resource.TestCheckResourceAttr("aws_budget.foo", "budget_type", "COST"),
		resource.TestCheckResourceAttr("aws_budget.foo", "limit_amount", limit),
		resource.TestCheckResourceAttr("aws_budget.foo", "limit_unit", "USD"),
		resource.TestCheckResourceAttr("aws_budget.foo", "include_tax", "true"),
		resource.TestCheckResourceAttr("aws_budget.foo", "include_subscriptions", "false"),
		resource.TestCheckResourceAttr("aws_budget.foo", "include_blended", "false"),
		resource.TestCheckResourceAttr("aws_budget.foo", "time_period_start", "2017-01-01_12:00"),
		resource.TestCheckResourceAttr("aws_budget.foo", "time_period_end", "2018-01-01_12:00"),
		resource.TestCheckResourceAttr("aws_budget.foo", "time_unit", "MONTHLY"),
		testBudgetExists(name, limit, filterKey, filterValue, provider),
	)
}

func testBudgetExists(budgetName, limit, filterKey, filterValue string, provider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		b, err := describeBudget(budgetName, provider.Meta())
		if err != nil {
			return fmt.Errorf("Describebudget error: %v", err)
		}

		if b.Budget == nil {
			return fmt.Errorf("No budget returned %v in %v", b.Budget, b)
		}

		if *b.Budget.BudgetLimit.Amount != limit {
			return fmt.Errorf("budget limit incorrectly set %v != %v", limit, *b.Budget.BudgetLimit.Amount)
		}

		if v, ok := b.Budget.CostFilters[filterKey]; !ok || *v[len(v)-1] != filterValue {
			return fmt.Errorf("cost filter not set properly: %v", b.Budget.CostFilters)
		}
		return nil
	}
}

func testCheckBudgetDestroy(budgetName string, provider *schema.Provider) error {
	if budgetExists(budgetName, provider.Meta()) {
		return fmt.Errorf("Budget '%s' was not deleted properly", budgetName)
	}

	return nil
}

func testBudgetConfig_basic(name, limit, filterKey, filterValue string) string {
	return fmt.Sprintf(`
resource "aws_budget" "foo" {
	budget_name = "%s"
	budget_type = "COST"
 	limit_amount = "%s"
 	limit_unit = "USD"
	include_tax = "true"
	include_subscriptions = "false"
	include_blended = "false"
	time_period_start = "2017-01-01_12:00" 
	time_period_end = "2018-01-01_12:00"
 	time_unit = "MONTHLY"
	cost_filters {
		%s = "%s"
	}
}`, name, limit, filterKey, filterValue)
}
