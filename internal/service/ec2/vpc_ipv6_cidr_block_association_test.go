// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCIPv6CIDRBlockAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var associationSecondary, associationTertiary awstypes.VpcIpv6CidrBlockAssociation
	resource1Name := "aws_vpc_ipv6_cidr_block_association.secondary_cidr"
	resource2Name := "aws_vpc_ipv6_cidr_block_association.tertiary_cidr"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIPv6CIDRBlockAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv6CIDRBlockAssociationConfig_amazonProvided(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVPCIPv6CIDRBlockAssociationExists(ctx, resource1Name, &associationSecondary),
					testAccCheckVPCIPv6CIDRBlockAssociationExists(ctx, resource2Name, &associationTertiary),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resource1Name, plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(resource2Name, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resource1Name, tfjsonpath.New("ip_source"), knownvalue.StringExact("amazon")),
					statecheck.ExpectKnownValue(resource2Name, tfjsonpath.New("ip_source"), knownvalue.StringExact("amazon")),
					statecheck.ExpectKnownValue(resource1Name, tfjsonpath.New("ipv6_address_attribute"), knownvalue.StringExact("public")),
					statecheck.ExpectKnownValue(resource2Name, tfjsonpath.New("ipv6_address_attribute"), knownvalue.StringExact("public")),
					statecheck.ExpectKnownValue(resource1Name, tfjsonpath.New("ipv6_pool"), knownvalue.StringExact("Amazon")),
					statecheck.ExpectKnownValue(resource2Name, tfjsonpath.New("ipv6_pool"), knownvalue.StringExact("Amazon")),
				},
			},
			{
				ResourceName:      resource1Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resource2Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCIPv6CIDRBlockAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var associationSecondary, associationTertiary awstypes.VpcIpv6CidrBlockAssociation
	resource1Name := "aws_vpc_ipv6_cidr_block_association.secondary_cidr"
	resource2Name := "aws_vpc_ipv6_cidr_block_association.tertiary_cidr"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIPv6CIDRBlockAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv6CIDRBlockAssociationConfig_amazonProvided(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv6CIDRBlockAssociationExists(ctx, resource1Name, &associationSecondary),
					testAccCheckVPCIPv6CIDRBlockAssociationExists(ctx, resource2Name, &associationTertiary),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCIPv6CIDRBlockAssociation(), resource1Name),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCIPv6CIDRBlockAssociation_ipamBasic(t *testing.T) {
	ctx := acctest.Context(t)
	var associationSecondary awstypes.VpcIpv6CidrBlockAssociation
	resourceName := "aws_vpc_ipv6_cidr_block_association.secondary_cidr"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIPv6CIDRBlockAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv6CIDRBlockAssociationConfig_ipam(rName, 56),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv6CIDRBlockAssociationExists(ctx, resourceName, &associationSecondary),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ip_source"), knownvalue.StringExact("amazon")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ipv6_address_attribute"), knownvalue.StringExact("public")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ipv6_pool"), knownvalue.StringExact("IPAM Managed")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ipv6_cidr_block"), knownvalue.StringRegexp(regexache.MustCompile(`/56$`))),
				},
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("Not Found: %s", resourceName)
					}

					return fmt.Sprintf("%s,%s,%s", rs.Primary.ID, rs.Primary.Attributes["ipv6_ipam_pool_id"], rs.Primary.Attributes["ipv6_netmask_length"]), nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCIPv6CIDRBlockAssociation_ipamBasicExplicitCIDR(t *testing.T) {
	ctx := acctest.Context(t)
	var associationSecondary awstypes.VpcIpv6CidrBlockAssociation
	resourceName := "aws_vpc_ipv6_cidr_block_association.secondary_cidr"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIPv6CIDRBlockAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv6CIDRBlockAssociationConfig_ipamExplicit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv6CIDRBlockAssociationExists(ctx, resourceName, &associationSecondary),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ip_source"), knownvalue.StringExact("amazon")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ipv6_address_attribute"), knownvalue.StringExact("public")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ipv6_pool"), knownvalue.StringExact("IPAM Managed")),
				},
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("Not Found: %s", resourceName)
					}

					return fmt.Sprintf("%s,%s", rs.Primary.ID, rs.Primary.Attributes["ipv6_ipam_pool_id"]), nil
				},
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVPCIPv6CIDRBlockAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_ipv6_cidr_block_association" {
				continue
			}

			_, _, err := tfec2.FindVPCIPv6CIDRBlockAssociationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 VPC IPv6 CIDR Block Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCIPv6CIDRBlockAssociationExists(ctx context.Context, n string, v *awstypes.VpcIpv6CidrBlockAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, _, err := tfec2.FindVPCIPv6CIDRBlockAssociationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCIPv6CIDRBlockAssociationConfig_amazonProvided(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipv6_cidr_block_association" "secondary_cidr" {
  vpc_id                           = aws_vpc.test.id
  assign_generated_ipv6_cidr_block = true
}

resource "aws_vpc_ipv6_cidr_block_association" "tertiary_cidr" {
  vpc_id                           = aws_vpc.test.id
  assign_generated_ipv6_cidr_block = true
}
`, rName)
}

func testAccVPCIPv6CIDRBlockAssociationConfig_ipam(rName string, netmaskLength int) string {
	return acctest.ConfigCompose(testAccVPCConfig_baseIPAMIPv6(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block          = "10.1.0.0/16"
  ipv6_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv6_netmask_length = 56

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}

resource "aws_vpc_ipv6_cidr_block_association" "secondary_cidr" {
  ipv6_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv6_netmask_length = %[2]d
  vpc_id              = aws_vpc.test.id
}
`, rName, netmaskLength))
}

func testAccVPCIPv6CIDRBlockAssociationConfig_ipamExplicit(rName string) string {
	return acctest.ConfigCompose(testAccVPCConfig_baseIPAMIPv6(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block          = "10.1.0.0/16"
  ipv6_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv6_netmask_length = 56

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}

resource "aws_vpc_ipv6_cidr_block_association" "secondary_cidr" {
  ipv6_ipam_pool_id = aws_vpc_ipam_pool.test.id
  ipv6_cidr_block   = cidrsubnets(aws_vpc_ipam_pool_cidr.test.cidr, 4, 4)[1]
  vpc_id            = aws_vpc.test.id
}
`, rName))
}
