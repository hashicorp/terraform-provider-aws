// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayRouteTablePropagation_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayRouteTablePropagation
	resourceName := "aws_ec2_transit_gateway_route_table_propagation.test"
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
		CheckDestroy:             testAccCheckTransitGatewayRouteTablePropagationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTablePropagationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTablePropagationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceType),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, transitGatewayVpcAttachmentResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", transitGatewayRouteTableResourceName, names.AttrID),
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

func testAccTransitGatewayRouteTablePropagation_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayRouteTablePropagation
	resourceName := "aws_ec2_transit_gateway_route_table_propagation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTablePropagationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTablePropagationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTablePropagationExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGatewayRouteTablePropagation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayRouteTablePropagation_attachmentChange(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayRouteTablePropagation
	resourceName := "aws_ec2_transit_gateway_route_table_propagation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTablePropagationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTablePropagationConfig_attachmentChange(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTablePropagationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, "aws_ec2_transit_gateway_vpc_attachment.test.0", names.AttrID),
				),
			},
			{
				Config: testAccTransitGatewayRouteTablePropagationConfig_attachmentChange(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTablePropagationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayAttachmentID, "aws_ec2_transit_gateway_vpc_attachment.test.1", names.AttrID),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func testAccTransitGatewayRouteTablePropagtion_recreatedDXGateway(t *testing.T, semaphore tfsync.Semaphore) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var a awstypes.TransitGatewayRouteTablePropagation
	resourceName := "aws_ec2_transit_gateway_route_table_propagation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBGPASN := sdkacctest.RandIntRange(4200000000, 4294967294)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTablePropagationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTablePropagationConfig_recreationByDXGateway(rName, rBGPASN, []string{"10.255.255.0/30"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTablePropagationExists(ctx, resourceName, &a),
				),
			},
			{
				Config: testAccTransitGatewayRouteTablePropagationConfig_recreationByDXGateway(rName, rBGPASN, []string{"10.255.255.0/30", "10.255.255.8/30"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTablePropagationExists(ctx, resourceName, &a),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func testAccCheckTransitGatewayRouteTablePropagationExists(ctx context.Context, n string, v *awstypes.TransitGatewayRouteTablePropagation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindTransitGatewayRouteTablePropagationByTwoPartKey(ctx, conn, rs.Primary.Attributes["transit_gateway_route_table_id"], rs.Primary.Attributes[names.AttrTransitGatewayAttachmentID])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayRouteTablePropagationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_route_table_propagation" {
				continue
			}

			_, err := tfec2.FindTransitGatewayRouteTablePropagationByTwoPartKey(ctx, conn, rs.Primary.Attributes["transit_gateway_route_table_id"], rs.Primary.Attributes[names.AttrTransitGatewayAttachmentID])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Transit Gateway Route Table Propagation %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTransitGatewayRouteTablePropagationConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

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

func testAccTransitGatewayRouteTablePropagationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayRouteTablePropagationConfig_base(rName), `
resource "aws_ec2_transit_gateway_route_table_propagation" "test" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
}
`)
}

func testAccTransitGatewayRouteTablePropagationConfig_recreationByDXGateway(rName string, rBGPASN int, allowedPrefixes []string) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  amazon_side_asn = "%[2]d"
  name            = %[1]q
}

resource "aws_ec2_transit_gateway" "test" {
  default_route_table_association = "disable"
  default_route_table_propagation = "disable"

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

resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = aws_dx_gateway.test.id
  associated_gateway_id = aws_ec2_transit_gateway.test.id

  allowed_prefixes = ["%[3]s"]
}

data "aws_ec2_transit_gateway_dx_gateway_attachment" "test" {
  transit_gateway_id = aws_dx_gateway_association.test.associated_gateway_id
  dx_gateway_id      = aws_dx_gateway_association.test.dx_gateway_id
}

resource "aws_ec2_transit_gateway_route_table_association" "test" {
  transit_gateway_attachment_id  = data.aws_ec2_transit_gateway_dx_gateway_attachment.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
}

resource "aws_ec2_transit_gateway_route_table_propagation" "test" {
  transit_gateway_attachment_id  = data.aws_ec2_transit_gateway_dx_gateway_attachment.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
}
`, rName, rBGPASN, strings.Join(allowedPrefixes, `", "`))
}

func testAccTransitGatewayRouteTablePropagationConfig_attachmentChange(rName string, attachmentIndex int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  count = 2

  cidr_block = "10.${count.index}.0.0/16"

  tags = {
    Name = "%[1]s-${count.index}"
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test[count.index].id
  cidr_block        = "10.${count.index}.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "%[1]s-${count.index}"
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
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
    Name = "%[1]s-${count.index}"
  }
}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_route_table_propagation" "test" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test[%[2]d].id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
}
`, rName, attachmentIndex)
}
