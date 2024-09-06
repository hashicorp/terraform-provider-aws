// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSyntheticsRuntimeVersionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_synthetics_runtime_versions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuntimeVersionsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "version_names.#", 0),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "version_names.*", "syn-1.0"),
				),
			},
		},
	})
}

func TestAccSyntheticsRuntimeVersionsDataSource_skipDeprecated(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_synthetics_runtime_versions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuntimeVersionsDataSourceConfig_skipDeprecated(false),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "version_names.#", 0),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "version_names.*", "syn-1.0"),
				),
			},
			{
				Config: testAccRuntimeVersionsDataSourceConfig_skipDeprecated(true),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "version_names.#", 0),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "version_names.*", "syn-1.0"),
				),
				ExpectError: regexache.MustCompile(`no TypeSet element "version_names.*", with value "syn-1.0" in state`),
			},
		},
	})
}

func testAccRuntimeVersionsDataSourceConfig_basic() string {
	return `
data "aws_synthetics_runtime_versions" "test" {}
`
}

func testAccRuntimeVersionsDataSourceConfig_skipDeprecated(skipDeprecated bool) string {
	return fmt.Sprintf(`
data "aws_synthetics_runtime_versions" "test" {
  skip_deprecated = %[1]t
}
`, skipDeprecated)
}
