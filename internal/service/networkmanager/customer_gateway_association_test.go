// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerCustomerGatewayAssociation_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccCustomerGatewayAssociation_basic,
		acctest.CtDisappears: testAccCustomerGatewayAssociation_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccCustomerGatewayAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_customer_gateway_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerGatewayAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayAssociationExists(ctx, resourceName),
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

func testAccCustomerGatewayAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_customer_gateway_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerGatewayAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfnetworkmanager.ResourceCustomerGatewayAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCustomerGatewayAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_customer_gateway_association" {
				continue
			}

			globalNetworkID, customerGatewayARN, err := tfnetworkmanager.CustomerGatewayAssociationParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfnetworkmanager.FindCustomerGatewayAssociationByTwoPartKey(ctx, conn, globalNetworkID, customerGatewayARN)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager Customer Gateway Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCustomerGatewayAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Customer Gateway Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn(ctx)

		globalNetworkID, customerGatewayARN, err := tfnetworkmanager.CustomerGatewayAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfnetworkmanager.FindCustomerGatewayAssociationByTwoPartKey(ctx, conn, globalNetworkID, customerGatewayARN)

		return err
	}
}

func testAccCustomerGatewayAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_device" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  site_id           = aws_networkmanager_site.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = 65534
  ip_address = "12.1.2.3"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = aws_customer_gateway.test.type
  static_routes_only  = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_transit_gateway_registration" "test" {
  global_network_id   = aws_networkmanager_global_network.test.id
  transit_gateway_arn = aws_ec2_transit_gateway.test.arn

  depends_on = [aws_vpn_connection.test]
}

resource "aws_networkmanager_customer_gateway_association" "test" {
  global_network_id    = aws_networkmanager_global_network.test.id
  customer_gateway_arn = aws_customer_gateway.test.arn
  device_id            = aws_networkmanager_device.test.id

  depends_on = [aws_networkmanager_transit_gateway_registration.test]
}
`, rName)
}
