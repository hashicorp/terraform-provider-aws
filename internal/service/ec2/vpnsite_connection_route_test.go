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

func TestAccSiteVPNConnectionRoute_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPNConnectionRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionRouteConfig_basic(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionRouteExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccSiteVPNConnectionRoute_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPNConnectionRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionRouteConfig_basic(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionRouteExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPNConnectionRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVPNConnectionRouteDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpn_connection_route" {
				continue
			}

			_, err := tfec2.FindVPNConnectionRouteByTwoPartKey(ctx, conn, rs.Primary.Attributes["vpn_connection_id"], rs.Primary.Attributes["destination_cidr_block"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 VPN Connection Route %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccVPNConnectionRouteExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindVPNConnectionRouteByTwoPartKey(ctx, conn, rs.Primary.Attributes["vpn_connection_id"], rs.Primary.Attributes["destination_cidr_block"])

		return err
	}
}

func testAccSiteVPNConnectionRouteConfig_basic(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "182.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection_route" "test" {
  destination_cidr_block = "172.168.10.0/24"
  vpn_connection_id      = aws_vpn_connection.test.id
}
`, rName, rBgpAsn)
}
