package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAwsBudget_basic(t *testing.T) {
	rInt := acctest.RandInt()
	name := fmt.Sprintf("test-budget-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		//	CheckDestroy: func() { testCheckBudgetDestroy(name) },
		Steps: []resource.TestStep{
			{
				Config: testBudgetConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("aws_budget.foo", name),
				),
			},
		},
	})
}

func testCheckBudgetDestroy(s *terraform.State) error {
	return fmt.Errorf("not yet implemented")
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
