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

func testAccTransitGatewayRoute_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayRoute
	resourceName := "aws_ec2_transit_gateway_route.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
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
		CheckDestroy:             testAccCheckTransitGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteConfig_destinationCIDRBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "blackhole", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, transitGatewayVpcAttachmentResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", transitGatewayResourceName, "association_default_route_table_id"),
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

func testAccTransitGatewayRoute_basic_ipv6(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayRoute
	resourceName := "aws_ec2_transit_gateway_route.test_ipv6"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
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
		CheckDestroy:             testAccCheckTransitGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteConfig_destinationCIDRBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", "2001:db8::/56"),
					resource.TestCheckResourceAttr(resourceName, "blackhole", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, transitGatewayVpcAttachmentResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", transitGatewayResourceName, "association_default_route_table_id"),
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

func testAccTransitGatewayRoute_blackhole(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayRoute
	resourceName := "aws_ec2_transit_gateway_route.test_blackhole"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteConfig_destinationCIDRBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "blackhole", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayAttachmentID, ""),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", transitGatewayResourceName, "association_default_route_table_id"),
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

func testAccTransitGatewayRoute_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayRoute
	resourceName := "aws_ec2_transit_gateway_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteConfig_destinationCIDRBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGatewayRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayRoute_disappears_TransitGatewayAttachment(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayRoute
	resourceName := "aws_ec2_transit_gateway_route.test"
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
		CheckDestroy:             testAccCheckTransitGatewayRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteConfig_destinationCIDRBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGatewayVPCAttachment(), transitGatewayVpcAttachmentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTransitGatewayRouteExists(ctx context.Context, n string, v *awstypes.TransitGatewayRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindTransitGatewayStaticRoute(ctx, conn, rs.Primary.Attributes["transit_gateway_route_table_id"], rs.Primary.Attributes["destination_cidr_block"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayRouteDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_route" {
				continue
			}

			_, err := tfec2.FindTransitGatewayStaticRoute(ctx, conn, rs.Primary.Attributes["transit_gateway_route_table_id"], rs.Primary.Attributes["destination_cidr_block"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Transit Gateway Route %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTransitGatewayRouteConfig_destinationCIDRBlock(rName string) string {
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

resource "aws_ec2_transit_gateway_route" "test" {
  destination_cidr_block         = "0.0.0.0/0"
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}

resource "aws_ec2_transit_gateway_route" "test_ipv6" {
  destination_cidr_block         = "2001:db8::/56"
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}

resource "aws_ec2_transit_gateway_route" "test_blackhole" {
  destination_cidr_block         = "10.1.0.0/16"
  blackhole                      = true
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}
`, rName))
}
