// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package waf_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFIPSetDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.ipset"
	datasourceName := "data.aws_waf_ipset.ipset"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccIPSetDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`no matching WAF IPSet found`),
			},
			{
				Config: testAccIPSetDataSourceConfig_name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccIPSetDataSourceConfig_name(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = %[1]q
}

data "aws_waf_ipset" "ipset" {
  name = aws_waf_ipset.ipset.name
}
`, name)
}

const testAccIPSetDataSourceConfig_nonExistent = `
data "aws_waf_ipset" "ipset" {
  name = "tf-acc-test-does-not-exist"
}
`
