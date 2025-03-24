// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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

func TestAccVPCIPv4CIDRBlockAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var associationSecondary, associationTertiary awstypes.VpcCidrBlockAssociation
	resource1Name := "aws_vpc_ipv4_cidr_block_association.secondary_cidr"
	resource2Name := "aws_vpc_ipv4_cidr_block_association.tertiary_cidr"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIPv4CIDRBlockAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv4CIDRBlockAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv4CIDRBlockAssociationExists(ctx, resource1Name, &associationSecondary),
					testAccCheckVPCIPv4CIDRBlockAssociationExists(ctx, resource2Name, &associationTertiary),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resource1Name, plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(resource2Name, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resource1Name, tfjsonpath.New(names.AttrCIDRBlock), knownvalue.StringExact("172.2.0.0/16")),
					statecheck.ExpectKnownValue(resource2Name, tfjsonpath.New(names.AttrCIDRBlock), knownvalue.StringExact("170.2.0.0/16")),
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

func TestAccVPCIPv4CIDRBlockAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var associationSecondary, associationTertiary awstypes.VpcCidrBlockAssociation
	resource1Name := "aws_vpc_ipv4_cidr_block_association.secondary_cidr"
	resource2Name := "aws_vpc_ipv4_cidr_block_association.tertiary_cidr"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIPv4CIDRBlockAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv4CIDRBlockAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv4CIDRBlockAssociationExists(ctx, resource1Name, &associationSecondary),
					testAccCheckVPCIPv4CIDRBlockAssociationExists(ctx, resource2Name, &associationTertiary),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPCIPv4CIDRBlockAssociation(), resource1Name),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCIPv4CIDRBlockAssociation_ipamBasic(t *testing.T) {
	ctx := acctest.Context(t)
	var associationSecondary awstypes.VpcCidrBlockAssociation
	resourceName := "aws_vpc_ipv4_cidr_block_association.secondary_cidr"
	ipamPoolResourceName := "aws_vpc_ipam_pool.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIPv4CIDRBlockAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv4CIDRBlockAssociationConfig_ipam(rName, 28),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv4CIDRBlockAssociationExists(ctx, resourceName, &associationSecondary),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCIDRBlock), knownvalue.StringRegexp(regexache.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}/`+`28$`))),
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

					return fmt.Sprintf("%s,%s,%s", rs.Primary.ID, rs.Primary.Attributes["ipv4_ipam_pool_id"], rs.Primary.Attributes["ipv4_netmask_length"]), nil
				},
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCIPv4CIDRBlockAssociationConfig_ipam(rName, 28),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPC(), vpcResourceName),
					testAccCheckVPCIPv4CIDRBlockAssociationWaitVPCIPAMPoolAllocationDeleted(ctx, ipamPoolResourceName, vpcResourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCIPv4CIDRBlockAssociation_ipamBasicExplicitCIDR(t *testing.T) {
	ctx := acctest.Context(t)
	var associationSecondary awstypes.VpcCidrBlockAssociation
	resourceName := "aws_vpc_ipv4_cidr_block_association.secondary_cidr"
	ipamPoolResourceName := "aws_vpc_ipam_pool.test"
	vpcResourceName := "aws_vpc.test"
	cidr := "172.2.0.32/28"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCIPv4CIDRBlockAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCIPv4CIDRBlockAssociationConfig_ipamExplicit(rName, cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCIPv4CIDRBlockAssociationExists(ctx, resourceName, &associationSecondary),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCIDRBlock), knownvalue.StringExact(cidr)),
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

					return fmt.Sprintf("%s,%s", rs.Primary.ID, rs.Primary.Attributes["ipv4_ipam_pool_id"]), nil
				},
				ImportStateVerify: true,
			},
			// Work around "Error: waiting for IPAM Pool CIDR (...) delete: unexpected state 'failed-deprovision', wanted target ''. last error: : The CIDR has one or more allocations".
			{
				Config: testAccVPCIPv4CIDRBlockAssociationConfig_ipamExplicit(rName, cidr),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPC(), vpcResourceName),
					testAccCheckVPCIPv4CIDRBlockAssociationWaitVPCIPAMPoolAllocationDeleted(ctx, ipamPoolResourceName, vpcResourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPCIPv4CIDRBlockAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpc_ipv4_cidr_block_association" {
				continue
			}

			_, _, err := tfec2.FindVPCCIDRBlockAssociationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 VPC IPv4 CIDR Block Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVPCIPv4CIDRBlockAssociationExists(ctx context.Context, n string, v *awstypes.VpcCidrBlockAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, _, err := tfec2.FindVPCCIDRBlockAssociationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVPCIPv4CIDRBlockAssociationWaitVPCIPAMPoolAllocationDeleted(ctx context.Context, nIPAMPool, nVPC string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rsIPAMPool, ok := s.RootModule().Resources[nIPAMPool]
		if !ok {
			return fmt.Errorf("Not found: %s", nIPAMPool)
		}
		rsVPC, ok := s.RootModule().Resources[nVPC]
		if !ok {
			return fmt.Errorf("Not found: %s", nVPC)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		const (
			timeout = 35 * time.Minute // IPAM eventual consistency. It can take ~30 min to release allocations.
		)
		_, err := tfresource.RetryUntilNotFound(ctx, timeout, func() (any, error) {
			return tfec2.FindIPAMPoolAllocationsForVPC(ctx, conn, rsIPAMPool.Primary.ID, rsVPC.Primary.ID)
		})

		return err
	}
}

func testAccVPCIPv4CIDRBlockAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipv4_cidr_block_association" "secondary_cidr" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "172.2.0.0/16"
}

resource "aws_vpc_ipv4_cidr_block_association" "tertiary_cidr" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "170.2.0.0/16"
}
`, rName)
}

func testAccVPCIPv4CIDRBlockAssociationConfig_ipam(rName string, netmaskLength int) string {
	return acctest.ConfigCompose(testAccVPCConfig_baseIPAMIPv4(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipv4_cidr_block_association" "secondary_cidr" {
  ipv4_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv4_netmask_length = %[2]d
  vpc_id              = aws_vpc.test.id

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}
`, rName, netmaskLength))
}

func testAccVPCIPv4CIDRBlockAssociationConfig_ipamExplicit(rName, cidr string) string {
	return acctest.ConfigCompose(testAccVPCConfig_baseIPAMIPv4(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipv4_cidr_block_association" "secondary_cidr" {
  ipv4_ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr_block        = %[2]q
  vpc_id            = aws_vpc.test.id

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}
`, rName, cidr))
}
