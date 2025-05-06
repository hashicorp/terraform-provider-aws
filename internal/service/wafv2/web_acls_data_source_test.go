// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFV2WebACLsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_wafv2_web_acls.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "names.#", 1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "names.*", "aws_wafv2_web_acl.test1", names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "names.*", "aws_wafv2_web_acl.test2", names.AttrName),
				),
			},
		},
	})
}

func testAccWebACLsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test1" {
  name  = "%[1]s-test1"
  scope = "REGIONAL"

  default_action {
    block {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-rule-metric-name"
    sampled_requests_enabled   = false
  }
}
resource "aws_wafv2_web_acl" "test2" {
  name  = "%[1]s-test2"
  scope = "REGIONAL"

  default_action {
    block {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-rule-metric-name"
    sampled_requests_enabled   = false
  }
}

data "aws_wafv2_web_acls" "test" {
  scope      = "REGIONAL"
  depends_on = [aws_wafv2_web_acl.test1, aws_wafv2_web_acl.test2]
}
`, rName)
}
