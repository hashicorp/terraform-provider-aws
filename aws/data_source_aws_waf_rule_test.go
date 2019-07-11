package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsWafRule_Basic(t *testing.T) {
	name := "tf-acc-test"
	resourceName := "aws_waf_rule.wafrule"
	datasourceName := "data.aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafRuleConfig_NonExistent,
				ExpectError: regexp.MustCompile(`WAF Rules not found`),
			},
			{
				Config: testAccDataSourceAwsWafRuleConfig_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWafRuleConfig_Name(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule" "wafrule" {
  name        = "%s"
  metric_name = "WafruleTest"

  predicates {
    data_id = "${aws_waf_ipset.test.id}"
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_waf_ipset" "test" {
  name              = "%s"

  ip_set_descriptors {
    type  = "IPV4"
    value = "10.0.0.0/8"
  }
}

data "aws_waf_rule" "wafrule" {
  name = "${aws_waf_rule.wafrule.name}"
}
`, name, name)
}

const testAccDataSourceAwsWafRuleConfig_NonExistent = `
data "aws_waf_rule" "wafrule" {
  name = "tf-acc-test-does-not-exist"
}
`
