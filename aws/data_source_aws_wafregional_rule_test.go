package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsWafRegionalRule_Basic(t *testing.T) {
	name := "tf-acc-test"
	resourceName := "aws_wafregional_rule.wafrule"
	datasourceName := "data.aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafRegionalRuleConfig_NonExistent,
				ExpectError: regexp.MustCompile(`WAF Rule not found`),
			},
			{
				Config: testAccDataSourceAwsWafRegionalRuleConfig_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWafRegionalRuleConfig_Name(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "wafrule" {
  name        = "%s"
  metric_name = "WafruleTest"

  predicate {
    data_id = "${aws_wafregional_ipset.test.id}"
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_wafregional_ipset" "test" {
  name              = "%s"

  ip_set_descriptor {
    type  = "IPV4"
    value = "10.0.0.0/8"
  }
}

data "aws_wafregional_rule" "wafrule" {
  name = "${aws_wafregional_rule.wafrule.name}"
}
`, name, name)
}

const testAccDataSourceAwsWafRegionalRuleConfig_NonExistent = `
data "aws_wafregional_rule" "wafrule" {
  name = "tf-acc-test-does-not-exist"
}
`
