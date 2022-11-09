package route53resolver_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRoute53ResolverRuleGroupAssociationDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_route53_resolver_firewall_rule_group_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupAssociationDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "route53resolver", regexp.MustCompile(`firewall-rule-group-association/.+`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creator_request_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "firewall_rule_group_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "modification_time"),
					resource.TestMatchResourceAttr(dataSourceName, "mutation_protection", regexp.MustCompile(`ENABLED|DISABLED`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "priority"),
					resource.TestMatchResourceAttr(dataSourceName, "status", regexp.MustCompile(`COMPLETE|DELETING|UPDATING`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "status_message"),
					resource.TestCheckResourceAttrSet(dataSourceName, "vpc_id"),
				),
			},
		},
	})
}

func testAccRuleGroupAssociationDataSourceConfig_basic() string {
	return `
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = "test"
}

resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = "test"
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  priority               = 200
  vpc_id                 = aws_vpc.test.id
}

data "aws_route53_resolver_firewall_rule_group_association" "test" {
  id = aws_route53_resolver_firewall_rule_group_association.test.id
}

`
}
