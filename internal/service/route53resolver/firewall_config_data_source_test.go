package route53resolver_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRoute53ResolverFirewallConfigDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_route53_resolver_firewall_config.test"
	resourceName := "aws_route53_resolver_firewall_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
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
	return acctest.ConfigCompose(testAccFirewallConfigConfig_basic(rName), `
data "aws_route53_resolver_firewall_config" "test" {
  resource_id = aws_vpc.test.id
}
`)
}
