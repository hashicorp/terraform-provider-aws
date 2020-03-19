package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsWafRegionalSubscribedRuleGroup_Basic(t *testing.T) {
	datasourceName := "data.aws_wafregional_subscribed_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafRegionalSubscribedRuleGroupNonExistent,
				ExpectError: regexp.MustCompile(`WAF Regional Subscribed Rule Group not found`),
			},
			{
				Config: testAccDataSourceAwsWafRegionalSubscribedRuleGroupConfigName,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", "Fortinet Managed Rules for AWS WAF - SQLi/XSS"),
					resource.TestCheckResourceAttr(datasourceName, "metric_name", "FortinetAWSXSSAndSQLiRuleset"),
				),
			},
		},
	})
}

const testAccDataSourceAwsWafRegionalSubscribedRuleGroupConfigName = `
data "aws_wafregional_subscribed_rule_group" "test" {
  name = "Fortinet Managed Rules for AWS WAF - SQLi/XSS"
}
`

const testAccDataSourceAwsWafRegionalSubscribedRuleGroupNonExistent = `
data "aws_wafregional_subscribed_rule_group" "test" {
  name = "tf-acc-test-does-not-exist"
}
`
