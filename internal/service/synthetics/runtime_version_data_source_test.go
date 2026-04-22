// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package synthetics_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSyntheticsRuntimeVersionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_synthetics_runtime_version.test"
	prefix := "syn-nodejs-puppeteer"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuntimeVersionDataSourceConfig_basic(prefix),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrHasPrefix(dataSourceName, "version_name", prefix),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, dataSourceName, "version_name"),
					resource.TestCheckNoResourceAttr(dataSourceName, "deprecation_date"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrSet(dataSourceName, "release_date"),
				),
			},
		},
	})
}

func TestAccSyntheticsRuntimeVersionDataSource_deprecatedVersion(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_synthetics_runtime_version.test"
	prefix := "syn-nodejs-puppeteer"
	version := "3.0"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SyntheticsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuntimeVersionDataSourceConfig_version(prefix, version),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrHasPrefix(dataSourceName, "version_name", fmt.Sprintf("%s-%s", prefix, version)),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, dataSourceName, "version_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "deprecation_date"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrSet(dataSourceName, "release_date"),
				),
			},
		},
	})
}

func testAccRuntimeVersionDataSourceConfig_basic(prefix string) string {
	return fmt.Sprintf(`
data "aws_synthetics_runtime_version" "test" {
  prefix = %[1]q
  latest = true
}
`, prefix)
}

func testAccRuntimeVersionDataSourceConfig_version(prefix, version string) string {
	return fmt.Sprintf(`
data "aws_synthetics_runtime_version" "test" {
  prefix  = %[1]q
  version = %[2]q
}
`, prefix, version)
}
