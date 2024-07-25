// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strconv"
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

func TestAccVPCNATGateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttrSet(resourceName, "allocation_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttr(resourceName, "connectivity_type", "public"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tags.#", acctest.Ct0),
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

func TestAccVPCNATGateway_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceNATGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCNATGateway_ConnectivityType_private(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_connectivityType(rName, "private"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "allocation_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrAssociationID, ""),
					resource.TestCheckResourceAttr(resourceName, "connectivity_type", "private"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", acctest.Ct0),
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

func TestAccVPCNATGateway_privateIP(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_privateIP(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "allocation_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrAssociationID, ""),
					resource.TestCheckResourceAttr(resourceName, "connectivity_type", "private"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttr(resourceName, "private_ip", "10.0.0.8"),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", acctest.Ct0),
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

func TestAccVPCNATGateway_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
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
				Config: testAccVPCNATGatewayConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCNATGatewayConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCNATGateway_secondaryAllocationIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	eipResourceName := "aws_eip.secondary"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_secondaryAllocationIDs(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secondary_allocation_ids.*", eipResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCNATGatewayConfig_secondaryAllocationIDs(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", acctest.Ct0),
				),
			},
			{
				Config: testAccVPCNATGatewayConfig_secondaryAllocationIDs(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secondary_allocation_ids.*", eipResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPCNATGateway_secondaryPrivateIPAddressCount(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	secondaryPrivateIpAddressCount := 3

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddressCount(rName, secondaryPrivateIpAddressCount),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", strconv.Itoa(secondaryPrivateIpAddressCount)),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", strconv.Itoa(secondaryPrivateIpAddressCount)),
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

func TestAccVPCNATGateway_secondaryPrivateIPAddresses(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	eipResourceName := "aws_eip.secondary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secondary_allocation_ids.*", eipResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", acctest.Ct0),
				),
			},
			{
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secondary_allocation_ids.*", eipResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.5"),
				),
			},
		},
	})
}

func TestAccVPCNATGateway_SecondaryPrivateIPAddresses_private(t *testing.T) {
	ctx := acctest.Context(t)
	var natGateway awstypes.NatGateway
	resourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNATGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses_private(rName, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "5"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.6"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.9"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses_private(rName, 7),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", "7"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", "7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.6"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.9"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.10"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.11"),
				),
			},
			{
				Config: testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses_private(rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNATGatewayExists(ctx, resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "secondary_allocation_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_address_count", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ip_addresses.#", acctest.Ct4),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.6"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "secondary_private_ip_addresses.*", "10.0.1.8"),
				),
			},
		},
	})
}

func testAccCheckNATGatewayDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_nat_gateway" {
				continue
			}

			_, err := tfec2.FindNATGatewayByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 NAT Gateway %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNATGatewayExists(ctx context.Context, n string, v *awstypes.NatGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 NAT Gateway ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindNATGatewayByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccNATGatewayConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "private" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.2.0/24"
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCNATGatewayConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), `
resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.public.id

  depends_on = [aws_internet_gateway.test]
}
`)
}

func testAccVPCNATGatewayConfig_connectivityType(rName, connectivityType string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  connectivity_type = %[2]q
  subnet_id         = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName, connectivityType))
}

func testAccVPCNATGatewayConfig_privateIP(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  connectivity_type = "private"
  private_ip        = "10.0.0.8"
  subnet_id         = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCNATGatewayConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.public.id

  tags = {
    %[1]q = %[2]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, tagKey1, tagValue1))
}

func testAccVPCNATGatewayConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.public.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVPCNATGatewayConfig_secondaryAllocationIDs(rName string, hasSecondary bool) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_eip" "secondary" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test" {
  allocation_id            = aws_eip.test.id
  subnet_id                = aws_subnet.public.id
  secondary_allocation_ids = %[2]t ? [aws_eip.secondary.id] : null

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName, hasSecondary))
}

func testAccVPCNATGatewayConfig_secondaryPrivateIPAddressCount(rName string, secondaryPrivateIpAddressCount int) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  connectivity_type                  = "private"
  subnet_id                          = aws_subnet.public.id
  secondary_private_ip_address_count = %[2]d

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName, secondaryPrivateIpAddressCount))
}

func testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses(rName string, hasSecondary bool) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_eip" "secondary" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test" {
  allocation_id                  = aws_eip.test.id
  subnet_id                      = aws_subnet.private.id
  secondary_allocation_ids       = %[2]t ? [aws_eip.secondary.id] : null
  secondary_private_ip_addresses = %[2]t ? ["10.0.1.5"] : null

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName, hasSecondary))
}

func testAccVPCNATGatewayConfig_secondaryPrivateIPAddresses_private(rName string, n int) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  connectivity_type              = "private"
  subnet_id                      = aws_subnet.private.id
  secondary_private_ip_addresses = [for n in range(%[2]d) : "10.0.1.${5 + n}"]

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName, n))
}
