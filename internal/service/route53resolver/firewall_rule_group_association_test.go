// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ResolverFirewallRuleGroupAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v route53resolver.FirewallRuleGroupAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_rule_group_id", "aws_route53_resolver_firewall_rule_group.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "mutation_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "101"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRoute53ResolverFirewallRuleGroupAssociation_name(t *testing.T) {
	ctx := acctest.Context(t)
	var v route53resolver.FirewallRuleGroupAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNewName := sdkacctest.RandomWithPrefix("tf-acc-test2")
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallRuleGroupAssociationConfig_basic(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNewName),
				),
			},
		},
	})
}

func TestAccRoute53ResolverFirewallRuleGroupAssociation_mutationProtection(t *testing.T) {
	ctx := acctest.Context(t)
	var v route53resolver.FirewallRuleGroupAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleGroupAssociationConfig_mutationProtection(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mutation_protection", "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallRuleGroupAssociationConfig_mutationProtection(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "mutation_protection", "DISABLED"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverFirewallRuleGroupAssociation_priority(t *testing.T) {
	ctx := acctest.Context(t)
	var v route53resolver.FirewallRuleGroupAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleGroupAssociationConfig_priority(rName, 101),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "101"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallRuleGroupAssociationConfig_priority(rName, 200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "200"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverFirewallRuleGroupAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v route53resolver.FirewallRuleGroupAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53resolver.ResourceFirewallRuleGroupAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53ResolverFirewallRuleGroupAssociation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v route53resolver.FirewallRuleGroupAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallRuleGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallRuleGroupAssociationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallRuleGroupAssociationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccFirewallRuleGroupAssociationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallRuleGroupAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckFirewallRuleGroupAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_resolver_firewall_rule_group_association" {
				continue
			}

			_, err := tfroute53resolver.FindFirewallRuleGroupAssociationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53 Resolver Firewall Rule Group Association still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFirewallRuleGroupAssociationExists(ctx context.Context, n string, v *route53resolver.FirewallRuleGroupAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route53 Resolver Firewall Rule Group Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn(ctx)

		output, err := tfroute53resolver.FindFirewallRuleGroupAssociationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccFirewallRuleGroupAssociationConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 0), fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
}
`, rName))
}

func testAccFirewallRuleGroupAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFirewallRuleGroupAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[1]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  mutation_protection    = "DISABLED"
  priority               = 101
  vpc_id                 = aws_vpc.test.id
}
`, rName))
}

func testAccFirewallRuleGroupAssociationConfig_mutationProtection(rName, mutationProtection string) string {
	return acctest.ConfigCompose(testAccFirewallRuleGroupAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[1]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  mutation_protection    = %[2]q
  priority               = 101
  vpc_id                 = aws_vpc.test.id
}
`, rName, mutationProtection))
}

func testAccFirewallRuleGroupAssociationConfig_priority(rName string, priority int) string {
	return acctest.ConfigCompose(testAccFirewallRuleGroupAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[1]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  priority               = %[2]d
  vpc_id                 = aws_vpc.test.id
}
`, rName, priority))
}

func testAccFirewallRuleGroupAssociationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccFirewallRuleGroupAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[1]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  priority               = 101
  vpc_id                 = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccFirewallRuleGroupAssociationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccFirewallRuleGroupAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group_association" "test" {
  name                   = %[1]q
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.test.id
  priority               = 101
  vpc_id                 = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
