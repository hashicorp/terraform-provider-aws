// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCSubnetCIDRReservation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var res awstypes.SubnetCidrReservation
	resourceName := "aws_ec2_subnet_cidr_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetCIDRReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetCIDRReservationConfig_testIPv4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetCIDRReservationExists(ctx, resourceName, &res),
					resource.TestCheckResourceAttr(resourceName, names.AttrCIDRBlock, "10.1.1.16/28"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttr(resourceName, "reservation_type", names.AttrPrefix),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSubnetCIDRReservationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSubnetCIDRReservation_ipv6(t *testing.T) {
	ctx := acctest.Context(t)
	var res awstypes.SubnetCidrReservation
	resourceName := "aws_ec2_subnet_cidr_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetCIDRReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetCIDRReservationConfig_testIPv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetCIDRReservationExists(ctx, resourceName, &res),
					resource.TestCheckResourceAttr(resourceName, "reservation_type", "explicit"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSubnetCIDRReservationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSubnetCIDRReservation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var res awstypes.SubnetCidrReservation
	resourceName := "aws_ec2_subnet_cidr_reservation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetCIDRReservationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnetCIDRReservationConfig_testIPv4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetCIDRReservationExists(ctx, resourceName, &res),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceSubnetCIDRReservation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSubnetCIDRReservationExists(ctx context.Context, n string, v *awstypes.SubnetCidrReservation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Subnet CIDR Reservation ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindSubnetCIDRReservationBySubnetIDAndReservationID(ctx, conn, rs.Primary.Attributes[names.AttrSubnetID], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSubnetCIDRReservationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_subnet_cidr_reservation" {
				continue
			}

			_, err := tfec2.FindSubnetCIDRReservationBySubnetIDAndReservationID(ctx, conn, rs.Primary.Attributes[names.AttrSubnetID], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Subnet CIDR Reservation %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSubnetCIDRReservationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		subnetId := rs.Primary.Attributes[names.AttrSubnetID]
		return fmt.Sprintf("%s:%s", subnetId, rs.Primary.ID), nil
	}
}

func testAccVPCSubnetCIDRReservationConfig_testIPv4(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_subnet_cidr_reservation" "test" {
  cidr_block       = "10.1.1.16/28"
  description      = "test"
  reservation_type = "prefix"
  subnet_id        = aws_subnet.test.id
}
`, rName)
}

func testAccVPCSubnetCIDRReservationConfig_testIPv6(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block      = "10.1.1.0/24"
  vpc_id          = aws_vpc.test.id
  ipv6_cidr_block = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_subnet_cidr_reservation" "test" {
  cidr_block       = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 12, 17)
  reservation_type = "explicit"
  subnet_id        = aws_subnet.test.id
}
`, rName)
}
