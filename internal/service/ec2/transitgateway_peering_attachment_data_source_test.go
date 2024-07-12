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

func testAccTransitGatewayPeeringAttachmentDataSource_Filter_sameAccount(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
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
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentDataSourceConfig_filterSameAccount(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", dataSourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_region", dataSourceName, "peer_region"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", dataSourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrState, dataSourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, dataSourceName, names.AttrTransitGatewayID),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentDataSource_Filter_differentAccount(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

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
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentDataSourceConfig_filterDifferentAccount(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(transitGatewayResourceName, names.AttrOwnerID, dataSourceName, "peer_account_id"),
					resource.TestCheckResourceAttr(dataSourceName, "peer_region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", dataSourceName, names.AttrTransitGatewayID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrState, dataSourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, dataSourceName, "peer_transit_gateway_id"),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentDataSource_ID_sameAccount(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
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
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentDataSourceConfig_idSameAccount(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", dataSourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_region", dataSourceName, "peer_region"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", dataSourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, dataSourceName, names.AttrTransitGatewayID),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentDataSource_ID_differentAccount(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

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
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentDataSourceConfig_iDDifferentAccount(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(transitGatewayResourceName, names.AttrOwnerID, dataSourceName, "peer_account_id"),
					resource.TestCheckResourceAttr(dataSourceName, "peer_region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", dataSourceName, names.AttrTransitGatewayID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrState, dataSourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, dataSourceName, "peer_transit_gateway_id"),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentDataSource_Tags(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
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
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentDataSourceConfig_tagsSameAccount(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", dataSourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_region", dataSourceName, "peer_region"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", dataSourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrState, dataSourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, dataSourceName, names.AttrTransitGatewayID),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentDataSourceConfig_filterSameAccount(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_sameAccount(rName), `
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_peering_attachment.test.id]
  }
}
`)
}

func testAccTransitGatewayPeeringAttachmentDataSourceConfig_idSameAccount(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_sameAccount(rName), `
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  id = aws_ec2_transit_gateway_peering_attachment.test.id
}
`)
}

func testAccTransitGatewayPeeringAttachmentDataSourceConfig_tagsSameAccount(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_tags1(rName, "Name", rName), `
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  tags = {
    Name = aws_ec2_transit_gateway_peering_attachment.test.tags["Name"]
  }
}
`)
}

func testAccTransitGatewayPeeringAttachmentDataSourceConfig_filterDifferentAccount(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_differentAccount(rName), `
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  provider = "awsalternate"

  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_peering_attachment.test.id]
  }
}
`)
}

func testAccTransitGatewayPeeringAttachmentDataSourceConfig_iDDifferentAccount(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayPeeringAttachmentConfig_differentAccount(rName), `
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  provider = "awsalternate"
  id       = aws_ec2_transit_gateway_peering_attachment.test.id
}
`)
}
