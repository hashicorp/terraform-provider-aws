package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccTransitGatewayConnectPeerDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_connect_peer.test"
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayConnect(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerFilterDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "bgp_asn", dataSourceName, "bgp_asn"),
					resource.TestCheckResourceAttrPair(resourceName, "inside_cidr_blocks", dataSourceName, "inside_cidr_blocks"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_address", dataSourceName, "peer_address"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_address", dataSourceName, "transit_gateway_address"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", dataSourceName, "transit_gateway_attachment_id"),
				),
			},
		},
	})
}

func testAccTransitGatewayConnectPeerDataSource_ID(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_connect_peer.test"
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayConnect(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerIDDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "bgp_asn", dataSourceName, "bgp_asn"),
					resource.TestCheckResourceAttrPair(resourceName, "inside_cidr_blocks", dataSourceName, "inside_cidr_blocks"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_address", dataSourceName, "peer_address"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_address", dataSourceName, "transit_gateway_address"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", dataSourceName, "transit_gateway_attachment_id"),
				),
			},
		},
	})
}

func testAccTransitGatewayConnectPeerFilterDataSourceConfig() string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() + `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {
  transit_gateway_cidr_blocks = ["10.20.30.0/24"]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
  inside_cidr_blocks            = ["169.254.200.0/29"]
  peer_address                  = "1.1.1.1"
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id
}

data "aws_ec2_transit_gateway_connect_peer" "test" {
  filter {
    name   = "transit-gateway-connect-peer-id"
    values = [aws_ec2_transit_gateway_connect_peer.test.id]
  }
}
`
}

func testAccTransitGatewayConnectPeerIDDataSourceConfig() string {
	return acctest.ConfigAvailableAZsNoOptInDefaultExclude() + `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {
  transit_gateway_cidr_blocks = ["10.20.30.0/24"]
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
  inside_cidr_blocks            = ["169.254.200.0/29"]
  peer_address                  = "1.1.1.1"
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id
}

data "aws_ec2_transit_gateway_connect_peer" "test" {
  id = aws_ec2_transit_gateway_connect_peer.test.id
}
`
}
