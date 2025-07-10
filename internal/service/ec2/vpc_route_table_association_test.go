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

func TestAccVPCRouteTableAssociation_Subnet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rta awstypes.RouteTableAssociation
	resourceName := "aws_route_table_association.test"
	resourceNameRouteTable := "aws_route_table.test"
	resourceNameSubnet := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteTableAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTableAssociationConfig_subnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(ctx, resourceName, &rta),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", resourceNameRouteTable, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, resourceNameSubnet, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteTabAssocImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRouteTableAssociation_Subnet_changeRouteTable(t *testing.T) {
	ctx := acctest.Context(t)
	var rta awstypes.RouteTableAssociation
	resourceName := "aws_route_table_association.test"
	resourceNameRouteTable1 := "aws_route_table.test"
	resourceNameRouteTable2 := "aws_route_table.test2"
	resourceNameSubnet := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteTableAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTableAssociationConfig_subnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(ctx, resourceName, &rta),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", resourceNameRouteTable1, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, resourceNameSubnet, names.AttrID),
				),
			},
			{
				Config: testAccVPCRouteTableAssociationConfig_subnetChange(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(ctx, resourceName, &rta),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", resourceNameRouteTable2, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, resourceNameSubnet, names.AttrID),
				),
			},
		},
	})
}

func TestAccVPCRouteTableAssociation_Gateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rta awstypes.RouteTableAssociation
	resourceName := "aws_route_table_association.test"
	resourceNameRouteTable := "aws_route_table.test"
	resourceNameGateway := "aws_internet_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteTableAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTableAssociationConfig_gateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(ctx, resourceName, &rta),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", resourceNameRouteTable, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", resourceNameGateway, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteTabAssocImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRouteTableAssociation_Gateway_changeRouteTable(t *testing.T) {
	ctx := acctest.Context(t)
	var rta awstypes.RouteTableAssociation
	resourceName := "aws_route_table_association.test"
	resourceNameRouteTable1 := "aws_route_table.test"
	resourceNameRouteTable2 := "aws_route_table.test2"
	resourceNameGateway := "aws_internet_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteTableAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTableAssociationConfig_gateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(ctx, resourceName, &rta),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", resourceNameRouteTable1, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", resourceNameGateway, names.AttrID),
				),
			},
			{
				Config: testAccVPCRouteTableAssociationConfig_gatewayChange(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(ctx, resourceName, &rta),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", resourceNameRouteTable2, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", resourceNameGateway, names.AttrID),
				),
			},
		},
	})
}

func TestAccVPCRouteTableAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var rta awstypes.RouteTableAssociation
	resourceName := "aws_route_table_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteTableAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTableAssociationConfig_subnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(ctx, resourceName, &rta),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceRouteTableAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRouteTableAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route_table_association" {
				continue
			}

			_, err := tfec2.FindRouteTableAssociationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route table association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRouteTableAssociationExists(ctx context.Context, n string, v *awstypes.RouteTableAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		association, err := tfec2.FindRouteTableAssociationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *association

		return nil
	}
}

func testAccRouteTabAssocImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		var target string
		if rs.Primary.Attributes[names.AttrSubnetID] != "" {
			target = rs.Primary.Attributes[names.AttrSubnetID]
		} else if rs.Primary.Attributes["gateway_id"] != "" {
			target = rs.Primary.Attributes["gateway_id"]
		}
		return fmt.Sprintf("%s/%s", target, rs.Primary.Attributes["route_table_id"]), nil
	}
}

func testAccRouteTableAssociationConfigBaseVPC(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.1.0/24"

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
`, rName)
}

func testAccVPCRouteTableAssociationConfig_subnet(rName string) string {
	return acctest.ConfigCompose(testAccRouteTableAssociationConfigBaseVPC(rName), fmt.Sprintf(`
resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "10.0.0.0/8"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  subnet_id      = aws_subnet.test.id
}
`, rName))
}

func testAccVPCRouteTableAssociationConfig_subnetChange(rName string) string {
	return acctest.ConfigCompose(testAccRouteTableAssociationConfigBaseVPC(rName), fmt.Sprintf(`
resource "aws_route_table" "test2" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "10.0.0.0/8"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test2.id
  subnet_id      = aws_subnet.test.id
}
`, rName))
}

func testAccVPCRouteTableAssociationConfig_gateway(rName string) string {
	return acctest.ConfigCompose(testAccRouteTableAssociationConfigBaseVPC(rName), fmt.Sprintf(`
resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block           = aws_subnet.test.cidr_block
    network_interface_id = aws_network_interface.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  gateway_id     = aws_internet_gateway.test.id
}
`, rName))
}

func testAccVPCRouteTableAssociationConfig_gatewayChange(rName string) string {
	return acctest.ConfigCompose(testAccRouteTableAssociationConfigBaseVPC(rName), fmt.Sprintf(`
resource "aws_route_table" "test2" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block           = aws_subnet.test.cidr_block
    network_interface_id = aws_network_interface.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test2.id
  gateway_id     = aws_internet_gateway.test.id
}
`, rName))
}
