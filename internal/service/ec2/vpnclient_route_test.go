// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccClientVPNRoute_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.ClientVpnRoute
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_route.test"
	endpointResourceName := "aws_ec2_client_vpn_endpoint.test"
	subnetResourceName := "aws_subnet.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNRouteConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "client_vpn_endpoint_id", endpointResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "origin", "add-route"),
					resource.TestCheckResourceAttrPair(resourceName, "target_vpc_subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "type", "Nat"),
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

func testAccClientVPNRoute_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.ClientVpnRoute
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNRouteConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNRouteExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceClientVPNRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccClientVPNRoute_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v ec2.ClientVpnRoute
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNRouteConfig_description(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClientVPNRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "test client VPN route"),
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

func testAccCheckClientVPNRouteDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_client_vpn_route" {
				continue
			}

			endpointID, targetSubnetID, destinationCIDR, err := tfec2.ClientVPNRouteParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfec2.FindClientVPNRouteByThreePartKey(ctx, conn, endpointID, targetSubnetID, destinationCIDR)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Client VPN Route %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClientVPNRouteExists(ctx context.Context, name string, v *ec2.ClientVpnRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Client VPN Route ID is set")
		}

		endpointID, targetSubnetID, destinationCIDR, err := tfec2.ClientVPNRouteParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		output, err := tfec2.FindClientVPNRouteByThreePartKey(ctx, conn, endpointID, targetSubnetID, destinationCIDR)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClientVPNRouteConfig_base(t *testing.T, rName string, subnetCount int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), testAccClientVPNEndpointConfig_basic(t, rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = %[2]d

  availability_zone       = data.aws_availability_zones.available.names[count.index]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}
`, rName, subnetCount))
}

func testAccClientVPNRouteConfig_basic(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccClientVPNRouteConfig_base(t, rName, 1), `
resource "aws_ec2_client_vpn_route" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_network_association.test.client_vpn_endpoint_id
  destination_cidr_block = "0.0.0.0/0"
  target_vpc_subnet_id   = aws_subnet.test[0].id
}

resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test[0].id
}
`)
}

func testAccClientVPNRouteConfig_description(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccClientVPNRouteConfig_base(t, rName, 1), `
resource "aws_ec2_client_vpn_route" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_network_association.test.client_vpn_endpoint_id
  destination_cidr_block = "0.0.0.0/0"
  target_vpc_subnet_id   = aws_subnet.test[0].id
  description            = "test client VPN route"
}

resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test[0].id
}
`)
}
