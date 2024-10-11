// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

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

func testAccTransitGatewayPrefixListReference_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	managedPrefixListResourceName := "aws_ec2_managed_prefix_list.test"
	resourceName := "aws_ec2_transit_gateway_prefix_list_reference.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			testAccPreCheckManagedPrefixList(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayPrefixListReferenceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPrefixListReferenceConfig_blackhole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccTransitGatewayPrefixListReferenceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "blackhole", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", managedPrefixListResourceName, names.AttrID),
					acctest.CheckResourceAttrAccountID(resourceName, "prefix_list_owner_id"),
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

func testAccTransitGatewayPrefixListReference_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_transit_gateway_prefix_list_reference.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			testAccPreCheckManagedPrefixList(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayPrefixListReferenceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPrefixListReferenceConfig_blackhole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccTransitGatewayPrefixListReferenceExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGatewayPrefixListReference(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayPrefixListReference_disappears_TransitGateway(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_transit_gateway_prefix_list_reference.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			testAccPreCheckManagedPrefixList(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayPrefixListReferenceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPrefixListReferenceConfig_blackhole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccTransitGatewayPrefixListReferenceExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGateway(), transitGatewayResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayPrefixListReference_TransitGatewayAttachmentID(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_transit_gateway_prefix_list_reference.test"
	transitGatewayVpcAttachmentResourceName1 := "aws_ec2_transit_gateway_vpc_attachment.test.0"
	transitGatewayVpcAttachmentResourceName2 := "aws_ec2_transit_gateway_vpc_attachment.test.1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
			testAccPreCheckManagedPrefixList(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayPrefixListReferenceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPrefixListReferenceConfig_attachmentID(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccTransitGatewayPrefixListReferenceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "blackhole", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, transitGatewayVpcAttachmentResourceName1, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayPrefixListReferenceConfig_attachmentID(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccTransitGatewayPrefixListReferenceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "blackhole", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, transitGatewayVpcAttachmentResourceName2, names.AttrID),
				),
			},
		},
	})
}

func testAccCheckTransitGatewayPrefixListReferenceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_prefix_list_reference" {
				continue
			}

			_, err := tfec2.FindTransitGatewayPrefixListReferenceByTwoPartKey(ctx, conn, rs.Primary.Attributes["transit_gateway_route_table_id"], rs.Primary.Attributes["prefix_list_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Transit Gateway Prefix List Reference %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTransitGatewayPrefixListReferenceExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindTransitGatewayPrefixListReferenceByTwoPartKey(ctx, conn, rs.Primary.Attributes["transit_gateway_route_table_id"], rs.Primary.Attributes["prefix_list_id"])

		return err
	}
}

func testAccTransitGatewayPrefixListReferenceConfig_blackhole(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_prefix_list_reference" "test" {
  blackhole                      = true
  prefix_list_id                 = aws_ec2_managed_prefix_list.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}
`, rName)
}

func testAccTransitGatewayPrefixListReferenceConfig_attachmentID(rName string, index int) string {
	return fmt.Sprintf(`
variable "index" {
  default = %[2]d
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_vpc" "test" {
  count = 2

  cidr_block = "10.${count.index}.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  cidr_block = cidrsubnet(aws_vpc.test[count.index].cidr_block, 8, 0)
  vpc_id     = aws_vpc.test[count.index].id

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
  count = 2

  subnet_ids         = [aws_subnet.test[count.index].id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test[count.index].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_prefix_list_reference" "test" {
  prefix_list_id                 = aws_ec2_managed_prefix_list.test.id
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test[var.index].id
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}
`, rName, index)
}
