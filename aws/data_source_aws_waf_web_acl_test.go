package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsWafRegionalWebAcl_Basic(t *testing.T) {
	name := "tf-acc-test"
	resourceName := "aws_wafregional_web_acl.web_acl"
	datasourceName := "data.aws_wafregional_web_acl.web_acl"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafRegionalWebAclConfig_NonExistent,
				ExpectError: regexp.MustCompile(`web ACLs not found`),
			},
			{
				Config: testAccDataSourceAwsWafRegionalWebAclConfig_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWafRegionalWebAclConfig_Name(name string) string {
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
resource "aws_wafregional_web_acl" "web_acl" {
  name        = "%s"
  metric_name = "tfWebACL"
  default_action {
    type = "ALLOW"
  }
  rule {
    action {
      type = "BLOCK"
    }
    priority = 1
    rule_id  = "${aws_wafregional_rule.wafrule.id}"
    type     = "REGULAR"
  }
}
data "aws_wafregional_web_acl" "web_acl" {
  name = "${aws_wafregional_web_acl.web_acl.name}"
}
`, name, name, name)
}

const testAccDataSourceAwsWafRegionalWebAclConfig_NonExistent = `
data "aws_wafregional_web_acl" "web_acl" {
  name = "tf-acc-test-does-not-exist"
}
`
