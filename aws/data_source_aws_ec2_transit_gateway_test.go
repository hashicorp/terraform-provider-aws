package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSEc2TransitGatewayDataSource_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"DxGatewayAttachment": {
			"Filter":                         testAccAWSEc2TransitGatewayDxGatewayAttachmentDataSource_filter,
			"TransitGatewayIdAndDxGatewayId": testAccAWSEc2TransitGatewayDxGatewayAttachmentDataSource_TransitGatewayIdAndDxGatewayId,
		},
		"Gateway": {
			"Filter": testAccAWSEc2TransitGatewayDataSource_Filter,
			"ID":     testAccAWSEc2TransitGatewayDataSource_ID,
		},
		"PeeringAttachment": {
			"FilterSameAccount":      testAccAWSEc2TransitGatewayPeeringAttachmentDataSource_Filter_sameAccount,
			"FilterDifferentAccount": testAccAWSEc2TransitGatewayPeeringAttachmentDataSource_Filter_differentAccount,
			"IDSameAccount":          testAccAWSEc2TransitGatewayPeeringAttachmentDataSource_ID_sameAccount,
			"IDDifferentAccount":     testAccAWSEc2TransitGatewayPeeringAttachmentDataSource_ID_differentAccount,
			"Tags":                   testAccAWSEc2TransitGatewayPeeringAttachmentDataSource_Tags,
		},
		"RouteTable": {
			"Filter": testAccAWSEc2TransitGatewayRouteTableDataSource_Filter,
			"ID":     testAccAWSEc2TransitGatewayRouteTableDataSource_ID,
		},
		"RouteTables": {
			"basic":  testAccDataSourceAwsEc2TransitGatewayRouteTables_basic,
			"Filter": testAccDataSourceAwsEc2TransitGatewayRouteTables_Filter,
		},
		"VpcAttachment": {
			"Filter": testAccAWSEc2TransitGatewayVpcAttachmentDataSource_Filter,
			"ID":     testAccAWSEc2TransitGatewayVpcAttachmentDataSource_ID,
		},
		"VpnAttachment": {
			"Filter":                             testAccAWSEc2TransitGatewayVpnAttachmentDataSource_filter,
			"TransitGatewayIdAndVpnConnectionId": testAccAWSEc2TransitGatewayVpnAttachmentDataSource_TransitGatewayIdAndVpnConnectionId,
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

func testAccAWSEc2TransitGatewayDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway.test"
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayDataSourceConfigFilter(),
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
					resource.TestCheckResourceAttrPair(resourceName, "vpn_ecmp_support", dataSourceName, "vpn_ecmp_support"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGatewayDataSource_ID(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway.test"
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayDataSourceConfigID(),
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
					resource.TestCheckResourceAttrPair(resourceName, "vpn_ecmp_support", dataSourceName, "vpn_ecmp_support"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGatewayDataSourceConfigFilter() string {
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

func testAccAWSEc2TransitGatewayDataSourceConfigID() string {
	return `
resource "aws_ec2_transit_gateway" "test" {}

data "aws_ec2_transit_gateway" "test" {
  id = aws_ec2_transit_gateway.test.id
}
`
}
