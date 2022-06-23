package route53resolver_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRoute53ResolverFirewallRulesDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_route53_resolver_firewall_rules.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRulesDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					// resource.TestCheckResourceAttr(dataSourceName, "provisioning_artifact_details.0.guidance", servicecatalog.ProvisioningArtifactGuidanceDefault),
					// resource.TestCheckResourceAttr(dataSourceName, "provisioning_artifact_details.0.name", rName),
					// resource.TestCheckResourceAttrPair(dataSourceName, "product_id", "aws_servicecatalog_product.test", "id"),
					// resource.TestCheckResourceAttr(dataSourceName, "provisioning_artifact_details.0.type", servicecatalog.ProductTypeCloudFormationTemplate),
				),
			},
		},
	})
}

func testAccFirewallRulesDataSourceBaseConfig() string {
	return `
resource "aws_route53_resolver_firewall_domain_list" "test01" {
  name    = "test01"
  domains = ["test01.com."]
}

resource "aws_route53_resolver_firewall_domain_list" "test02" {
  name    = "test02"
  domains = ["test02.com."]
}

resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = "test"
}

resource "aws_route53_resolver_firewall_rule" "test01" {
  name                    = "test01"
  action                  = "ALLOW"
  firewall_domain_list_id = aws_route53_resolver_firewall_domain_list.test01.id
  firewall_rule_group_id  = aws_route53_resolver_firewall_rule_group.test.id
  priority                = 100
}

resource "aws_route53_resolver_firewall_rule" "test02" {
  name                    = "test02"
  action                  = "BLOCK"
  block_override_dns_type = "CNAME"
  block_override_domain   = "test02.com."
  block_override_ttl      = 1
  block_response          = "OVERRIDE"
  firewall_domain_list_id = aws_route53_resolver_firewall_domain_list.test02.id
  firewall_rule_group_id  = aws_route53_resolver_firewall_rule_group.test.id
  priority                = 101
}
`
}

func testAccFirewallRulesDataSourceConfig_basic() string {
	return acctest.ConfigCompose(testAccFirewallRulesDataSourceBaseConfig(), `
data "aws_route53_resolver_firewall_rules" "test" {
  id = aws_route53_resolver_firewall_rule_group.test.id
}
`)
}
