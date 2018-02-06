package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsLbSSLPolicy_basic(t *testing.T) {
	policyName := "ELBSecurityPolicy-TLS-1-2-2017-01"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsLbSSLPolicyConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_lb_ssl_policy.basic", "name"),
				),
			},
			{
				Config: testAccDataSourceAwsLbSSLPolicyConfig_name(policyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_lb_ssl_policy.tls", "name", policyName),
				),
			},
		},
	})
}

const testAccDataSourceAwsLbSSLPolicyConfig_basic = `data "aws_lb_ssl_policy" "basic" {}`

func testAccDataSourceAwsLbSSLPolicyConfig_name(pname string) string {
	return fmt.Sprintf(`
data "aws_lb_ssl_policy" "tls" {
  name = "%s"
}
`, pname)
}
