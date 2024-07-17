// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
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
			testAccPreCheckTransitGateway(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue("data.aws_ec2_transit_gateway_peering_attachments.all", "ids.#", 1),
					resource.TestCheckResourceAttr("data.aws_ec2_transit_gateway_peering_attachments.by_attachment_id", "ids.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_sameAccount(rName), `
data "aws_ec2_transit_gateway_peering_attachments" "all" {
  depends_on = [aws_ec2_transit_gateway_peering_attachment.test]
}

data "aws_ec2_transit_gateway_peering_attachments" "by_attachment_id" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_peering_attachment.test.id]
  }
}
`)
}
