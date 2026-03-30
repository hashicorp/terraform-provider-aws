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

func TestAccSiteVPNGatewayRoutePropagation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpn_gateway_route_propagation.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPNGatewayRoutePropagationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPNGatewayRoutePropagationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNGatewayRoutePropagationExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccSiteVPNGatewayRoutePropagation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpn_gateway_route_propagation.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPNGatewayRoutePropagationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPNGatewayRoutePropagationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNGatewayRoutePropagationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceVPNGatewayRoutePropagation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPNGatewayRoutePropagationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route Table VPN Gateway route propagation ID is set")
		}

		routeTableID, gatewayID, err := tfec2.VPNGatewayRoutePropagationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		return tfec2.FindVPNGatewayRoutePropagationExists(ctx, conn, routeTableID, gatewayID)
	}
}

func testAccCheckVPNGatewayRoutePropagationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpn_gateway_route_propagation" {
				continue
			}

			routeTableID, gatewayID, err := tfec2.VPNGatewayRoutePropagationParseID(rs.Primary.ID)

			if err != nil {
				return err
			}

			conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

			err = tfec2.FindVPNGatewayRoutePropagationExists(ctx, conn, routeTableID, gatewayID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route Table (%s) VPN Gateway (%s) route propagation still exists", routeTableID, gatewayID)
		}

		return nil
	}
}

func testAccVPNGatewayRoutePropagationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_route_propagation" "test" {
  vpn_gateway_id = aws_vpn_gateway.test.id
  route_table_id = aws_route_table.test.id
}
`, rName)
}
