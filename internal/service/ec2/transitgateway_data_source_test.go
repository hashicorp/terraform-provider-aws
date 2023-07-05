// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccTransitGatewayDataSource_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Attachment": {
			"Filter": testAccTransitGatewayAttachmentDataSource_Filter,
			"ID":     testAccTransitGatewayAttachmentDataSource_ID,
		},
		"Attachments": {
			"Filter": testAccTransitGatewayAttachmentsDataSource_Filter,
		},
		"Connect": {
			"Filter": testAccTransitGatewayConnectDataSource_Filter,
			"ID":     testAccTransitGatewayConnectDataSource_ID,
		},
		"ConnectPeer": {
			"Filter": testAccTransitGatewayConnectPeerDataSource_Filter,
			"ID":     testAccTransitGatewayConnectPeerDataSource_ID,
		},
		"DxGatewayAttachment": {
			"Filter":                         testAccTransitGatewayDxGatewayAttachmentDataSource_filter,
			"TransitGatewayIdAndDxGatewayId": testAccTransitGatewayDxGatewayAttachmentDataSource_TransitGatewayIdAndDxGatewayID,
		},
		"Gateway": {
			"Filter": testAccTransitGatewayDataSource_Filter,
			"ID":     testAccTransitGatewayDataSource_ID,
		},
		"MulticastDomain": {
			"Filter": testAccTransitGatewayMulticastDomainDataSource_Filter,
			"ID":     testAccTransitGatewayMulticastDomainDataSource_ID,
		},
		"PeeringAttachment": {
			"FilterSameAccount":      testAccTransitGatewayPeeringAttachmentDataSource_Filter_sameAccount,
			"FilterDifferentAccount": testAccTransitGatewayPeeringAttachmentDataSource_Filter_differentAccount,
			"IDSameAccount":          testAccTransitGatewayPeeringAttachmentDataSource_ID_sameAccount,
			"IDDifferentAccount":     testAccTransitGatewayPeeringAttachmentDataSource_ID_differentAccount,
			"Tags":                   testAccTransitGatewayPeeringAttachmentDataSource_Tags,
		},
		"RouteTable": {
			"Filter": testAccTransitGatewayRouteTableDataSource_Filter,
			"ID":     testAccTransitGatewayRouteTableDataSource_ID,
		},
		"RouteTables": {
			"basic":  testAccTransitGatewayRouteTablesDataSource_basic,
			"Filter": testAccTransitGatewayRouteTablesDataSource_filter,
			"Tags":   testAccTransitGatewayRouteTablesDataSource_tags,
			"Empty":  testAccTransitGatewayRouteTablesDataSource_empty,
		},
		"RouteTableAssociations": {
			"Filter": testAccTransitGatewayRouteTableAssociationsDataSource_filter,
			"basic":  testAccTransitGatewayRouteTableAssociationsDataSource_basic,
		},
		"RouteTablePropagations": {
			"Filter": testAccTransitGatewayRouteTablePropagationsDataSource_filter,
			"basic":  testAccTransitGatewayRouteTablePropagationsDataSource_basic,
		},
		"VpcAttachment": {
			"Filter": testAccTransitGatewayVPCAttachmentDataSource_Filter,
			"ID":     testAccTransitGatewayVPCAttachmentDataSource_ID,
		},
		"VpcAttachments": {
			"Filter": testAccTransitGatewayVPCAttachmentsDataSource_Filter,
		},
		"VpnAttachment": {
			"Filter":                             testAccTransitGatewayVPNAttachmentDataSource_filter,
			"TransitGatewayIdAndVpnConnectionId": testAccTransitGatewayVPNAttachmentDataSource_idAndVPNConnectionID,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccTransitGatewayDataSource_Filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_transit_gateway.test"
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "amazon_side_asn", dataSourceName, "amazon_side_asn"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "association_default_route_table_id", dataSourceName, "association_default_route_table_id"),
					resource.TestCheckResourceAttrPair(resourceName, "auto_accept_shared_attachments", dataSourceName, "auto_accept_shared_attachments"),
					resource.TestCheckResourceAttrPair(resourceName, "default_route_table_association", dataSourceName, "default_route_table_association"),
					resource.TestCheckResourceAttrPair(resourceName, "default_route_table_propagation", dataSourceName, "default_route_table_propagation"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "dns_support", dataSourceName, "dns_support"),
					resource.TestCheckResourceAttrPair(resourceName, "multicast_support", dataSourceName, "multicast_support"),
					resource.TestCheckResourceAttrPair(resourceName, "owner_id", dataSourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(resourceName, "propagation_default_route_table_id", dataSourceName, "propagation_default_route_table_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_cidr_blocks.#", dataSourceName, "transit_gateway_cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_ecmp_support", dataSourceName, "vpn_ecmp_support"),
				),
			},
		},
	})
}

func testAccTransitGatewayDataSource_ID(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_transit_gateway.test"
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayDataSourceConfig_id(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "amazon_side_asn", dataSourceName, "amazon_side_asn"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "association_default_route_table_id", dataSourceName, "association_default_route_table_id"),
					resource.TestCheckResourceAttrPair(resourceName, "auto_accept_shared_attachments", dataSourceName, "auto_accept_shared_attachments"),
					resource.TestCheckResourceAttrPair(resourceName, "default_route_table_association", dataSourceName, "default_route_table_association"),
					resource.TestCheckResourceAttrPair(resourceName, "default_route_table_propagation", dataSourceName, "default_route_table_propagation"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "dns_support", dataSourceName, "dns_support"),
					resource.TestCheckResourceAttrPair(resourceName, "owner_id", dataSourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(resourceName, "propagation_default_route_table_id", dataSourceName, "propagation_default_route_table_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_cidr_blocks.#", dataSourceName, "transit_gateway_cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_ecmp_support", dataSourceName, "vpn_ecmp_support"),
				),
			},
		},
	})
}

func testAccTransitGatewayDataSourceConfig_filter(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_transit_gateway" "test" {
  filter {
    name   = "transit-gateway-id"
    values = [aws_ec2_transit_gateway.test.id]
  }
}
`, rName)
}

func testAccTransitGatewayDataSourceConfig_id(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_transit_gateway" "test" {
  id = aws_ec2_transit_gateway.test.id
}
`, rName)
}
