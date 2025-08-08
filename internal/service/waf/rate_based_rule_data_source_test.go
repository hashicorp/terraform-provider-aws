// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccWAFRateBasedRuleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_rate_based_rule.wafrule"
	datasourceName := "data.aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRateBasedRuleDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`no matching WAF Rate Based Rule found`),
			},
			{
				Config: testAccRateBasedRuleDataSourceConfig_name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccRateBasedRuleDataSourceConfig_name(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = "WafruleTest"
  rate_key    = "IP"
  rate_limit  = 2000
}

data "aws_waf_rate_based_rule" "wafrule" {
  name = aws_waf_rate_based_rule.wafrule.name
}
`, name)
}

const testAccRateBasedRuleDataSourceConfig_nonExistent = `
data "aws_waf_rate_based_rule" "wafrule" {
  name = "tf-acc-test-does-not-exist"
}
`
