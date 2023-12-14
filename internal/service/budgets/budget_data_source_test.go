// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package budgets_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/budgets"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccBudgetsBudgetDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var budget budgets.Budget
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_budgets_budget.test"
	dataSourceName := "data.aws_budgets_budget.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, budgets.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBudgetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBudgetDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccBudgetExists(ctx, resourceName, &budget),
					acctest.CheckResourceAttrAccountID(dataSourceName, "account_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "calculated_spend.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "budget_limit.#"),
				),
			},
		},
	})
}

func testAccBudgetDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_budgets_budget" "test" {
  name         = %[1]q
  budget_type  = "RI_UTILIZATION"
  limit_amount = "100.0"
  limit_unit   = "PERCENTAGE"
  time_unit    = "QUARTERLY"

  cost_filter {
    name   = "Service"
    values = ["Amazon Redshift"]
  }
}

data "aws_budgets_budget" "test" {
  name = aws_budgets_budget.test.name
}
`, rName)
}
