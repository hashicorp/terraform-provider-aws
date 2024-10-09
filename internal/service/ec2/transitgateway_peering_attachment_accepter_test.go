// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayPeeringAttachmentAccepter_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment awstypes.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
	peeringAttachmentName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentAccepterConfig_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", transitGatewayResourceNamePeer, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceNamePeer, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, transitGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, peeringAttachmentName, names.AttrID),
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

func testAccTransitGatewayPeeringAttachmentAccepter_tags(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment awstypes.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentAccepterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentAccepterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayPeeringAttachmentAccepterConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTransitGatewayPeeringAttachmentAccepterConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentAccepter_differentAccount(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment awstypes.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
	peeringAttachmentName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentAccepterConfig_differentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", transitGatewayResourceNamePeer, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceNamePeer, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, transitGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, peeringAttachmentName, names.AttrID),
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
