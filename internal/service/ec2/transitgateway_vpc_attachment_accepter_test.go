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

func testAccTransitGatewayVPCAttachmentAccepter_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment_accepter.test"
	vpcAttachmentName := "aws_ec2_transit_gateway_vpc_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	vpcResourceName := "aws_vpc.test"
	callerIdentityDatasourceName := "data.aws_caller_identity.creator"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "appliance_mode_support", string(awstypes.ApplianceModeSupportValueDisable)),
					resource.TestCheckResourceAttr(resourceName, "dns_support", string(awstypes.DnsSupportValueEnable)),
					resource.TestCheckResourceAttr(resourceName, "ipv6_support", string(awstypes.Ipv6SupportValueDisable)),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, transitGatewayResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, vpcAttachmentName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_owner_id", callerIdentityDatasourceName, names.AttrAccountID),
				),
			},
			{
				Config:            testAccTransitGatewayVPCAttachmentAccepterConfig_basic(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayVPCAttachmentAccepter_tags(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment_accepter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config:            testAccTransitGatewayVPCAttachmentAccepterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachmentAccepter_TransitGatewayDefaultRouteTableAssociationAndPropagation(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway awstypes.TransitGateway
	var transitGatewayVpcAttachment awstypes.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment_accepter.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", acctest.CtFalse),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentAssociated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", acctest.CtFalse),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentPropagated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", acctest.CtTrue),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentAssociated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentPropagated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachmentAccepterConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_ec2_transit_gateway.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_caller_identity.creator.account_id
  resource_share_arn = aws_ram_resource_share.test.id
}

# VPC attachment creator.
data "aws_caller_identity" "creator" {
  provider = "awsalternate"
}

resource "aws_vpc" "test" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  provider = "awsalternate"

  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTransitGatewayVPCAttachmentAccepterConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentAccepterConfig_base(rName), `
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}
`)
}

func testAccTransitGatewayVPCAttachmentAccepterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentAccepterConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccTransitGatewayVPCAttachmentAccepterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentAccepterConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccTransitGatewayVPCAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName string, association, propagation bool) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentAccepterConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    Name = %[1]q
  }

  transit_gateway_default_route_table_association = %[2]t
  transit_gateway_default_route_table_propagation = %[3]t
}
`, rName, association, propagation))
}
