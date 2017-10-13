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

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		//CheckDestroy: testCheckBudgetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testBudgetConfig_basic(fmt.Sprintf("test-budget-%d", rInt)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput(
						"aws_budget.foo", fmt.Sprintf("test-budget-%d", rInt)),
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
	name = "%s"
	type = "COST"
 	limit_amount = "100"
 	limit_unit = "USD"
	include_tax = "true"
	include_subscriptions = "false"
	include_blended = "false"
	time_period_start = 1477353600000
	time_period_end = 1477958399000
 	time_unit = "MONTHLY"
}`, name)
}
