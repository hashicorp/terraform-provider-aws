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
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ResolverRuleAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var assn route53resolver.ResolverRuleAssociation
	resourceName := "aws_route53_resolver_rule_association.test"
	vpcResourceName := "aws_vpc.test"
	ruleResourceName := "aws_route53_resolver_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAssociationConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleAssociationExists(ctx, resourceName, &assn),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_rule_id", ruleResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
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

func TestAccRoute53ResolverRuleAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var assn route53resolver.ResolverRuleAssociation
	resourceName := "aws_route53_resolver_rule_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAssociationConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleAssociationExists(ctx, resourceName, &assn),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53resolver.ResourceRuleAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53ResolverRuleAssociation_Disappears_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var assn route53resolver.ResolverRuleAssociation
	resourceName := "aws_route53_resolver_rule_association.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleAssociationConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleAssociationExists(ctx, resourceName, &assn),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPC(), vpcResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRuleAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_resolver_rule_association" {
				continue
			}

			_, err := tfroute53resolver.FindResolverRuleAssociationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53 Resolver Rule Association still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRuleAssociationExists(ctx context.Context, n string, v *route53resolver.ResolverRuleAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route53 Resolver Rule Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn(ctx)

		output, err := tfroute53resolver.FindResolverRuleAssociationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRuleAssociationConfig_basic(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_resolver_rule" "test" {
  domain_name = %[2]q
  name        = %[1]q
  rule_type   = "SYSTEM"
}

resource "aws_route53_resolver_rule_association" "test" {
  name             = %[1]q
  resolver_rule_id = aws_route53_resolver_rule.test.id
  vpc_id           = aws_vpc.test.id
}
`, rName, domainName)
}
