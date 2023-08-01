// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRoute53ResolverFirewallDomainListDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_route53_resolver_firewall_domain_list.test"
	resourceName := "aws_route53_resolver_firewall_domain_list.test"
	domainName := acctest.RandomFQDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallDomainListDataSourceConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creator_request_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "firewall_domain_list_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domain_count", resourceName, "domains.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "modification_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
					resource.TestCheckResourceAttrSet(dataSourceName, "status_message"),
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
