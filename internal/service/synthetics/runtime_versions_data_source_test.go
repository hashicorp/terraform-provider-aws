// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package synthetics_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSyntheticsRuntimeVersionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_synthetics_runtime_versions.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuntimeVersionsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "runtime_versions.#", 0),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "runtime_versions.*", map[string]string{
						"version_name": "syn-1.0",
					}),
				),
			},
		},
	})
}

func testAccRuntimeVersionsDataSourceConfig_basic() string {
	return `
data "aws_synthetics_runtime_versions" "test" {}
`
}
