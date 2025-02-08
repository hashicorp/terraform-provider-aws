// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRoute53ResolverRuleAssociationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_route53_resolver_rule_associations.test"
	domainName := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAssociationsDataSourceConfig_vpcFilter(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "associations.#", "2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "associations.1.vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "associations.1.resolver_rule_id", "aws_route53_resolver_rule.test", "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "associations.1.name", "aws_route53_resolver_rule.test", "name"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverRuleAssociationsDataSource_multipleFilters(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_route53_resolver_rule_associations.test"
	domainName := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAssociationsDataSourceConfig_vpcAndRuleFilter(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "associations.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "associations.0.vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "associations.0.resolver_rule_id", "aws_route53_resolver_rule.test", "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "associations.0.name", "aws_route53_resolver_rule.test", "name"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverRuleAssociationsDataSource_emptyList(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_route53_resolver_rule_associations.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAssociationsDataSourceConfig_noResults(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(dataSourceName, "associations#"),
				),
			},
		},
	})
}

func testAccRuleAssociationsDataSourceConfig_vpcFilter(rName, domainName string) string {
	return acctest.ConfigCompose(testAccRuleAssociationConfig_basic(rName, domainName), `
data "aws_route53_resolver_rule_associations" "test" {
  filter {
    name   = "VPCId"
    values = [aws_route53_resolver_rule_association.test.vpc_id]
  }
}
`)
}

func testAccRuleAssociationsDataSourceConfig_vpcAndRuleFilter(rName, domainName string) string {
	return acctest.ConfigCompose(testAccRuleAssociationConfig_basic(rName, domainName), `
data "aws_route53_resolver_rule_associations" "test" {
  filter {
    name   = "ResolverRuleId"
    values = [aws_route53_resolver_rule_association.test.resolver_rule_id]
  }

  filter {
    name   = "VPCId"
    values = [aws_route53_resolver_rule_association.test.vpc_id]
  }

  depends_on = [aws_route53_resolver_rule_association.test]
}
`)
}

func testAccRuleAssociationsDataSourceConfig_noResults() string {
	return `
data "aws_route53_resolver_rule_associations" "test" {
  filter {
    name   = "VPCId"
    values = ["vpc-00000000"]
  }
}
`
}
