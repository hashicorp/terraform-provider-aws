// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2OutpostsLocalGatewayRoute_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rInt := acctest.RandIntRange(t, 0, 255)
	destinationCidrBlock := fmt.Sprintf("172.16.%d.0/24", rInt)
	localGatewayRouteTableDataSourceName := "data.aws_ec2_local_gateway_route_table.test"
	localGatewayVirtualInterfaceGroupDataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.test"
	resourceName := "aws_ec2_local_gateway_route.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocalGatewayRouteDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteConfig_destinationCIDRBlock(destinationCidrBlock),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteExists(ctx, t, resourceName),
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
	rInt := acctest.RandIntRange(t, 0, 255)
	destinationCidrBlock := fmt.Sprintf("172.16.%d.0/24", rInt)
	resourceName := "aws_ec2_local_gateway_route.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocalGatewayRouteDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteConfig_destinationCIDRBlock(destinationCidrBlock),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocalGatewayRouteExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceLocalGatewayRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLocalGatewayRouteExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		_, err := tfec2.FindLocalGatewayRouteByTwoPartKey(ctx, conn, rs.Primary.Attributes["local_gateway_route_table_id"], rs.Primary.Attributes["destination_cidr_block"])

		return err
	}
}

func testAccCheckLocalGatewayRouteDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_local_gateway_route" {
				continue
			}

			_, err := tfec2.FindLocalGatewayRouteByTwoPartKey(ctx, conn, rs.Primary.Attributes["local_gateway_route_table_id"], rs.Primary.Attributes["destination_cidr_block"])

			if retry.NotFound(err) {
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
