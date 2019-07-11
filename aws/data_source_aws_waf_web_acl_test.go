package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsWafWebAcl_Basic(t *testing.T) {
	name := "tf-acc-test"
	resourceName := "aws_waf_web_acl.web_acl"
	datasourceName := "data.aws_waf_web_acl.web_acl"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafWebAclConfig_NonExistent,
				ExpectError: regexp.MustCompile(`web ACLs not found`),
			},
			{
				Config: testAccDataSourceAwsWafWebAclConfig_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWafWebAclConfig_Name(name string) string {
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

resource "aws_waf_web_acl" "web_acl" {
  name        = "%s"
  metric_name = "tfWebACL"

  default_action {
    type = "ALLOW"
  }

  rules {
    action {
      type = "BLOCK"
    }

    priority = 1
    rule_id  = "${aws_waf_rule.wafrule.id}"
    type     = "REGULAR"
  }
}

data "aws_waf_web_acl" "web_acl" {
  name = "${aws_waf_web_acl.web_acl.name}"
}

`, name, name, name)
}

const testAccDataSourceAwsWafWebAclConfig_NonExistent = `
data "aws_waf_web_acl" "web_acl" {
  name = "tf-acc-test-does-not-exist"
}
`
