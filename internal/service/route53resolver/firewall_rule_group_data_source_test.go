package route53resolver_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRoute53ResolverFirewallRuleGroupDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_route53_resolver_firewall_rule_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleGroupDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "route53resolver", regexp.MustCompile(`firewall-rule-group/.+`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creator_request_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "modification_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "owner_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "rule_count"),
					resource.TestMatchResourceAttr(dataSourceName, "status", regexp.MustCompile(`COMPLETE|DELETING|UPDATING`)),
					resource.TestMatchResourceAttr(dataSourceName, "share_status", regexp.MustCompile(`NOT_SHARED|SHARED_WITH_ME|SHARED_BY_ME`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "status_message"),
				),
			},
		},
	})
}

func testAccFirewallRuleGroupDataSourceConfig_basic() string {
	return `
resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = "test"
}

data "aws_route53_resolver_firewall_rule_group" "test" {
  id = aws_route53_resolver_firewall_rule_group.test.id
}

`
}
