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

func testAccTransitGatewayPeeringAttachment_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment awstypes.TransitGatewayPeeringAttachment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_sameAccount(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct0),
					acctest.CheckResourceAttrAccountID(resourceName, "peer_account_id"),
					resource.TestCheckResourceAttr(resourceName, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceNamePeer, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, transitGatewayResourceName, names.AttrID),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentConfig_sameAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachment_options(t *testing.T, semaphore tfsync.Semaphore) {
	acctest.Skip(t, "IncorrectState: You cannot create a dynamic peering attachment")

	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment awstypes.TransitGatewayPeeringAttachment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_options_sameAccount(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.dynamic_routing", "enable"),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentConfig_options_sameAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachment_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment awstypes.TransitGatewayPeeringAttachment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGatewayPeeringAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachment_tags(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment awstypes.TransitGatewayPeeringAttachment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachment_differentAccount(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGatewayPeeringAttachment awstypes.TransitGatewayPeeringAttachment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentConfig_differentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(ctx, resourceName, &transitGatewayPeeringAttachment),
					// Test that the peer account ID != the primary (request) account ID
					func(s *terraform.State) error {
						if acctest.CheckResourceAttrAccountID(resourceName, "peer_account_id") == nil {
							return fmt.Errorf("peer_account_id attribute incorrectly to the requester's account ID")
						}
						return nil
					},
					resource.TestCheckResourceAttr(resourceName, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceNamePeer, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, transitGatewayResourceName, names.AttrID),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentConfig_differentAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTransitGatewayPeeringAttachmentExists(ctx context.Context, n string, v *awstypes.TransitGatewayPeeringAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Peering Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindTransitGatewayPeeringAttachmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayPeeringAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_peering_attachment" {
				continue
			}

			_, err := tfec2.FindTransitGatewayPeeringAttachmentByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Transit Gateway Peering Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTransitGatewayPeeringAttachmentConfig_base(rName string) string {
	return fmt.Sprintf(`
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
`, rName)
}

func testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		testAccTransitGatewayPeeringAttachmentConfig_base(rName),
	)
}

func testAccTransitGatewayPeeringAttachmentConfig_differentAccount_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountAlternateRegionProvider(),
		testAccTransitGatewayPeeringAttachmentConfig_base(rName),
	)
}

func testAccTransitGatewayPeeringAttachmentConfig_sameAccount(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
}
`, acctest.AlternateRegion()))
}

func testAccTransitGatewayPeeringAttachmentConfig_options_sameAccount(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id

  options {
    dynamic_routing = "enable"
  }
}
`, acctest.AlternateRegion()))
}

func testAccTransitGatewayPeeringAttachmentConfig_differentAccount(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_differentAccount_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_account_id         = aws_ec2_transit_gateway.peer.owner_id
  peer_region             = %[2]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, acctest.AlternateRegion()))
}

func testAccTransitGatewayPeeringAttachmentConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, acctest.AlternateRegion(), tagKey1, tagValue1))
}

func testAccTransitGatewayPeeringAttachmentConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, acctest.AlternateRegion(), tagKey1, tagValue1, tagKey2, tagValue2))
}
