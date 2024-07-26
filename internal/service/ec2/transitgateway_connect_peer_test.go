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
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayConnectPeer_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"
	transitGatewayConnectResourceName := "aws_ec2_transit_gateway_connect.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayConnect(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayConnectPeerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", "64512"),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_peer_address"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "bgp_transit_gateway_addresses.#", 0),
					resource.TestCheckResourceAttr(resourceName, "inside_cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "peer_address", "1.1.1.1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "transit_gateway_address"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, transitGatewayConnectResourceName, names.AttrID),
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

func testAccTransitGatewayConnectPeer_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayConnect(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayConnectPeerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGatewayConnectPeer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayConnectPeer_bgpASN(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayConnect(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayConnectPeerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerConfig_bgpASN2(rName, "4294967294"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", "4294967294"),
				),
			},
		},
	})
}

func testAccTransitGatewayConnectPeer_insideCIDRBlocks(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayConnect(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayConnectPeerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerConfig_insideCIDRBlocks2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "inside_cidr_blocks.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "inside_cidr_blocks.*", "169.254.200.0/29"),
					resource.TestCheckTypeSetElemAttr(resourceName, "inside_cidr_blocks.*", "fd00::/125"),
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

func testAccTransitGatewayConnectPeer_tags(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayConnect(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayConnectPeerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(ctx, resourceName, &v),
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
				Config: testAccTransitGatewayConnectPeerConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTransitGatewayConnectPeerConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccTransitGatewayConnectPeer_TransitGatewayAddress(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayConnectPeer
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayConnect(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayConnectPeerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerConfig_address(rName, "10.20.30.200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectPeerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_address", "10.20.30.200"),
				),
			},
		},
	})
}

func testAccCheckTransitGatewayConnectPeerExists(ctx context.Context, n string, v *awstypes.TransitGatewayConnectPeer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Connect Peer ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindTransitGatewayConnectPeerByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayConnectPeerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_connect_peer" {
				continue
			}

			_, err := tfec2.FindTransitGatewayConnectPeerByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Transit Gateway Connect Peer %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTransitGatewayConnectPeerConfig_basic(rName string) string {
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
}
`, rName))
}

func testAccTransitGatewayConnectPeerConfig_bgpASN2(rName string, bgpAsn string) string {
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
  bgp_asn                       = %[2]q
  inside_cidr_blocks            = ["169.254.200.0/29"]
  peer_address                  = "1.1.1.1"
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, bgpAsn))
}

func testAccTransitGatewayConnectPeerConfig_tags1(rName, tagKey1, tagValue1 string) string {
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
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccTransitGatewayConnectPeerConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccTransitGatewayConnectPeerConfig_insideCIDRBlocks2(rName string) string {
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
  inside_cidr_blocks            = ["169.254.200.0/29", "fd00::/125"]
  peer_address                  = "1.1.1.1"
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTransitGatewayConnectPeerConfig_address(rName, transitGatewayAddress string) string {
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
  transit_gateway_address       = %[2]q
  inside_cidr_blocks            = ["169.254.200.0/29"]
  peer_address                  = "1.1.1.1"
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, transitGatewayAddress))
}
