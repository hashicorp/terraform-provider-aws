// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayPeeringAttachmentsDataSource_Filter(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGatewayVPCAttachment(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ec2_transit_gateway_peering_attachments.by_attachment_id", "ids.#", "1"),
					resource.TestCheckResourceAttr("data.aws_ec2_transit_gateway_peering_attachments.by_gateway_id", "ids.#", "2"),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test1" {
  amazon_side_asn = "64512"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test2" {
  amazon_side_asn = "64513"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_peering_attachment" "test1" {
  transit_gateway_id      = aws_ec2_transit_gateway.test1.id
  peer_transit_gateway_id = aws_ec2_transit_gateway.test2.id

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_transit_gateway_peering_attachments" "by_attachment_id" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_peering_attachment.test1.id]
  }
}

data "aws_ec2_transit_gateway_peering_attachments" "by_gateway_id" {
  filter {
    name   = "transit-gateway-id"
    values = [aws_ec2_transit_gateway.test1.id]
  }
}
`, rName))
}
