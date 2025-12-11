// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package billing_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBillingViewsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_billing_views.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BillingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccViewsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "billing_view.#", "1"),
					acctest.CheckResourceAttrGlobalARN(ctx, dataSourceName, "billing_view.0.arn", "billing", "billingview/primary"),
					resource.TestCheckResourceAttr(dataSourceName, "billing_view.0.billing_view_type", "PRIMARY"),
					resource.TestCheckResourceAttr(dataSourceName, "billing_view.0.name", "Primary View"),
				),
			},
			{
				Config: testAccViewsDataSourceConfig_noArguments(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "billing_view.*", map[string]string{
						"billing_view_type": "PRIMARY",
						names.AttrName:      "Primary View",
					}),
				),
			},
		},
	})
}

func testAccViewsDataSourceConfig_basic() string {
	return `
data "aws_billing_views" "test" {
  billing_view_types = ["PRIMARY"]
}
`
}

func testAccViewsDataSourceConfig_noArguments() string {
	return `
data "aws_billing_views" "test" {}
`
}
