package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccTransitGatewayDataSource_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
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

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccTransitGatewayDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway.test"
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayFilterDataSourceConfig(),
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
	dataSourceName := "data.aws_ec2_transit_gateway.test"
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayIDDataSourceConfig(),
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

func testAccTransitGatewayFilterDataSourceConfig() string {
	return `
resource "aws_ec2_transit_gateway" "test" {}

data "aws_ec2_transit_gateway" "test" {
  filter {
    name   = "transit-gateway-id"
    values = [aws_ec2_transit_gateway.test.id]
  }
}
`
}

func testAccTransitGatewayIDDataSourceConfig() string {
	return `
resource "aws_ec2_transit_gateway" "test" {}

data "aws_ec2_transit_gateway" "test" {
  id = aws_ec2_transit_gateway.test.id
}
`
}
