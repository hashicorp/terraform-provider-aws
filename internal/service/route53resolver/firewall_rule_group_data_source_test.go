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

func TestAccRoute53ResolverFirewallRuleGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_route53_resolver_firewall_rule_group.test"
	resourceName := "aws_route53_resolver_firewall_rule_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(dataSourceName, "creator_request_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "firewall_rule_group_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(dataSourceName, "modification_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(dataSourceName, "rule_count", acctest.Ct0),
					resource.TestCheckResourceAttrPair(dataSourceName, "share_status", resourceName, "share_status"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatusMessage),
				),
			},
		},
	})
}

func testAccFirewallRuleGroupDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFirewallRuleGroupConfig_basic(rName), `
data "aws_route53_resolver_firewall_rule_group" "test" {
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
}
`)
}
