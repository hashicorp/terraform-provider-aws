// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53profiles_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/route53profiles/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfroute53profiles "github.com/hashicorp/terraform-provider-aws/internal/service/route53profiles"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ProfilesResourceAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourceAssociation awstypes.ProfileResourceAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53profiles_resource_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAssociationExists(ctx, resourceName, &resourceAssociation),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccRoute53ProfilesResourceAssociation_firewallRuleGroup(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourceAssociation awstypes.ProfileResourceAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53profiles_resource_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssociationConfig_firewallRuleGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAssociationExists(ctx, resourceName, &resourceAssociation),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccRoute53ProfilesResourceAssociation_resolverRule(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourceAssociation awstypes.ProfileResourceAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53profiles_resource_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssociationConfig_resolverRule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAssociationExists(ctx, resourceName, &resourceAssociation),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccRoute53ProfilesResourceAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var resourceAssociation awstypes.ProfileResourceAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53profiles_resource_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAssociationExists(ctx, resourceName, &resourceAssociation),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfroute53profiles.Route53ProfileResourceAssocation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourceAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ProfilesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53profiles_resource_association" {
				continue
			}

			_, err := tfroute53profiles.FindResourceAssociationByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Route53Profiles, create.ErrActionCheckingDestroyed, tfroute53profiles.ResNameResourceAssociation, rs.Primary.ID, err)
			}

			return create.Error(names.Route53Profiles, create.ErrActionCheckingDestroyed, tfroute53profiles.ResNameResourceAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckResourceAssociationExists(ctx context.Context, name string, association *awstypes.ProfileResourceAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Route53Profiles, create.ErrActionCheckingExistence, tfroute53profiles.ResNameResourceAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Route53Profiles, create.ErrActionCheckingExistence, tfroute53profiles.ResNameResourceAssociation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ProfilesClient(ctx)
		resp, err := tfroute53profiles.FindResourceAssociationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.Route53Profiles, create.ErrActionCheckingExistence, tfroute53profiles.ResNameResourceAssociation, rs.Primary.ID, err)
		}

		*association = *resp

		return nil
	}
}

func testAccResourceAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_route53_zone" "test" {
  name = "test.com"

  vpc {
    vpc_id = aws_vpc.test.id
  }
}

resource "aws_route53profiles_resource_association" "test" {
  name         = %[1]q
  profile_id   = aws_route53profiles_profile.test.id
  resource_arn = aws_route53_zone.test.arn
}
`, rName)
}

func testAccResourceAssociationConfig_firewallRuleGroup(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}

resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
}


resource "aws_route53profiles_resource_association" "test" {
  name                = %[1]q
  profile_id          = aws_route53profiles_profile.test.id
  resource_arn        = aws_route53_resolver_firewall_rule_group.test.arn
  resource_properties = jsonencode({ "priority" = 102 })
}
`, rName)
}

func testAccResourceAssociationConfig_resolverRule(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}

resource "aws_route53_resolver_rule" "test" {
  domain_name = "subdomain.test.com"
  rule_type   = "SYSTEM"
}

resource "aws_route53profiles_resource_association" "test" {
  name         = %[1]q
  profile_id   = aws_route53profiles_profile.test.id
  resource_arn = aws_route53_resolver_rule.test.arn
}
`, rName)
}
