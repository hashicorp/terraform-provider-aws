// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"fmt"
	"strconv"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ResolverRulesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dsResourceName := "data.aws_route53_resolver_rules.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRulesDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dsResourceName, "resolver_rule_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(dsResourceName, "resolver_rule_ids.*", "rslvr-autodefined-rr-internet-resolver"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverRulesDataSource_resolverEndpointID(t *testing.T) {
	ctx := acctest.Context(t)
	domainName1 := acctest.RandomDomainName()
	domainName2 := acctest.RandomDomainName()
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ds1ResourceName := "data.aws_route53_resolver_rules.by_resolver_endpoint_id"
	ds2ResourceName := "data.aws_route53_resolver_rules.by_resolver_endpoint_id_rule_type_share_status"
	ds3ResourceName := "data.aws_route53_resolver_rules.by_invalid_owner_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRulesDataSourceConfig_resolverEndpointID(rName1, rName2, domainName1, domainName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ds1ResourceName, "resolver_rule_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(ds2ResourceName, "resolver_rule_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(ds3ResourceName, "resolver_rule_ids.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccRoute53ResolverRulesDataSource_nameRegex(t *testing.T) {
	ctx := acctest.Context(t)
	dsResourceName := "data.aws_route53_resolver_rules.test"
	rCount := 3
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRulesDataSourceConfig_nameRegex(rCount, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dsResourceName, "resolver_rule_ids.#", strconv.Itoa(rCount)),
				),
			},
		},
	})
}

func TestAccRoute53ResolverRulesDataSource_nonExistentNameRegex(t *testing.T) {
	ctx := acctest.Context(t)
	dsResourceName := "data.aws_route53_resolver_rules.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRulesDataSourceConfig_nonExistentNameRegex,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dsResourceName, "resolver_rule_ids.#", acctest.Ct0),
				),
			},
		},
	})
}

const testAccRulesDataSourceConfig_basic = `
# The default Internet Resolver rule.
data "aws_route53_resolver_rules" "test" {
  owner_id     = "Route 53 Resolver"
  rule_type    = "RECURSIVE"
  share_status = "NOT_SHARED"
}
`

func testAccRulesDataSourceConfig_resolverEndpointID(rName1, rName2, domainName1, domainName2 string) string {
	return acctest.ConfigCompose(testAccRuleConfig_resolverEndpointBase(rName1), fmt.Sprintf(`
resource "aws_route53_resolver_rule" "forward" {
  domain_name = %[3]q
  rule_type   = "FORWARD"
  name        = %[1]q

  resolver_endpoint_id = aws_route53_resolver_endpoint.test[1].id

  target_ip {
    ip = "192.0.2.7"
  }
}

resource "aws_route53_resolver_rule" "recursive" {
  domain_name = %[4]q
  rule_type   = "RECURSIVE"
  name        = %[2]q
}

data "aws_route53_resolver_rules" "by_resolver_endpoint_id" {
  owner_id             = aws_route53_resolver_rule.forward.owner_id
  resolver_endpoint_id = aws_route53_resolver_rule.forward.resolver_endpoint_id
}

data "aws_route53_resolver_rules" "by_resolver_endpoint_id_rule_type_share_status" {
  owner_id             = aws_route53_resolver_rule.recursive.owner_id
  resolver_endpoint_id = aws_route53_resolver_rule.recursive.resolver_endpoint_id
  rule_type            = aws_route53_resolver_rule.recursive.rule_type
  share_status         = aws_route53_resolver_rule.recursive.share_status
}

data "aws_route53_resolver_rules" "by_invalid_owner_id" {
  owner_id     = "000000000000"
  share_status = "SHARED_WITH_ME"
}
`, rName1, rName2, domainName1, domainName2))
}

func testAccRulesDataSourceConfig_nameRegex(rCount int, rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_rule" "test" {
  count       = %[1]d
  domain_name = "%[2]s.example.org"
  name        = "%[2]s-${count.index}-rule"
  rule_type   = "SYSTEM"
}

data "aws_route53_resolver_rules" "test" {
  name_regex = "%[2]s-.*-rule"

  depends_on = [aws_route53_resolver_rule.test]
}
`, rCount, rName)
}

const testAccRulesDataSourceConfig_nonExistentNameRegex = `
data "aws_route53_resolver_rules" "test" {
  name_regex = "dne-regex"
}
`
