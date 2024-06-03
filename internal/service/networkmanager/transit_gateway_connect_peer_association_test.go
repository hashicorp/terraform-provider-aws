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
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerTransitGatewayConnectPeerAssociation_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:          testAccTransitGatewayConnectPeerAssociation_basic,
		acctest.CtDisappears:     testAccTransitGatewayConnectPeerAssociation_disappears,
		"disappears_ConnectPeer": testAccTransitGatewayConnectPeerAssociation_Disappears_connectPeer,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccTransitGatewayConnectPeerAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_transit_gateway_connect_peer_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayConnectPeerAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerAssociationExists(ctx, resourceName),
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

func testAccTransitGatewayConnectPeerAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_transit_gateway_connect_peer_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayConnectPeerAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfnetworkmanager.ResourceTransitGatewayConnectPeerAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayConnectPeerAssociation_Disappears_connectPeer(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_transit_gateway_connect_peer_association.test"
	connetPeerResourceName := "aws_ec2_transit_gateway_connect_peer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayConnectPeerAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGatewayConnectPeer(), connetPeerResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTransitGatewayConnectPeerAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_customer_gateway_association" {
				continue
			}

			globalNetworkID, connectPeerARN, err := tfnetworkmanager.TransitGatewayConnectPeerAssociationParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfnetworkmanager.FindTransitGatewayConnectPeerAssociationByTwoPartKey(ctx, conn, globalNetworkID, connectPeerARN)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager Transit Gateway Connect Peer Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTransitGatewayConnectPeerAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Transit Gateway Connect Peer Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn(ctx)

		globalNetworkID, connectPeerARN, err := tfnetworkmanager.TransitGatewayConnectPeerAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfnetworkmanager.FindTransitGatewayConnectPeerAssociationByTwoPartKey(ctx, conn, globalNetworkID, connectPeerARN)

		return err
	}
}

func testAccTransitGatewayConnectPeerAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  transit_gateway_cidr_blocks = ["10.20.30.0/24"]

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
  inside_cidr_blocks            = ["169.254.200.0/29"]
  peer_address                  = "1.1.1.1"
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id

  tags = {
    Name = %[1]q
  }
}

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

resource "aws_networkmanager_transit_gateway_registration" "test" {
  global_network_id   = aws_networkmanager_global_network.test.id
  transit_gateway_arn = aws_ec2_transit_gateway.test.arn

  depends_on = [aws_ec2_transit_gateway_connect_peer.test]
}

resource "aws_networkmanager_transit_gateway_connect_peer_association" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  device_id         = aws_networkmanager_device.test.id

  transit_gateway_connect_peer_arn = aws_ec2_transit_gateway_connect_peer.test.arn

  depends_on = [aws_networkmanager_transit_gateway_registration.test]
}
`, rName))
}
