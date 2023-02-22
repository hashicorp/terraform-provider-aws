package route53resolver_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRoute53ResolverRuleGroupAssociationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_route53_resolver_firewall_rule_group_association.test"
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupAssociationDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creator_request_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "firewall_rule_group_id", resourceName, "firewall_rule_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "firewall_rule_group_association_id", resourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "modification_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "mutation_protection", resourceName, "mutation_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "priority", resourceName, "priority"),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
					resource.TestCheckResourceAttrSet(dataSourceName, "status_message"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func testAccRuleGroupAssociationDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFirewallRuleGroupAssociationConfig_basic(rName), `
data "aws_route53_resolver_firewall_rule_group_association" "test" {
  firewall_rule_group_association_id = aws_route53_resolver_firewall_rule_group_association.test.id
}
`)
}
