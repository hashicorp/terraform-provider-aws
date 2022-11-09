package route53resolver_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRoute53ResolverFirewallConfigDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_route53_resolver_firewall_config.test"
	resourceName := "aws_route53_resolver_firewall_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfigDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "firewall_fail_open", resourceName, "firewall_fail_open"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "resource_id", resourceName, "resource_id"),
				),
			},
		},
	})
}

func testAccFirewallConfigDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_resolver_firewall_config" "test" {
  resource_id        = aws_vpc.test.id
  firewall_fail_open = "ENABLED"
}

data "aws_route53_resolver_firewall_config" "test" {
  resource_id = aws_vpc.test.id
}
`, rName)
}
