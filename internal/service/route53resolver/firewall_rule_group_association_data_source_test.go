// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ResolverRuleGroupAssociationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_route53_resolver_firewall_rule_group_association.test"
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupAssociationDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(dataSourceName, "creator_request_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "firewall_rule_group_id", resourceName, "firewall_rule_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "firewall_rule_group_association_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(dataSourceName, "modification_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "mutation_protection", resourceName, "mutation_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatusMessage),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
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
