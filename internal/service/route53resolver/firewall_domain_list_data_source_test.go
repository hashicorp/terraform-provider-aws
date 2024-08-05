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

func TestAccRoute53ResolverFirewallDomainListDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_route53_resolver_firewall_domain_list.test"
	resourceName := "aws_route53_resolver_firewall_domain_list.test"
	domainName := acctest.RandomFQDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallDomainListDataSourceConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(dataSourceName, "creator_request_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "firewall_domain_list_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "domain_count", resourceName, "domains.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "modification_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatusMessage),
				),
			},
		},
	})
}

func testAccFirewallDomainListDataSourceConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(testAccFirewallDomainListConfig_domains(rName, domain), `
data "aws_route53_resolver_firewall_domain_list" "test" {
  firewall_domain_list_id = aws_route53_resolver_firewall_domain_list.test.id
}
`)
}
