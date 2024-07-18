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

func TestAccEC2OutpostsLocalGatewayRoute_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rInt := sdkacctest.RandIntRange(0, 255)
	destinationCidrBlock := fmt.Sprintf("172.16.%d.0/24", rInt)
	localGatewayRouteTableDataSourceName := "data.aws_ec2_local_gateway_route_table.test"
	localGatewayVirtualInterfaceGroupDataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.test"
	resourceName := "aws_ec2_local_gateway_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocalGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteConfig_destinationCIDRBlock(destinationCidrBlock),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidrBlock),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_route_table_id", localGatewayRouteTableDataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_virtual_interface_group_id", localGatewayVirtualInterfaceGroupDataSourceName, names.AttrID),
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

func TestAccEC2OutpostsLocalGatewayRoute_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rInt := sdkacctest.RandIntRange(0, 255)
	destinationCidrBlock := fmt.Sprintf("172.16.%d.0/24", rInt)
	resourceName := "aws_ec2_local_gateway_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocalGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteConfig_destinationCIDRBlock(destinationCidrBlock),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceLocalGatewayRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLocalGatewayRouteExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindLocalGatewayRouteByTwoPartKey(ctx, conn, rs.Primary.Attributes["local_gateway_route_table_id"], rs.Primary.Attributes["destination_cidr_block"])

		return err
	}
}

func testAccCheckLocalGatewayRouteDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_local_gateway_route" {
				continue
			}

			_, err := tfec2.FindLocalGatewayRouteByTwoPartKey(ctx, conn, rs.Primary.Attributes["local_gateway_route_table_id"], rs.Primary.Attributes["destination_cidr_block"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Local Gateway Route still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccOutpostsLocalGatewayRouteConfig_destinationCIDRBlock(destinationCidrBlock string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}

resource "aws_ec2_local_gateway_route" "test" {
  destination_cidr_block                   = %[1]q
  local_gateway_route_table_id             = data.aws_ec2_local_gateway_route_table.test.id
  local_gateway_virtual_interface_group_id = data.aws_ec2_local_gateway_virtual_interface_group.test.id
}
`, destinationCidrBlock)
}
