// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
)

func TestAccMetaRegionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_regions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionsDataSourceConfig_empty(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "names.#", 0),
				),
			},
		},
	})
}

func TestAccMetaRegionsDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_regions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionsDataSourceConfig_optInStatusFilter("opt-in-not-required"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "names.#", 0),
				),
			},
		},
	})
}

func TestAccMetaRegionsDataSource_allRegions(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_regions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionsDataSourceConfig_allRegions(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "names.#", 0),
				),
			},
		},
	})
}

func TestAccMetaRegionsDataSource_nonExistentRegion(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_regions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionsDataSourceConfig_nonExistentRegion(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccRegionsDataSourceConfig_empty() string {
	return `
data "aws_regions" "test" {}
`
}

func testAccRegionsDataSourceConfig_allRegions() string {
	return `
data "aws_regions" "test" {
  all_regions = "true"
}
`
}

func testAccRegionsDataSourceConfig_optInStatusFilter(filter string) string {
	return fmt.Sprintf(`
data "aws_regions" "test" {
  filter {
    name   = "opt-in-status"
    values = [%[1]q]
  }
}
`, filter)
}

func testAccRegionsDataSourceConfig_nonExistentRegion() string {
	return `
data "aws_regions" "test" {
  filter {
    name   = "region-name"
    values = ["us-east-4"]
  }
}
`
}
