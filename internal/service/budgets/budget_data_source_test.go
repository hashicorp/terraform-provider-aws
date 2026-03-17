// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package budgets_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBudgetsBudgetDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_budgets_budget.test"
	dataSourceName := "data.aws_budgets_budget.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BudgetsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBudgetDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrAccountID(ctx, dataSourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(dataSourceName, "calculated_spend.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "budget_limit.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsKey1, resourceName, acctest.CtTagsKey1),
					resource.TestCheckResourceAttrPair(dataSourceName, "billing_view_arn", resourceName, "billing_view_arn"),
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
  tags = {
    "key1" = "value1updated"
  }
}

data "aws_budgets_budget" "test" {
  name = aws_budgets_budget.test.name
}
`, rName)
}
