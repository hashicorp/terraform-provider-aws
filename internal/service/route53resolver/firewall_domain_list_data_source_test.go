package route53resolver_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRoute53ResolverFirewallDomainListDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_route53_resolver_firewall_domain_list.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallDomainListDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "route53resolver", regexp.MustCompile(`firewall-domain-list/.+`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creator_request_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "domain_count"),
					resource.TestCheckResourceAttrSet(dataSourceName, "modification_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "name"),
					resource.TestMatchResourceAttr(dataSourceName, "status", regexp.MustCompile(`COMPLETE|COMPLETE_IMPORT_FAILED|IMPORTING|DELETING`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "status_message"),
				),
			},
		},
	})
}

func testAccFirewallDomainListDataSourceConfig_basic() string {
	return `
resource "aws_route53_resolver_firewall_domain_list" "test" {
  name    = "example"
  domains = ["example.com.", "test.com."]
}

data "aws_route53_resolver_firewall_domain_list" "test" {
  firewall_domain_list_id = aws_route53_resolver_firewall_domain_list.test.id
}

`
}
