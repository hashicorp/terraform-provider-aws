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

func testAccTransitGatewayMulticastGroupSource_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayMulticastGroup
	resourceName := "aws_ec2_transit_gateway_multicast_group_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayMulticastGroupSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastGroupSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayMulticastGroupSourceExists(ctx, resourceName, &v),
				),
			},
		},
	})
}

func testAccTransitGatewayMulticastGroupSource_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayMulticastGroup
	resourceName := "aws_ec2_transit_gateway_multicast_group_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayMulticastGroupSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastGroupSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayMulticastGroupSourceExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGatewayMulticastGroupSource(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayMulticastGroupSource_Disappears_domain(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayMulticastGroup
	resourceName := "aws_ec2_transit_gateway_multicast_group_source.test"
	domainResourceName := "aws_ec2_transit_gateway_multicast_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayMulticastGroupSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastGroupSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayMulticastGroupSourceExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGatewayMulticastDomain(), domainResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTransitGatewayMulticastGroupSourceExists(ctx context.Context, n string, v *awstypes.TransitGatewayMulticastGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindTransitGatewayMulticastGroupSourceByThreePartKey(ctx, conn, rs.Primary.Attributes["transit_gateway_multicast_domain_id"], rs.Primary.Attributes["group_ip_address"], rs.Primary.Attributes[names.AttrNetworkInterfaceID])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayMulticastGroupSourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_multicast_group_source" {
				continue
			}

			_, err := tfec2.FindTransitGatewayMulticastGroupSourceByThreePartKey(ctx, conn, rs.Primary.Attributes["transit_gateway_multicast_domain_id"], rs.Primary.Attributes["group_ip_address"], rs.Primary.Attributes[names.AttrNetworkInterfaceID])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Transit Gateway Multicast Group Source %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTransitGatewayMulticastGroupSourceConfig_basic(rName string) string {
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
  multicast_support = "enable"

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

resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  static_sources_support = "enable"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_domain_association" "test" {
  subnet_id                           = aws_subnet.test.id
  transit_gateway_attachment_id       = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain.test.id
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_group_source" "test" {
  group_ip_address                    = "224.0.0.1"
  network_interface_id                = aws_network_interface.test.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain_association.test.transit_gateway_multicast_domain_id
}
`, rName))
}
