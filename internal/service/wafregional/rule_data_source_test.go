// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFRegionalRuleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafregional_rule.wafrule"
	datasourceName := "data.aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRuleDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`no matching WAF Regional Rule found`),
			},
			{
				Config: testAccRuleDataSourceConfig_name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccRuleDataSourceConfig_name(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = "WafruleTest"
}

data "aws_wafregional_rule" "wafrule" {
  name = aws_wafregional_rule.wafrule.name
}
`, name)
}

const testAccRuleDataSourceConfig_nonExistent = `
data "aws_wafregional_rule" "wafrule" {
  name = "tf-acc-test-does-not-exist"
}
`
