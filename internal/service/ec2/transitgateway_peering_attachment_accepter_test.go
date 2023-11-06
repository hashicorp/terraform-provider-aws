// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccTransitGatewayPeeringAttachmentAccepter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
	peeringAttachmentName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentAccepterConfig_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", transitGatewayResourceNamePeer, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceNamePeer, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", peeringAttachmentName, "id"),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentAccepterConfig_sameAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentAccepter_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentAccepterConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentAccepterConfig_tags1(rName, "key1", "value1"),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayPeeringAttachmentAccepterConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTransitGatewayPeeringAttachmentAccepterConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentAccepter_differentAccount(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
	peeringAttachmentName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentAccepterConfig_differentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", transitGatewayResourceNamePeer, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceNamePeer, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", peeringAttachmentName, "id"),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentAccepterConfig_differentAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentAccepterConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "peer" {
  provider = "awsalternate"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  provider = "awsalternate"

  peer_account_id         = aws_ec2_transit_gateway.test.owner_id
  peer_region             = data.aws_region.current.name
  peer_transit_gateway_id = aws_ec2_transit_gateway.test.id
  transit_gateway_id      = aws_ec2_transit_gateway.peer.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTransitGatewayPeeringAttachmentAccepterConfig_sameAccount(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		testAccTransitGatewayPeeringAttachmentAccepterConfig_base(rName),
		`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_peering_attachment.test.id
}
`)
}

func testAccTransitGatewayPeeringAttachmentAccepterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		testAccTransitGatewayPeeringAttachmentAccepterConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_peering_attachment.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccTransitGatewayPeeringAttachmentAccepterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		testAccTransitGatewayPeeringAttachmentAccepterConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_peering_attachment.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccTransitGatewayPeeringAttachmentAccepterConfig_differentAccount(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountAlternateRegionProvider(),
		testAccTransitGatewayPeeringAttachmentAccepterConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_peering_attachment.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
