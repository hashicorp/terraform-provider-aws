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

func TestAccWAFRegionalSubscribedRuleGroupDataSource_basic(t *testing.T) {
	if os.Getenv("WAF_SUBSCRIBED_RULE_GROUP_NAME") == "" {
		t.Skip("Environment variable WAF_SUBSCRIBED_RULE_GROUP_NAME is not set")
	}

	ruleGroupName := os.Getenv("WAF_SUBSCRIBED_RULE_GROUP_NAME")

	if os.Getenv("WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME") == "" {
		t.Skip("Environment variable WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME is not set")
	}

	metricName := os.Getenv("WAF_SUBSCRIBED_RULE_GROUP_METRIC_NAME")

	datasourceName := "data.aws_wafregional_subscribed_rule_group.rulegroup"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, wafregional.EndpointsID),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccSubscribedRuleGroupDataSourceConfig_nonexistent,
				ExpectError: regexp.MustCompile(`no matches found`),
			},
			{
				Config: testAccSubscribedRuleGroupDataSourceConfig_name(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, "metric_name", metricName),
				),
			},
			{
				Config: testAccSubscribedRuleGroupDataSourceConfig_metricName(metricName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, "metric_name", metricName),
				),
			},
			{
				Config: testAccSubscribedRuleGroupDataSourceConfig_nameAndMetricName(ruleGroupName, metricName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(datasourceName, "metric_name", metricName),
				),
			},
			{
				Config:      testAccDataSourceSubscribedRuleGroupDataSourceConfig_nameAndMismatchingMetricName(ruleGroupName),
				ExpectError: regexp.MustCompile(`no matches found`),
			},
		},
	})
}

func testAccSubscribedRuleGroupDataSourceConfig_name(name string) string {
	return fmt.Sprintf(`
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  name = %[1]q
}
`, name)
}

func testAccSubscribedRuleGroupDataSourceConfig_metricName(metricName string) string {
	return fmt.Sprintf(`
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  metric_name = %[1]q
}
`, metricName)
}

func testAccSubscribedRuleGroupDataSourceConfig_nameAndMetricName(name string, metricName string) string {
	return fmt.Sprintf(`
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  name        = %[1]q
  metric_name = %[2]q
}
`, name, metricName)
}

func testAccDataSourceSubscribedRuleGroupDataSourceConfig_nameAndMismatchingMetricName(name string) string {
	return fmt.Sprintf(`
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  name        = %[1]q
  metric_name = "tf-acc-test-does-not-exist"
}
`, name)
}

const testAccSubscribedRuleGroupDataSourceConfig_nonexistent = `
data "aws_wafregional_subscribed_rule_group" "rulegroup" {
  name = "tf-acc-test-does-not-exist"
}
`
