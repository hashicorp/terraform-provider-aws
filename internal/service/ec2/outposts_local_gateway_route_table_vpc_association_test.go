// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2OutpostsLocalGatewayRouteTableVPCAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	localGatewayRouteTableDataSourceName := "data.aws_ec2_local_gateway_route_table.test"
	resourceName := "aws_ec2_local_gateway_route_table_vpc_association.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocalGatewayRouteTableVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteTableVPCAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVPCAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_id", localGatewayRouteTableDataSourceName, "local_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_route_table_id", localGatewayRouteTableDataSourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccEC2OutpostsLocalGatewayRouteTableVPCAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_local_gateway_route_table_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocalGatewayRouteTableVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteTableVPCAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVPCAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceLocalGatewayRouteTableVPCAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayRouteTableVPCAssociation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_local_gateway_route_table_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocalGatewayRouteTableVPCAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteTableVPCAssociationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVPCAssociationExists(ctx, resourceName),
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
				Config: testAccOutpostsLocalGatewayRouteTableVPCAssociationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVPCAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccOutpostsLocalGatewayRouteTableVPCAssociationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteTableVPCAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckLocalGatewayRouteTableVPCAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindLocalGatewayRouteTableVPCAssociationByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckLocalGatewayRouteTableVPCAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_local_gateway_route_table_vpc_association" {
				continue
			}

			_, err := tfec2.FindLocalGatewayRouteTableVPCAssociationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Local Gateway Route Table VPC Association still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccLocalGatewayRouteTableVPCAssociationBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  outpost_arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccOutpostsLocalGatewayRouteTableVPCAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccLocalGatewayRouteTableVPCAssociationBaseConfig(rName),
		`
resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.test.id
}
`)
}

func testAccOutpostsLocalGatewayRouteTableVPCAssociationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccLocalGatewayRouteTableVPCAssociationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccOutpostsLocalGatewayRouteTableVPCAssociationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccLocalGatewayRouteTableVPCAssociationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
