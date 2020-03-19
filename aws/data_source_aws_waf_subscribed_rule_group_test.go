package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsWafSubscribedRuleGroup_Basic(t *testing.T) {
	datasourceName := "data.aws_waf_subscribed_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafSubscribedRuleGroupNonExistent,
				ExpectError: regexp.MustCompile(`WAF Subscribed Rule Group not found`),
			},
			{
				Config: testAccDataSourceAwsWafSubscribedRuleGroupConfigName,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", "Fortinet Managed Rules for AWS WAF - SQLi/XSS"),
					resource.TestCheckResourceAttr(datasourceName, "metric_name", "FortinetAWSXSSAndSQLiRuleset"),
				),
			},
		},
	})
}

const testAccDataSourceAwsWafSubscribedRuleGroupConfigName = `
data "aws_waf_subscribed_rule_group" "test" {
  name = "Fortinet Managed Rules for AWS WAF - SQLi/XSS"
}
`

const testAccDataSourceAwsWafSubscribedRuleGroupNonExistent = `
data "aws_waf_subscribed_rule_group" "test" {
  name = "tf-acc-test-does-not-exist"
}
`
