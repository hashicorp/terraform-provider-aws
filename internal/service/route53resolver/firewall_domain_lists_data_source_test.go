// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ResolverFirewallDomainListsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dsResourceName := "data.aws_route53_resolver_firewall_domain_lists.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallDomainListsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsResourceName, "firewall_domain_lists.#"),
					acctest.CheckResourceAttrGreaterThanValue(dsResourceName, "firewall_domain_lists.#", 0),
					// Check that our created resource appears in the list
					resource.TestCheckTypeSetElemNestedAttrs(dsResourceName, "firewall_domain_lists.*", map[string]string{
						"name": rName,
					}),
					// Check that AWS managed malware domain list appears in the list
					resource.TestCheckTypeSetElemNestedAttrs(dsResourceName, "firewall_domain_lists.*", map[string]string{
						"name":               "AWSManagedDomainsMalwareDomainList",
						"managed_owner_name": "Route 53 Resolver DNS Firewall",
					}),
				),
			},
		},
	})
}

func testAccFirewallDomainListsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_domain_list" "test" {
  name = %[1]q
}

data "aws_route53_resolver_firewall_domain_lists" "test" {
  depends_on = [aws_route53_resolver_firewall_domain_list.test]
}
`, rName)
}
