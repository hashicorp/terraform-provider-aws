package route53resolver_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRoute53ResolverFirewallRulesDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_route53_resolver_firewall_rules.test"

	propagationSleep := func() resource.TestCheckFunc {
		return func(s *terraform.State) error {
			log.Print("[DEBUG] Test: Sleep to allow firewall rule to be visible in the list.")
			time.Sleep(5 * time.Second)
			return nil
		}
	}

	fqdn := acctest.RandomFQDomainName()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRulesDataSourceBaseConfig(fqdn, rName),
				Check:  propagationSleep(),
			},
			{
				Config: testAccFirewallRulesDataSourceConfig_basic(fqdn, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "firewall_rules.0.action"),
					resource.TestCheckResourceAttrSet(dataSourceName, "firewall_rules.0.block_override_ttl"),
					resource.TestCheckResourceAttrSet(dataSourceName, "firewall_rules.0.creation_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "firewall_rules.0.creator_request_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "firewall_rules.0.firewall_domain_list_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "firewall_rules.0.firewall_rule_group_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "firewall_rules.0.modification_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "firewall_rules.0.priority"),
					resource.TestCheckResourceAttr(dataSourceName, "firewall_rules.0.name", rName),
				),
			},
		},
	})
}

func testAccFirewallRulesDataSourceBaseConfig(fqdn, rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_domain_list" "test" {
  name    = %[2]q
  domains = [%[1]q]
}

resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[2]q
}

resource "aws_route53_resolver_firewall_rule" "test" {
  name                    = %[2]q
  action                  = "ALLOW"
  firewall_domain_list_id = aws_route53_resolver_firewall_domain_list.test.id
  firewall_rule_group_id  = aws_route53_resolver_firewall_rule_group.test.id
  priority                = 100
}
`, fqdn, rName)
}

func testAccFirewallRulesDataSourceConfig_basic(fqdn, rName string) string {
	return acctest.ConfigCompose(testAccFirewallRulesDataSourceBaseConfig(fqdn, rName), `
data "aws_route53_resolver_firewall_rules" "test" {
  id = aws_route53_resolver_firewall_rule_group.test.id
}
`)
}
