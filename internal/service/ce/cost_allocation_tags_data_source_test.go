// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCECostAllocationTagsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ce_cost_allocation_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCostAllocationTagsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "tags.#"),
				),
			},
		},
	})
}

func testAccCostAllocationTagsDataSourceConfig_basic() string {
	return `
data "aws_ce_cost_allocation_tags" "test" {}
`
}
