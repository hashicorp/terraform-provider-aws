package wafregional_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccWafRegionalSubscribedRuleGroupDataSource_Basic(t *testing.T) {
	if os.Getenv("AWS_WAF_SUBSCRIBED_RULE_GROUP_NAME") == "" {
		t.Skip("Environment variable AWS_WAF_SUBSCRIBED_RULE_GROUP_NAME is not set")
	}

	ruleGroupName := os.Getenv("AWS_WAF_SUBSCRIBED_RULE_GROUP_NAME")

	if os.Getenv("AWS_WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME") == "" {
		t.Skip("Environment variable AWS_WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME is not set")
	}

	metricName := os.Getenv("AWS_WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME")

	datasourceName := "data.aws_wafregional_subscribed_rule_group.rulegroup"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, wafregional.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config:      testAccSubscribedRuleGroupDataSourceConfig_Nonexistent,
				ExpectError: regexp.MustCompile(`WAF Subscribed Rule Group not found`),
			},
			{
				Config: testAccSubscribedRuleGroupDataSourceConfig_Name(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, "metric_name", metricName),
				),
			},
			{
				Config: testAccSubscribedRuleGroupDataSourceConfig_MetricName(metricName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, "metric_name", metricName),
				),
			},
			{
				Config: testAccSubscribedRuleGroupDataSourceConfig_NameAndMetricName(ruleGroupName, metricName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, "metric_name", metricName),
				),
			},
			{
				Config:      testAccDataSourceSubscribedRuleGroupDataSourceConfig_NameAndMismatchingMetricName(ruleGroupName),
				ExpectError: regexp.MustCompile(`WAF Subscribed Rule Group not found`),
			},
		},
	})
}

func testAccSubscribedRuleGroupDataSourceConfig_Name(name string) string {
	return fmt.Sprintf(`
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  name = %[1]q
}
`, name)
}

func testAccSubscribedRuleGroupDataSourceConfig_MetricName(metricName string) string {
	return fmt.Sprintf(`
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  metric_name = %[1]q
}
`, metricName)
}

func testAccSubscribedRuleGroupDataSourceConfig_NameAndMetricName(name string, metricName string) string {
	return fmt.Sprintf(`
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  name = %[1]q
  metric_name = %[2]q
}
`, name, metricName)
}

func testAccDataSourceSubscribedRuleGroupDataSourceConfig_NameAndMismatchingMetricName(name string) string {
	return fmt.Sprintf(`
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  name = %[1]q
  metric_name = "tf-acc-test-does-not-exist"
}
`, name)
}

const testAccSubscribedRuleGroupDataSourceConfig_Nonexistent = `
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  name = "tf-acc-test-does-not-exist"
}
`
