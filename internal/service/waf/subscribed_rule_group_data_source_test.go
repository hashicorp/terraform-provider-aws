// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFSubscribedRuleGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("WAF_SUBSCRIBED_RULE_GROUP_NAME") == "" {
		t.Skip("Environment variable WAF_SUBSCRIBED_RULE_GROUP_NAME is not set")
	}

	ruleGroupName := os.Getenv("WAF_SUBSCRIBED_RULE_GROUP_NAME")

	if os.Getenv("WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME") == "" {
		t.Skip("Environment variable WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME is not set")
	}

	metricName := os.Getenv("WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME")

	datasourceName := "data.aws_waf_subscribed_rule_group.rulegroup"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, waf.ServiceID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		CheckDestroy:             nil,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSubscribedRuleGroupDataSourceConfig_nonexistent,
				ExpectError: regexache.MustCompile(`no matches found`),
			},
			{
				Config: testAccSubscribedRuleGroupDataSourceConfig_name(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, names.AttrMetricName, metricName),
				),
			},
			{
				Config: testAccSubscribedRuleGroupDataSourceConfig_metricName(metricName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, names.AttrMetricName, metricName),
				),
			},
			{
				Config: testAccSubscribedRuleGroupDataSourceConfig_nameAndMetricName(ruleGroupName, metricName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, names.AttrMetricName, metricName),
				),
			},
			{
				Config:      testAccSubscribedRuleGroupDataSourceConfig_nameAndMismatchingMetricName(ruleGroupName),
				ExpectError: regexache.MustCompile(`no matches found`),
			},
		},
	})
}

func testAccSubscribedRuleGroupDataSourceConfig_name(name string) string {
	return fmt.Sprintf(`
data "aws_waf_subscribed_rule_group" "rulegroup" {
  name = %[1]q
}
`, name)
}

func testAccSubscribedRuleGroupDataSourceConfig_metricName(metricName string) string {
	return fmt.Sprintf(`
data "aws_waf_subscribed_rule_group" "rulegroup" {
  metric_name = %[1]q
}
`, metricName)
}

func testAccSubscribedRuleGroupDataSourceConfig_nameAndMetricName(name string, metricName string) string {
	return fmt.Sprintf(`
data "aws_waf_subscribed_rule_group" "rulegroup" {
  name        = %[1]q
  metric_name = %[2]q
}
`, name, metricName)
}

func testAccSubscribedRuleGroupDataSourceConfig_nameAndMismatchingMetricName(name string) string {
	return fmt.Sprintf(`
data "aws_waf_subscribed_rule_group" "rulegroup" {
  name        = %[1]q
  metric_name = "tf-acc-test-does-not-exist"
}
`, name)
}

const testAccSubscribedRuleGroupDataSourceConfig_nonexistent = `
data "aws_waf_subscribed_rule_group" "rulegroup" {
  name = "tf-acc-test-does-not-exist"
}
`
