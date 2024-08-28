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

func testAccTransitGatewayRouteTableAssociation_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayRouteTableAssociation
	resourceName := "aws_ec2_transit_gateway_route_table_association.test"
	transitGatewayRouteTableResourceName := "aws_ec2_transit_gateway_route_table.test"
	transitGatewayVpcAttachmentResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTableAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "replace_existing_association", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceType),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, transitGatewayVpcAttachmentResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", transitGatewayRouteTableResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replace_existing_association"},
			},
		},
	})
}

func testAccTransitGatewayRouteTableAssociation_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayRouteTableAssociation
	resourceName := "aws_ec2_transit_gateway_route_table_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTableAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableAssociationExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGatewayRouteTableAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayRouteTableAssociation_replaceExistingAssociation(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayRouteTableAssociation
	resourceName := "aws_ec2_transit_gateway_route_table_association.test"
	transitGatewayRouteTableResourceName := "aws_ec2_transit_gateway_route_table.test"
	transitGatewayVpcAttachmentResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTableAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAssociationConfig_replaceExistingAssociation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "replace_existing_association", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceType),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, transitGatewayVpcAttachmentResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", transitGatewayRouteTableResourceName, names.AttrID),
					resource.TestCheckResourceAttr(transitGatewayVpcAttachmentResourceName, "transit_gateway_default_route_table_association", acctest.CtTrue),
					resource.TestCheckResourceAttr(transitGatewayVpcAttachmentResourceName, "transit_gateway_default_route_table_propagation", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replace_existing_association"},
			},
		},
	})
}

func testAccCheckTransitGatewayRouteTableAssociationExists(ctx context.Context, n string, v *awstypes.TransitGatewayRouteTableAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindTransitGatewayRouteTableAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["transit_gateway_route_table_id"], rs.Primary.Attributes[names.AttrTransitGatewayAttachmentID])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayRouteTableAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_route_table_association" {
				continue
			}

			_, err := tfec2.FindTransitGatewayRouteTableAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["transit_gateway_route_table_id"], rs.Primary.Attributes[names.AttrTransitGatewayAttachmentID])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Transit Gateway Route Table Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTransitGatewayRouteTableAssociationConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}


resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTransitGatewayRouteTableAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayRouteTableAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids                                      = aws_subnet.test[*].id
  transit_gateway_default_route_table_association = false
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  vpc_id                                          = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_route_table_association" "test" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
}
`, rName))
}

func testAccTransitGatewayRouteTableAssociationConfig_replaceExistingAssociation(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayRouteTableAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_route_table_association" "test" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id

  replace_existing_association = true
}
`, rName))
}
